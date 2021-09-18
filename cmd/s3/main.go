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

func initLogger() {
	// Add klog flags to cli for verbosity flag
	klog.InitFlags(nil)

	// Attach flags to the command line so they can be parsed
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	// Hide all klog args since they are useless
	pflag.CommandLine.MarkHidden("add_dir_header")
	pflag.CommandLine.MarkHidden("alsologtostderr")
	pflag.CommandLine.MarkHidden("log_backtrace_at")
	pflag.CommandLine.MarkHidden("log_dir")
	pflag.CommandLine.MarkHidden("log_file")
	pflag.CommandLine.MarkHidden("log_file_max_size")
	pflag.CommandLine.MarkHidden("logtostderr")
	pflag.CommandLine.MarkHidden("one_output")
	pflag.CommandLine.MarkHidden("skip_headers")
	pflag.CommandLine.MarkHidden("skip_log_headers")
	pflag.CommandLine.MarkHidden("stderrthreshold")
	pflag.CommandLine.MarkHidden("vmodule")
}

func main() {

	rootCmd.AddCommand(
		sync.NewCommand(),
	)

	initLogger()

	if err := rootCmd.Execute(); err != nil {
		klog.Errorf("Error executing cmd. %s", err)
		os.Exit(1)
	}
}
