package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Create accepts a channel of files to create
func (f *FileSystem) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 500)

	go downloadS3Files(f.S3Config, f.DestinationDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func downloadS3Files(s3Config *config.S3Config, destinationDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU*2; w++ {
		wg.Add(1)
		go handleDownloadS3ObjectNew(
			wg,
			s3Config,
			destinationDir,
			inputFiles,
			outputFileInfo,
		)
	}
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("downloaded all s3 objects\n")
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, newConfig *config.S3Config, destinationDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
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
