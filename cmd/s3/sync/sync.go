package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/sync"
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

		destination := &s3.S3{
			S3Prefix:       sourceStr,
			S3Config:       newConfig,
			DestinationDir: destinationStr,
		}

		source := &filesystem.FileSystem{
			SourceDir:      sourceStr,
			S3Config:       newConfig,
			DestinationDir: destinationStr,
		}

		return sync.Run(newConfig, source, destination)
	},
}

func NewCommand() *cobra.Command {
	// Attach the cli indicator default flags
	syncCmd.Flags().StringP("config", "", "/etc/s3/s3Config.yml", "S3 Config")
	syncCmd.Flags().StringP("source", "", "s3://TIME_SERIES_INTRADAY_V2/1min/", "Prefix to filter the contents of the bucket")
	syncCmd.Flags().StringP("destination", "", "./", "Destination to download s3 contents too")

	return syncCmd
}
