package core

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/mritd/mmh/pkg/common"
	"github.com/mritd/promptx"
)

// findServerByName find server from config by server name
func findServerByName(name string) (*Server, error) {
	for _, s := range getServers() {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, errors.New("server not found")
}

// findServersByTag find servers from config by server tag
func findServersByTag(tag string) (Servers, error) {
	var servers Servers
	for _, s := range getServers() {
		tmpServer := s
		for _, t := range tmpServer.Tags {
			if tag == t {
				servers = append(servers, tmpServer)
			}
		}
	}
	if len(servers) == 0 {
		return nil, errors.New("server not found")
	}
	return servers, nil
}

// getServers merge basic context servers and current context servers
func getServers() Servers {
	var servers Servers
	bss := setDefaultValue(basicConfig.Servers, basicConfig.Basic)
	sort.Sort(bss)
	servers = append(servers, bss...)
	if currentConfig.configPath != basicConfig.configPath {
		css := setDefaultValue(currentConfig.Servers, currentConfig.Basic)
		sort.Sort(css)
		servers = append(servers, css...)
	}
	return servers
}

// setDefaultValue set the default config value to the given servers
func setDefaultValue(servers Servers, basic BasicServerConfig) Servers {
	var ss Servers
	for _, s := range servers {
		if s.User == "" {
			s.User = basic.User
			if s.User == "" {
				s.User = "root"
			}
		}
		if s.Password == "" {
			s.Password = basic.Password
		}
		if s.PrivateKey == "" {
			s.PrivateKey = basic.PrivateKey
		}
		if s.PrivateKeyPassword == "" {
			s.PrivateKeyPassword = basic.PrivateKeyPassword
		}
		if s.KeyboardAuthCmd == "" {
			s.KeyboardAuthCmd = basic.KeyboardAuthCmd
		}
		if s.EnableAPI == "" {
			s.EnableAPI = basic.EnableAPI
		}
		if s.Environment == nil {
			s.Environment = basic.Environment
			if s.Environment == nil {
				s.Environment = make(map[string]string)
			}
		}
		if s.Port == 0 {
			s.Port = basic.Port
			if s.Port == 0 {
				s.Port = 22
			}
		}
		if s.ServerAliveInterval == 0 {
			s.ServerAliveInterval = basic.ServerAliveInterval
			if s.ServerAliveInterval == 0 {
				s.ServerAliveInterval = 10 * time.Second
			}
		}
		ss = append(ss, s)
	}
	return ss
}

// ListServers print server list
func ListServers() {
	t, _ := common.Template(listServersTpl)
	var buf bytes.Buffer
	common.CheckAndExit(t.Execute(&buf, getServers()))
	fmt.Println(buf.String())
}

// ServerDetail print single server detail
func ServerDetail(serverName string) {
	s, err := findServerByName(serverName)
	common.CheckAndExit(err)
	t, _ := common.Template(serverDetailTpl)
	var buf bytes.Buffer
	common.CheckAndExit(t.Execute(&buf, s))
	fmt.Println(buf.String())
}

// SingleLogin open a single server interactive terminal
// If running in the tmux environment, mmh will automatically update the tmux window name
func SingleLogin(name string) {
	s, err := findServerByName(name)
	common.CheckAndExit(err)
	var tmuxWinIndex, tmuxWinName string
	var tmuxAutoRename bool
	if common.Tmux() {
		tmuxWinIndex, tmuxWinName = common.TmuxWindowInfo()
		tmuxAutoRename = common.TmuxAutomaticRename()
		common.TmuxSetWindowName(tmuxWinIndex, s.Name)
		common.TmuxSetAutomaticRename(tmuxWinIndex, false)
	}

	common.PrintErrWithPrefix("\n😱", s.Terminal())
	if common.Tmux() {
		common.TmuxSetWindowName(tmuxWinIndex, tmuxWinName)
		common.TmuxSetAutomaticRename(tmuxWinIndex, tmuxAutoRename)
	}
}

// SingleInteractiveLogin display a list of interactive servers and then call SingleLogin
func SingleInteractiveLogin() {
	cfg := &promptx.SelectConfig{
		SelectPrompt: "Login Server",
		SelectedTpl:  interactiveLoginSelectedTpl,
		ActiveTpl:    interactiveLoginActiveTpl,
		InactiveTpl:  interactiveLoginInactiveTpl,
		DetailsTpl:   interactiveLoginDetailsTpl,
		DisPlaySize:  9,
	}

	ss := getServers()
	s := &promptx.Select{
		Items:  ss,
		Config: cfg,
	}
	SingleLogin(ss[s.Run()].Name)
}
