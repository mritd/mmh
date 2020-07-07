package core

import (
	"bytes"
	"errors"
	"fmt"
	osexec "os/exec"
	"sort"
	"strings"

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
	bss := setDefaultValue(basicConfig.Servers)
	sort.Sort(bss)
	servers = append(servers, bss...)
	if currentConfig.configPath != basicConfig.configPath {
		css := setDefaultValue(currentConfig.Servers)
		sort.Sort(css)
		servers = append(servers, css...)
	}
	return servers
}

func setDefaultValue(servers Servers) Servers {
	var ss Servers
	for _, s := range servers {
		if s.User == "" {
			s.User = basicConfig.Basic.User
			if s.User == "" {
				s.User = "root"
			}
		}
		if s.Password == "" {
			s.Password = basicConfig.Basic.Password
		}
		if s.PrivateKey == "" {
			s.PrivateKey = basicConfig.Basic.PrivateKey
		}
		if s.PrivateKeyPassword == "" {
			s.PrivateKeyPassword = basicConfig.Basic.PrivateKeyPassword
		}
		if s.Port == 0 {
			s.Port = basicConfig.Basic.Port
			if s.Port == 0 {
				s.Port = 22
			}
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

func SingleLogin(name string) {
	s, err := findServerByName(name)
	common.CheckAndExit(err)
	var winName string
	if s.TmuxSupport {
		winName = getTmuxWindowName()
		setTmuxWindowName(s.Name, false)
		defer setTmuxWindowName(winName, s.TmuxAutoRename)
	}
	common.PrintErr(s.Terminal())
}

func SingleInteractiveLogin() {
	cfg := &promptx.SelectConfig{
		SelectPrompt: "Login Server",
		SelectedTpl:  interactiveLoginSelectedTpl,
		ActiveTpl:    interactiveLoginActiveTpl,
		InactiveTpl:  interactiveLoginInactiveTpl,
		DetailsTpl:   interactiveLoginDetailsTpl,
		DisPlaySize:  9,
	}

	s := &promptx.Select{
		Items:  getServers(),
		Config: cfg,
	}
	SingleLogin(currentConfig.Servers[s.Run()].Name)
}

func setTmuxWindowName(name string, autoRename bool) {
	cmd := osexec.Command("tmux", "rename-window", name)
	common.CheckAndExit(cmd.Run())
	if autoRename {
		cmd = osexec.Command("tmux", "set-window", "automatic-rename", "on")
		common.CheckAndExit(cmd.Run())
	}
}

func getTmuxWindowName() string {
	cmd := osexec.Command("tmux", "display-message", "-p", "#W")
	bs, err := cmd.CombinedOutput()
	common.CheckAndExit(err)
	return strings.TrimSpace(string(bs))
}
