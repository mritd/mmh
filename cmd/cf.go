package cmd

import (
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:     "cf",
	Aliases: []string{"mcf", "mcx"},
	Short:   "change current context",
	Long:    "change current context.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			core.InteractiveSetConfig()
		} else {
			_ = cmd.Help()
		}
	},
}

var cfListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list context",
	Long:  "list context",
	Run:   func(cmd *cobra.Command, args []string) { core.ListConfig() },
}

var cfSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set context",
	Long:  "set context",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.SetConfig(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	cfCmd.AddCommand(cfListCmd, cfSetCmd)
	rootCmd.AddCommand(cfCmd)
}
