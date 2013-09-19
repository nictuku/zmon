package tcp

import (
	"time"
	"net"
	"net/url"
)

type tcpProbe struct {
	hostport string
}

func (p *tcpProbe) Check() error {
	conn, err := net.DialTimeout("tcp", p.hostport, 10*time.Second)
	if err != nil {
		return err
	}
	// TODO: Keep the connection open and use that as a status check?
	return conn.Close()
}

func (p *tcpProbe) Scheme() string {
	return "tcp"
}

func New(url url.URL) *tcpProbe {
	return &tcpProbe{url.Host}
}
