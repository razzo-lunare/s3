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
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// List all s3 objects and passes them to the next step
func (s *S3) List() (<-chan *betav1.FileInfo, error) {
	asciiterm.PrintfWarn("Listing all objects in s3://%s", s.S3Path)
	fileInfo := make(chan *betav1.FileInfo, 500)

	go listS3Files(s.Config, s.S3Path, fileInfo)

	return fileInfo, nil
}

func listS3Files(s3Config *config.S3, s3Prefix string, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	s3Prefixes := make(chan string, 10001)
	goRoutineStatus := NewGoRoutineStatus(numCPU, s3Prefixes)
	jobTimer := average.New()

	for instanceID := 0; instanceID < numCPU; instanceID++ {
		go handleListS3ObjectRecursive(
			instanceID,
			jobTimer,
			goRoutineStatus,
			s3Config,
			s3Prefixes,
			outputFileInfo,
		)
	}

	// send the first prefix to the pool then the rest will search recursively
	s3Prefixes <- s3Prefix

	goRoutineStatus.Wait()

	// Close recursive list goroutines to clean up them
	close(s3Prefixes)

	// Notify the next step in the pipeline there are no more files listed
	close(outputFileInfo)

	asciiterm.PrintfInfo("Finished Listing all files in s3. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleListS3ObjectRecursive(id int, jobTimer *average.JobAverageInt, goRoutineStatus *GoRoutineStatus, newConfig *config.S3, s3Prefixes chan string, outputFileInfo chan<- *betav1.FileInfo) {

	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for s3PrefixJob := range s3Prefixes {
		goRoutineStatus.StartTask(id)
		jobTimer.StartTimer()
		klog.V(2).Infof("List S3: %s", s3PrefixJob)

		opts := minio.ListObjectsOptions{
			UseV1:        false,
			WithVersions: false,
			Prefix:       s3PrefixJob,
			Recursive:    false,
			MaxKeys:      5000,
		}

		for object := range s3Client.ListObjects(context.Background(), newConfig.BucketName, opts) {
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

		jobTimer.EndTimer()
		goRoutineStatus.FinishTask(id)
	}
}

func isDir(objectPath string) bool {
	return strings.HasSuffix(objectPath, "/")
}
