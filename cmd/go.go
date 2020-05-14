package cmd

import (
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:     "go SERVER",
	Aliases: []string{"mgo"},
	Short:   "login server",
	Long:    "login server.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.SingleLogin(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(goCmd)
}
