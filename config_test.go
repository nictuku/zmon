package main

import (
	"net/url"
	"testing"

	"github.com/nictuku/zmon/probes/tcp"
)

func TestDecode(t *testing.T) {
	values, err := url.ParseQuery("tcp=localhost:20&sa=&st=yves.junqueira%40gmail.com&sf=root%40cetico.org")
	if err != nil {
		t.Fatalf("ParseQuery failed: %v", err)
	}
	cfg := Decode(values)
	scheme := cfg.Probes[0].Scheme()
	if scheme != "tcp" {
		t.Fatalf("Probe not using the expected schema. Got %v, wanted 'tcp'.", scheme)
	}
	if len(cfg.Probes) != 1 {
		t.Fatalf("Wrong count of probes found. Got %d, wanted 1.", len(cfg.Probes))
	}
}

func TestEncoding(t *testing.T) {
	cfg := ServiceConfig{
		Probes: []Probe{tcp.New(url.URL{Host: "localhost:20"})},
		esc: escalator{
			Notificators: []notificator{&smtpNotification{"", "root@cetico.org", "yves.junqueira@gmail.com"}}},
	}
	t.Logf("%v", Encode(cfg))
}
