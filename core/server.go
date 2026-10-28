package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/olekukonko/tablewriter"

	"github.com/mritd/mmh/common"
)

// serverGroup servers loaded from one config file
type serverGroup struct {
	configPath string
	servers    Servers
}

// name returns the config file name without suffix, e.g. basic or hk-h20
func (g serverGroup) name() string {
	if g.configPath == "" {
		return "-"
	}
	return strings.TrimSuffix(filepath.Base(g.configPath), filepath.Ext(g.configPath))
}

// serverGroups returns servers grouped by config file; servers from basic.yaml
// always come first
func serverGroups(serverSort bool) []serverGroup {
	if serverSort {
		sort.Sort(basicConfig.Servers)
	}
	groups := []serverGroup{{configPath: basicConfig.configPath, servers: basicConfig.Servers}}
	if currentConfig.configPath != basicConfig.configPath {
		if serverSort {
			sort.Sort(currentConfig.Servers)
		}
		groups = append(groups, serverGroup{configPath: currentConfig.configPath, servers: currentConfig.Servers})
	}
	return groups
}

// ListServers merge basic context servers and current context servers
func ListServers(serverSort bool) Servers {
	var servers Servers
	for _, group := range serverGroups(serverSort) {
		servers = append(servers, group.servers...)
	}
	return servers
}

// PrintServers print server list
func PrintServers(serverSort bool) {
	fmt.Print(renderServers(serverGroups(serverSort)))
}

// renderServers renders server groups as a table, a border line separates the
// servers of different config files
func renderServers(groups []serverGroup) string {
	var buf bytes.Buffer

	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Name", "Group", "User", "Tags", "Address"})
	// keep one table row per server, so the separator line can be inserted at
	// a known position
	table.SetAutoWrapText(false)

	var seps []int
	rows, hasRows := 0, false
	for _, group := range groups {
		if len(group.servers) == 0 {
			continue
		}
		if hasRows {
			seps = append(seps, rows)
		}
		hasRows = true
		for _, s := range group.servers {
			table.Append([]string{s.Name, group.name(), s.User, fmt.Sprint(s.Tags), fmt.Sprintf("%s:%d", s.Address, s.Port)})
		}
		rows += len(group.servers)
	}
	table.Render()

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// line 0 is the top border, line 1 the header, line 2 the header border,
	// the table rows start at line 3
	if len(lines) < 3 {
		return buf.String()
	}
	sepLine := lines[2]
	for i := len(seps) - 1; i >= 0; i-- {
		idx := 3 + seps[i]
		if idx >= len(lines) {
			continue
		}
		lines = append(lines[:idx], append([]string{sepLine}, lines[idx:]...)...)
	}
	return strings.Join(lines, "\n") + "\n"
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
