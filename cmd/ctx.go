package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var ctxCmd = &cobra.Command{
	Use:     "ctx",
	Short:   "change current context",
	Aliases: []string{"mcx"},
	Long: `
change current context.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			mmh.InteractiveSetContext()
		} else {
			_ = cmd.Help()
		}
	},
}

var ctxListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list context",
	Long: `
list context`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.ListContexts()
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
		mmh.SetContext(args[0])
	},
}

func init() {
	ctxCmd.AddCommand(ctxListCmd, ctxSetCmd)
	RootCmd.AddCommand(ctxCmd)
}
