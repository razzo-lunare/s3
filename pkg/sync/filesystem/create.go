package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// Create accepts a channel of files to create
func (f *FileSystem) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go downloadS3Files(f.SyncDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func downloadS3Files(syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}

	for w := 1; w <= numCPU*2; w++ {
		wg.Add(1)
		go handleDownloadS3ObjectNew(
			wg,
			syncDir,
			inputFiles,
			outputFileInfo,
		)
	}
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("downloaded all s3 objects\n")
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		klog.V(2).Infof("Create S3: %s", fileJob.Name)

		tickerDestinationFile := syncDir + fileJob.Name

		tickerDir := filepath.Dir(tickerDestinationFile)

		err := createDir(tickerDir)
		if err != nil {
			klog.Error(err)
			continue
		}

		localFile, err := os.Create(tickerDestinationFile)
		if err != nil {
			klog.Error(err)
			continue
		}
		if fileJob.Content != nil {
			if _, err = io.Copy(localFile, fileJob.Content); err != nil {
				klog.Error(err)
				continue
			}
		} else {
			// klog.Error("File Content to download was nil: ", fileJob.Name)
		}

		localFile.Close()

		outputFileInfo <- fileJob
	}

	wg.Done()
}

var (
	dirCache = map[string]interface{}{}
)

func createDir(pathToCreate string) error {
	// cache the directories we have checked so we don't have to stat them multiple times
	// don't lock the cache map because we don't care about the state of the data.
	if _, ok := dirCache[pathToCreate]; !ok {

		// Create the directory if it doesn't exist locally
		_, err := os.Stat(pathToCreate)
		if err != nil && os.IsNotExist(err) {
			err = os.MkdirAll(pathToCreate, 0700)
			if err != nil {
				return fmt.Errorf("error making dir, %w", err)
			}
		}

		dirCache[pathToCreate] = nil
	}

	return nil
}
