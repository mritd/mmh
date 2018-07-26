package mmh

import (
	"strings"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/spf13/viper"
)

func SingleLogin(name string) {
	var servers []Server
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))
	for _, s := range servers {
		// Ignore case
		if strings.ToLower(name) == strings.ToLower(s.Name) {
			s.Connect()
		}
	}
}

func Run(args []string) {
	if len(args) == 1 {
		SingleLogin(args[0])
	}
}
