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
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
)

// server basic config
type Basic struct {
	User                string        `yaml:"user" mapstructure:"user"`
	Password            string        `yaml:"password" mapstructure:"password"`
	PrivateKey          string        `yaml:"privatekey" mapstructure:"privatekey"`
	PrivateKeyPassword  string        `yaml:"privatekey_password" mapstructure:"privatekey_password"`
	Port                int           `yaml:"port" mapstructure:"port"`
	Proxy               string        `yaml:"proxy" mapstructure:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval" mapstructure:"server_alive_interval"`
}

// server config
type Server struct {
	Name                string        `yaml:"name" mapstructure:"name"`
	Tags                []string      `yaml:"tags" mapstructure:"tags"`
	User                string        `yaml:"user" mapstructure:"user"`
	Password            string        `yaml:"password" mapstructure:"password"`
	PrivateKey          string        `yaml:"privatekey" mapstructure:"privatekey"`
	PrivateKeyPassword  string        `yaml:"privatekey_password" mapstructure:"privatekey_password"`
	Address             string        `yaml:"address" mapstructure:"address"`
	Port                int           `yaml:"port" mapstructure:"port"`
	Proxy               string        `yaml:"proxy" mapstructure:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval" mapstructure:"server_alive_interval"`
	proxyCount          int
}

// server tags
type Tags []string

// mmh context
type Context struct {
	Name       string `yaml:"name" mapstructure:"name"`
	ConfigPath string `yaml:"config_path" mapstructure:"config_path"`
}

// mmh context list
type Contexts struct {
	Context       []Context     `yaml:"context" mapstructure:"context"`
	Current       string        `yaml:"current" mapstructure:"current"`
	TimeStamp     time.Time     `yaml:"timestamp" mapstructure:"timestamp"`
	TimeOut       time.Duration `yaml:"timeout" mapstructure:"timeout"`
	AutoDowngrade string        `yaml:"auto_downgrade" mapstructure:"auto_downgrade"`
}

// mmh context detail
type ContextDetail struct {
	Name           string
	ConfigPath     string
	CurrentContext bool
}

// mmh context details
type ContextDetails []ContextDetail

func (cd ContextDetails) Len() int {
	return len(cd)
}
func (cd ContextDetails) Less(i, j int) bool {
	return cd[i].Name < cd[j].Name
}
func (cd ContextDetails) Swap(i, j int) {
	cd[i], cd[j] = cd[j], cd[i]
}

// mmh servers
type Servers []*Server

func (servers Servers) Len() int {
	return len(servers)
}
func (servers Servers) Less(i, j int) bool {
	return servers[i].Name < servers[j].Name
}
func (servers Servers) Swap(i, j int) {
	servers[i], servers[j] = servers[j], servers[i]
}

// context config example
func ExampleContexts() Contexts {
	return Contexts{
		Context: []Context{
			{
				Name:       "default",
				ConfigPath: "./default.yaml",
			},
		},
		AutoDowngrade: "default",
		Current:       "default",
		TimeStamp:     time.Now(),
		TimeOut:       0,
	}
}

// basic config example
func ExampleBasic() Basic {
	home, _ := homedir.Dir()
	return Basic{
		User:                "root",
		Port:                22,
		PrivateKey:          filepath.Join(home, ".ssh", "id_rsa"),
		PrivateKeyPassword:  "",
		Password:            "",
		Proxy:               "",
		ServerAliveInterval: 0,
	}
}

// server config example
func ExampleServers() Servers {
	home, _ := homedir.Dir()
	return Servers{
		{
			Name:                "prod11",
			User:                "root",
			Tags:                []string{"prod"},
			Address:             "10.10.4.11",
			Port:                22,
			Password:            "password",
			Proxy:               "prod12",
			ServerAliveInterval: 20 * time.Second,
		},
		{
			Name:                "prod12",
			User:                "root",
			Tags:                []string{"prod"},
			Address:             "10.10.4.12",
			Port:                22,
			PrivateKey:          filepath.Join(home, ".ssh", "id_rsa"),
			PrivateKeyPassword:  "password",
			ServerAliveInterval: 10 * time.Second,
		},
	}
}

// tags config example
func ExampleTags() Tags {
	return Tags{
		"prod",
	}
}
