package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var copy2Group bool
var copyDir bool

var mcp = &cobra.Command{
	Use:     "mcp [-g] [-r] FILE/DIR|SERVER:PATH SERVER:PATH|FILE/DIR",
	Short:   "copies files between hosts on a network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			core.Copy(args, copy2Group)
		}
	},
}

func init() {
	cmds["mcp"] = mcp
	mcp.PersistentFlags().BoolVarP(&copy2Group, "group", "g", false, "multi-server copy")
	mcp.PersistentFlags().BoolVarP(&copyDir, "dir", "r", false, "useless flag")
}
