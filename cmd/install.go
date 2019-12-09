package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var installDir string
var installCmd = &cobra.Command{
	Use:    "install",
	Short:  "install",
	Long:   "install mmh.",
	PreRun: func(cmd *cobra.Command, args []string) { mmh.LoadConfig() },
	Run: func(cmd *cobra.Command, args []string) {
		mmh.Install(installDir)
	},
}

func init() {
	installCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "install dir")
	rootCmd.AddCommand(installCmd)
}
