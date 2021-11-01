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
	klog.V(1).Infof("Listing all objects in s3://%s\n", s.S3Path)

	fileInfo := make(chan *betav1.FileInfo, 500)

	go listS3Files(s.Config, s.S3Path, fileInfo)

	return fileInfo, nil
}

func listS3Files(s3Config *config.S3, s3Prefix string, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU() * 3
	s3Prefixes := make(chan string, 10001)
	goRoutineStatus := NewGoRoutineStatus(numCPU, s3Prefixes)
	jobTimer := average.New()

	initialS3Prefix := s3Prefix

	for instanceID := 0; instanceID < numCPU; instanceID++ {
		go handleListS3ObjectRecursive(
			instanceID,
			jobTimer,
			goRoutineStatus,
			s3Config,
			initialS3Prefix,
			s3Prefixes,
			outputFileInfo,
		)
	}

	// send the first prefix to the pool then the rest will search recursively
	s3Prefixes <- s3Prefix

	asciiterm.PrintfInfo("Listing S3 Objects '%s'\n", s3Prefix)
	goRoutineStatus.Wait()

	// Close recursive list goroutines to clean up them
	close(s3Prefixes)

	// Notify the next step in the pipeline there are no more files listed
	close(outputFileInfo)

	asciiterm.PrintfInfo("Finished Listing all files in s3. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleListS3ObjectRecursive(id int, jobTimer *average.JobAverageInt, goRoutineStatus *GoRoutineStatus, newConfig *config.S3, initialS3Prefix string, s3Prefixes chan string, outputFileInfo chan<- *betav1.FileInfo) {

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
			// Skip object if it's the current object
			if s3PrefixJob == object.Key {
				continue
			}

			if isDir(object.Key) {
				newPrefix := object.Key
				s3Prefixes <- newPrefix

				continue
			}
			objectName := object.Key
			// Trim the initial prefix from the full prefix so the destination is relative to the source path. Remove
			// the s3PrefixJob from the object name to ensure the comparison to the destination/source is accurate
			// for example: --source filesystem://../fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-01-05/     --destination s3://fortuna-stock-data-new/test
			//  we have to make sure we copy the content after the specified `source` to the destination. Refer to the examples below,
			// notice only the contents of the source directory are copied to the specified destination location

			//  e.g.
			//       source file name == '../fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-01-05/GME'
			//       source flag      == '--source filesystem://../fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/'
			//       destination flag == '--destination s3://fortuna-stock-data-new/test'
			//       s3 destination   == '/test/2021-01-05/GME'

			//  e.g.
			//       source file name == '../fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-01-05/GME'
			//       source flag      == '--source filesystem://../fortuna-stock-data/TIME_SERIES_INTRADAY_V2/1min/2021-01-05/'
			//       destination flag == '--destination s3://fortuna-stock-data-new/test'
			//       s3 destination   == '/test/GME'
			relativeObjectName := strings.TrimPrefix(objectName, initialS3Prefix)

			newFile := &betav1.FileInfo{
				Name: relativeObjectName,
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
