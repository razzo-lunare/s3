package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/razzo-lunare/fortuna/pkg/config"
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

		startDateStr, err := cmd.Flags().GetString("start-date")
		if err != nil {
			return err
		}
		endDateStr, err := cmd.Flags().GetString("end-date")
		if err != nil {
			return err
		}
		newConfig, err := config.NewConfig(configPath)
		if err != nil {
			return fmt.Errorf("Gathering config, %s", err)
		}

		return s3.Sync(newConfig, startDateStr, endDateStr)
	},
}
