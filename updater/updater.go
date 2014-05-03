package updater

import (
	"fmt"
	"log"
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
	url := updateURL()
	log.Println("Updating zmon binary from", url)
	err, errRecover := update.New().FromUrl(updateURL())
	if err != nil {
		if errRecover != nil {
			return fmt.Errorf("WARNING: Update failed and could recover the old binary%v", err)
		}
		return fmt.Errorf("Update failed: %v\n", err)
	}
	log.Println("Update succeeded")
	return nil
}
