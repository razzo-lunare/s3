package s3

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
)

// VerifyFiles if the file doesn't exist on the filesystem or the md5 sum doesn't match
// make that the file needs to be downloaded
func DownloadS3Files(fortunaConfig *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo) <-chan *FileInfo {
	outputFileInfo := make(chan *FileInfo)

	go downloadS3Files(fortunaConfig, destinationDir, inputFiles, outputFileInfo)

	return outputFileInfo
}

func downloadS3Files(fortunaConfig *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo, outputFileInfo chan<- *FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU*2; w++ {
		wg.Add(1)
		go handleDownloadS3ObjectNew(
			wg,
			fortunaConfig,
			destinationDir,
			inputFiles,
			outputFileInfo,
		)
	}
	wg.Wait()

	close(outputFileInfo)

	asciiterm.PrintfInfo("downloaded all s3 objects\n")
	time.Sleep(1 * time.Second)
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, newConfig *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo, outputFileInfo chan<- *FileInfo) {
	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})

	if err != nil {
		fmt.Println(err)
	}

	for fileJob := range inputFiles {
		// TODO add ticker destination as a CLI flag!!
		tickerDestinationFile := destinationDir + fileJob.Name

		tickerDir := filepath.Dir(tickerDestinationFile)
		file, err := os.Open(tickerDir)
		if err != nil {
			if err == os.ErrNotExist {
				err = os.MkdirAll(tickerDir, os.ModeAppend)
				if err != nil {
					fmt.Println("error making dir, ", err)

					continue
				}
			}
		}
		file.Close()

		opts := minio.GetObjectOptions{}

		err = s3Client.FGetObject(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, fileJob.Name, tickerDestinationFile, opts)
		if err != nil {
			fmt.Println("error making dir, ", err)

			continue
		}

		outputFileInfo <- fileJob
	}

	wg.Done()
}
