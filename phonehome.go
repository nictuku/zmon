package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	heartBeatURL    = "https://analytics.artell.ai/"
	heartBeatPeriod = time.Second * 30
	waitTime        = time.Second * 30
	debug           = true

	// Prevent the execution of multiple agents. Also useful for occasional debugging.
	httpServerPort = 9999
)

func heartBeat() {
	s := newServerInfo()
	resp, err := http.PostForm(heartBeatURL, s.Values())
	if err != nil {
		log.Printf("Request to %v err %v", heartBeatURL, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Response from %v: err %v", heartBeatURL, err)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("Response from %v: %v", heartBeatURL, resp.Status)
		if debug {
			log.Println(string(body))
		}
		return
	}
	log.Printf("Response from %v: %v", heartBeatURL, resp.Status)
}

type serverInfo struct {
	hostname string
	sshPort  int
}

func (s *serverInfo) Values() url.Values {
	param := make(url.Values)
	param.Set("hostname", s.hostname)
	param.Set("sshPort", strconv.Itoa(s.sshPort))
	return param
}

func newServerInfo() *serverInfo {
	h, _ := os.Hostname()
	if h == "" {
		h = "unknown"
	}
	return &serverInfo{
		hostname: h,
		sshPort:  22,
	}

}

// phoneHome sends stats to http://mothership.pw once a while.
func contactMothership() {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", httpServerPort), nil)
		if err != nil {
			log.Fatal("could not start http server, check that no other agent is running: ", err)
		}
	}()
	tick := time.Tick(heartBeatPeriod)
	for {
		heartBeat()
		<-tick
	}
}
