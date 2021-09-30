package sync

import (
	"fmt"
	"log"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/sync/filesystem"
	"github.com/razzo-lunare/s3/pkg/sync/s3"
	"github.com/razzo-lunare/s3/pkg/utils/str"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "compare upstream and local files and only download the difference",
	RunE: func(cmd *cobra.Command, args []string) error {

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		sourceStr, err := cmd.Flags().GetString("source")
		if err != nil {
			return err
		}

		destinationStr, err := cmd.Flags().GetString("destination")
		if err != nil {
			return err
		}

		newConfig, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("Gathering config, %s", err)
		}

		source := GetSyncObject(sourceStr, newConfig)
		destination := GetSyncObject(destinationStr, newConfig)

		return sync.Run(newConfig, source, destination)
	},
}

// GetSyncObject returns an implementation of a SyncObject either s3 or filesystem
func GetSyncObject(objectOrFileInput string, config *config.S3Config) betav1.SyncObject {

	// "s3://" or "filesystem://"
	rx := regexp.MustCompile(`(?P<InputType>^[A-Za-z1-9]+)://`)
	regexGroups := str.RegexGroups(rx, objectOrFileInput)

	switch regexGroups["InputType"] {
	case "s3":
		rx := regexp.MustCompile(`s3://(?P<BucketName>[A-Za-z1-9-]+)/?(?P<S3Path>.*)`)
		regexGroups := str.RegexGroups(rx, objectOrFileInput)

		if regexGroups["BucketName"] == "" {
			log.Fatalf("No bucket name found in s3 argument\n")
		}
		bucketConfig, err := config.GetBucketCreds(regexGroups["BucketName"])
		if err != nil {
			return nil
		}

		return &s3.S3{
			// The S3Path can be empty in the regex since it could point the root of the repo
			S3Path: regexGroups["S3Path"],
			Config: bucketConfig,
		}
	case "filesystem":
		rx := regexp.MustCompile(`filesystem://(?P<FileSystemPath>.*)`)
		regexGroups := str.RegexGroups(rx, objectOrFileInput)
		return &filesystem.FileSystem{
			SyncDir: regexGroups["FileSystemPath"],
		}
	default:
		log.Fatalf("Invalid Sync Input Type %s", objectOrFileInput)

		panic("This code is never hit but the compiler doesn't know that :p")
	}
}

// NewCommand attaches the sync cmd to the partent
func NewCommand() *cobra.Command {
	// Attach the cli indicator default flags
	syncCmd.Flags().StringP("config", "", "/etc/s3/s3Config.yml", "S3 Config")
	syncCmd.Flags().StringP("source", "", "s3://TIME_SERIES_INTRADAY_V2/1min/", "Prefix to filter the contents of the bucket")
	syncCmd.Flags().StringP("destination", "", "./", "Destination to download s3 contents too")

	return syncCmd
}
