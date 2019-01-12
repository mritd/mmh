/*
 * Copyright 2018 mritd <mritd1234@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mritd/promptx/util"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/mmh"
	"github.com/mritd/mmh/utils"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "mmh",
	Short: "A simple Multi-server ssh tool",
	Long: `
A simple Multi-server ssh tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.InteractiveLogin()
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
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

	// set main config file
	mainConfigFile := filepath.Join(cfgDir, "main.yaml")

	if _, err = os.Stat(cfgDir); os.IsNotExist(err) {
		// create config dir
		utils.CheckAndExit(os.MkdirAll(cfgDir, 0755))
		// create default context config file
		defaultCtxCfg := filepath.Join(cfgDir, "default.yaml")
		// write main config
		utils.CheckAndExit(mmh.MainConfigExample().SetConfigPath(mainConfigFile).Write())
		// write context config
		utils.CheckAndExit(mmh.ContextConfigExample().SetConfigPath(defaultCtxCfg).Write())
	} else if err != nil {
		utils.CheckAndExit(err)
	}

	// load main config
	utils.CheckAndExit(mmh.MainCfg.Load(mainConfigFile))

	// if timeout, context will downgrade
	if !mmh.MainCfg.Contexts.TimeStamp.IsZero() && mmh.MainCfg.Contexts.TimeOut != 0 && mmh.MainCfg.Contexts.AutoDowngrade != "" {
		if time.Now().After(mmh.MainCfg.Contexts.TimeStamp.Add(mmh.MainCfg.Contexts.TimeOut)) && mmh.MainCfg.Contexts.Current != mmh.MainCfg.Contexts.AutoDowngrade {
			fmt.Printf("🍁 context timeout, auto downgrade => [%s]\n", mmh.MainCfg.Contexts.AutoDowngrade)
			mmh.MainCfg.Contexts.Current = mmh.MainCfg.Contexts.AutoDowngrade
			util.CheckAndExit(mmh.MainCfg.Write())
		}
	}

	// check context
	if len(mmh.MainCfg.Contexts.Context) == 0 {
		utils.Exit("get context failed", 1)
	}

	// get current use context
	ctx, ok := mmh.MainCfg.Contexts.FindContextByName(mmh.MainCfg.Contexts.Current)
	if !ok {
		utils.Exit(fmt.Sprintf("could not found current context: %s\n", mmh.MainCfg.Contexts.Current), 1)
	}

	var ctxConfigFile string
	if filepath.IsAbs(ctx.ConfigPath) {
		ctxConfigFile = ctx.ConfigPath
	} else {
		ctxConfigFile = filepath.Join(cfgDir, ctx.ConfigPath)
	}
	if _, err = os.Stat(ctxConfigFile); os.IsNotExist(err) {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s not found\n", mmh.MainCfg.Contexts.Current, ctx.ConfigPath), 1)
	} else if err != nil {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s load failed: %s\n", mmh.MainCfg.Contexts.Current, ctx.ConfigPath, err.Error()), 1)
	}

	// load current context
	utils.CheckAndExit(mmh.ContextCfg.Load(ctxConfigFile))
}
