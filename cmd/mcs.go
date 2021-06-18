package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var serverSort bool
var mcs = &cobra.Command{
	Use:   "mcs [SERVER_NAME]",
	Short: "Print server list",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.PrintServerDetail(args[0])
		} else {
			core.PrintServers(serverSort)
		}
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		var res []string
		for _, s := range core.ListServers(true) {
			res = append(res, fmt.Sprintf("%s\tfrom %s(%s)", s.Name, filepath.Base(s.ConfigPath), s.Name))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	mcs.PersistentFlags().BoolVarP(&serverSort, "sort", "s", false, "sort server list")
	mcs.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
	rootCmd.AddCommand(mcs)
}
