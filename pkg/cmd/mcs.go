package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var serverSort bool
var mcs = &cobra.Command{
	Use:   "mcs [SERVER_NAME]",
	Short: "print server list",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.PrintServerDetail(args[0])
		} else {
			core.PrintServers(serverSort)
		}
	},
}

func init() {
	cmds["mcs"] = mcs
	mcs.PersistentFlags().BoolVarP(&serverSort, "sort", "s", false, "sort server list")
}
