package http

import (
	"net/http"
	"net/url"
	"time"

	// Needed until the fix for https://code.google.com/p/go/issues/detail?id=3362 is released.
	httpclient "github.com/dmichael/go-httptimeoutclient"
)

// New creates a new HTTP probe that sends a GET to the specified URL.
func New(url string) *httpProbe {
	client := httpclient.NewTimeoutClient(5*time.Second, 5*time.Second)
	return &httpProbe{url, client}
}

type httpProbe struct {
	url    string
	client *http.Client
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
	client := httpclient.NewTimeoutClient(5*time.Second, 5*time.Second)
	probes := make([]*httpProbe, 0, len(urls))
	for _, h := range urls {
		probes = append(probes, &httpProbe{h, client})
	}
	return probes
}

func FromURL(url url.URL) *httpProbe {
	client := httpclient.NewTimeoutClient(5*time.Second, 5*time.Second)
	return &httpProbe{url.Host, client}
}
