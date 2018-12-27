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
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Server command",
	Aliases: []string{"mms"},
	Long: `
Server command.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
}

var serverAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add ssh server",
	Long: `
Add ssh server.`,
	Run: func(cmd *cobra.Command, args []string) {
		mmh.AddServer()
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
}

var serverDelCmd = &cobra.Command{
	Use:   "del SERVER1 SERVER2...",
	Short: "Delete ssh server",
	Long: `
Delete ssh server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			mmh.DeleteServer(args)
		} else {
			_ = cmd.Help()
		}
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
}

var serverListCmd = &cobra.Command{
	Use:   "ls [SERVER_NAME]",
	Short: "List ssh server",
	Long: `
List ssh server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			mmh.ServerDetail(args[0])
		} else {
			mmh.ListServers()
		}
	},
	PreRun:  mmh.UpdateContextTimestampTask,
	PostRun: mmh.UpdateContextTimestamp,
}

func init() {
	serverCmd.AddCommand(serverAddCmd, serverDelCmd, serverListCmd)
	RootCmd.AddCommand(serverCmd)
}
