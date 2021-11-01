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
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Get downloads the content for incoming FileInfo and passes them
// to the next step
func (s *S3) Get(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go getS3Files(s.Config, s.S3Path, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func getS3Files(s3Config *config.S3, S3PathPrefix string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU; w++ {
		wg.Add(1)
		go handleGetS3Object(
			wg,
			jobTimer,
			s3Config,
			S3PathPrefix,
			inputFiles,
			outputFileInfo,
		)
	}
	// wait for all items to be verified
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Gathered file content for the files that need to be downloaded. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleGetS3Object(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, newConfig *config.S3, S3PathPrefix string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	s3Client, err := minio.New(newConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.AccessKeyID, newConfig.AccessKey, ""),
		Secure: true,
	})
	if err != nil {
		klog.Error(err)
		return
	}

	for fileJob := range inputFiles {
		jobTimer.StartTimer()

		// Append the s3 prefix to the job filename so it's a full path to the resource in s3
		fullObjectPath := S3PathPrefix + fileJob.Name

		klog.V(2).Infof("Get S3 obj name: %s bucket name: %s", fullObjectPath, newConfig.BucketName)

		s3Object, err := s3Client.GetObject(
			context.Background(),
			newConfig.BucketName,
			fullObjectPath,
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

		jobTimer.EndTimer()
	}

	wg.Done()
}
