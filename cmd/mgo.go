package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var mgo = &cobra.Command{
	Use:   "mgo SERVER_NAME",
	Short: "Login server",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.SingleLogin(args[0])
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
	mgo.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
	rootCmd.AddCommand(mgo)
}
