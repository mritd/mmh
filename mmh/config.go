package mmh

import (
	"errors"
	"strconv"
	"strings"

	"path/filepath"

	"fmt"

	"bytes"
	"text/template"

	"sort"

	"github.com/mritd/mmh/utils"
	"github.com/mritd/promptx"
)

var (
	Main           MainConfig
	BasicContext   ContextConfig
	CurrentContext ContextConfig
	MaxProxy       int
)

// error def
var (
	inputEmptyErr    = errors.New("input is empty")
	inputTooLongErr  = errors.New("input length must be <= 12")
	serverExistErr   = errors.New("server name exist")
	notNumberErr     = errors.New("only number support")
	proxyNotFoundErr = errors.New("proxy server not found")
)

// find context by name
func FindContextByName(name string) (Context, bool) {
	for _, ctx := range Main.Contexts {
		if name == ctx.Name {
			return ctx, true
		}
	}
	return Context{}, false
}

// find server by name
func findServerByName(name string) *ServerConfig {

	for _, s := range getServers() {
		if s.Name == name {
			return s
		}
	}
	return nil
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

// list contexts
func ListContexts() {

	tpl := `  Name          Path
---------------------------------
{{ range . }}{{ if .IsContext }}» {{ .Name | ListLayout }}{{ else }}  {{ .Name | ListLayout }}{{ end }}  {{ .ConfigPath }}
{{ end }}`

	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTags,
	})
	_, _ = t.Parse(tpl)

	var ctxList []struct {
		Context
		IsContext bool
	}

	sort.Sort(Main.Contexts)
	for _, c := range Main.Contexts {
		ctxList = append(ctxList, struct {
			Context
			IsContext bool
		}{
			Context:   c,
			IsContext: c.Name == Main.Current})
	}

	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, ctxList))
	fmt.Println(buf.String())
}

// set current context
func UseContext(ctxName string) {
	_, ok := FindContextByName(ctxName)
	if !ok {
		utils.Exit(fmt.Sprintf("context [%s] not found", ctxName), 1)
	}
	Main.Current = ctxName
	utils.CheckAndExit(Main.Write())
}

// print layout func
func listLayout(name string) string {
	if len(name) < 12 {
		return fmt.Sprintf("%-12s", name)
	} else {
		return fmt.Sprintf("%-12s", utils.ShortenString(name, 12))
	}
}

// merge tags
func mergeTags(tags []string) string {
	return strings.Join(tags, ",")
}
