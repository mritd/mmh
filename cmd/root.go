// Copyright © 2018 mritd <mritd1234@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"os"

	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/pkg/mmh"
	"github.com/mritd/mmh/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var singleServer bool
var RootCmd = &cobra.Command{
	Use:   "mmh",
	Short: "A simple Multi-user ssh tool",
	Long: `
A simple Multi-user ssh tool.`,
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
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mmh.yaml)")
	RootCmd.PersistentFlags().BoolVarP(&singleServer, "single", "s", false, "Single server")
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		utils.CheckAndExit(err)
		cfgFile = home + string(filepath.Separator) + ".mmh.yaml"
		viper.SetConfigFile(cfgFile)

		if _, err := os.Stat(cfgFile); err != nil {
			os.Create(cfgFile)
			viper.Set(mmh.SERVERS, mmh.ServersExample())
			viper.Set(mmh.TAGS, mmh.TagsExample())
			viper.WriteConfig()
		}

	}
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
