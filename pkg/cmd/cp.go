package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var copy2Group bool
var copyDir bool

var cpCmd = &cobra.Command{
	Use:     "cp [-g] [-r] FILE/DIR|SERVER:PATH SERVER:PATH|FILE/DIR",
	Aliases: []string{"mcp"},
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
	cpCmd.PersistentFlags().BoolVarP(&copy2Group, "group", "g", true, "multi-server copy")
	cpCmd.PersistentFlags().BoolVarP(&copyDir, "dir", "r", false, "useless flag")
	rootCmd.AddCommand(cpCmd)
}
