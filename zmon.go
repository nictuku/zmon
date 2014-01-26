// zmon notifies the local server admin when there's a problem.
// Design: http://goo.gl/l1Y36T
package main

import (
	"bytes"
	"container/ring"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"time"

	"bitbucket.org/kisom/gopush/pushover"
)

// listenPort is used a simple lock mechanism for zmon. If it can't listen to
// this port, interrupt the startup.
const listenPort = "127.0.0.1:61510"

// TODO: Prober must not warn after the first error.
// TODO: Warn user when service restored?

type Probe interface {
	// Check must never take more than 10s.
	Check() error
	Scheme() string
	Encode(url.Values)
}

type ServiceConfig struct {
	Frequency time.Duration
	Probes    []Probe
	esc       escalator
}

type escalator struct {
	lastEscalation     time.Time
	escalationInterval time.Duration
	// queued holds messages that will be sent at some point, even if they are old. When the
	// queue gets to 20 messages, older ones are dropped.
	queued       *ring.Ring
	Notificators []notificator
}

func (e *escalator) escalate(err error) {
	e.queued = e.queued.Next()
	e.queued.Value = notification{time.Now(), err}
	if time.Since(e.lastEscalation) > e.escalationInterval {
		// Merge all queued notifications.
		// Optimization todo: cache msg.
		msg := make([]byte, 0, 200)
		e.queued.Do(func(v interface{}) {
			if notif, ok := v.(notification); ok {
				msg = append(msg, []byte(notif.String())...)
			}

		})
		for _, n := range e.Notificators {
			if err := n.notify(msg); err != nil {
				log.Println("notification error:", err)
				log.Println("Would have written: %q", string(msg))
			} else {
				e.queued = ring.New(maxNotificationLines)
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

var subject = []byte("Subject:Alert from zmon")

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

func decodePushoverNotification(v url.Values) *pushoverNotification {
	p := &pushoverNotification{
		pt: v.Get("pt"),
	}
	if p.pt == "" {
		return nil
	}
	p.identity = pushover.Authenticate(pushoverKey, p.pt)
	return p
}

type pushoverNotification struct {
	pt       string
	identity pushover.Identity
}

func (p *pushoverNotification) notify(msg []byte) error {
	sent := pushover.Notify(p.identity, string(msg))
	if !sent {
		return fmt.Errorf("pushover notification failed.")
	}
	return nil
}
func (p *pushoverNotification) encode(v url.Values) {
	// Using Set and not Add. If more than one config of this type are
	// found, use only the last one.
	v.Set("pt", p.pt)
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
	if fd, err := net.Listen("tcp", listenPort); err != nil {
		// Assume that this is a "address already in use" error and just exit without
		// printing anything to avoid excessive logging. If there was a nice way to test for
		// that error I'd use it.
		os.Exit(1)
	} else {
		defer fd.Close()
	}

	input, err := url.ParseQuery(os.Args[1])
	if err != nil {
		log.Fatalf("Input parse error: %v", err)
	}

	monitoring := Decode(input)

	t := time.Tick(monitoring.Frequency)
	for {
		for _, probe := range monitoring.Probes {
			if err := probe.Check(); err != nil {
				monitoring.esc.escalate(fmt.Errorf("%v: %v", probe.Scheme(), err))
			} else {
				// DEBUG
				// log.Println(probe.Scheme(), "went fine")
			}
		}
		<-t
	}
}
