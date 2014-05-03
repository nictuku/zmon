package http

import (
	"net/http"
	"net/url"
	"time"
)

// New creates a new HTTP probe that sends a GET to the specified URL.
func New(url string) *httpProbe {
	client := http.Client{Timeout: 5 * time.Second}
	return &httpProbe{url, client}
}

type httpProbe struct {
	url    string
	client http.Client
}

func (p *httpProbe) Check() error {
	_, err := p.client.Get(p.url)
	return err
}

func (p *httpProbe) Scheme() string {
	return "http"
}

func (p *httpProbe) Encode(v url.Values) {
	v.Add("http", p.url)
}

func Decode(v url.Values) []*httpProbe {
	urls, ok := v["http"]
	if !ok {
		return nil
	}
	client := http.Client{Timeout: 5 * time.Second}
	probes := make([]*httpProbe, 0, len(urls))
	for _, h := range urls {
		probes = append(probes, &httpProbe{h, client})
	}
	return probes
}

func FromURL(url url.URL) *httpProbe {
	client := http.Client{Timeout: 5 * time.Second}
	return &httpProbe{url.Host, client}
}
