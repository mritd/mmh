package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/mritd/mmh/common"
)

// ListServers merge basic context servers and current context servers
func ListServers(serverSort bool) Servers {
	var servers Servers
	if serverSort {
		sort.Sort(basicConfig.Servers)
	}
	servers = append(servers, basicConfig.Servers...)
	if currentConfig.configPath != basicConfig.configPath {
		if serverSort {
			sort.Sort(currentConfig.Servers)
		}
		servers = append(servers, currentConfig.Servers...)
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
