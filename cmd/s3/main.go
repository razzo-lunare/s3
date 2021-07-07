package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

var rootCmd = &cobra.Command{
	Use:   "s3",
	Short: "Simple fast s3 cli",
}

func init() {
	// Add klog flags to cli for verbosity flag
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
}

func main() {
	rootCmd.AddCommand(syncCmd)

	// Attach the cli indicator default flags
	syncCmd.Flags().StringP("config", "", "/etc/fortuna/fortuna.yml", "Fortuna Config")
	syncCmd.Flags().StringP("start-date", "", "2021-01-01", "Date on when to start the simulation")
	syncCmd.Flags().StringP("end-date", "", "TODAY", "Date on when to start the simulation")

	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Error executing cmd. %s", err)
		os.Exit(1)
	}
}
