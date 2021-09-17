package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/razzo-lunare/s3/pkg/config"
	"github.com/razzo-lunare/s3/pkg/s3"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "compare upstream and local files and only download the difference",
	RunE: func(cmd *cobra.Command, args []string) error {

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		s3Prefix, err := cmd.Flags().GetString("s3-prefix")
		if err != nil {
			return err
		}
		destinationDir, err := cmd.Flags().GetString("destination-dir")
		if err != nil {
			return err
		}

		newConfig, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("Gathering config, %s", err)
		}

		return s3.Sync(newConfig, s3Prefix, destinationDir)
	},
}

func NewCommand() *cobra.Command {
	// Attach the cli indicator default flags
	syncCmd.Flags().StringP("config", "", "/etc/s3/s3Config.yml", "S3 Config")
	syncCmd.Flags().StringP("s3-prefix", "", "TIME_SERIES_INTRADAY_V2/1min/", "Prefix to filter the contents of the bucket")
	syncCmd.Flags().StringP("destination-dir", "", "./", "Destination to download s3 contents too")

	return syncCmd
}
