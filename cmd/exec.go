package cmd

import (
	"strings"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var singleServer bool

var execCmd = &cobra.Command{
	Use:     "exec SERVER COMMAND",
	Aliases: []string{"mec"},
	Short:   "batch exec command",
	Long:    "batch exec command.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			cmd := strings.Join(args[1:], " ")
			core.Exec(cmd, args[0], singleServer, false)
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&singleServer, "single", "s", false, "single server")
}
