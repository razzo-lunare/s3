package s3

import (
	"runtime"
	"sync"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/utils/average"
)

// Verify checks to see if the FileInfo exists
func (s *S3) Verify(inputFiles <-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	outputFileInfo := make(chan *betav1.FileInfo, 10001)

	// Verify for s3 is a bit different. I first wrote this so it would verify each file
	// one at a time when they flow through but the api calls to s3 were really slow. Instead
	// we are going to use the list function since it's great at list and gathering the md5s.
	// Then verify() can compare the inputFiles with the files from s3 that come in through List()
	s3Files, err := s.List()
	if err != nil {
		return nil, err
	}

	// Start a single background thread to manage them all
	go verifyS3Files(s.Config, inputFiles, s3Files, outputFileInfo)

	// Return the new pipeline channel
	return outputFileInfo, nil
}

func verifyS3Files(s3Config *config.S3, inputFiles <-chan *betav1.FileInfo, s3Files <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {
	klog.V(1).Info("verifying s3 data")

	numCPU := runtime.NumCPU()
	wg := &sync.WaitGroup{}
	jobTimer := average.New()

	// Load the s3 cache variable. We wait
	// here until the cache is fully loaded
	// so we don't have to deal with cache misses.
	// the code would be much faster if we find a clean way
	// to handle this.
	handleLoadS3FileCache(s3Files)

	for w := 1; w <= numCPU; w++ {
		wg.Add(1)

		go handleVerifyS3Object(
			wg,
			jobTimer,
			s3Config,
			inputFiles,
			outputFileInfo,
		)
	}

	// wait for all items to be verified
	asciiterm.PrintfInfo("Verifying Files\n")

	wg.Wait()
	close(outputFileInfo)

	asciiterm.PrintfInfo("Identified files that need to be downloaded. Time: %.2f Jobs: %d\n", jobTimer.GetAverage(), jobTimer.Count)
}

var (
	s3FileCache map[string]string = map[string]string{}
)

func handleLoadS3FileCache(s3Files <-chan *betav1.FileInfo) {
	for fileJob := range s3Files {
		klog.V(2).Infof("S3 Cache add: %s -> %s", fileJob.Name, fileJob.MD5)

		s3FileCache[fileJob.Name] = fileJob.MD5
	}
}

// handleVerifyS3Object gathers the files in the S3
func handleVerifyS3Object(wg *sync.WaitGroup, jobTimer *average.JobAverageInt, newConfig *config.S3, inputFiles <-chan *betav1.FileInfo, outputFileInfo chan<- *betav1.FileInfo) {

	for fileJob := range inputFiles {
		jobTimer.StartTimer()

		if fileMD5, ok := s3FileCache[fileJob.Name]; ok {
			if fileJob.MD5 != fileMD5 {
				// File exists in S3 but the MD5 doesn't match
				klog.V(1).Infof("VERIFY MD5 missmatch: %s -> %s != %s", fileJob.Name, fileJob.MD5, fileMD5)
				outputFileInfo <- fileJob
			}
		} else {
			klog.V(1).Infof("VERIFY Doesn't Exist S3: %s -> %s", fileJob.Name, fileJob.MD5)
			// File doesn't exist in S3
			outputFileInfo <- fileJob
		}

		jobTimer.EndTimer()
	}

	// Notify parent proccess that all items have been verified
	wg.Done()
}
