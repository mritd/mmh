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
	"encoding/base64"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var bannerBase64 = "ICAgICBfICAgICAgICAgICBfICAgICAgXwogICAgfCB8ICAgICAgICAgfCB8ICAgIHwgfAogX19ffCB8X18gIF8gICBffCB8XyBfX3wgfCBfX19fXyAgICAgIF9fXyBfXyAgICBfIF9fICAgX19fX18gICAgICBfXwovIF9ffCAnXyBcfCB8IHwgfCBfXy8gX2AgfC8gXyBcIFwgL1wgLyAvICdfIFwgIHwgJ18gXCAvIF8gXCBcIC9cIC8gLwpcX18gXCB8IHwgfCB8X3wgfCB8fCAoX3wgfCAoXykgXCBWICBWIC98IHwgfCB8IHwgfCB8IHwgKF8pIFwgViAgViAvCnxfX18vX3wgfF98XF9fLF98XF9fXF9fLF98XF9fXy8gXF8vXF8vIHxffCB8X3wgfF98IHxffFxfX18vIFxfL1xfLw=="

var versionTpl = `%s

Name: mmh
Version: %s
Arch: %s
BuildTime: %s
CommitID: %s
`

var (
	Version   string
	BuildTime string
	CommitID  string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Long: `
Print version.`,
	Run: func(cmd *cobra.Command, args []string) {
		banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
		fmt.Printf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildTime, CommitID)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
