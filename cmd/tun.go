package cmd

import (
	"github.com/mritd/mmh/mmh"
	"github.com/spf13/cobra"
)

var tunLocalAddr, tunRemoteAddr string

var tunCmd = &cobra.Command{
	Use:     "tun SERVER_NAME",
	Aliases: []string{"mtun"},
	Short:   "SSH tunnel",
	Long: `
SSH tunnel.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			mmh.Tunnel(args[0], tunLocalAddr, tunRemoteAddr)
		} else {
			_ = cmd.Help()
		}
	},
}

func init() {
	tunCmd.Flags().StringVarP(&tunLocalAddr, "local", "l", "", "local address")
	tunCmd.Flags().StringVarP(&tunRemoteAddr, "remote", "r", "", "local address")
	_ = tunCmd.MarkFlagRequired("local")
	_ = tunCmd.MarkFlagRequired("remote")
	RootCmd.AddCommand(tunCmd)
}
