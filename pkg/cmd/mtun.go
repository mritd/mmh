package cmd

import (
	"github.com/mritd/mmh/pkg/core"
	"github.com/spf13/cobra"
)

var tunLeftAddr, tunRightAddr string
var tunReverse bool

var mtun = &cobra.Command{
	Use:     "tun SERVER",
	Short:   "ssh tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.Tunnel(args[0], tunLeftAddr, tunRightAddr, tunReverse)
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	cmds["mtun"] = mtun
	mtun.Flags().StringVarP(&tunLeftAddr, "left", "l", "", "left address")
	mtun.Flags().StringVarP(&tunRightAddr, "right", "r", "", "right address")
	mtun.PersistentFlags().BoolVar(&tunReverse, "reverse", false, "reverse tcp tunnel(right to left)")
	_ = mtun.MarkFlagRequired("left")
	_ = mtun.MarkFlagRequired("right")
}
