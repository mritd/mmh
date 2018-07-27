package mmh

import (
	"strings"

	"fmt"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/mritd/promptx"
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

func InteractiveLogin() {
	var servers []Server
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))

	cfg := &promptx.SelectConfig{
		ActiveTpl:    `»  {{ .Name | cyan }}: {{ .User | cyan }}{{ "@" | cyan }}{{ .Address | cyan }}`,
		InactiveTpl:  `  {{ .Name | white }}: {{ .User | white }}{{ "@" | white }}{{ .Address | white }}`,
		SelectPrompt: "Login Server",
		SelectedTpl:  `{{ "» " | green }}{{ .Name | green }}: {{ .User | green }}{{ "@" | green }}{{ .Address | green }}`,
		DisPlaySize:  9,
		DetailsTpl: `
--------- Login Server ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "User:" | faint }} {{ .User | faint }}
{{ "Address:" | faint }} {{ .Address | faint }}{{ ":" | faint }}{{ .Port | faint }}`,
	}

	s := &promptx.Select{
		Items:  servers,
		Config: cfg,
	}
	idx := s.Run()
	SingleLogin(servers[idx].Name)
}
