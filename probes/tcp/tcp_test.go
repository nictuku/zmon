package tcp

import (
	"net/url"
	"testing"
)

func TestDecode(t *testing.T) {
	v, err := url.ParseQuery("tcp=localhost:200&tcp=localhost:300")
	if err != nil {
		t.Fatalf("parse query %v", err)
	}
	p := Decode(v)
	if len(p) != 2 {
		t.Fatalf("tcp Decode failure")
	}

	// This assumes order is preserved but that's not guaranteed by the documentation.
	if p[1].hostport != "localhost:300" {
		t.Fatalf("nope")
	}
}
