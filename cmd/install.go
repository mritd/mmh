package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var installDir string

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install",
	Long:  "install mmh.",
	Run:   func(cmd *cobra.Command, args []string) { mmh.Install(installDir) },
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall",
	Long:  "uninstall mmh.",
	Run:   func(cmd *cobra.Command, args []string) { mmh.Uninstall(installDir) },
}

func init() {
	installCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "install dir")
	uninstallCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "uninstall dir")
	rootCmd.AddCommand(installCmd, uninstallCmd)
}
