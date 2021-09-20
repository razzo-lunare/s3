package sync

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
	"github.com/razzo-lunare/s3/pkg/sync/filesystem"
	"github.com/razzo-lunare/s3/pkg/sync/s3"
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

		destination := GetSyncObject("s3", sourceStr, destinationStr, newConfig)

		source := GetSyncObject("filesystem", sourceStr, destinationStr, newConfig)

		return sync.Run(newConfig, source, destination)
	},
}

func GetSyncObject(objType string, objectSource string, objectDestination string, config *config.S3Config) betav1.SyncObject {
	// TODO infer the objType from the source and destination string

	switch objType {
	case "s3":
		return &s3.S3{
			S3Prefix:       objectSource,
			S3Config:       config,
			DestinationDir: objectDestination,
		}
	case "filesystem":
		return &filesystem.FileSystem{
			SourceDir:      objectSource,
			S3Config:       config,
			DestinationDir: objectDestination,
		}
	default:
		log.Fatalf("Invalid Sync Object Type %s", objType)

		panic("This code is never hit but the compiler doesn't know that :p")
	}
}

func NewCommand() *cobra.Command {
	// Attach the cli indicator default flags
	syncCmd.Flags().StringP("config", "", "/etc/s3/s3Config.yml", "S3 Config")
	syncCmd.Flags().StringP("source", "", "s3://TIME_SERIES_INTRADAY_V2/1min/", "Prefix to filter the contents of the bucket")
	syncCmd.Flags().StringP("destination", "", "./", "Destination to download s3 contents too")

	return syncCmd
}
