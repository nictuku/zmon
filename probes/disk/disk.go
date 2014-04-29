package disk

import (
	"fmt"
	"net/url"
	"syscall"
)

// New creates a new disk probe that uses the specified mount point, e.g: / or /var.
func New(mountPoint string) *diskProbe {
	return &diskProbe{mountPoint}
}

type diskProbe struct {
	mountpoint string
}

const MaxFull = 0.90

func (p *diskProbe) Check() error {
	buf := new(syscall.Statfs_t)
	err := syscall.Statfs("/", buf)
	if err != nil {
		return err
	}
	usedRatio := 1 - (float64(buf.Bfree) / float64(buf.Blocks))
	if usedRatio > MaxFull {
		return fmt.Errorf("Partition at %q almost full at %.2f%% (min: %.2f%%)", p.mountpoint, usedRatio*100, MaxFull*100)
	}
	return nil
}

func (p *diskProbe) Scheme() string {
	return "disk"
}

func (p *diskProbe) Encode(v url.Values) {
	v.Add("disk", p.mountpoint)
}

func Decode(v url.Values) []*diskProbe {
	mountpoints, ok := v["disk"]
	if !ok {
		return nil
	}
	probes := make([]*diskProbe, 0, len(mountpoints))
	for _, h := range mountpoints {
		probes = append(probes, &diskProbe{h})
	}
	return probes
}
