package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var execGroup bool

var mec = &cobra.Command{
	Use:   "mec [OPTIONS] SERVER|TAG COMMAND",
	Short: "Batch exec command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			cs := strings.Join(args[1:], " ")
			core.Exec(cs, args[0], execGroup, false)
		}
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, fmt.Sprintf("%s\tfrom %s(%s)", s.Name, filepath.Base(s.ConfigPath), s.Name))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	mec.PersistentFlags().BoolVarP(&execGroup, "tag", "t", false, "server tag")
	rootCmd.AddCommand(mec)
}
