package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "server command",
	Aliases: []string{"mcs"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.ServerDetail(args[0])
		} else {
			core.ListServers()
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
