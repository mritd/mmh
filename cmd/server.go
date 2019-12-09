package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "server command",
	Aliases: []string{"mcs"},
	Long:    "server command.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var serverListCmd = &cobra.Command{
	Use:   "ls [SERVER_NAME]",
	Short: "list ssh server",
	Long:  "list ssh server.",
	Run: func(cmd *cobra.Command, args []string) {
		mmh.LoadConfig()
		if len(args) == 1 {
			mmh.PrintServerDetail(args[0])
		} else {
			mmh.ListServers()
		}
	},
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add ssh server",
	Long:  "add ssh server.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			mmh.AddServer()
		} else {
			_ = cmd.Help()
		}
	},
}

var serverDelCmd = &cobra.Command{
	Use:   "del SERVER_NAME",
	Short: "delete ssh server",
	Long:  "delete ssh server.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			mmh.DeleteServer(args)
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverDelCmd)
	rootCmd.AddCommand(serverCmd)
}
