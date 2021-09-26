package s3

import (
	"context"
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
func (s *S3) Get(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 500)

	go getS3Files(s.Config, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func getS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleGetS3Object(
			wg,
			s3Config,
			inputFiles,
			outputFileInfo,
		)
	}
	// wait for all items to be verified
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Gathered file content for the files that need to be downloaded\n")
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleGetS3Object(wg *sync.WaitGroup, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		klog.Error(err)
		return
	}

	for fileJob := range inputFiles {
		klog.V(2).Infof("Get S3: %s", fileJob.Name)

		s3Object, err := s3Client.GetObject(
			context.Background(),
			newConfig.DigitalOceanS3StockDataBucketName,
			fileJob.Name,
			minio.GetObjectOptions{},
		)

		if err != nil {
			klog.Error("error making dir, ", err)

			continue
		}
		if s3Object == nil {
			klog.Error("Why was this nil?? ", fileJob.Name)
			continue
		}
		newFileContent := &betav1.FileInfo{
			Name:    fileJob.Name,
			MD5:     fileJob.MD5,
			Content: s3Object,
		}

		outputFileInfo <- newFileContent
	}

	wg.Done()
}
