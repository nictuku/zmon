package main

import (
	"container/ring"
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"bitbucket.org/kisom/gopush/pushover"

	"github.com/nictuku/zmon/probes/disk"
	"github.com/nictuku/zmon/probes/http"
	"github.com/nictuku/zmon/probes/tcp"
)

type Prober struct {
	// Type of the probe: "disk", tcp", "http", etc.
	Type string
	// Target is the resource to be probed. The format depends on the type.
	Target string
	// Interval when the probe should run
	Freq time.Duration
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
	// Authentication details.
	// For PushOver, this is the pushover application key.
	// For SMTP, it's the server and login details, in the format "user[:password][@server][:port]".
	// The "user" string is used as the From address. Only 'user' is required. Password is currenly unused even if specified.
	Auth string
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
		user, server := parseSMTPAuth(n.Auth)
		return &smtpNotification{
			addr: server,
			from: user,
			to:   n.Destination,
		}
	case "pushover":
		return &pushoverNotification{
			identity: pushover.Authenticate(pushoverKey, n.Auth),
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

// ReadConfig reads the mothership configuration from $HOME/.zmon/zmon.json and returns
// the parsed Config.
func ReadConf() (cfg Config, err error) {
	file, err := os.Open(confPath())
	if err != nil {
		return cfg, err
	}
	decoder := json.NewDecoder(file)
	// TODO: Sanity check the config. (e.g: missing notificators, probes)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, err
	}
	if len(cfg.Notification) == 0 {
		log.Fatal("No notification settings found. Exiting")
	}
	return cfg, nil
}
