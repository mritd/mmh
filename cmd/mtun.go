package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var tunLeftAddr, tunRightAddr string
var tunReverse bool

var mtun = &cobra.Command{
	Use:   "mtun SERVER -l LEFT_ADDR -r RIGHT_ADDR [OPTIONS]",
	Short: "Open a ssh tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && tunLeftAddr != "" && tunRightAddr != "" {
			core.Tunnel(args[0], tunLeftAddr, tunRightAddr, tunReverse)
		} else {
			_ = cmd.Help()
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, fmt.Sprintf("%s\tfrom %s(%s)", s.Name, filepath.Base(s.ConfigPath), s.Name))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	mtun.PersistentFlags().StringVarP(&tunLeftAddr, "left", "l", "", "left address")
	mtun.PersistentFlags().StringVarP(&tunRightAddr, "right", "r", "", "right address")
	mtun.PersistentFlags().BoolVar(&tunReverse, "reverse", false, "reverse tcp tunnel(right to left)")
	_ = mtun.RegisterFlagCompletionFunc("left", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"127.0.0.1"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	_ = mtun.RegisterFlagCompletionFunc("right", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"127.0.0.1"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	rootCmd.AddCommand(mtun)
}
