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
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

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
	SuRoot              bool          `yaml:"su_root" mapstructure:"su_root"`
	UseSudo             bool          `yaml:"use_sudo" mapstructure:"use_sudo"`
	RootPassword        string        `yaml:"root_password" mapstructure:"root_password"`
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
type ContextList struct {
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
func MainConfigExample() MainConfig {
	return MainConfig{
		Contexts: ContextList{
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
		},
	}
}

func ContextConfigExample() ContextConfig {
	return ContextConfig{
		Basic:   BasicExample(),
		Servers: ServersExample(),
		Tags:    TagsExample(),
	}
}

// basic config example
func BasicExample() Basic {
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
func ServersExample() Servers {
	home, _ := homedir.Dir()
	return Servers{
		{
			Name:                "prod11",
			User:                "root",
			Tags:                []string{"prod"},
			Address:             "10.10.4.11",
			Port:                22,
			Password:            "password",
			SuRoot:              true,
			UseSudo:             true,
			RootPassword:        "root",
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
func TagsExample() Tags {
	return Tags{
		"prod",
	}
}

// main.yaml config struct
type MainConfig struct {
	configPath string
	Contexts   ContextList `yaml:"context_list"`
}

// set config file path
func (cfg MainConfig) SetConfigPath(configPath string) MainConfig {
	cfg.configPath = configPath
	return cfg
}

// write config to yaml file
func (cfg MainConfig) Write() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.configPath, out, 0644)
}

// load config from yaml file
func (cfg *MainConfig) Load(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, cfg)
	return err
}

// context config(eg: default.yaml)
type ContextConfig struct {
	configPath string
	Basic      Basic   `yaml:"basic"`
	Servers    Servers `yaml:"servers"`
	Tags       Tags    `yaml:"tags"`
}

// set config file path
func (cfg ContextConfig) SetConfigPath(configPath string) ContextConfig {
	cfg.configPath = configPath
	return cfg
}

// write config to yaml file
func (cfg ContextConfig) Write() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.configPath, out, 0644)
}

// load config from yaml file
func (cfg *ContextConfig) Load(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, cfg)
}
