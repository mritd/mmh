package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:     "ctx",
	Short:   "change current context",
	Aliases: []string{"mcx"},
	Long: `
change current context.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			mmh.LoadConfig()
			mmh.InteractiveSetConfig()
		} else {
			_ = cmd.Help()
		}
	},
}

var cfListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list context",
	Long: `
list context`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.LoadConfig()
		mmh.ListConfig()
	},
}

var ctxSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set context",
	Long: `
set context`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
			return
		}
		mmh.LoadConfig()
		mmh.SetConfig(args[0])
	},
}

func init() {
	cfCmd.AddCommand(cfListCmd, ctxSetCmd)
	RootCmd.AddCommand(cfCmd)
}
