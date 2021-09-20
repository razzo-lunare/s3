package sync

import (
	"fmt"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

func Run(newConfig *config.S3Config, source betav1.SyncObject, destination betav1.SyncObject) error {
	source.List()
	// Pipeline to download s3 files through multiple pools of goroutines
	s3Files, err := source.List()
	if err != nil {
		return err
	}

	s3FilesToDownload, err := destination.Verify(s3Files)
	if err != nil {
		return err
	}

	downloadedFiles, err := destination.Create(s3FilesToDownload)
	if err != nil {
		return err
	}

	count := 0
	for range downloadedFiles {
		asciiterm.PrintfWarn("Downloaded s3 object Count: %d", count)

		count++
	}
	// Add 1 more line after the PrintfWarn above
	fmt.Println()

	return nil
}
