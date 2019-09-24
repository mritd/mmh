package mmh

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/mritd/mmh/utils"
	"github.com/mritd/promptx"
)

// find server by name
func findServerByName(name string) (*ServerConfig, error) {

	for _, s := range getServers() {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, errors.New("server not found")
}

// find servers by tag
func findServersByTag(tag string) Servers {

	ss := Servers{}

	for _, s := range getServers() {
		tmpServer := s
		for _, t := range tmpServer.Tags {
			if tag == t {
				ss = append(ss, tmpServer)
			}
		}
	}
	return ss
}

// getServers merge basic context servers and current context servers
func getServers() Servers {
	var servers Servers
	for _, s := range BasicContext.Servers {
		if s.User == "" {
			s.User = BasicContext.Basic.User
		}
		if s.Password == "" {
			s.Password = BasicContext.Basic.Password
			if s.PrivateKey == "" {
				s.PrivateKey = BasicContext.Basic.PrivateKey
			}
			if s.PrivateKeyPassword == "" {
				s.PrivateKeyPassword = BasicContext.Basic.PrivateKeyPassword
			}
		}
		if s.Port == 0 {
			s.Port = BasicContext.Basic.Port
		}
		if s.ServerAliveInterval == 0 {
			s.ServerAliveInterval = BasicContext.Basic.ServerAliveInterval
		}
		servers = append(servers, s)
	}

	if CurrentContext.configPath != BasicContext.configPath {
		for _, s := range CurrentContext.Servers {
			if s.User == "" {
				s.User = CurrentContext.Basic.User
			}
			if s.Password == "" {
				s.Password = CurrentContext.Basic.Password
				if s.PrivateKey == "" {
					s.PrivateKey = CurrentContext.Basic.PrivateKey
				}
				if s.PrivateKeyPassword == "" {
					s.PrivateKeyPassword = CurrentContext.Basic.PrivateKeyPassword
				}
			}
			if s.Port == 0 {
				s.Port = CurrentContext.Basic.Port
			}
			if s.ServerAliveInterval == 0 {
				s.ServerAliveInterval = CurrentContext.Basic.ServerAliveInterval
			}
			servers = append(servers, s)
		}
	}

	return servers
}

// list servers
func ListServers() {

	tpl := `Name          User          Tags          Address
-------------------------------------------------------------
{{range . }}{{ .Name | ListLayout }}  {{ .User | ListLayout }}  {{ .Tags | MergeTag | ListLayout }}  {{ .Address }}:{{ .Port }}
{{end}}`
	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTags,
	})

	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, getServers()))
	fmt.Println(buf.String())
}

// print single server detail
func PrintServerDetail(serverName string) {
	s, err := findServerByName(serverName)
	utils.CheckAndExit(err)

	tpl := `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | MergeTag }}
Proxy: {{ .Proxy }}`
	t := template.New("").Funcs(map[string]interface{}{"MergeTag": mergeTags})
	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, s))
	fmt.Println(buf.String())
}

func SingleLogin(name string) {
	s, err := findServerByName(name)
	utils.CheckAndExit(err)
	utils.CheckAndExit(s.Terminal())
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
