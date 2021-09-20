package filesystem

import "github.com/razzo-lunare/s3/pkg/config"

type FileSystem struct {
	SourceDir      string
	S3Config       *config.S3Config
	DestinationDir string
}
