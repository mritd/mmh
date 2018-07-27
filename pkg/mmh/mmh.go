package mmh

import (
	"strings"

	"fmt"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/spf13/viper"
)

func SingleLogin(name string) {
	serverExist := false
	var servers []Server
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))
	for _, s := range servers {
		// Ignore case
		if strings.ToLower(name) == strings.ToLower(s.Name) {
			serverExist = true
			s.Connect()
		}
	}

	if !serverExist {
		fmt.Println("Server not found!")
	}
}
