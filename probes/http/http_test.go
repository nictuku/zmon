package http

import (
	"net/url"
	"testing"
)

func TestDecode(t *testing.T) {
	wantProbes := []string{"http://localhost:3000/", "http://gmail.com/"}
	v, err := url.ParseQuery("http=http%3A%2F%2Flocalhost%3A3000%2F&http=http%3A%2F%2Fgmail.com%2F")
	if err != nil {
		t.Fatalf("parse query %v", err)
	}
	p := Decode(v)
	if len(p) != 2 {
		t.Fatalf("http Decode failure")
	}
	// Order is not guaranteed by the documentation.
	for _, want := range wantProbes {
		found := false
		for _, got := range p {
			if got.url == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("not found: %v", want)
		}
	}
}
