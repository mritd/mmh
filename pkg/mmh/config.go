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
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

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

const (
	KeyServers              = "servers"
	KeyBasic                = "basic"
	KeyTags                 = "tags"
	KeyContext              = "context"
	KeyContextUse           = "context_use"
	KeyContextUseTime       = "context_use_time"
	KeyContextTimeout       = "context_timeout"
	KeyContextAutoDowngrade = "context_auto_downgrade"
)

var (
	MainViper   = viper.New()
	CtxViper    = viper.New()
	AllContexts Contexts
	initOnce    sync.Once
	allTags     []string
	basic       Basic
	servers     Servers
	serversMap  = make(map[string]*Server)
	tagsMap     = make(map[string][]*Server)
	maxProxy    = 0
)

var (
	inputEmptyErr    = errors.New("input is empty")
	inputTooLongErr  = errors.New("input length must be <= 12")
	serverExistErr   = errors.New("server name exist")
	notNumberErr     = errors.New("only number support")
	proxyNotFoundErr = errors.New("proxy server not found")
)

func InitConfig() {

	initOnce.Do(
		func() {

			// set default max proxy
			if maxProxy == 0 {
				maxProxy = CtxViper.GetInt("maxProxy")
			}

			// get home dir
			home, err := homedir.Dir()
			utils.CheckAndExit(err)

			// set default basic config
			CtxViper.SetDefault(KeyBasic, Basic{
				User:               "root",
				Port:               22,
				PrivateKey:         filepath.Join(home, ".ssh", "id_rsa"),
				PrivateKeyPassword: "",
				Password:           "",
				Proxy:              "",
			})

			// init basic config
			utils.CheckAndExit(CtxViper.UnmarshalKey(KeyBasic, &basic))

			// init servers
			utils.CheckAndExit(CtxViper.UnmarshalKey(KeyServers, &servers))
			sort.Sort(servers)

			// init servers map
			for _, s := range servers {
				serversMap[strings.ToLower(s.Name)] = s
			}

			// init tags
			utils.CheckAndExit(CtxViper.UnmarshalKey(KeyTags, &allTags))

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

func ServerAdd() {

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
	CtxViper.Set(KeyTags, allTags)

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
	CtxViper.Set(KeyServers, servers)
	utils.CheckAndExit(CtxViper.WriteConfig())
}

func ServerDelete(serverNames []string) {

	var deletesIdx []int

	for _, serverName := range serverNames {
		for i, s := range servers {
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
		servers = append(servers[:del-i], servers[del-i+1:]...)
	}

	// save config
	sort.Sort(servers)
	CtxViper.Set(KeyServers, servers)
	utils.CheckAndExit(CtxViper.WriteConfig())

}

func ServerList() {

	tpl := `Name          User          Tags          Address
-------------------------------------------------------------
{{range . }}{{ .Name | ListLayout }}  {{ .User | ListLayout }}  {{ .Tags | MergeTag | ListLayout }}  {{ .Address }}:{{ .Port }}
{{end}}`
	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTag,
	})

	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, servers))
	fmt.Println(buf.String())
}

func ServerDetail(serverName string) {
	s := findServerByName(serverName)
	if s == nil {
		utils.Exit("server not found!", 1)
	}
	tpl := `Name: {{ .Name }}
User: {{ .User }}
Address: {{ .Address }}:{{ .Port }}
Tags: {{ .Tags | MergeTag }}
Proxy: {{ .Proxy }}`
	t := template.New("").Funcs(map[string]interface{}{"MergeTag": mergeTag})
	_, _ = t.Parse(tpl)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, s))
	fmt.Println(buf.String())
}

func ContextList() {

	tpl := `  Name          Path
---------------------------------
{{ range . }}{{ if .CurrentContext }}» {{ .Name | ListLayout }}{{ else }}  {{ .Name | ListLayout }}{{ end }}  {{ .ConfigPath }}
{{ end }}`

	t := template.New("").Funcs(map[string]interface{}{
		"ListLayout": listLayout,
		"MergeTag":   mergeTag,
	})
	_, _ = t.Parse(tpl)

	currentContext := MainViper.GetString(KeyContextUse)

	var ctxList contextDetails
	for k, v := range AllContexts {
		ctxList = append(ctxList, contextDetail{
			Name:           k,
			ConfigPath:     v.ConfigPath,
			CurrentContext: k == currentContext,
		})
	}
	sort.Sort(ctxList)
	var buf bytes.Buffer
	utils.CheckAndExit(t.Execute(&buf, ctxList))
	fmt.Println(buf.String())
}

func ContextUse(ctxName string) {
	_, ok := AllContexts[ctxName]
	if !ok {
		utils.Exit(fmt.Sprintf("context [%s] not found", ctxName), 1)
	}
	MainViper.Set(KeyContextUse, ctxName)
	utils.CheckAndExit(MainViper.WriteConfig())
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

func UpdateContextTimestamp(_ *cobra.Command, _ []string) {

	home, _ := homedir.Dir()
	pidFile := filepath.Join(home, ".mmh", ".pid")
	_ = os.Remove(pidFile)

	MainViper.Set(KeyContextUseTime, time.Now())
	utils.CheckAndExit(MainViper.WriteConfig())
}

func UpdateContextTimestampTask(_ *cobra.Command, _ []string) {

	// get context
	contextUseTime := MainViper.GetTime(KeyContextUseTime)
	contextTimeout := MainViper.GetDuration(KeyContextTimeout)
	contextAutoDowngrade := MainViper.GetString(KeyContextAutoDowngrade)

	// if context auto downgrade is open
	if !contextUseTime.IsZero() && contextTimeout != 0 && contextAutoDowngrade != "" {

		home, _ := homedir.Dir()
		pid := strconv.Itoa(os.Getpid())
		pidFile := filepath.Join(home, ".mmh", ".pid")

		go func() {
			for {
				select {
				case <-time.Tick(contextTimeout - 3*time.Second):

					if _, err := os.Stat(pidFile); os.IsNotExist(err) {
						// write current pid to pid file
						utils.CheckAndExit(ioutil.WriteFile(pidFile, []byte(pid), 0644))
					} else {
						p, err := ioutil.ReadFile(pidFile)
						if err != nil {
							fmt.Println(err)
						}

						// check pid
						if string(p) == pid {
							MainViper.Set(KeyContextUseTime, time.Now())
							err = MainViper.WriteConfig()
							if err != nil {
								fmt.Println(err)
							}
						}
					}

				}
			}
		}()

	}

}
