package core

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/olekukonko/tablewriter"

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

// PrintServers print server list
func PrintServers(serverSort bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "User", "Tags", "Address"})
	ss := ListServers(serverSort)
	for _, s := range ss {
		table.Append([]string{s.Name, s.User, fmt.Sprint(s.Tags), fmt.Sprintf("%s:%d", s.Address, s.Port)})
	}
	table.Render()
}

// PrintServerDetail print single server detail
func PrintServerDetail(serverName string) {
	s, err := findServerByName(serverName)
	common.CheckAndExit(err)

	table := tablewriter.NewWriter(os.Stdout)
	table.Append([]string{"NAME", s.Name})
	table.Append([]string{"USER", s.User})
	table.Append([]string{"ADDR", fmt.Sprintf("%s:%d", s.Address, s.Port)})
	table.Append([]string{"PROXY", s.Proxy})
	table.Append([]string{"CONFIG", s.ConfigPath})
	table.Render()
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
