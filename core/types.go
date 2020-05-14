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
	User                string        `yaml:"user"`
	Password            string        `yaml:"password"`
	PrivateKey          string        `yaml:"private_key"`
	PrivateKeyPassword  string        `yaml:"private_key_password"`
	Port                int           `yaml:"port"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`
	TmuxSupport         string        `yaml:"tmux_support"`
	TmuxAutoRename      string        `yaml:"tmux_auto_rename"`
}

// server config
type ServerConfig struct {
	Name                string        `yaml:"name"`
	Tags                []string      `yaml:"tags"`
	User                string        `yaml:"user"`
	Password            string        `yaml:"password"`
	SuRoot              bool          `yaml:"su_root"`
	UseSudo             bool          `yaml:"use_sudo"`
	NoPasswordSudo      bool          `yaml:"no_password_sudo"`
	RootPassword        string        `yaml:"root_password"`
	PrivateKey          string        `yaml:"private_key"`
	PrivateKeyPassword  string        `yaml:"private_key_password"`
	Address             string        `yaml:"address"`
	Port                int           `yaml:"port"`
	Proxy               string        `yaml:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`
	TmuxSupport         string        `yaml:"tmux_support"`
	TmuxAutoRename      string        `yaml:"tmux_auto_rename"`
	proxyCount          int
}

// server tags
type Tags []string

// mmh servers
type Servers []*ServerConfig

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

type ConfigInfo []struct {
	Name      string
	Path      string
	IsCurrent bool
}

func (info ConfigInfo) Len() int {
	return len(info)
}
func (info ConfigInfo) Less(i, j int) bool {
	return info[i].Name < info[j].Name
}
func (info ConfigInfo) Swap(i, j int) {
	info[i], info[j] = info[j], info[i]
}

// basic config example
func BasicServerExample() BasicServerConfig {
	home, _ := homedir.Dir()
	return BasicServerConfig{
		User:                "root",
		Port:                22,
		PrivateKey:          filepath.Join(home, ".ssh", "id_rsa"),
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

// context config example
func ConfigExample() *Config {
	return &Config{
		Basic:    BasicServerExample(),
		Servers:  ServersExample(),
		Tags:     TagsExample(),
		MaxProxy: 5,
	}
}
