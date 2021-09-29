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
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Verify checks to see if the FileInfo exists
func (s *S3) Verify(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	// Start a single background thread to manage them all
	go verifyS3Files(s.Config, inputFiles, outputFileInfo)

	// Return the new pipeline channel
	return outputFileInfo, nil
}

func verifyS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleVerifyS3Object(
			wg,
			jobTimer,
			s3Config,
			inputFiles,
			outputFileInfo,
		)
	}

	// wait for all items to be verified
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Identified files that need to be downloaded. Time: %f\n", jobTimer.GetAverage())
}

// handleVerifyS3Object gathers the files in the S3
func handleVerifyS3Object(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal("setting up s3 client: ", err)
	}

	for fileJob := range inputFiles {
		jobTimer.StartTimer()
		klog.V(2).Infof("Verify S3: %s -> %s", fileJob.Name, fileJob.MD5)
		opts := minio.StatObjectOptions{}

		object, err := s3Client.StatObject(context.Background(), newConfig.BucketName, fileJob.Name, opts)

		name := ""
		md5 := ""

		if err != nil {
			klog.Errorf("stat s3 object: %s, %s: %s", newConfig.BucketName, fileJob.Name, err)
			name = fileJob.Name
			md5 = "FILE_NOT_FOUND THIS WILL TRIGGER A NEW FILE TO BE UPLOADED :("
		} else {
			name = object.Key
			md5 = object.ETag
		}

		if fileJob.MD5 != md5 && fileJob.Name == name {
			// download the file from s3
			outputFileInfo <- fileJob
			continue
		}
		jobTimer.EndTimer()
	}

	// Notify parent proccess that all items have been verified
	wg.Done()
}
