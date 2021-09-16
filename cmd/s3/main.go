package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/cmd/s3/sync"
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
	rootCmd.AddCommand(
		sync.NewCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Error executing cmd. %s", err)
		os.Exit(1)
	}
}
