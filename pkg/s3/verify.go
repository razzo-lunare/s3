package s3

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
)

// VerifyFiles if the file doesn't exist on the filesystem or the md5 sum doesn't match
// make that the file needs to be downloaded
func VerifyS3Files(s3Config *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo) <-chan *FileInfo {
	outputFileInfo := make(chan *FileInfo, 500)

	go verifyS3Files(s3Config, destinationDir, inputFiles, outputFileInfo)

	return outputFileInfo
}

func verifyS3Files(s3Config *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo, outputFileInfo chan<- *FileInfo) {
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

	asciiterm.PrintfInfo("Identified files that need to be downloaded, %d\n", len(outputFileInfo))
}

// handleListS3Object gathers the files in the S3
func handleVerifyS3Object(wg *sync.WaitGroup, newConfig *config.S3Config, destinationDir string, inputFiles <-chan *FileInfo, outputFileInfo chan<- *FileInfo) {

	for fileJob := range inputFiles {

		stockFile := destinationDir + fileJob.Name
		// fmt.Println("File: ", stockFile)
		// S3 OBJECT DOESN'T EXIT ON THE FILESYSTEM
		if _, err := os.Stat(stockFile); errors.Is(err, fs.ErrNotExist) {
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
		if fileOnDiskMd5 != fileJob.md5 {
			fmt.Println("UPDATE: %s" + fileJob.Name)

			// download the file from s3
			outputFileInfo <- fileJob
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
