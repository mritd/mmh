package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var uninstallDir string
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall",
	Long:  "uninstall mmh.",
	Run: func(cmd *cobra.Command, args []string) {
		mmh.Uninstall(uninstallDir)
	},
}

func init() {
	uninstallCmd.PersistentFlags().StringVar(&uninstallDir, "dir", "/usr/local/bin", "uninstall dir")
	rootCmd.AddCommand(uninstallCmd)
}
