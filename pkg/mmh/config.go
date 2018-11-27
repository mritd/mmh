/*
 * Copyright 2018 mritd <mritd1234@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mmh

import (
	"errors"
	"strings"
	"sync"

	"strconv"

	"path/filepath"

	"fmt"

	"bytes"
	"text/template"

	"sort"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/pkg/utils"
	"github.com/mritd/promptx"
	"github.com/spf13/viper"
)

const keyServers = "servers"
const keyBasic = "basic"
const keyTags = "tags"
const keyContext = "context"
const keyCurrentContext = "current_context"

var (
	initOnce   sync.Once
	allTags    []string
	basic      Basic
	servers    Servers
	serversMap = make(map[string]*Server)
	tagsMap    = make(map[string][]*Server)
	maxProxy   = viper.GetInt("maxProxy")
)

var (
	inputEmptyErr    = errors.New("input is empty")
	inputTooLongErr  = errors.New("input length must be <= 12")
	serverExistErr   = errors.New("server name exist")
	notNumberErr     = errors.New("only number support")
	proxyNotFoundErr = errors.New("proxy server not found")
)

func WriteExampleConfig() {

	home, err := homedir.Dir()
	utils.CheckAndExit(err)

	// create main config
	viper.Set(keyContext, []string{"default"})
	viper.Set(keyCurrentContext, "default")
	utils.CheckAndExit(viper.WriteConfig())

	viper.SetConfigFile("default.yaml")
	viper.Set(keyBasic, Basic{
		User:               "root",
		Port:               22,
		PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
		PrivateKeyPassword: "",
		Password:           "",
		Proxy:              "",
	})
	viper.Set(keyServers, []Server{
		{
			Name:     "prod11",
			User:     "root",
			Tags:     []string{"prod"},
			Address:  "10.10.4.11",
			Port:     22,
			Password: "password",
			Proxy:    "prod12",
		},
		{
			Name:               "prod12",
			User:               "root",
			Tags:               []string{"prod"},
			Address:            "10.10.4.12",
			Port:               22,
			PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
			PrivateKeyPassword: "password",
		},
	})
	viper.Set(keyTags, []string{
		"prod",
		"test",
	})
	viper.Set("MaxProxy", 5)
	viper.WriteConfig()
}

func InitConfig() {

	initOnce.Do(
		func() {

			// set default max proxy
			if maxProxy == 0 {
				maxProxy = 5
			}

			// get home dir
			home, err := homedir.Dir()
			utils.CheckAndExit(err)

			// set default basic config
			viper.SetDefault(keyBasic, Basic{
				User:               "root",
				Port:               22,
				PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
				PrivateKeyPassword: "",
				Password:           "",
				Proxy:              "",
			})

			// init basic config
			utils.CheckAndExit(viper.UnmarshalKey(keyBasic, &basic))

			// init servers
			utils.CheckAndExit(viper.UnmarshalKey(keyServers, &servers))
			sort.Sort(servers)

			// init servers map
			for _, s := range servers {
				serversMap[strings.ToLower(s.Name)] = s
			}

			// init tags
			utils.CheckAndExit(viper.UnmarshalKey(keyTags, &allTags))

			// init tags group
			for _, tag := range allTags {
				var tmpServers []*Server
				for _, server := range servers {
					for _, stag := range server.Tags {
						if tag == stag {
							tmpServers = append(tmpServers, server)
							break
						}
					}
				}
				tagsMap[tag] = tmpServers
			}
		})

}

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
		newTag := tag
		servers := tagsMap[newTag]
		if len(servers) == 0 {
			allTags = append(allTags, newTag)
		}
	}
	viper.Set(keyTags, allTags)

	// ssh user
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// allow empty
		return nil

	}, "User:")

	user := p.Run()
	if strings.TrimSpace(user) == "" {
		user = basic.User
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
		port = basic.Port
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
			privateKey = basic.PrivateKey
		}

		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "PrivateKey Password:")
		privateKeyPassword = p.Run()
		if strings.TrimSpace(privateKeyPassword) == "" {
			privateKeyPassword = basic.PrivateKeyPassword
		}
	} else {
		// use password
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// allow empty
			return nil

		}, "Password:")
		password = p.Run()
		if strings.TrimSpace(password) == "" {
			password = basic.Password
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
	if strings.TrimSpace(proxy) == "" {
		proxy = basic.Proxy
	}

	// create server
	server := Server{
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
	servers = append(servers, &server)
	sort.Sort(servers)
	viper.Set(keyServers, servers)
	utils.CheckAndExit(viper.WriteConfig())
}

func DeleteServer(name string) {

	delIdx := -1
	for i, s := range servers {
		if strings.ToLower(s.Name) == strings.ToLower(name) {
			delIdx = i
		}
	}

	if delIdx == -1 {
		utils.Exit("Server not found!", 1)
	} else {
		servers = append(servers[:delIdx], servers[delIdx+1:]...)
		sort.Sort(servers)
		viper.Set(keyServers, servers)
		utils.CheckAndExit(viper.WriteConfig())
	}

}

func ListServers() {

	tpl := `Name          User          Tags          Address
-------------------------------------------------------------
{{range . }}{{ .Name | ListLayout }}  {{ .User | ListLayout }}  {{ .Tags | MergeTag | ListLayout }}  {{ .Address }}:{{ .Port }}
{{end}}`
	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTag,
	})

	t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, servers))
	fmt.Println(buf.String())
}

func ListServer(serverName string) {
	s := findServerByName(serverName)
	if s == nil {
		fmt.Println("server not found!")
		return
	}
	tpl := `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | MergeTag }}
Proxy: {{ .Proxy }}`
	t := template.New("").Funcs(map[string]interface{}{"MergeTag": mergeTag})
	t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, s))
	fmt.Println(buf.String())
}

func listLayout(name string) string {
	if len(name) < 12 {
		return fmt.Sprintf("%-12s", name)
	} else {
		return fmt.Sprintf("%-12s", utils.ShortenString(name, 12))
	}
}

func mergeTag(tags []string) string {
	return strings.Join(tags, ",")
}

func findServerByName(name string) *Server {
	return serversMap[strings.ToLower(name)]
}
