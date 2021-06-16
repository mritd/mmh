package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var copy2Group bool

var mcp = &cobra.Command{
	Use:   "mcp [-g] [-r] FILE/DIR|SERVER:PATH SERVER:PATH|FILE/DIR",
	Short: "copies files between hosts on a network",
	Run: func(cmd *cobra.Command, args []string) {
		if completionShell != "" {
			GenCompletion(cmd, completionShell)
			return
		}

		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			core.Copy(args, copy2Group)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var res []string
		for _, s := range core.ListServers(true) {
			res = append(res, s.Name)
		}
		return res, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	cmds["mcp"] = mcp
	mcp.PersistentFlags().BoolVarP(&copy2Group, "group", "g", false, "multi-server copy")
	mcp.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
}
