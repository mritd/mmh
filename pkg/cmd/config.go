package cmd

import (
	"fmt"
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"mcx"},
	Short:   "change config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
		}
		core.SetConfig(args[0])
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		var res []string
		for _, info := range core.Configs {
			res = append(res, fmt.Sprintf("%s    %s", info.Name, info.Path))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(cfCmd)
}
