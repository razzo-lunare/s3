package filesystem

import (
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
	outputFileInfo := make(chan *betav1.FileInfo, 500)

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

var (
	dirCache     = map[string]interface{}{}
	dirCacheLock = sync.Mutex{}
)

// handleListS3Object gathers the files in the S3
func handleDownloadS3ObjectNew(wg *sync.WaitGroup, syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		klog.V(2).Infof("Create S3: %s", fileJob.Name)
		// TODO add ticker destination as a CLI flag!!
		tickerDestinationFile := syncDir + fileJob.Name

		tickerDir := filepath.Dir(tickerDestinationFile)
		// TODO: this is un-efficient but we need to make sure the directories exist...
		// maybe we should cache the directories we have checked
		// if _, ok := dirCache[tickerDir]; !ok {
		_, err := os.Stat(tickerDir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(tickerDir, 0700)
				if err != nil {
					klog.Error("error making dir, ", err)

					continue
				}
			}
		}
		// dirCacheLock.Lock()
		// dirCache[tickerDir] = nil
		// }

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
