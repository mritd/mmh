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
BuildDate: %s
CommitID: %s
`

var (
	Version   string
	BuildDate string
	CommitID  string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Long:  "show version.",
	Run: func(cmd *cobra.Command, args []string) {
		banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
		fmt.Printf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildDate, CommitID)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
