package cmd

import (
	"fmt"
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var mcx = &cobra.Command{
	Use:   "mcx",
	Short: "change config file",
	Run: func(cmd *cobra.Command, args []string) {
		if completionShell != "" {
			GenCompletion(cmd, completionShell)
			return
		}

		if len(args) < 1 {
			_ = cmd.Help()
			return
		}

		core.SetConfig(args[0])
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		var res []string
		for _, info := range core.Configs {
			res = append(res, fmt.Sprintf("%s\t%s", info.Name, info.Path))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	cmds["mcx"] = mcx
	mcx.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
}
