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
	mmh.MainViper.SetConfigFile(mainConfigFile)

	if _, err = os.Stat(cfgDir); os.IsNotExist(err) {
		// create config dir
		utils.CheckAndExit(os.MkdirAll(cfgDir, 0755))
		// create main config file
		_, err = os.Create(mainConfigFile)
		util.CheckAndExit(err)
		// create default context config file
		defaultCtxCfg := filepath.Join(cfgDir, "default.yaml")
		_, err = os.Create(defaultCtxCfg)
		util.CheckAndExit(err)
		// write default config
		writeExampleConfig(cfgDir)
	} else if err != nil {
		utils.CheckAndExit(err)
	}

	// load main config
	mmh.MainViper.AutomaticEnv()
	utils.CheckAndExit(mmh.MainViper.ReadInConfig())
	utils.CheckAndExit(mmh.MainViper.UnmarshalKey(mmh.KeyContexts, &mmh.ContextsCfg))

	// if timeout, context will downgrade
	if !mmh.ContextsCfg.TimeStamp.IsZero() && mmh.ContextsCfg.TimeOut != 0 && mmh.ContextsCfg.AutoDowngrade != "" {
		if time.Now().After(mmh.ContextsCfg.TimeStamp.Add(mmh.ContextsCfg.TimeOut)) && mmh.ContextsCfg.Current != mmh.ContextsCfg.AutoDowngrade {
			fmt.Printf("🍁 context timeout, auto downgrade => [%s]\n", mmh.ContextsCfg.AutoDowngrade)
			mmh.ContextsCfg.Current = mmh.ContextsCfg.AutoDowngrade
			mmh.MainViper.Set(mmh.KeyContexts, mmh.ContextsCfg)
			utils.CheckAndExit(mmh.MainViper.WriteConfig())
		}
	}

	// check context
	if len(mmh.ContextsCfg.Context) == 0 {
		utils.Exit("get context failed", 1)
	}

	// get current use context
	ctx, ok := mmh.ContextsCfg.FindContextByName(mmh.ContextsCfg.Current)
	if !ok {
		utils.Exit(fmt.Sprintf("could not found current context: %s\n", mmh.ContextsCfg.Current), 1)
	}

	var ctxConfig string
	if filepath.IsAbs(ctx.ConfigPath) {
		ctxConfig = ctx.ConfigPath
	} else {
		ctxConfig = filepath.Join(cfgDir, ctx.ConfigPath)
	}
	if _, err = os.Stat(ctxConfig); os.IsNotExist(err) {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s not found\n", mmh.ContextsCfg.Current, ctx.ConfigPath), 1)
	} else if err != nil {
		utils.Exit(fmt.Sprintf("current context [%s] config file %s load failed: %s\n", mmh.ContextsCfg.Current, ctx.ConfigPath, err.Error()), 1)
	}

	mmh.CtxViper.SetConfigFile(ctxConfig)

	// load current context
	mmh.CtxViper.AutomaticEnv()
	utils.CheckAndExit(mmh.CtxViper.ReadInConfig())
	utils.CheckAndExit(mmh.CtxViper.UnmarshalKey(mmh.KeyBasic, &mmh.BasicCfg))
	utils.CheckAndExit(mmh.CtxViper.UnmarshalKey(mmh.KeyServers, &mmh.ServersCfg))
	utils.CheckAndExit(mmh.CtxViper.UnmarshalKey(mmh.KeyTags, &mmh.TagsCfg))
	utils.CheckAndExit(mmh.CtxViper.UnmarshalKey(mmh.KeyMaxProxy, &mmh.MaxProxy))
}

// write example config to config file
func writeExampleConfig(cfgDir string) {

	// write main example config
	mmh.MainViper.Set(mmh.KeyContexts, mmh.ExampleContexts())
	utils.CheckAndExit(mmh.MainViper.WriteConfig())

	// write server example config
	mmh.CtxViper.SetConfigFile(filepath.Join(cfgDir, "default.yaml"))
	mmh.CtxViper.Set(mmh.KeyBasic, mmh.ExampleBasic())
	mmh.CtxViper.Set(mmh.KeyServers, mmh.ExampleServers())
	mmh.CtxViper.Set(mmh.KeyTags, mmh.ExampleTags())
	mmh.CtxViper.Set(mmh.KeyMaxProxy, 5)
	utils.CheckAndExit(mmh.CtxViper.WriteConfig())
}
