package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Server command",
	Aliases: []string{"mcs"},
	Long: `
Server command.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var serverListCmd = &cobra.Command{
	Use:   "ls [SERVER_NAME]",
	Short: "List ssh server",
	Long: `
List ssh server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			mmh.PrintServerDetail(args[0])
		} else {
			mmh.ListServers()
		}
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	RootCmd.AddCommand(serverCmd)
}
