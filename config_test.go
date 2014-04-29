package main

import (
	"bytes"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/nictuku/zmon/probes/tcp"
)

func TestConfig(t *testing.T) {
	c := Config{
		[]Prober{{"http", "http://localhost:4040", 5 * time.Second}},
		[]Notificator{{"pushover", "fooopushoverkey"}},
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("Got: %v", string(b))
	dec := json.NewDecoder(bytes.NewReader(b))
	var c2 Config
	if err := dec.Decode(&c2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c, c2) {
		t.Fatalf("Config parsing wanted %v, got %v", c, c2)
	}
}

// TODO: Deprecate these.

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
