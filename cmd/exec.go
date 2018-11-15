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
	"strings"

	"github.com/mritd/mmh/pkg/mmh"
	"github.com/spf13/cobra"
)

var singleExecServer bool
var execCmd = &cobra.Command{
	Use:     "exec SERVER_TAG CMD",
	Aliases: []string{"mec"},
	Short:   "Batch exec command",
	Long: `
Batch exec command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Help()
		} else {
			cmd := strings.Join(args[1:], " ")
			mmh.Exec(args[0], cmd, singleExecServer)
		}
	},
}

func init() {
	RootCmd.AddCommand(execCmd)
	execCmd.PersistentFlags().BoolVarP(&singleExecServer, "single", "s", false, "Single server")
}
