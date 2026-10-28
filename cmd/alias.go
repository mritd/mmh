package cmd

import (
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var (
	aliasShell  string
	aliasFile   string
	aliasBin    string
	aliasRemove bool
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Inject shell aliases for mmh subcommands",
	Long: `Inject shell aliases for mmh subcommands.

The aliases (mcp, mcs, mcx, mec, mgo, mping, mtun) are written between managed
marker lines into the config file of the current shell (detected from $SHELL),
so this command can be executed repeatedly without duplicating anything.

Use --shell to target one or more shells (comma separated), or --file to target
an arbitrary config file. Use --remove to delete the managed block again.`,
	Example: `  mmh alias                  # inject into the current shell's config file
  mmh alias --shell zsh,bash # inject into both zsh and bash config files
  mmh alias --file ~/.zshrc  # inject into the specified file
  mmh alias --remove         # remove the injected aliases`,
	Run: func(cmd *cobra.Command, args []string) {
		core.InstallAlias(aliasShell, aliasFile, aliasBin, aliasRemove)
	},
}

func init() {
	aliasCmd.PersistentFlags().StringVarP(&aliasShell, "shell", "s", "", "target shells, comma separated: zsh,bash,fish,sh (default: detect from $SHELL)")
	aliasCmd.PersistentFlags().StringVarP(&aliasFile, "file", "f", "", "target config file, overrides --shell")
	aliasCmd.PersistentFlags().StringVar(&aliasBin, "bin", "mmh", "mmh command used in the injected aliases")
	aliasCmd.PersistentFlags().BoolVarP(&aliasRemove, "remove", "r", false, "remove the injected aliases")
	rootCmd.AddCommand(aliasCmd)
}
