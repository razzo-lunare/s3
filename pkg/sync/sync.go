package sync

import (
	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
)

func Run(newConfig *config.S3Config, source string, destination string) error {

	// Pipeline to download s3 files through multiple pools of goroutines
	s3Files := ListS3Files(newConfig, source)
	s3FilesToDownload := VerifyS3Files(newConfig, destination, s3Files)
	downloadedFiles := DownloadS3Files(newConfig, destination, s3FilesToDownload)

	count := 0
	for range downloadedFiles {
		asciiterm.PrintfWarn("Downloaded s3 object Count: %d", count)

		count++
	}

	return nil
}
