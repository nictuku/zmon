// obamad notifies the local server admin when there's a problem.
// Design: http://goo.gl/l1Y36T
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"time"

	"github.com/nictuku/obamad/probes/tcp"
)

// TODO: Prober must not warn after the first error.
// TODO: Warn user when service restored?

type Probe interface {
	// Check must never take more than 10s.
	Check() error
	Scheme() string
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
	notificators []notificator
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
		for _, n := range e.notificators {
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
}

type smtpNotification struct {
	addr string
	from string
	to   []string
}

var subject = []byte("Subject:Alert from obamad")

func (s *smtpNotification) notify(msg []byte) error {
	msg = bytes.Join([][]byte{subject, msg}, []byte("\n\n"))
	if s.addr == "" {
		return localSendMail(s.from, s.to, msg)
	}
	return smtp.SendMail(s.addr, nil, s.from, s.to, msg)
}

type notification struct {
	time time.Time
	m    error
}

func (n notification) String() string {
	return fmt.Sprintf("%v: %v\n", n.time, n.m)
}

func main() {

	tcp22 := tcp.New(url.URL{Scheme: "tcp", Host: "localhost:TA-ERRADO-MERMAO"})

	localMonitoring := &serviceConfig{
		frequency: 5 * time.Second,
		probes:    []Probe{tcp22},
	}

	esc := &escalator{
		escalationInterval: 30 * time.Minute,
		queued:             make([]notification, 0, 10),
		// XXX
		notificators: []notificator{
			// If addr is empty, uses localhost:25 and doesn't try to use TLS.
			&smtpNotification{"", "root@cetico.org", []string{"yves.junqueira@gmail.com"}},

			// This would use TLS, so it won't work with self-signed certificates.
			// &smtpNotification{"something:25", "root@cetico.org", []string{"yves.junqueira@gmail.com"}},
		},
	}

	t := time.Tick(localMonitoring.frequency)
	for {
		for _, probe := range localMonitoring.probes {
			if err := probe.Check(); err != nil {
				esc.escalate(err)
			} else {
				// DEBUG
				log.Println(probe.Scheme(), "went fine")
			}
		}
		<-t
	}
}
