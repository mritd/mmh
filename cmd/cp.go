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

var singleCPServer bool
var cpCmd = &cobra.Command{
	Use:     "cp FILE/DIR|SERVER_TAG:PATH SERVER_NAME:PATH|FILE/DIR",
	Aliases: []string{"mcp"},
	Short:   "Copies files between hosts on a network",
	Long: `
Copies files between hosts on a network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Help()
		} else {
			mmh.Copy(args[0], args[1], singleCPServer)
		}
	},
}

func init() {
	RootCmd.AddCommand(cpCmd)
	cpCmd.PersistentFlags().BoolVarP(&singleCPServer, "single", "s", false, "Single server")
}
