// zmon notifies the local server admin when there's a problem.
// Design: http://goo.gl/l1Y36T
package main

import (
	"bytes"
	"container/ring"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"bitbucket.org/kisom/gopush/pushover"
)

type escalator struct {
	lastEscalation     time.Time
	escalationInterval time.Duration
	// queued holds messages that will be sent at some point, even if they are old. When the
	// queue gets to 20 messages, older ones are dropped.
	queued    *ring.Ring
	Notifiers []notifier
}

func (e *escalator) escalate(err error) {
	e.queued = e.queued.Next()
	e.queued.Value = notification{time.Now(), hostname, err}
	if time.Since(e.lastEscalation) > e.escalationInterval {
		// Merge all queued notifications.
		// Optimization todo: cache msg.
		msg := make([]byte, 0, 200)
		e.queued.Do(func(v interface{}) {
			if notif, ok := v.(notification); ok {
				msg = append(msg, []byte(notif.String())...)
			}

		})
		for _, n := range e.Notifiers {
			if err := n.notify(msg); err != nil {
				log.Println("notification error:", err)
				log.Printf("Would have written: %q", string(msg))
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

type notifier interface {
	notify(msg []byte) error
}

type smtpNotification struct {
	// If addr is empty, uses localhost:25 and doesn't try to use TLS.
	addr string
	from string
	to   string
}

var subject = []byte("Subject:Alert from zmon")

func (s *smtpNotification) notify(msg []byte) error {
	msg = bytes.Join([][]byte{subject, msg}, []byte("\n\n"))
	if s.addr == "" {
		return localSendMail(s.from, []string{s.to}, msg)
	}
	return smtp.SendMail(s.addr, nil, s.from, []string{s.to}, msg)
}

type pushoverNotification struct {
	identity pushover.Identity
}

func (p *pushoverNotification) notify(msg []byte) error {
	sent := pushover.Notify(p.identity, string(msg))
	if !sent {
		return fmt.Errorf("pushover notification failed.")
	}
	return nil
}

type notification struct {
	time time.Time
	host string
	m    error
}

func (n notification) String() string {
	return fmt.Sprintf("%v: @%v: %v\n", n.time.Format("2006-01-02 15:04:05 -0700 MST"), n.host, n.m)
}
