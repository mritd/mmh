package cmd

import (
	"strings"

	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var execGroup bool

var execCmd = &cobra.Command{
	Use:     "exec SERVER|TAG COMMAND",
	Aliases: []string{"mec"},
	Short:   "batch exec command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			cs := strings.Join(args[1:], " ")
			core.Exec(cs, args[0], execGroup, false)
		}
	},
}

func init() {
	execCmd.PersistentFlags().BoolVarP(&execGroup, "group", "g", false, "multi-server exec")
	rootCmd.AddCommand(execCmd)
}
