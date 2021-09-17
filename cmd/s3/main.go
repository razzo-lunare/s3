package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/cmd/s3/sync"
)

var rootCmd = &cobra.Command{
	Use:   "s3",
	Short: "Simple fast s3 cli",
}

func main() {

	rootCmd.AddCommand(
		sync.NewCommand(),
	)
	flagset := flag.CommandLine
	verbosity := 3
	flagset.IntVar(&verbosity, "v", 0, "number for the log level verbosity")
	// Add klog flags to cli for verbosity flag
	klog.InitFlags(flagset)

	// flag.Set("v", fmt.Sprintf("%d", verbosity))
	flag.Parse()

	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Error executing cmd. %s", err)
		os.Exit(1)
	}
}
