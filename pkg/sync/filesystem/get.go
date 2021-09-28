package filesystem

import (
	"bytes"
	"os"
	"runtime"
	"sync"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Get gathers the content for incoming FileInfo and passes them
// to the next step

// Get downloads the content for incoming FileInfo and passes them
// to the next step
func (s *FileSystem) Get(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go getS3Files(inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func getS3Files(inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleGetS3Object(
			wg,
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
func handleGetS3Object(wg *sync.WaitGroup, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		klog.V(2).Infof("Get S3: %s", fileJob.Name)

		fileContent, err := os.ReadFile(fileJob.Name)
		if err != nil {
			klog.Error("FileSystem Get() read file error: ", err)
		}

		newFileContent := &betav1.FileInfo{
			Name:    fileJob.Name,
			MD5:     fileJob.MD5,
			Content: bytes.NewReader(fileContent),
		}

		outputFileInfo <- newFileContent
	}

	wg.Done()
}
