package cmd

import (
	"github.com/mritd/mmh/utils"
	"github.com/spf13/cobra"
)

var uninstallDir string
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall mmh",
	Long: `
Uninstall mmh.`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.Uninstall(uninstallDir)
	},
}

func init() {
	uninstallCmd.PersistentFlags().StringVar(&uninstallDir, "dir", "/usr/bin", "uninstall dir")
	RootCmd.AddCommand(uninstallCmd)
}
