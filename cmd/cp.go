package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var groupCPServer bool
var copyDir bool
var cpCmd = &cobra.Command{
	Use:     "cp [-r] FILE/DIR|SERVER_TAG:PATH SERVER_NAME:PATH|FILE/DIR",
	Aliases: []string{"mcp"},
	Short:   "Copies files between hosts on a network",
	Long: `
Copies files between hosts on a network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			mmh.Copy(args, groupCPServer)
		}
	},
}

func init() {
	RootCmd.AddCommand(cpCmd)
	cpCmd.PersistentFlags().BoolVarP(&groupCPServer, "group", "g", false, "multi-server copy")
	cpCmd.PersistentFlags().BoolVarP(&copyDir, "dir", "r", false, "useless flag")
}
