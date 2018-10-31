/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
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

	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/pkg/mmh"
	"github.com/mritd/mmh/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
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
	cobra.OnInitialize(initConfig, mmh.InitTagsGroup)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mmh.yaml)")
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		utils.CheckAndExit(err)
		cfgFile = path.Join(home, ".mmh.yaml")
		viper.SetConfigFile(cfgFile)

		if _, err := os.Stat(cfgFile); err != nil {
			os.Create(cfgFile)
			viper.Set(mmh.SERVERS, mmh.ServersExample())
			viper.Set(mmh.TAGS, mmh.TagsExample())
			viper.Set("MaxProxy", 5)
			viper.WriteConfig()
		}

	}
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
