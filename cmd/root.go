package cmd

import (
	"fmt"
	"os"

	"github.com/mritd/mmh/common"
	"github.com/mritd/mmh/core"

	"github.com/spf13/cobra"
)

var completionShell string

var BuildCmd string
var cmds = make(map[string]*cobra.Command, 10)

func Execute() {
	core.LoadConfig()

	targetCmd, ok := cmds[BuildCmd]
	if !ok {
		common.Exit(fmt.Sprintf("target cmd [%s] not found\n", BuildCmd), 1)
	}

	if err := targetCmd.Execute(); err != nil {
		common.Exit(err.Error(), -1)
	}
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
