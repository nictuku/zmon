// obamad notifies the local server admin when there's a problem.
// Design: http://goo.gl/l1Y36T
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"time"
)

// TODO: Prober must not warn after the first error.
// TODO: Warn user when service restored?

type Probe interface {
	// Check must never take more than 10s.
	Check() error
	Scheme() string
	Encode(url.Values)
}

type serviceConfig struct {
	frequency time.Duration
	probes    []Probe
	esc       escalator
}

type escalator struct {
	lastEscalation     time.Time
	escalationInterval time.Duration
	// queued holds messages that will be sent at some point, even if they are old.
	queued       []notification
	Notificators []notificator
}

func (e *escalator) escalate(err error) {
	e.queued = append(e.queued, notification{time.Now(), err})
	if time.Since(e.lastEscalation) > e.escalationInterval {
		// Merge all queued notifications.
		// Optimization todo: cache msg.
		msg := make([]byte, 0, len(e.queued)*len(e.queued[0].m.Error()))
		for _, n := range e.queued {
			msg = append(msg, []byte(n.String())...)
		}
		for _, n := range e.Notificators {
			if err := n.notify(msg); err != nil {
				log.Println("notification error:", err)
			} else {
				// Zero-out the queue but preserve its allocated space to avoid reallocations.
				e.queued = e.queued[:0]
				e.lastEscalation = time.Now()
				log.Println("escalation successful")
				return
			}
		}
		log.Println("IMPORTANT: all escalation methods failed.")
	}
}

type notificator interface {
	notify(msg []byte) error
	encode(v url.Values)
}

type smtpNotification struct {
	// If addr is empty, uses localhost:25 and doesn't try to use TLS.
	addr string
	from string
	to   string
}

func (s *smtpNotification) encode(v url.Values) {
	// Using Set and not Add. If more than one config of this type are
	// found, use only the last one.
	v.Set("sa", s.addr)
	v.Set("sf", s.from)
	v.Set("st", s.to)
}

var subject = []byte("Subject:Alert from obamad")

func (s *smtpNotification) notify(msg []byte) error {
	msg = bytes.Join([][]byte{subject, msg}, []byte("\n\n"))
	if s.addr == "" {
		return localSendMail(s.from, []string{s.to}, msg)
	}
	return smtp.SendMail(s.addr, nil, s.from, []string{s.to}, msg)
}

func decodeSMTPNotification(v url.Values) *smtpNotification {
	s := &smtpNotification{
		addr: v.Get("sa"),
		from: v.Get("sf"),
		to:   v.Get("st"),
	}
	if s.from == "" || s.to == "" {
		return nil
	}
	return s
}

type notification struct {
	time time.Time
	m    error
}

func (n notification) String() string {
	return fmt.Sprintf("%v: %v\n", n.time, n.m)
}

func main() {

	if len(os.Args) != 2 {
		log.Fatalf("Not enough arguments\nUsage: %v <encoded config string>", os.Args[0])
	}

	input, err := url.ParseQuery(os.Args[1])
	if err != nil {
		log.Fatalf("Input parse error: %v", err)
	}

	monitoring := Decode(input)

	t := time.Tick(monitoring.frequency)
	for {
		for _, probe := range monitoring.probes {
			if err := probe.Check(); err != nil {
				monitoring.esc.escalate(err)
			} else {
				// DEBUG
				log.Println(probe.Scheme(), "went fine")
			}
		}
		<-t
	}
}
