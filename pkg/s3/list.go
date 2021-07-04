package s3

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/razzo-lunare/fortuna/pkg/config"
)

// func ListS3Files1() <-chan int {
// 	out := make(chan int, len(nums))
// 	for _, n := range nums {
// 		out <- n
// 	}
// 	close(out)
// 	return out
// }

func ListS3Files(fortunaConfig *config.FortunaConfig, weekDay <-chan string) <-chan *FileInfo {
	fileInfo := make(chan *FileInfo)

	go listS3Files(fortunaConfig, weekDay, fileInfo)

	return fileInfo
}

func listS3Files(fortunaConfig *config.FortunaConfig, inputWeekDay <-chan string, outputFileInfo chan<- *FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleListS3Object1(
			wg,
			fortunaConfig,
			inputWeekDay,
			outputFileInfo,
		)
	}
	wg.Wait()
	// fmt.Println("Done with: file list")
	close(outputFileInfo)
}

// handleListS3Object gathers the files in the S3
func handleListS3Object1(wg *sync.WaitGroup, newConfig *config.FortunaConfig, inputWeekday <-chan string, outputFileInfo chan<- *FileInfo) {

	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for weekDayListJob := range inputWeekday {

		opts := minio.ListObjectsOptions{
			UseV1:        false,
			WithVersions: false,
			Prefix:       fmt.Sprintf("TIME_SERIES_INTRADAY_V2/1min/%s/", weekDayListJob),
			// Recursive:    true,
			MaxKeys: 1000,
		}

		for object := range s3Client.ListObjects(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, opts) {
			newFile := &FileInfo{
				Name: object.Key,
				md5:  object.ETag,
			}

			if object.Err != nil {
				fmt.Println(object.Err)
				continue
			}

			outputFileInfo <- newFile
		}
	}
	wg.Done()

	// List all objects from a bucket-name with a matching prefix.
}
