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
	"os"
	"path/filepath"

	"github.com/mritd/promptx/util"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/pkg/mmh"
	"github.com/mritd/mmh/pkg/utils"
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
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		utils.Exit(err.Error(), -1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, mmh.InitConfig)
}

func initConfig() {

	home, err := homedir.Dir()
	utils.CheckAndExit(err)
	cfgDir := filepath.Join(home, ".mmh")
	mainConfigFile := filepath.Join(cfgDir, "main.yaml")
	mmh.MainViper.SetConfigFile(mainConfigFile)

	if _, err = os.Stat(cfgDir); err != nil {
		// create config dir
		utils.CheckAndExit(os.MkdirAll(cfgDir, 0755))
		// create config file
		_, err = os.Create(mainConfigFile)
		util.CheckAndExit(err)
		// create default context config file
		defaultCtxCfg := filepath.Join(cfgDir, "default.yaml")
		_, err = os.Create(defaultCtxCfg)
		util.CheckAndExit(err)
		// write default config
		writeExampleConfig(cfgDir)
	}

	// load main config
	mmh.MainViper.AutomaticEnv()
	util.CheckAndExit(mmh.MainViper.ReadInConfig())

	// get context
	currentContext := mmh.MainViper.GetString(mmh.KeyCurrentContext)
	mmh.CtxViper.SetConfigFile(filepath.Join(cfgDir, currentContext+".yaml"))

	// load context config
	mmh.CtxViper.AutomaticEnv()
	util.CheckAndExit(mmh.CtxViper.ReadInConfig())
}

func writeExampleConfig(cfgDir string) {

	// ignore this error, because it is already check
	home, _ := homedir.Dir()

	// write main config
	mmh.MainViper.Set(mmh.KeyContext, mmh.Contexts{
		{
			Name:          "default",
			IsRemote:      false,
			RemoteAddress: "",
		},
	})
	mmh.MainViper.Set(mmh.KeyCurrentContext, "default")
	utils.CheckAndExit(mmh.MainViper.WriteConfig())

	// write context config
	mmh.CtxViper.SetConfigFile(filepath.Join(cfgDir, "default.yaml"))
	mmh.CtxViper.Set(mmh.KeyBasic, mmh.Basic{
		User:               "root",
		Port:               22,
		PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
		PrivateKeyPassword: "",
		Password:           "",
		Proxy:              "",
	})
	mmh.CtxViper.Set(mmh.KeyServers, []mmh.Server{
		{
			Name:     "prod11",
			User:     "root",
			Tags:     []string{"prod"},
			Address:  "10.10.4.11",
			Port:     22,
			Password: "password",
			Proxy:    "prod12",
		},
		{
			Name:               "prod12",
			User:               "root",
			Tags:               []string{"prod"},
			Address:            "10.10.4.12",
			Port:               22,
			PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
			PrivateKeyPassword: "password",
		},
	})
	mmh.CtxViper.Set(mmh.KeyTags, []string{
		"prod",
		"test",
	})
	mmh.CtxViper.Set("MaxProxy", 5)
	utils.CheckAndExit(mmh.CtxViper.WriteConfig())
}
