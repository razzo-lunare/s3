package s3

import (
	"sync"
	"time"

	"github.com/razzo-lunare/fortuna/pkg/config"
	"github.com/razzo-lunare/fortuna/pkg/constants"
	"github.com/razzo-lunare/fortuna/pkg/utils/market"
	"github.com/razzo-lunare/s3/pkg/asciiterm"
)

type FileInfo struct {
	Name string
	md5  string
}

func Sync(newConfig *config.FortunaConfig, startDateStr string, endDateStr string) error {
	weekdays := make(chan string, 10001)
	var wgList sync.WaitGroup

	s3Files := ListS3Files(newConfig, weekdays)
	s3FilesToDownload := VerifyS3Files(newConfig, s3Files)
	downloadedFiles := DownloadS3Files(newConfig, s3FilesToDownload)

	// todo move this ugly date logic into it's own function
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

	lastNWeekdays := market.GetWeekdays(startDate, endDate)
	for weekday := range lastNWeekdays {
		wgList.Add(1)
		weekdays <- weekday
	}
	close(weekdays)
	asciiterm.PrintfInfo("%s, %d\n", "generate all stock market dates in user specified range", len(lastNWeekdays))

	count := 0
	for range downloadedFiles {
		asciiterm.PrintfWarn("Download s3 object Count: %d", count)

		count++
	}

	return nil
}
