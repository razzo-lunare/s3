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
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Create accepts a channel of files to create
func (f *FileSystem) Create(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go downloadS3Files(f.SyncDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func downloadS3Files(syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU() * 3
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU; w++ {
		wg.Add(1)
		go handleDownloadS3ObjectNew(
			wg,
			jobTimer,
			syncDir,
			inputFiles,
			outputFileInfo,
		)
	}
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("downloaded all s3 objects. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		jobTimer.StartTimer()
		klog.V(2).Infof("Create file: %s", fileJob.Name)

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

		if fileJob.Content != nil && localFile != nil {
			if written, err := io.Copy(localFile, fileJob.Content); err != nil {
				localFile.Close()
				klog.Errorf("File: %s. Written Count %d Error: %s", tickerDestinationFile, written, err)
				continue
			}
		} else {
			klog.Error("File Content to download was nil: ", fileJob.Name)
		}

		localFile.Close()
		klog.V(1).Info("Created: ", fileJob.Name)

		outputFileInfo <- fileJob
		jobTimer.EndTimer()
	}

	wg.Done()
}

var (
	dirCache     = map[string]interface{}{}
	dirCacheLock = sync.Mutex{}
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

		dirCacheLock.Lock()
		dirCache[pathToCreate] = nil
		dirCacheLock.Unlock()
	}

	return nil
}
