package tcp

import (
	"net/url"
	"testing"
)

func TestDecode(t *testing.T) {
	wantProbes := []string{"localhost:300", "localhost:200"}
	v, err := url.ParseQuery("tcp=localhost:200&tcp=localhost:300")
	if err != nil {
		t.Fatalf("parse query %v", err)
	}
	p := Decode(v)
	if len(p) != 2 {
		t.Fatalf("tcp Decode failure")
	}
	// Order is not guaranteed by the documentation.
	for _, want := range wantProbes {
		found := false
		for _, got := range p {
			if got.hostport == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("not found: %v", want)
		}
	}
}
