package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	c := Config{
		[]Prober{{"http", "http://localhost:4040", 5 * time.Second}},
		[]Notificator{{"pushover", "userdestination", "fooopushoverkey"}},
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

func TestConcreteConfig(t *testing.T) {
	c := Config{
		[]Prober{{"http", "http://localhost:4040", 5 * time.Second}},
		[]Notificator{{"pushover", "userdestination", "fooopushoverkey"}},
	}
	fmt.Printf("Escalator: %+v", c.escalator())
	fmt.Printf("Escalator Noti: %+v", c.escalator().Notifiers)
}

func TestSMTPAuthParse(t *testing.T) {
	tests := []struct {
		input      string
		user       string
		serverport string
	}{{"nictuku", "nictuku", ""},
		{"nictuku@server", "nictuku", "server"},
		{"nictuku@server:port", "nictuku", "server:port"},
		{"nictuku:password", "nictuku", ""},
		{"nictuku:password@server", "nictuku", "server"},
		{"nictuku:password@server:port", "nictuku", "server:port"},
	}
	for _, test := range tests {
		user, serverport := parseSMTPAuth(test.input)
		if user != test.user {
			t.Errorf("parseSMTPAuth(%v), user = %v; wanted %v", test.input, user, test.user)
		}
		if serverport != test.serverport {
			t.Errorf("parseSMTPAuth(%v), serverport = %v; wanted %v", test.input, serverport, test.serverport)
		}
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
