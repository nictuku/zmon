package updater

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	update "github.com/inconshreveable/go-update"
)

func updateURL() string {
	// TODO: Use HTTPS.
	return fmt.Sprintf("http://zmon.org/static/%v/%v/zmon", runtime.GOOS, runtime.GOARCH)
}

// SelfUpdate updates the zmon binary. It has several limitations:
//
// - It always performs updates, unconditionally.
//
// - It downloads the entire binary, not patches.
//
// - Binaries are not compressed.
//
// - There is no forced shutdown of old binaries.
func SelfUpdate() error {
	resp, err := http.Get(updateURL())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := update.Apply(resp.Body, update.Options{}); err != nil {
		return fmt.Errorf("Update failed: %v\n", err)
	}
	log.Println("Update succeeded")
	return nil
}
