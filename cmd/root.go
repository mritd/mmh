package cmd

import (
	"os"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "mmh",
	Short:            "a simple multi-server ssh tool",
	Long:             "a simple multi-server ssh tool.",
	TraverseChildren: true,
	Run:              func(cmd *cobra.Command, args []string) { core.InteractiveLogin() },
}

func Execute() {
	core.Aliases = findAllAliases(rootCmd)
	core.LoadConfig()

	subCmd, _, err := rootCmd.Find([]string{filepath.Base(os.Args[0])})
	if err == nil && subCmd.Name() != rootCmd.Name() {
		if len(os.Args) > 1 {
			rootCmd.SetArgs(append([]string{subCmd.Name()}, os.Args[1:]...))
		} else {
			rootCmd.SetArgs([]string{subCmd.Name()})
		}
	}

	if err := rootCmd.Execute(); err != nil {
		core.Exit(err.Error(), -1)
	}
}

func findAllAliases(cmd *cobra.Command) []string {
	var aliases []string
	if cmd.HasSubCommands() {
		cmds := cmd.Commands()
		for _, c := range cmds {
			if len(c.Aliases) > 0 {
				aliases = append(aliases, c.Aliases...)
			}
			if c.HasSubCommands() {
				as := findAllAliases(c)
				if len(as) > 0 {
					aliases = append(aliases, as...)
				}
			}
		}
	} else {
		if len(cmd.Aliases) > 0 {
			aliases = append(aliases, cmd.Aliases...)
		}
	}

	return aliases
}
