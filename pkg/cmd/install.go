package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var installDir string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install mmh",
	Run:   func(cmd *cobra.Command, args []string) { core.Install(installDir) },
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall mmh",
	Run:   func(cmd *cobra.Command, args []string) { core.Uninstall(installDir) },
}

func init() {
	installCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "install dir")
	uninstallCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "uninstall dir")
	rootCmd.AddCommand(installCmd, uninstallCmd)
}
