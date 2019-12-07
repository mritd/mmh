package cmd

import (
	"strings"

	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var singleServer bool
var execCmd = &cobra.Command{
	Use:     "exec SERVER_TAG CMD",
	Aliases: []string{"mec"},
	Short:   "batch exec command",
	Long: `
batch exec command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
		} else {
			cmd := strings.Join(args[1:], " ")
			mmh.LoadConfig()
			mmh.Exec(args[0], cmd, singleServer, false)
		}
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&singleServer, "single", "s", false, "single server")
}
