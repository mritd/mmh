package core

import (
	"bytes"
	"errors"
	"fmt"
	osexec "os/exec"
	"strings"
	"text/template"

	"github.com/mritd/promptx"
)

// findServerByName find server from config by server name
func findServerByName(name string) (*ServerConfig, error) {
	for _, s := range getServers() {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, errors.New("server not found")
}

// findServersByTag find servers from config by server tag
func findServersByTag(tag string) Servers {
	var servers Servers
	for _, s := range getServers() {
		tmpServer := s
		for _, t := range tmpServer.Tags {
			if tag == t {
				servers = append(servers, tmpServer)
			}
		}
	}
	return servers
}

// getServers merge basic context servers and current context servers
func getServers() Servers {
	var servers Servers
	for _, s := range basicConfig.Servers {
		if s.User == "" {
			s.User = basicConfig.Basic.User
		}
		if s.Password == "" {
			s.Password = basicConfig.Basic.Password
			if s.PrivateKey == "" {
				s.PrivateKey = basicConfig.Basic.PrivateKey
			}
			if s.PrivateKeyPassword == "" {
				s.PrivateKeyPassword = basicConfig.Basic.PrivateKeyPassword
			}
		}
		if s.Port == 0 {
			s.Port = basicConfig.Basic.Port
		}
		if s.ServerAliveInterval == 0 {
			s.ServerAliveInterval = basicConfig.Basic.ServerAliveInterval
		}
		if !s.TmuxSupport && basicConfig.Basic.TmuxSupport {
			s.TmuxSupport = basicConfig.Basic.TmuxSupport
		}
		if !s.TmuxAutoRename && basicConfig.Basic.TmuxAutoRename {
			s.TmuxAutoRename = basicConfig.Basic.TmuxAutoRename
		}
		servers = append(servers, s)
	}

	if currentConfig.configPath != basicConfig.configPath {
		for _, s := range currentConfig.Servers {
			if s.User == "" {
				s.User = currentConfig.Basic.User
			}
			if s.Password == "" {
				s.Password = currentConfig.Basic.Password
				if s.PrivateKey == "" {
					s.PrivateKey = currentConfig.Basic.PrivateKey
				}
				if s.PrivateKeyPassword == "" {
					s.PrivateKeyPassword = currentConfig.Basic.PrivateKeyPassword
				}
			}
			if s.Port == 0 {
				s.Port = currentConfig.Basic.Port
			}
			if s.ServerAliveInterval == 0 {
				s.ServerAliveInterval = currentConfig.Basic.ServerAliveInterval
			}
			if !s.TmuxSupport && basicConfig.Basic.TmuxSupport {
				s.TmuxSupport = basicConfig.Basic.TmuxSupport
			}
			if !s.TmuxAutoRename && basicConfig.Basic.TmuxAutoRename {
				s.TmuxAutoRename = basicConfig.Basic.TmuxAutoRename
			}
			servers = append(servers, s)
		}
	}

	return servers
}

// ListServers print server list
func ListServers() {

	tpl := `Name            User            Tags            Address
-----------------------------------------------------------------
{{range . }}{{ .Name | listLayout }}  {{ .User | listLayout }}  {{ .Tags | mergeTags | listLayout }}  {{ .Address }}:{{ .Port }}
{{end}}`
	t := template.New("").Funcs(map[string]interface{}{
		"listLayout": listLayout,
		"mergeTags":  mergeTags,
	})

	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	checkAndExit(t.Execute(&buf, getServers()))
	fmt.Println(buf.String())
}

// PrintServerDetail print single server detail
func PrintServerDetail(serverName string) {
	s, err := findServerByName(serverName)
	checkAndExit(err)

	tpl := `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | MergeTag }}
Proxy: {{ .Proxy }}`
	t := template.New("").Funcs(map[string]interface{}{"MergeTag": mergeTags})
	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	checkAndExit(t.Execute(&buf, s))
	fmt.Println(buf.String())
}

func SingleLogin(name string) {
	s, err := findServerByName(name)
	checkAndExit(err)
	var winName string
	if s.TmuxSupport {
		winName = getTmuxWindowName()
		setTmuxWindowName(s.Name, false)
		defer setTmuxWindowName(winName, s.TmuxAutoRename)
	}
	printErr(s.Terminal())
}

// InteractiveLogin interactive login server
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
		Items:  currentConfig.Servers,
		Config: cfg,
	}
	SingleLogin(currentConfig.Servers[s.Run()].Name)
}

func setTmuxWindowName(name string, autoRename bool) {
	cmd := osexec.Command("tmux", "rename-window", name)
	checkAndExit(cmd.Run())
	if autoRename {
		cmd = osexec.Command("tmux", "set-window", "automatic-rename", "on")
		checkAndExit(cmd.Run())
	}
}

func getTmuxWindowName() string {
	cmd := osexec.Command("tmux", "display-message", "-p", "#W")
	bs, err := cmd.CombinedOutput()
	checkAndExit(err)
	return strings.TrimSpace(string(bs))
}
