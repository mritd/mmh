package cmd

import (
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var copy2Group bool

var mcp = &cobra.Command{
	Use:   "mcp [-t] FILE/DIR|SERVER:PATH SERVER:PATH|FILE/DIR",
	Short: "Copies files between hosts on a network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			core.Copy(args, copy2Group)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, s.Name)
		}
		return res, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	mcp.PersistentFlags().BoolVarP(&copy2Group, "tag", "t", false, "server tag")
	rootCmd.AddCommand(mcp)
}
