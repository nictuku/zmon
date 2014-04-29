package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	c := Config{
		[]Prober{{"http", "http://localhost:4040", 5}},
		[]Notificator{{"pushover", "userdestination", ""},
			{"smtp", "user@example.com", "from@example.com"}},
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("got")
	fmt.Println(string(b))
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
		[]Prober{{"http", "http://localhost:4040", 5}},
		[]Notificator{{"pushover", "userdestination", "fooopushoverkey"}},
	}
	fmt.Printf("Escalator: %+v", c.newEscalator())
	fmt.Printf("Escalator Noti: %+v", c.newEscalator().Notifiers)
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
