package s3

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
)

func ListS3Files(s3Config *config.S3Config, s3Prefix string) <-chan *FileInfo {
	asciiterm.PrintfWarn("Listing all files in s3 under: %s", s3Prefix)
	fileInfo := make(chan *FileInfo, 500)

	go listS3Files(s3Config, s3Prefix, fileInfo)

	return fileInfo
}

func listS3Files(s3Config *config.S3Config, s3Prefix string, outputFileInfo chan<- *FileInfo) {

	numCPU := runtime.NumCPU()
	// wg := &sync.WaitGroup{}

	s3Prefixes := make(chan string, 10001)
	goRoutineStatus := NewGoRoutineStatus(numCPU / 2)
	for w := 0; w < numCPU/2; w++ {
		go handleListS3ObjectRecursive(
			w,
			goRoutineStatus,
			s3Config,
			s3Prefixes,
			outputFileInfo,
		)
	}

	// send the first prefix to the pool then the rest will search recursively
	s3Prefixes <- s3Prefix

	// TODO: This is slow... I need to figure out how to speed it up
	// Verify no jobs come in for 5 seconds
	totalIsAllDones := 0
	for {
		if goRoutineStatus.IsAllDone(s3Prefixes) {
			totalIsAllDones++
			// Take into account any false alarms
			if totalIsAllDones == 2 {
				break
			}
			time.Sleep(200 * time.Millisecond)
		} else {
			totalIsAllDones = 0
		}
	}

	// Kill all list goroutines since there is no more data
	close(s3Prefixes)
	close(outputFileInfo)

	asciiterm.PrintfInfo("Finished Listing all files in s3\n")
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleListS3ObjectRecursive(id int, goRoutineStatus *GoRoutineStatus, newConfig *config.S3Config, s3Prefixes chan string, outputFileInfo chan<- *FileInfo) {

	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for s3PrefixJob := range s3Prefixes {

		goRoutineStatus.SetStateRunning(id)

		opts := minio.ListObjectsOptions{
			UseV1:        false,
			WithVersions: false,
			Prefix:       s3PrefixJob,
			Recursive:    false,
			MaxKeys:      5000,
		}

		for object := range s3Client.ListObjects(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, opts) {
			if isDir(object.Key) {

				newPrefix := object.Key
				s3Prefixes <- newPrefix

				continue
			}

			newFile := &FileInfo{
				Name: object.Key,
				md5:  object.ETag,
			}

			if object.Err != nil {
				fmt.Println(object.Err)
				continue
			}

			outputFileInfo <- newFile
		}

		goRoutineStatus.SetStateDone(id)

	}

}

func isDir(objectPath string) bool {
	return strings.HasSuffix(objectPath, "/")
}

type GoRoutineStatus struct {
	state []int
}

func NewGoRoutineStatus(numberOfGoRoutines int) *GoRoutineStatus {
	return &GoRoutineStatus{
		state: make([]int, numberOfGoRoutines),
	}
}

func (g *GoRoutineStatus) SetStateRunning(goRoutineID int) {
	g.state[goRoutineID] = 1
}

func (g *GoRoutineStatus) SetStateDone(goRoutineID int) {
	g.state[goRoutineID] = 0
}

func (g *GoRoutineStatus) IsAllDone(thing chan string) bool {
	if len(thing) != 0 {
		return false
	}

	for _, state := range g.state {
		if state == 1 {
			return false
		}
	}
	return true
}
