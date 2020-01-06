package mmh

import (
	"bytes"
	"errors"
	"fmt"
	osexec "os/exec"
	"path/filepath"
	"sort"
	"strconv"
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
	for _, s := range BasicConfig.Servers {
		if s.User == "" {
			s.User = BasicConfig.Basic.User
		}
		if s.Password == "" {
			s.Password = BasicConfig.Basic.Password
			if s.PrivateKey == "" {
				s.PrivateKey = BasicConfig.Basic.PrivateKey
			}
			if s.PrivateKeyPassword == "" {
				s.PrivateKeyPassword = BasicConfig.Basic.PrivateKeyPassword
			}
		}
		if s.Port == 0 {
			s.Port = BasicConfig.Basic.Port
		}
		if s.ServerAliveInterval == 0 {
			s.ServerAliveInterval = BasicConfig.Basic.ServerAliveInterval
		}
		if s.TmuxSupport == "" {
			s.TmuxSupport = BasicConfig.Basic.TmuxSupport
		}
		if s.TmuxAutoRename == "" {
			s.TmuxAutoRename = BasicConfig.Basic.TmuxAutoRename
		}
		servers = append(servers, s)
	}

	if CurrentConfig.configPath != BasicConfig.configPath {
		for _, s := range CurrentConfig.Servers {
			if s.User == "" {
				s.User = CurrentConfig.Basic.User
			}
			if s.Password == "" {
				s.Password = CurrentConfig.Basic.Password
				if s.PrivateKey == "" {
					s.PrivateKey = CurrentConfig.Basic.PrivateKey
				}
				if s.PrivateKeyPassword == "" {
					s.PrivateKeyPassword = CurrentConfig.Basic.PrivateKeyPassword
				}
			}
			if s.Port == 0 {
				s.Port = CurrentConfig.Basic.Port
			}
			if s.ServerAliveInterval == 0 {
				s.ServerAliveInterval = CurrentConfig.Basic.ServerAliveInterval
			}
			if s.TmuxSupport == "" {
				s.TmuxSupport = BasicConfig.Basic.TmuxSupport
			}
			if s.TmuxAutoRename == "" {
				s.TmuxAutoRename = BasicConfig.Basic.TmuxAutoRename
			}
			servers = append(servers, s)
		}
	}

	return servers
}

// add server
func AddServer() {

	// name
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("name is empty")
		} else if len(line) > 12 {
			return errors.New("name too long")
		}

		if _, err := findServerByName(string(line)); err == nil {
			return errors.New("name already exist")
		}
		return nil

	}, "Name:")

	name := p.Run()

	// tags
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// allow empty
		return nil
	}, "Tags:")

	// if it is a new tag, write the configuration file
	inputTags := strings.Fields(p.Run())
	for _, tag := range inputTags {
		tagExist := false
		for _, extTag := range CurrentConfig.Tags {
			if tag == extTag {
				tagExist = true
			}
		}
		if !tagExist {
			CurrentConfig.Tags = append(CurrentConfig.Tags, tag)
		}
	}

	// ssh user
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// allow empty
		return nil

	}, "User:")

	user := p.Run()
	if strings.TrimSpace(user) == "" {
		user = CurrentConfig.Basic.User
	}

	// server address
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("server address is empty")
		}
		return nil

	}, "Address:")

	address := p.Run()

	// server port
	var port int
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			if _, err := strconv.Atoi(string(line)); err != nil {
				return err
			}
		}
		return nil

	}, "Port:")

	portStr := p.Run()
	if strings.TrimSpace(portStr) == "" {
		port = CurrentConfig.Basic.Port
	} else {
		port, _ = strconv.Atoi(portStr)
	}

	// auth method
	var password, privateKey, privateKeyPassword string
	cfg := &promptx.SelectConfig{
		ActiveTpl:    "»  {{ . | cyan }}",
		InactiveTpl:  "  {{ . | white }}",
		SelectPrompt: "Auth Method",
		SelectedTpl:  "{{ \"» \" | green }}{{\"Method:\" | cyan }} {{ . | faint }}",
		DisPlaySize:  9,
		DetailsTpl: `
--------- SSH Auth Method ----------
{{ "Method:" | faint }}	{{ . }}`,
	}

	s := &promptx.Select{
		Items: []string{
			"PrivateKey",
			"Password",
		},
		Config: cfg,
	}

	idx := s.Run()

	// use private key
	if idx == 0 {
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "PrivateKey:")

		privateKey = p.Run()
		if strings.TrimSpace(privateKey) == "" {
			privateKey = CurrentConfig.Basic.PrivateKey
		}

		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "PrivateKey Password:")
		privateKeyPassword = p.Run()
		if strings.TrimSpace(privateKeyPassword) == "" {
			privateKeyPassword = CurrentConfig.Basic.PrivateKeyPassword
		}
	} else {
		// use password
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "Password:")
		password = p.Run()
		if strings.TrimSpace(password) == "" {
			password = CurrentConfig.Basic.Password
		}
	}

	// server proxy
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			if _, err := findServerByName(string(line)); err != nil {
				return errors.New("proxy server not found")
			}
		}
		return nil

	}, "Proxy:")

	proxy := p.Run()

	// create server
	server := ServerConfig{
		Name:               name,
		Tags:               inputTags,
		User:               user,
		Address:            address,
		Port:               port,
		PrivateKey:         privateKey,
		PrivateKeyPassword: privateKeyPassword,
		Password:           password,
		Proxy:              proxy,
	}

	// Save
	CurrentConfig.Servers = append(CurrentConfig.Servers, &server)
	sort.Sort(CurrentConfig.Servers)
	checkAndExit(CurrentConfig.Write())
}

// delete server
func DeleteServer(serverNames []string) {

	var deletesIdx []int

	for _, serverName := range serverNames {
		for i, s := range CurrentConfig.Servers {
			matched, err := filepath.Match(serverName, s.Name)
			// server name may contain special characters
			if err != nil {
				// check equal
				if strings.ToLower(s.Name) == strings.ToLower(serverName) {
					deletesIdx = append(deletesIdx, i)
				}
			} else {
				if matched {
					deletesIdx = append(deletesIdx, i)
				}
			}

		}

	}

	if len(deletesIdx) == 0 {
		Exit("server not found", 1)
	}

	// sort and delete
	sort.Ints(deletesIdx)
	for i, del := range deletesIdx {
		CurrentConfig.Servers = append(CurrentConfig.Servers[:del-i], CurrentConfig.Servers[del-i+1:]...)
	}

	// save config
	sort.Sort(CurrentConfig.Servers)
	checkAndExit(CurrentConfig.Write())

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
	if s.TmuxSupport == "true" {
		winName = getTmuxWindowName()
		setTmuxWindowName(s.Name, "false")
		defer setTmuxWindowName(winName, s.TmuxAutoRename)
	}
	checkAndExit(s.Terminal())
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
		Items:  CurrentConfig.Servers,
		Config: cfg,
	}
	idx := s.Run()
	SingleLogin(CurrentConfig.Servers[idx].Name)
}

func setTmuxWindowName(name, autoRename string) {
	cmd := osexec.Command("tmux", "rename-window", name)
	checkAndExit(cmd.Run())
	if autoRename != "false" && autoRename != "off" {
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
