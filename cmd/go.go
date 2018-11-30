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
	"github.com/mritd/mmh/pkg/mmh"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:     "go SERVER_NAME",
	Aliases: []string{"mgo"},
	Short:   "Login single server",
	Long: `
Login single server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			mmh.SingleLogin(args[0])
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	RootCmd.AddCommand(goCmd)
}
