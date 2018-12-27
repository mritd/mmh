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

package main

import (
	"os"
	"path/filepath"

	"github.com/mritd/mmh/cmd"
	"github.com/mritd/mmh/utils"
	"github.com/spf13/cobra"
)

func commandFor(basename string, rootCommand *cobra.Command) *cobra.Command {

	c, _, _ := rootCommand.Find([]string{basename})
	if c != nil {
		rootCommand.RemoveCommand(c)
		return c
	}
	return rootCommand
}

func main() {
	basename := filepath.Base(os.Args[0])
	utils.CheckAndExit(commandFor(basename, cmd.RootCmd).Execute())
}
