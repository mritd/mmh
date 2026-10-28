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
	script, err := genCompletion(cmd, shell)
	common.CheckAndExit(err)
	fmt.Print(script)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&completionShell, "completion", "", "generate shell completion")
	rootCmd.PersistentFlags().BoolVar(&core.IgnoreProxy, "no-proxy", false, "ignore socks5/http proxy settings and connect directly")
}
