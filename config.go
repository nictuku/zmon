package main

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gregdel/pushover"

	"github.com/nictuku/zmon/probes/disk"
	"github.com/nictuku/zmon/probes/http"
	"github.com/nictuku/zmon/probes/tcp"
)

var pushoverKey = os.Getenv("PUSHOVER_KEY")
var pushoverRecipient = os.Getenv("PUSHOVER_RECIPIENT")

type Prober struct {
	// Type of the probe: "disk", tcp", "http", etc.
	Type string
	// Target is the resource to be probed. For disk, this is the mount point being checked. For
	// TCP, it's "host:port". For HTTP, it's a URL like "http://localhost:4040/debug/vars".
	Target string
	// How often should the probe run, in seconds.
	IntervalSeconds int
}

func (p *Prober) Probe() Probe {
	switch p.Type {
	case "disk":
		return disk.New(p.Target)
	case "tcp":
		return tcp.New(p.Target)
	case "http":
		return http.New(p.Target)
	default:
		log.Printf("Ignoring unknown probe type %q", p.Type)
		return nil
	}
}

// Notificator represents a method to send notifications about problems.
// It's used when encoding and decoding the JSON configuration.
type Notificator struct {
	// smtp, pushover, etc.
	Type string
	// Where to send the notifications to.
	// For SMTP, the email address. For PushOver, it's the destination user's token.
	Destination string
	// Authentication details. For PushOver, this is the pushover application key. For SMTP, it's
	// the server and login details, in the format "user[:password][@server][:port]". The "user"
	// string is used as the From address. Only 'user' is required. Password is currenly unused even
	// if specified.
	From string `json:",omitempty"`
}

func parseSMTPAuth(auth string) (user, serverport string) {
	authp := strings.SplitN(auth, "@", 2)
	userpass := authp[0]
	if len(authp) == 2 {
		serverport = authp[1]
	}
	user = strings.SplitN(userpass, ":", 2)[0] // Password is ignored for now.
	return user, serverport
}

func (n *Notificator) notifier() notifier {
	switch n.Type {
	case "smtp":
		user, server := parseSMTPAuth(n.From)
		return &smtpNotification{
			addr: server,
			from: fmt.Sprintf("%v@%v", user, server),
			to:   n.Destination,
		}
	case "pushover":
		return &pushoverNotification{
			app:      pushover.New(pushoverKey),
			identity: pushover.NewRecipient(pushoverRecipient),
		}
	default:
		log.Printf("Ignoring unknown notifier type %q", n.Type)
		return nil
	}
}

// Config contains the zmon agent configuration.
type Config struct {
	Probes       []Prober
	Notification []Notificator
}

const maxNotificationLines = 20

func (c *Config) newEscalator() *escalator {
	notifiers := make([]notifier, len(c.Notification))
	for i, n := range c.Notification {
		notifiers[i] = n.notifier()
	}
	return &escalator{
		escalationInterval: 30 * time.Minute,
		queued:             ring.New(maxNotificationLines),
		Notifiers:          notifiers,
	}
}

func makeConfDir() string {
	dir := "/var/run/zmon"
	env := os.Environ()
	for _, e := range env {
		if strings.HasPrefix(e, "HOME=") {
			dir = strings.SplitN(e, "=", 2)[1]
			dir = path.Join(dir, ".zmon")
		}
	}
	// Ignore errors.
	os.MkdirAll(dir, 0750)

	if s, err := os.Stat(dir); err != nil {
		log.Fatal("stat config dir", err)
	} else if !s.IsDir() {
		log.Fatalf("Dir %v expected directory, got %v", dir, s)
	}
	return dir
}

func confPath() string {
	dir := makeConfDir()
	return path.Join(dir, "zmon.json")
}

var defaultCfg = Config{
	Probes: []Prober{
		{

			Type:            "disk",
			Target:          "/",
			IntervalSeconds: 5,
		},
		{
			Type:            "tcp",
			Target:          "localhost:22",
			IntervalSeconds: 5,
		},
		{
			Type: "http",

			Target:          "http://localhost:4040",
			IntervalSeconds: 5,
		},
	},
	Notification: []Notificator{
		{
			Type:        "pushover",
			Destination: "userdestination",
		},
		{
			Type:        "smtp",
			Destination: "user@example.com",
			From:        "zmon@example.com",
		},
	},
}

// ReadConfig reads the mothership configuration from $HOME/.zmon/zmon.json and returns
// the parsed Config.
func ReadConf() (cfg Config, err error) {
	file, err := os.Open(confPath())
	if err != nil {
		return defaultCfg, nil
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return defaultCfg, nil
	}
	if len(cfg.Notification) == 0 {
		log.Fatal("No notification settings found. Exiting")
	}
	for _, p := range cfg.Probes {
		if p.IntervalSeconds == 0 {
			log.Fatalf("Probe of type %q missing IntervalSeconds", p.Type)
		}
	}
	return cfg, nil
}
