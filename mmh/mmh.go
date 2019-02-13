package mmh

import (
	"github.com/mritd/mmh/utils"
	"github.com/mritd/promptx"
)

func SingleLogin(name string) {
	s := ContextCfg.Servers.FindServerByName(name)
	if s == nil {
		utils.Exit("server not found!", 1)
	} else {
		utils.CheckAndExit(s.Terminal())
	}
}

func InteractiveLogin() {

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
		Items:  ContextCfg.Servers,
		Config: cfg,
	}
	idx := s.Run()
	SingleLogin(ContextCfg.Servers[idx].Name)
}
