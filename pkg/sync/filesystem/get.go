package filesystem

import (
	"bytes"
	"os"
	"runtime"
	"sync"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Get gathers the content for incoming FileInfo and passes them
// to the next step

// Get downloads the content for incoming FileInfo and passes them
// to the next step
func (s *FileSystem) Get(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go getS3Files(s.SyncDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func getS3Files(syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU/2; w++ {
		wg.Add(1)
		go handleGetS3Object(
			wg,
			jobTimer,
			syncDir,
			inputFiles,
			outputFileInfo,
		)
	}
	// wait for all items to be verified
	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Gathered file content for the files that need to be downloaded. Time: %f\n", jobTimer.GetAverage())
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleGetS3Object(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		jobTimer.StartTimer()
		stockFile := syncDir + fileJob.Name
		klog.V(1).Infof("Get S3: %s", stockFile)

		file, err := os.Open(stockFile)
		if err != nil {
			klog.Errorf("FileSystem Get() open file %s: %s", stockFile, err)
			continue
		}
		fileStat, err := file.Stat()
		if err != nil {
			klog.Errorf("FileSystem Get() stat file %s: %s", stockFile, err)
			continue
		}
		file.Close()

		fileContent, err := os.ReadFile(stockFile)
		if err != nil {
			klog.Error("FileSystem Get() read file error: ", err)
			continue
		}

		newFileContent := &betav1.FileInfo{
			Name:        fileJob.Name,
			MD5:         fileJob.MD5,
			Content:     bytes.NewReader(fileContent),
			ContentSize: fileStat.Size(),
		}

		outputFileInfo <- newFileContent
		jobTimer.EndTimer()
	}

	wg.Done()
}
