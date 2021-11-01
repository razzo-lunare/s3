package s3

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Create accepts a channel of files to create and passes the fileinfo to the next step
func (s *S3) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go uploadS3Files(s.Config, inputFiles, outputFileInfo, s.S3Path)

	return outputFileInfo, nil
}

func uploadS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo, S3Path string) {
	numCPU := runtime.NumCPU() * 3
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU; w++ {
		wg.Add(1)
		go handleUploadS3ObjectNew1(
			wg,
			jobTimer,
			s3Config,
			inputFiles,
			outputFileInfo,
			S3Path,
		)
	}
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Uploaded all s3 objects. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleUploadS3ObjectNew1 gathers the files in the S3
func handleUploadS3ObjectNew1(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo, S3Path string) {
	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for fileJob := range inputFiles {
		jobTimer.StartTimer()

		destinationFullPath := S3Path + fileJob.Name

		klog.V(1).Infof("Create S3: %s %d", destinationFullPath, fileJob.ContentSize)
		uploadInfo, err := s3Client.PutObject(
			context.Background(),
			newConfig.BucketName,
			destinationFullPath,
			fileJob.Content,
			fileJob.ContentSize,
			minio.PutObjectOptions{ContentType: "application/json"},
		)
		if err != nil {
			fmt.Println("error making dir, ", err)

			continue
		}

		klog.V(1).Info("Uploaded: ", uploadInfo.Key)

		outputFileInfo <- fileJob
		jobTimer.EndTimer()
	}

	wg.Done()
}
