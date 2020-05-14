package cmd

import (
	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var tunLeftAddr, tunRightAddr string
var tunReverse bool

var tunCmd = &cobra.Command{
	Use:     "tun SERVER_NAME",
	Aliases: []string{"mtun"},
	Short:   "ssh tunnel",
	Long:    "ssh tunnel.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			core.Tunnel(args[0], tunLeftAddr, tunRightAddr, tunReverse)
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	tunCmd.Flags().StringVarP(&tunLeftAddr, "left", "l", "", "left address")
	tunCmd.Flags().StringVarP(&tunRightAddr, "right", "r", "", "right address")
	tunCmd.PersistentFlags().BoolVar(&tunReverse, "reverse", false, "reverse tcp tunnel(right to left)")
	_ = tunCmd.MarkFlagRequired("left")
	_ = tunCmd.MarkFlagRequired("right")
	rootCmd.AddCommand(tunCmd)
}
