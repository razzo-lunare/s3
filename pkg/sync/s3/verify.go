package s3

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"k8s.io/klog/v2"
)

// Verify checks to see if the FileInfo exists
func (s *S3) Verify(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 500)

	go verifyS3Files(s.S3Config, s.DestinationDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func verifyS3Files(s3Config *config.S3Config, destinationDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleVerifyS3Object(
			wg,
			s3Config,
			destinationDir,
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
func handleVerifyS3Object(wg *sync.WaitGroup, newConfig *config.S3Config, destinationDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for fileJob := range inputFiles {
		opts := minio.StatObjectOptions{}

		object, err := s3Client.StatObject(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, fileJob.Name, opts)
		if err != nil {
			klog.Error("stat s3 object: ", err)
			continue
		}

		newFile := &betav1.FileInfo{
			Name: object.Key,
			MD5:  object.ETag,
		}
		if fileJob != newFile {
			// download the file from s3
			outputFileInfo <- fileJob
			continue
		}
	}

	// Notify parent proccess that all items have been verified
	wg.Done()
}

func hashFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return "FILE_NOT_FOUND THIS WILL TRIGGER A NEW FILE TO DOWNLOAD", nil
		}

		return returnMD5String, err
	}

	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}
