package s3

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/razzo-lunare/fortuna/pkg/config"
	"github.com/razzo-lunare/fortuna/pkg/constants"
	"github.com/razzo-lunare/fortuna/pkg/utils/market"
)

type FileInfo struct {
	Name string
	md5  string
}

func Sync(newConfig *config.FortunaConfig, startDateStr string, endDateStr string) error {
	numCPU := runtime.NumCPU()
	downloadJobs := make(chan *FileInfo, 10001)
	listJobs := make(chan string, 10001)
	var wgList sync.WaitGroup
	var wgDownload sync.WaitGroup

	for w := 1; w <= numCPU; w++ {
		go handleListObject(
			newConfig,
			&wgList,
			&wgDownload,
			listJobs,
			downloadJobs,
		)
	}

	for w := 1; w <= numCPU*2; w++ {
		go handleDownloadObject(
			newConfig,
			&wgDownload,
			downloadJobs,
		)
	}

	startDate, err := time.Parse(constants.FortunaFileFormat, startDateStr)
	if err != nil {
		return err
	}
	endDate := time.Time{}
	if endDateStr == "TODAY" {
		// Create a string and parse it so the hour,minute and second are zero
		// to be consistant every time the app runs
		today := time.Now().Format(constants.FortunaFileFormat)
		endDate, err = time.Parse(constants.FortunaFileFormat, today)
		if err != nil {
			return err
		}
	} else {
		// TODO CHANGE THIS DATE TO constants.FortunaFileFormat
		endDate, err = time.Parse(constants.FortunaFileFormat, endDateStr)
		if err != nil {
			return err
		}
	}

	fmt.Println("Sending weekdays")
	lastNWeekdays := market.GetWeekdays(startDate, endDate)
	for weekday := range lastNWeekdays {
		wgList.Add(1)
		listJobs <- weekday
	}
	close(listJobs)
	fmt.Println("Waiting list")
	wgList.Wait()

	// 1. pass data to list jobs
	// 2. close list jobs
	// 3. wait list job close
	// 4. close download jobs
	// 5. wait download jobs

	// Send signal to notify channel there is no more input
	close(downloadJobs)
	fmt.Println("Waiting Download")

	// Wait for the rest of the jobs to finish
	wgDownload.Wait()

	return nil
}

func handleListObject(newConfig *config.FortunaConfig, wgList *sync.WaitGroup, wgDownload *sync.WaitGroup, weekDayListJobs <-chan string, downloadJobRequest chan<- *FileInfo) {

	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	for weekDayListJob := range weekDayListJobs {

		opts := minio.ListObjectsOptions{
			UseV1:        false,
			WithVersions: false,
			Prefix:       fmt.Sprintf("TIME_SERIES_INTRADAY_V2/1min/%s/", weekDayListJob),
			// Recursive:    true,
			MaxKeys: 1000,
		}

		for object := range s3Client.ListObjects(context.Background(), newConfig.DigitalOceanS3StockDataBucketName, opts) {
			newFile := &FileInfo{
				Name: object.Key,
				md5:  object.ETag,
			}

			if object.Err != nil {
				fmt.Println(object.Err)
				continue
			}
			wgDownload.Add(1)
			downloadJobRequest <- newFile
		}
		wgList.Done()
	}

	// List all objects from a bucket-name with a matching prefix.
}

func handleDownloadObject(newConfig *config.FortunaConfig, wg *sync.WaitGroup, fileJobs <-chan *FileInfo) {
	s3Client, err := minio.New(newConfig.DigitalOceanS3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(newConfig.DigitalOceanS3AccessKeyID, newConfig.DigitalOceanS3SecretAccessKey, ""),
		Secure: true,
	})

	if err != nil {
		fmt.Println(err)
	}

	for fileJob := range fileJobs {
		// todo this should be a config option
		stockFile := "../fortuna-stock-data1/" + fileJob.Name
		// check if file exists
		if _, err := os.Stat(stockFile); errors.Is(err, fs.ErrNotExist) {
			err = writeDataFile(
				s3Client,
				newConfig.DigitalOceanS3StockDataBucketName,
				fileJob.Name,
				stockFile,
			)
			if err != nil {
				fmt.Println("err ", err)
			}
			wg.Done()
			continue
		}

		fileOnDiskMd5, err := hash_file_md5(stockFile)
		if err != nil {
			fmt.Printf("Error calculating m5d of file on disk, %s\n", err)
		}

		if fileOnDiskMd5 != fileJob.md5 {
			fmt.Println("UPDATE: %s" + fileJob.Name)

			err = writeDataFile(
				s3Client,
				newConfig.DigitalOceanS3StockDataBucketName,
				fileJob.Name,
				stockFile,
			)
			if err != nil {
				fmt.Println("err ", err)
			}
		}
		wg.Done()
	}
}

func hash_file_md5(filePath string) (string, error) {
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

func writeDataFile(s3Client *minio.Client, s3bucketName string, s3FileObjectName string, tickerDestionationFile string) error {
	// fmt.Println("DOWNLOADING ", s3FileObjectName)

	tickerDir := filepath.Dir(tickerDestionationFile)
	file, err := os.Open(tickerDir)
	if err != nil {
		if err == os.ErrNotExist {
			err = os.MkdirAll(tickerDir, os.ModeAppend)
			if err != nil {
				return err
			}
		}
	}
	file.Close()

	opts := minio.GetObjectOptions{}

	err = s3Client.FGetObject(context.Background(), s3bucketName, s3FileObjectName, tickerDestionationFile, opts)
	if err != nil {
		return err
	}

	return nil
}
