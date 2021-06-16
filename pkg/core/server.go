package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/mritd/mmh/pkg/common"
)

// ListServers merge basic context servers and current context servers
func ListServers(serverSort bool) Servers {
	var servers Servers
	bss := setDefaultValue(basicConfig.Servers, basicConfig.Basic)
	sort.Sort(bss)
	servers = append(servers, bss...)
	if currentConfig.configPath != basicConfig.configPath {
		css := setDefaultValue(currentConfig.Servers, currentConfig.Basic)
		if serverSort {
			sort.Sort(css)
		}
		servers = append(servers, css...)
	}
	return servers
}

// findServerByName find server from config by server name
func findServerByName(name string) (*Server, error) {
	for _, s := range ListServers(false) {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, errors.New("server not found")
}

// findServersByTag find servers from config by server tag
func findServersByTag(tag string) (Servers, error) {
	var servers Servers
	for _, s := range ListServers(false) {
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
		if s.ExtAuth == "" {
			s.ExtAuth = basic.ExtAuth
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

// PrintServers print server list
func PrintServers(serverSort bool) {
	t, _ := common.ColorFuncTemplate(listServersTpl)
	var buf bytes.Buffer
	common.CheckAndExit(t.Execute(&buf, ListServers(serverSort)))
	fmt.Println(buf.String())
}

// PrintServerDetail print single server detail
func PrintServerDetail(serverName string) {
	s, err := findServerByName(serverName)
	common.CheckAndExit(err)
	t, _ := common.ColorFuncTemplate(serverDetailTpl)
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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		if common.Tmux() {
			common.TmuxSetWindowName(tmuxWinIndex, tmuxWinName)
			common.TmuxSetAutomaticRename(tmuxWinIndex, tmuxAutoRename)
		}
		os.Exit(1)
	}()

	common.PrintErrWithPrefix("\n😱", s.Terminal())
	if common.Tmux() {
		common.TmuxSetWindowName(tmuxWinIndex, tmuxWinName)
		common.TmuxSetAutomaticRename(tmuxWinIndex, tmuxAutoRename)
	}
}
