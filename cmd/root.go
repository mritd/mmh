package cmd

import (
	"os"
	"path/filepath"

	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "mmh",
	Short:            "a simple multi-server ssh tool",
	Long:             "a simple multi-server ssh tool.",
	TraverseChildren: true,
	Run:              func(cmd *cobra.Command, args []string) { mmh.InteractiveLogin() },
}

func Execute() {
	runCmd := rootCmd

	subCmd, _, _ := rootCmd.Find([]string{filepath.Base(os.Args[0])})
	if subCmd != nil {
		runCmd = subCmd
		rootCmd.SetArgs(append([]string{subCmd.Name()}, os.Args[1:]...))
	}

	if runCmd.Name() != "install" && runCmd.Name() != "uninstall" {
		mmh.LoadConfig()
	}

	if err := runCmd.Execute(); err != nil {
		mmh.Exit(err.Error(), -1)
	}
}
