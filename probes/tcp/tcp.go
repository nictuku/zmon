package tcp

import (
	"net"
	"net/url"
	"time"
)

// New creates a new TCP probe that attempts to connect to the specified host:port.
func New(hostport string) *tcpProbe {
	return &tcpProbe{hostport}
}

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

func (p *tcpProbe) Encode(v url.Values) {
	v.Add("tcp", p.hostport)
}

func Decode(v url.Values) []*tcpProbe {
	hostports, ok := v["tcp"]
	if !ok {
		return nil
	}
	probes := make([]*tcpProbe, 0, len(hostports))
	for _, h := range hostports {
		probes = append(probes, &tcpProbe{h})
	}
	return probes
}

func FromURL(url url.URL) *tcpProbe {
	return &tcpProbe{url.Host}
}
