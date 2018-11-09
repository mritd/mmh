/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
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
const keyTags = "tags"

var (
	servers    = Servers{}
	serversMap = make(map[string]*Server)
	tagsMap    = make(map[string][]*Server)
)

func WriteExampleConfig() {
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
			Name:      "prod12",
			User:      "root",
			Tags:      []string{"prod"},
			Address:   "10.10.4.12",
			Port:      22,
			PublicKey: "/Users/mritd/.ssh/id_rsa",
		},
	})
	viper.Set(keyTags, []string{
		"prod",
		"test",
	})
	viper.Set("MaxProxy", 5)
	viper.WriteConfig()
}

func InitServers() {
	// init servers
	utils.CheckAndExit(viper.UnmarshalKey(keyServers, &servers))

	// init servers map
	for _, s := range servers {
		serversMap[strings.ToLower(s.Name)] = s
	}

	// init tags group
	var tags []string
	utils.CheckAndExit(viper.UnmarshalKey(keyTags, &tags))
	for _, tag := range tags {
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
}

func AddServer() {

	// Name
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		} else if len(line) > 12 {
			return errors.New("Input length must <= 12!")
		}

		if s := findServerByName(string(line)); s != nil {
			return errors.New("Server name exist!")
		}
		return nil

	}, "Name:")

	name := p.Run()

	// Tags
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// Allow empty
		return nil
	}, "Tags:")

	tags := strings.Fields(p.Run())

	// SSH user
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// Allow empty
		return nil

	}, "User:")

	user := p.Run()
	if strings.TrimSpace(user) == "" {
		user = "root"
	}

	// Server address
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		}
		return nil

	}, "Address:")

	address := p.Run()

	// Server port
	var port int
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			_, err := strconv.Atoi(string(line))
			if err != nil {
				return errors.New("Only number support!")
			}
		}
		return nil

	}, "Port:")

	portStr := p.Run()
	if strings.TrimSpace(portStr) == "" {
		port = 22
	} else {
		port, _ = strconv.Atoi(portStr)
	}

	// Auth method
	var password, publicKey string
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
			"PublicKey",
			"Password",
		},
		Config: cfg,
	}

	idx := s.Run()
	if idx == 0 {
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			// Allow empty
			return nil

		}, "PublicKey:")

		publicKey = p.Run()
		if strings.TrimSpace(publicKey) == "" {
			home, err := homedir.Dir()
			utils.CheckAndExit(err)
			publicKey = home + string(filepath.Separator) + ".ssh" + string(filepath.Separator) + "id_rsa"
		}
	} else {
		p = promptx.NewDefaultPrompt(func(line []rune) error {
			if strings.TrimSpace(string(line)) == "" {
				return errors.New("Input is empty!")
			}
			return nil

		}, "Password:")
		password = p.Run()
	}

	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) != "" {
			if findServerByName(string(line)) == nil {
				return errors.New("Proxy server not found!")
			}
		}
		return nil

	}, "Proxy:")

	proxy := p.Run()

	server := Server{
		Name:      name,
		Tags:      tags,
		User:      user,
		Address:   address,
		Port:      port,
		PublicKey: publicKey,
		Password:  password,
		Proxy:     proxy,
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
	sort.Sort(servers)

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
		fmt.Println("Server not found!")
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
