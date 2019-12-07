package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/mritd/mmh/utils"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mmh",
	Short: "a simple multi-server ssh tool",
	Long: `
a simple multi-server ssh tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.LoadConfig()
		mmh.InteractiveLogin()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		utils.Exit(err.Error(), -1)
	}
}
