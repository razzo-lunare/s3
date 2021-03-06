package filesystem

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Verify checks to see if the FileInfo exists on the filesystem already and the md5 matches if it doesn't it
// passes them to the next step to be downloaded
func (f *FileSystem) Verify(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	go verifyS3Files(f.SyncDir, inputFiles, outputFileInfo)

	return outputFileInfo, nil
}

func verifyS3Files(syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	klog.V(1).Info("Identifying files that need to be downloaded")

	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	for w := 1; w <= numCPU; w++ {
		wg.Add(1)
		go handleVerifyS3Object(
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

	asciiterm.PrintfInfo("Identified files that need to be downloaded. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

// handleListS3Object gathers the files in the S3
func handleVerifyS3Object(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, syncDir string, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		klog.V(2).Infof("Verify S3: %s", fileJob.Name)
		jobTimer.StartTimer()
		stockFile := syncDir + fileJob.Name

		// S3 object doesn't exist on filesystem
		if _, err := os.Stat(stockFile); os.IsNotExist(err) {
			// download the file from s3
			outputFileInfo <- fileJob

			continue
		}

		fileOnDiskMd5, err := hashFileMd5(stockFile)
		if err != nil {
			fmt.Printf("Error calculating m5d of file on disk, %s\n", err)
			continue
		}

		// S3 OBJECT MD5 DOESN'T MATCH THE FILE ON THE FILESYSTEM
		if fileOnDiskMd5 != fileJob.MD5 {
			// download the file from s3
			outputFileInfo <- fileJob
		}
		jobTimer.EndTimer()
	}

	// Notify parent proccess that all items have been verified
	wg.Done()
}

func hashFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "FILE_NOT_FOUND THIS WILL TRIGGER A NEW FILE TO DOWNLOAD", nil
		}

		return returnMD5String, err
	}

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		file.Close()
		return "", err
	}

	file.Close()

	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil
}
