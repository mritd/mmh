package mmh

import (
	"github.com/mritd/mmh/utils"
	"github.com/mritd/promptx"
)

func SingleLogin(name string) {
	s := findServerByName(name)
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
		Items:  CurrentContext.Servers,
		Config: cfg,
	}
	idx := s.Run()
	SingleLogin(CurrentContext.Servers[idx].Name)
}

func InteractiveSetContext() {
	cfg := &promptx.SelectConfig{
		ActiveTpl:    `»  {{ .Name | cyan }}`,
		InactiveTpl:  `  {{ .Name | white }}`,
		SelectPrompt: "Context",
		SelectedTpl:  `{{ "» " | green }}{{ .Name | green }}`,
		DisPlaySize:  9,
		DetailsTpl: `
--------- Context ----------
{{ "Name:" | faint }} {{ .Name | faint }}
{{ "ConfigPath:" | faint }} {{ .ConfigPath | faint }}`,
	}

	s := &promptx.Select{
		Items:  Main.Contexts,
		Config: cfg,
	}
	idx := s.Run()
	UseContext(Main.Contexts[idx].Name)
}
