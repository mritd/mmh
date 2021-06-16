package cmd

import (
	"github.com/mritd/mmh/common"
	"github.com/mritd/mmh/core"
	"github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var completionShell string

var BuildCmd string
var cmds = make(map[string]*cobra.Command, 10)

func Execute() {
	core.LoadConfig()

	targetCmd, ok := cmds[BuildCmd]
	if !ok {
		logrus.Fatalf("target cmd [%s] not found", BuildCmd)
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