package s3

import "github.com/razzo-lunare/s3/pkg/config"

// S3 is the implementation for betav1/SyncObject for s3
type S3 struct {
	S3Path string
	Config *config.S3
}
