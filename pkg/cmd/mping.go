package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var mping = &cobra.Command{
	Use:   "mping SERVER",
	Short: "ping server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.Ping(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	cmds["mping"] = mping
}
