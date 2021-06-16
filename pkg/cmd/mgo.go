package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var mgo = &cobra.Command{
	Use:   "mgo SERVER_NAME",
	Short: "login server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.SingleLogin(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	cmds["mgo"] = mgo
}
