package s3

import "github.com/razzo-lunare/s3/pkg/config"

type S3 struct {
	S3Prefix       string
	S3Config       *config.S3Config
	DestinationDir string
}
