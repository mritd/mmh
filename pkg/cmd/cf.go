package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"mcx"},
	Short:   "change current context",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			core.InteractiveSetConfig()
		} else {
			core.SetConfig(args[0])
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var res []string
		for _, info := range core.Configs {
			res = append(res, info.Name)
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(cfCmd)
}
