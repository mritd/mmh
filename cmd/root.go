package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mritd/mmh/common"

	"github.com/mritd/mmh/core"

	"github.com/spf13/cobra"
)

var completionShell string
var showVersion bool

var rootCmd = &cobra.Command{
	Use:   "mmh",
	Short: "Modular ssh toolkit",
	Run: func(cmd *cobra.Command, args []string) {
		if completionShell != "" {
			GenCompletion(cmd, completionShell)
			return
		}
		if showVersion {
			banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
			fmt.Printf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildDate, CommitID)
			return
		}
		_ = cmd.Help()
	},
}

func Execute() {
	core.LoadConfig()

	// check bin name
	if os.Args[0] != rootCmd.Name() {
		// try to find the subcommand with the same name as bin and execute it
		subCmd, _, err := rootCmd.Find([]string{filepath.Base(os.Args[0])})
		if err == nil && subCmd.Name() != rootCmd.Name() {
			// if find a subcommand, we need to remove the subcommand from the parent command
			// to ensure that the '__complete' command takes effect
			rootCmd.RemoveCommand(subCmd)
			// re set args for the subcommand
			if len(os.Args) > 1 {
				subCmd.SetArgs(os.Args[1:])
			}
			// execute subcommand
			common.CheckAndExit(subCmd.Execute())
			return
		}
	}
	common.CheckAndExit(rootCmd.Execute())
}

func GenCompletion(cmd *cobra.Command, shell string) {
	switch shell {
	case "bash":
		_ = cmd.GenBashCompletion(os.Stdout)
	case "zsh":
		_ = cmd.GenZshCompletion(os.Stdout)
	case "fish":
		_ = cmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		_ = cmd.GenPowerShellCompletionWithDesc(os.Stdout)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
}
