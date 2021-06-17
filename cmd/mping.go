package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var mping = &cobra.Command{
	Use:   "mping SERVER",
	Short: "ping server",
	Run: func(cmd *cobra.Command, args []string) {
		if completionShell != "" {
			GenCompletion(cmd, completionShell)
			return
		}

		if len(args) == 1 {
			core.Ping(args[0])
		} else {
			_ = cmd.Help()
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
	cmds["mping"] = mping
	mping.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
}
