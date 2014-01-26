package main

import (
	"net/url"
	"testing"

	"github.com/nictuku/zmon/probes/tcp"
)

func TestDecode(t *testing.T) {
	_, err := url.ParseQuery("tcp=localhost:20&sa=&st=yves.junqueira%40gmail.com&sf=root%40cetico.org")
	if err != nil {
		t.Fatalf("url parse query error: %v", err)
	}
	// TODO: test this.
}

func TestEncoding(t *testing.T) {
	cfg := ServiceConfig{
		Probes: []Probe{tcp.New(url.URL{Host: "localhost:20"})},
		esc: escalator{
			Notificators: []notificator{&smtpNotification{"", "root@cetico.org", "yves.junqueira@gmail.com"}}},
	}
	t.Logf("%v", Encode(cfg))
}
