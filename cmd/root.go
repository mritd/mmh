package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/mmh"
	"github.com/mritd/mmh/utils"
	"github.com/spf13/cobra"
)

const MainConfigPath = "MMH_MAIN_CONFIG_PATH"

var RootCmd = &cobra.Command{
	Use:   "mmh",
	Short: "a simple multi-server ssh tool",
	Long: `
a simple multi-server ssh tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.InteractiveLogin()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		utils.Exit(err.Error(), -1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initial configuration
func initConfig() {

	// get user home dir
	home, err := homedir.Dir()
	utils.CheckAndExit(err)

	// get config dir
	cfgDir := filepath.Join(home, ".mmh")

	// load main config from env
	mainConfigPath := os.Getenv(MainConfigPath)
	if mainConfigPath == "" {
		// default to $HOME/.mmh/main.yaml
		mainConfigPath = filepath.Join(cfgDir, "main.yaml")
		if _, err = os.Stat(cfgDir); os.IsNotExist(err) {
			// create config dir
			utils.CheckAndExit(os.MkdirAll(cfgDir, 0755))
			// create default context config file
			defaultCtxCfg := filepath.Join(cfgDir, "default.yaml")
			// write example main config
			utils.CheckAndExit(mmh.MainConfigExample().WriteTo(mainConfigPath))
			// write example context config
			utils.CheckAndExit(mmh.ContextConfigExample().WriteTo(defaultCtxCfg))
		} else if err != nil {
			utils.CheckAndExit(err)
		}
	}

	// load main config
	utils.CheckAndExit(mmh.Main.LoadFrom(mainConfigPath))

	// check context
	if len(mmh.Main.Contexts) == 0 {
		utils.Exit("get context failed", 1)
	}

	// get current use context
	ctx, ok := mmh.FindContextByName(mmh.Main.Current)
	if !ok {
		utils.Exit(fmt.Sprintf("could not found current context: %s\n", mmh.Main.Current), 1)
	}

	var ctxConfigFile string
	if filepath.IsAbs(ctx.ConfigPath) {
		ctxConfigFile = ctx.ConfigPath
	} else {
		ctxConfigFile = filepath.Join(cfgDir, ctx.ConfigPath)
	}
	if _, err = os.Stat(ctxConfigFile); os.IsNotExist(err) {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s not found\n", mmh.Main.Current, ctx.ConfigPath), 1)
	} else if err != nil {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s load failed: %s\n", mmh.Main.Current, ctx.ConfigPath, err.Error()), 1)
	}

	// load basic context
	utils.CheckAndExit(mmh.CurrentContext.LoadFrom(ctxConfigFile))

	// get current use context
	basicCtx, ok := mmh.FindContextByName(mmh.Main.Basic)
	if ok {
		if filepath.IsAbs(basicCtx.ConfigPath) {
			ctxConfigFile = basicCtx.ConfigPath
		} else {
			ctxConfigFile = filepath.Join(cfgDir, basicCtx.ConfigPath)
		}
		if _, err = os.Stat(ctxConfigFile); os.IsNotExist(err) {
			utils.Exit(fmt.Sprintf("basic context [%s] config file %s not found\n", mmh.Main.Current, ctx.ConfigPath), 1)
		} else if err != nil {
			utils.Exit(fmt.Sprintf("basic context [%s] config file %s load failed: %s\n", mmh.Main.Current, ctx.ConfigPath, err.Error()), 1)
		}

		// load current context
		utils.CheckAndExit(mmh.BasicContext.LoadFrom(ctxConfigFile))
	}

}
