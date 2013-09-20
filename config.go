package main

import (
	"log"
	"net/url"
	"time"

	"github.com/nictuku/obamad/probes/disk"
	"github.com/nictuku/obamad/probes/tcp"
)

func Decode(input url.Values) *serviceConfig {
	probes := make([]Probe, 0, 2)

	// Try parsing the input for each probe type.
	tcpProbes := tcp.Decode(input)
	for _, p := range tcpProbes {
		probes = append(probes, p)
	}
	diskProbes := disk.Decode(input)
	for _, p := range diskProbes {
		probes = append(probes, p)
	}

	if len(probes) == 0 {
		log.Fatalf("No probes configured. Exiting.")
	}

	notificators := make([]notificator, 0, 1)

	smtpN := decodeSMTPNotification(input)
	if smtpN != nil {
		notificators = append(notificators, smtpN)
	}

	if len(notificators) == 0 {
		log.Fatal("No notification settings found. Exiting")
	}

	localMonitoring := &serviceConfig{
		frequency: 5 * time.Second,
		probes:    probes,
		esc: escalator{
			escalationInterval: 30 * time.Minute,
			queued:             make([]notification, 0, 10),
			Notificators:       notificators,
		},
	}
	return localMonitoring
}

func Encode(cfg serviceConfig) string {
	v := make(url.Values)
	for _, n := range cfg.esc.Notificators {
		n.encode(v)
	}
	for _, n := range cfg.probes {
		n.Encode(v)
	}
	return v.Encode()
}
