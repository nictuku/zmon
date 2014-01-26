package main

import (
	"container/ring"
	"log"
	"net/url"
	"time"

	"github.com/nictuku/zmon/probes/disk"
	"github.com/nictuku/zmon/probes/tcp"
)

const maxNotificationLines = 20

func Decode(input url.Values) *ServiceConfig {
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

	localMonitoring := &ServiceConfig{
		Frequency: 5 * time.Second,
		Probes:    probes,
		esc: escalator{
			escalationInterval: 30 * time.Minute,
			queued:             ring.New(maxNotificationLines),
			Notificators:       notificators,
		},
	}
	return localMonitoring
}

func Encode(cfg ServiceConfig) string {
	v := make(url.Values)
	for _, n := range cfg.esc.Notificators {
		n.encode(v)
	}
	for _, n := range cfg.Probes {
		n.Encode(v)
	}
	return v.Encode()
}
