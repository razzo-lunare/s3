package sync

import (
	"fmt"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Run kicks off the main goroutine pipeline to either upload or download s3 objects
func Run(newConfig *config.S3Config, source betav1.SyncObject, destination betav1.SyncObject) error {
	// Pipeline to download s3 files through multiple pools of goroutines

	// List the files/objects
	s3Files, err := source.List()
	if err != nil {
		return err
	}

	// Identify what files/objects need to be synced
	s3FilesToGet, err := destination.Verify(s3Files)
	if err != nil {
		return err
	}

	// Gather the file/object content that need to be synced
	s3FilesToDownload, err := source.Get(s3FilesToGet)
	if err != nil {
		return err
	}

	// Write the file/object content to the destination
	downloadedFiles, err := destination.Create(s3FilesToDownload)
	if err != nil {
		return err
	}

	// Display the status and wait for the pipeline to finish
	count := 0
	for range downloadedFiles {
		asciiterm.PrintfWarn("Downloaded s3 object Count: %d", count)

		count++
	}
	fmt.Println()

	return nil
}
