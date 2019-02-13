package cmd

import (
	"strings"

	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var singleExecServer bool
var execCmd = &cobra.Command{
	Use:     "exec SERVER_TAG CMD",
	Aliases: []string{"mec"},
	Short:   "Batch exec command",
	Long: `
Batch exec command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			cmd := strings.Join(args[1:], " ")
			mmh.Exec(args[0], cmd, singleExecServer)
		}
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
}

func init() {
	RootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&singleExecServer, "single", "s", false, "single server")
}
