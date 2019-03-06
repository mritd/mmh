package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var ctxCmd = &cobra.Command{
	Use:     "ctx",
	Short:   "Change current context",
	Aliases: []string{"mcx"},
	Long: `
Change current context.`,
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
	Short: "List context",
	Long: `
List context`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.ListContexts()
	},
}

var ctxUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Use context",
	Long: `
Use context`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
			return
		}
		mmh.UseContext(args[0])
	},
}

func init() {
	ctxCmd.AddCommand(ctxListCmd, ctxUseCmd)
	RootCmd.AddCommand(ctxCmd)
}
