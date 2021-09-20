package sync

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/types"
)

func ListS3Files(s3Config *config.S3Config, s3Prefix string) <-chan *types.FileInfo {
	asciiterm.PrintfWarn("Listing all files in s3 under: %s", s3Prefix)
	fileInfo := make(chan *types.FileInfo, 500)

	go listS3Files(s3Config, s3Prefix, fileInfo)

	return fileInfo
}

func listS3Files(s3Config *config.S3Config, s3Prefix string, outputFileInfo chan<- *types.FileInfo) {

	numCPU := runtime.NumCPU()

	s3Prefixes := make(chan string, 10001)
	goRoutineStatus := NewGoRoutineStatus(numCPU/2, s3Prefixes)
	for instanceID := 0; instanceID < numCPU/2; instanceID++ {
		go handleListS3ObjectRecursive(
			instanceID,
			goRoutineStatus,
			s3Config,
			s3Prefixes,
			outputFileInfo,
		)
	}

	// send the first prefix to the pool then the rest will search recursively
	s3Prefixes <- s3Prefix

	// TODO: This is slow... I need to figure out how to speed it up. This could
	// be cpu intensive.
	goRoutineStatus.Wait()

	// Close recursive list goroutines to clean up them
	close(s3Prefixes)

	// Notify the next step in the pipeline there are no more files listed
	close(outputFileInfo)

	asciiterm.PrintfInfo("Finished Listing all files in s3\n")
}

// handleListS3ObjectRecursive gathers the files in the S3
func handleListS3ObjectRecursive(id int, goRoutineStatus *GoRoutineStatus, newConfig *config.S3Config, s3Prefixes chan string, outputFileInfo chan<- *types.FileInfo) {

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

			newFile := &types.FileInfo{
				Name: object.Key,
				MD5:  object.ETag,
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
	channel chan string
	state   []int
}

func NewGoRoutineStatus(numberOfGoRoutines int, s3Prefixes chan string) *GoRoutineStatus {
	return &GoRoutineStatus{
		channel: s3Prefixes,
		state:   make([]int, numberOfGoRoutines),
	}
}

func (g *GoRoutineStatus) Wait() {
	// Verify no jobs are still running
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			if g.IsAllDone() {
				wg.Done()
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	wg.Wait()
}

func (g *GoRoutineStatus) SetStateRunning(goRoutineID int) {
	g.state[goRoutineID] = 1
}

func (g *GoRoutineStatus) SetStateDone(goRoutineID int) {
	g.state[goRoutineID] = 0
}

func (g *GoRoutineStatus) IsAllDone() bool {
	// If there is an item in the channel we are
	if len(g.channel) != 0 {
		return false
	}

	for _, state := range g.state {
		if state == 1 {
			return false
		}
	}

	return true
}
