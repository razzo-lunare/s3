package s3

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Create accepts a channel of files to create
func (s *S3) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 500)

	go downloadS3Files(s.S3Config, s.DestinationDir, inputFiles, outputFileInfo)

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
	for fileJob := range inputFiles {
		fmt.Println("Handle Creating file: ", fileJob.Name)
	}

	wg.Done()
}
