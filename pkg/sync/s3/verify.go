package s3

import (
	"context"
	"log"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Verify checks to see if the FileInfo exists
func (s *S3) Verify(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go verifyS3Files(s.Config, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func verifyS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleVerifyS3Object(
			wg,
			s3Config,
			inputFiles,
			outputFileInfo,
		)
	}
	// wait for all items to be verified
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Identified files that need to be downloaded\n")
}

// handleVerifyS3Object gathers the files in the S3
func handleVerifyS3Object(wg *sync.WaitGroup, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal("setting up s3 client: ", err)
	}

	for fileJob := range inputFiles {
		klog.V(2).Infof("Verify S3: %s -> %s", fileJob.Name, fileJob.MD5)
		opts := minio.StatObjectOptions{}

		object, err := s3Client.StatObject(context.Background(), newConfig.BucketName, fileJob.Name, opts)
		if err != nil {
			klog.Errorf("stat s3 object: %s, %s: %s", newConfig.BucketName, fileJob.Name, err)
			continue
		}

		newFile := &betav1.FileInfo{
			Name: object.Key,
			MD5:  object.ETag,
		}
		if fileJob.MD5 != newFile.MD5 && fileJob.Name != newFile.Name {
			// download the file from s3
			outputFileInfo <- fileJob
			continue
		}
	}

	// Notify parent proccess that all items have been verified
	wg.Done()
}
