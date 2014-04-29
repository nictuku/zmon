// zmon notifies the local server admin when there's a problem.
// Design: http://goo.gl/l1Y36T
package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"
)

// listenPort is used a simple lock mechanism for zmon. If it can't listen to
// this port, interrupt the startup.
const listenPort = "127.0.0.1:61510"

var hostname string

// TODO: Prober must not warn after the first error.
// TODO: Warn user when service restored?

type Probe interface {
	// Check must never take more than 10s.
	Check() error
	Scheme() string
	// TODO: Remove.
	Encode(url.Values)
}

type ServiceConfig struct {
	Frequency time.Duration
	Probes    []Probe
	esc       escalator
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Not enough arguments\nUsage: %v <encoded config string>", os.Args[0])
	}
	if fd, err := net.Listen("tcp", listenPort); err != nil {
		// Assume that this is a "address already in use" error and just exit without
		// printing anything to avoid excessive logging. If there was a nice way to test for
		// that error I'd use it.
		os.Exit(1)
	} else {
		defer fd.Close()
	}

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Printf("os.Hostname failed: %v. Using 'unknown' hostname", err)
		hostname = "unknown"
	}
	cfg, err := ReadConf()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	go phoneHome()

	e := make(chan error)
	for _, probe := range cfg.Probes {
		go probeCheck(probe, e)
	}
	escalator := cfg.escalator()
	for err := range e {
		escalator.escalate(err)
	}
}

func probeCheck(p Prober, e chan error) {
	probe := p.Probe()
	t := time.Tick(p.Freq)
	for _ = range t {
		if err := probe.Check(); err != nil {
			e <- fmt.Errorf("%v: %v", probe.Scheme(), err)
		} else {
			// DEBUG
			// log.Println(probe.Scheme(), "went fine")
		}
	}
}
