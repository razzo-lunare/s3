package s3

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

func (s *S3) List() (<-chan *betav1.FileInfo, error) {
	asciiterm.PrintfWarn("Listing all objects in s3://%s", s.S3Path)
	fileInfo := make(chan *betav1.FileInfo, 500)

	go listS3Files(s.Config, s.S3Path, fileInfo)

	return fileInfo, nil
}

func listS3Files(s3Config *config.S3, s3Prefix string, outputFileInfo chan<- *betav1.FileInfo) {

	numCPU := runtime.NumCPU()

	s3Prefixes := make(chan string, 10001)
	goRoutineStatus := NewGoRoutineStatus(numCPU/2, s3Prefixes)
	for instanceID := 0; instanceID < numCPU/2; instanceID++ {
		go handleListS3ObjectRecursive(
			instanceID,
			goRoutineStatus,
			s3Config,
			s3Prefixes,
			outputFileInfo,
		)
	}

	// send the first prefix to the pool then the rest will search recursively
	s3Prefixes <- s3Prefix

	// TODO: This is slow... I need to figure out how to speed it up. This could
	// be cpu intensive.
	goRoutineStatus.Wait()

	// Close recursive list goroutines to clean up them
	close(s3Prefixes)

	// Notify the next step in the pipeline there are no more files listed
	close(outputFileInfo)

	asciiterm.PrintfInfo("Finished Listing all files in s3\n")
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleListS3ObjectRecursive(id int, goRoutineStatus *GoRoutineStatus, newConfig *config.S3, s3Prefixes chan string, outputFileInfo chan<- *betav1.FileInfo) {

	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for s3PrefixJob := range s3Prefixes {
		klog.V(2).Infof("List S3: %s", s3PrefixJob)

		goRoutineStatus.SetStateRunning(id)

		opts := minio.ListObjectsOptions{
			UseV1:        false,
			WithVersions: false,
			Prefix:       s3PrefixJob,
			Recursive:    false,
			MaxKeys:      5000,
		}

		for object := range s3Client.ListObjects(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, opts) {
			if isDir(object.Key) {

				newPrefix := object.Key
				s3Prefixes <- newPrefix

				continue
			}

			newFile := &betav1.FileInfo{
				Name: object.Key,
				MD5:  object.ETag,
			}

			if object.Err != nil {
				fmt.Println(object.Err)
				continue
			}

			outputFileInfo <- newFile
		}

		goRoutineStatus.SetStateDone(id)

	}
}

func isDir(objectPath string) bool {
	return strings.HasSuffix(objectPath, "/")
}
