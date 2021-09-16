package s3

import (
	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
)

type FileInfo struct {
	Name string
	md5  string
}

func Sync(newConfig *config.S3Config, s3Prefix string, destinationDir string) error {

	// Pipeline to download s3 files through multiple pools of goroutines
	s3Files := ListS3Files(newConfig, s3Prefix)
	s3FilesToDownload := VerifyS3Files(newConfig, destinationDir, s3Files)
	downloadedFiles := DownloadS3Files(newConfig, destinationDir, s3FilesToDownload)

	count := 0
	for range downloadedFiles {
		asciiterm.PrintfWarn("Downloaded s3 object Count: %d", count)

		count++
	}

	return nil
}
