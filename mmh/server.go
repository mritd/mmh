package mmh

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/mritd/mmh/utils"
	"github.com/mritd/promptx"
)

// find server by name
func findServerByName(name string) *ServerConfig {

	for _, s := range getServers() {
		if s.Name == name {
			return s
		}
	}
	return nil
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

// add server
func AddServer() {

	// name
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return inputEmptyErr
		} else if len(line) > 12 {
			return inputTooLongErr
		}

		if s := findServerByName(string(line)); s != nil {
			return serverExistErr
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
		for _, extTag := range CurrentContext.Tags {
			if tag == extTag {
				tagExist = true
			}
		}
		if !tagExist {
			CurrentContext.Tags = append(CurrentContext.Tags, tag)
		}
	}

	// ssh user
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// allow empty
		return nil

	}, "User:")

	user := p.Run()
	if strings.TrimSpace(user) == "" {
		user = CurrentContext.Basic.User
	}

	// server address
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return inputEmptyErr
		}
		return nil

	}, "Address:")

	address := p.Run()

	// server port
	var port int
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			if _, err := strconv.Atoi(string(line)); err != nil {
				return notNumberErr
			}
		}
		return nil

	}, "Port:")

	portStr := p.Run()
	if strings.TrimSpace(portStr) == "" {
		port = CurrentContext.Basic.Port
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
			privateKey = CurrentContext.Basic.PrivateKey
		}

		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "PrivateKey Password:")
		privateKeyPassword = p.Run()
		if strings.TrimSpace(privateKeyPassword) == "" {
			privateKeyPassword = CurrentContext.Basic.PrivateKeyPassword
		}
	} else {
		// use password
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "Password:")
		password = p.Run()
		if strings.TrimSpace(password) == "" {
			password = CurrentContext.Basic.Password
		}
	}

	// server proxy
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			if findServerByName(string(line)) == nil {
				return proxyNotFoundErr
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
	CurrentContext.Servers = append(CurrentContext.Servers, &server)
	sort.Sort(CurrentContext.Servers)
	utils.CheckAndExit(CurrentContext.Write())
}

// delete server
func DeleteServer(serverNames []string) {

	var deletesIdx []int

	for _, serverName := range serverNames {
		for i, s := range CurrentContext.Servers {
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
		utils.Exit("server not found!", 1)
	}

	// sort and delete
	sort.Ints(deletesIdx)
	for i, del := range deletesIdx {
		CurrentContext.Servers = append(CurrentContext.Servers[:del-i], CurrentContext.Servers[del-i+1:]...)
	}

	// save config
	sort.Sort(CurrentContext.Servers)
	utils.CheckAndExit(CurrentContext.Write())

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
	s := findServerByName(serverName)
	if s == nil {
		utils.Exit("server not found!", 1)
	}
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
