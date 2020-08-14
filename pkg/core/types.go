package core

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"

	"gopkg.in/yaml.v2"
)

// server basic config
type BasicServerConfig struct {
	User                string            `yaml:"user"`
	Password            string            `yaml:"password"`
	PrivateKey          string            `yaml:"private_key"`
	PrivateKeyPassword  string            `yaml:"private_key_password"`
	KeyboardAuthCmd     string            `yaml:"keyboard_auth_cmd"`
	Environment         map[string]string `yaml:"environment"`
	EnableAPI           string            `yaml:"enable_api"`
	Port                int               `yaml:"port"`
	ServerAliveInterval time.Duration     `yaml:"server_alive_interval"`
}

// server config
type Server struct {
	Name                string            `yaml:"name"`
	Address             string            `yaml:"address"`
	Port                int               `yaml:"port"`
	User                string            `yaml:"user"`
	Password            string            `yaml:"password"`
	HookCmd             string            `yaml:"hook_cmd"`
	HookStdout          bool              `yaml:"hook_stdout"`
	KeyboardAuthCmd     string            `yaml:"keyboard_auth_cmd"`
	PrivateKey          string            `yaml:"private_key"`
	PrivateKeyPassword  string            `yaml:"private_key_password"`
	Environment         map[string]string `yaml:"environment"`
	EnableAPI           string            `yaml:"enable_api"`
	Proxy               string            `yaml:"proxy"`
	ServerAliveInterval time.Duration     `yaml:"server_alive_interval"`
	Tags                []string          `yaml:"tags"`
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
	Basic      BasicServerConfig `yaml:"basic"`
	MaxProxy   int               `yaml:"max_proxy"`
	Servers    Servers           `yaml:"servers"`
	Tags       Tags              `yaml:"tags"`
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
	home, _ := homedir.Dir()
	return BasicServerConfig{
		User:       "root",
		Port:       22,
		PrivateKey: filepath.Join(home, ".ssh", "id_rsa"),
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

// context config example
func ConfigExample() *Config {
	return &Config{
		Basic:    BasicServerExample(),
		Servers:  ServersExample(),
		Tags:     TagsExample(),
		MaxProxy: 5,
	}
}

type KeyBoardRequest struct {
	User        string   `json:"user"`
	Instruction string   `json:"instruction"`
	Questions   []string `json:"questions"`
	Echos       []bool   `json:"echos"`
}
