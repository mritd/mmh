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
			cmd := strings.Join(args[1:], " ")
			core.Exec(cmd, args[0], execGroup, false)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&execGroup, "group", "g", true, "multi-server exec")
}
