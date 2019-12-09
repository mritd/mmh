package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var cfCmd = &cobra.Command{
	Use:     "cf",
	Aliases: []string{"mcf", "mcx"},
	Short:   "change current context",
	Long:    "change current context.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			mmh.InteractiveSetConfig()
		} else {
			_ = cmd.Help()
		}
	},
}

var cfListCmd = &cobra.Command{
	Use:   "ls",
	Short: "list context",
	Long:  "list context",
	Run: func(cmd *cobra.Command, args []string) {
		mmh.ListConfig()
	},
}

var cfSetCmd = &cobra.Command{
	Use:   "set",
	Short: "set context",
	Long:  "set context",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
			return
		}
		mmh.SetConfig(args[0])
	},
}

func init() {
	cfCmd.AddCommand(cfListCmd, cfSetCmd)
	rootCmd.AddCommand(cfCmd)
}
