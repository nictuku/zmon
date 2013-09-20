package disk

import (
	"fmt"
	"net/url"
	"syscall"
)

type diskProbe struct {
	mountpoint string
}

const MinFree = 0.80

func (p *diskProbe) Check() error {
	buf := new(syscall.Statfs_t)
	err := syscall.Statfs("/", buf)
	if err != nil {
		return err
	}

	freeRatio := float64(buf.Bfree) / (float64(buf.Blocks))
	if freeRatio < MinFree {
		return fmt.Errorf("Partition at %q almost full at %.2f%% (min: %.2f%%)", p.mountpoint, freeRatio, MinFree)
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
