package core

import (
	"errors"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// server basic config
type BasicServerConfig struct {
	User                string            `yaml:"user,omitempty"`
	Password            string            `yaml:"password,omitempty"`
	PrivateKey          string            `yaml:"private_key,omitempty"`
	PrivateKeyPassword  string            `yaml:"private_key_password,omitempty"`
	KeyboardAuthCmd     string            `yaml:"keyboard_auth_cmd,omitempty"`
	Environment         map[string]string `yaml:"environment,omitempty"`
	EnableAPI           string            `yaml:"enable_api,omitempty"`
	ExtAuth             string            `yaml:"ext_auth,omitempty"`
	Port                int               `yaml:"port,omitempty"`
	ServerAliveInterval time.Duration     `yaml:"server_alive_interval,omitempty"`
}

// server config
type Server struct {
	Name                string            `yaml:"name,omitempty"`
	Address             string            `yaml:"address"`
	Port                int               `yaml:"port,omitempty"`
	User                string            `yaml:"user,omitempty"`
	Proxy               string            `yaml:"proxy,omitempty"`
	Password            string            `yaml:"password,omitempty"`
	PrivateKey          string            `yaml:"private_key,omitempty"`
	PrivateKeyPassword  string            `yaml:"private_key_password,omitempty"`
	HookCmd             string            `yaml:"hook_cmd,omitempty"`
	HookStdout          bool              `yaml:"hook_stdout,omitempty"`
	KeyboardAuthCmd     string            `yaml:"keyboard_auth_cmd,omitempty"`
	Environment         map[string]string `yaml:"environment,omitempty"`
	EnableAPI           string            `yaml:"enable_api,omitempty"`
	ExtAuth             string            `yaml:"ext_auth,omitempty"`
	ServerAliveInterval time.Duration     `yaml:"server_alive_interval,omitempty"`
	Tags                []string          `yaml:"tags,omitempty"`
}

// server tags
type Tags []string

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

// context config(eg: default.yaml)
type Config struct {
	configPath string
	Basic      BasicServerConfig `yaml:"basic,omitempty"`
	MaxProxy   int               `yaml:"max_proxy,omitempty"`
	Servers    Servers           `yaml:"servers"`
	Tags       Tags              `yaml:"tags,omitempty"`
}

// set config file path
func (cfg *Config) SetConfigPath(configPath string) {
	cfg.configPath = configPath
}

// write config
func (cfg *Config) Write() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cfg.configPath, out, 0644)
}

// write config to yaml file
func (cfg *Config) WriteTo(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Write()
}

// load config
func (cfg *Config) Load() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	buf, err := ioutil.ReadFile(cfg.configPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, cfg)
}

// load config from yaml file
func (cfg *Config) LoadFrom(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Load()
}

type ConfigInfo struct {
	Name      string
	Path      string
	IsCurrent bool
}

type ConfigList []ConfigInfo

func (info ConfigList) Len() int {
	return len(info)
}
func (info ConfigList) Less(i, j int) bool {
	return info[i].Name < info[j].Name
}
func (info ConfigList) Swap(i, j int) {
	info[i], info[j] = info[j], info[i]
}

// basic config example
func BasicServerExample() BasicServerConfig {
	return BasicServerConfig{
		User:     "root",
		Password: "password",
	}
}

// server config example
func ServersExample() Servers {
	return Servers{
		{
			Name:    "prod11",
			Address: "10.10.4.11",
			Proxy:   "prod12",
		},
		{
			Name:    "prod12",
			Address: "10.10.4.12",
		},
	}
}

// context config example
func ConfigExample() *Config {
	return &Config{
		Basic:    BasicServerExample(),
		Servers:  ServersExample(),
		MaxProxy: 5,
	}
}

type KeyBoardRequest struct {
	User        string   `json:"user"`
	Instruction string   `json:"instruction"`
	Questions   []string `json:"questions"`
	Echos       []bool   `json:"echos"`
}
