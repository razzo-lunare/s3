package s3

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"k8s.io/klog/v2"
)

// Create accepts a channel of files to create
func (s *S3) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go downloadS3Files(s.Config, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func downloadS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU*2; w++ {
		wg.Add(1)
		go handleDownloadS3ObjectNew(
			wg,
			s3Config,
			inputFiles,
			outputFileInfo,
		)
	}
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("downloaded all s3 objects\n")
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	log.Fatal("not implemented")
	for fileJob := range inputFiles {
		outputFileInfo <- fileJob
	}

	wg.Done()
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew1(wg *sync.WaitGroup, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})

	if err != nil {
		fmt.Println(err)
	}

	for fileJob := range inputFiles {
		klog.V(2).Infof("Create S3: %s", fileJob.Name)

		uploadInfo, err := s3Client.PutObject(
			context.Background(),
			newConfig.BucketName,
			fileJob.Name, fileJob.Content,
			0, // Not setting content size, this is a test :)
			minio.PutObjectOptions{ContentType: "application/octet-stream"},
		)
		if err != nil {
			fmt.Println("error making dir, ", err)

			continue
		}

		klog.Info("Uploaded: ", uploadInfo.Key)

		outputFileInfo <- fileJob
	}

	wg.Done()
}
