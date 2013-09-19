package main

import (
	"net/smtp"
)

// localSendMail uses the localhost SMTP server and does not attempt to use
// TLS.
func localSendMail(from string, to []string, msg []byte) error {

	// Connect to the remote SMTP server.
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		return err
	}
	// Set the sender and recipient.
	c.Mail(from)
	for _, r := range to {
		c.Rcpt(r)
	}
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	_, err = wc.Write(msg)
	if err != nil {
		return err
	}
	err = wc.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
