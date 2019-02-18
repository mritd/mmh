package mmh

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
	PrivateKey          string        `yaml:"privatekey"`
	PrivateKeyPassword  string        `yaml:"privatekey_password"`
	Port                int           `yaml:"port"`
	Proxy               string        `yaml:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`
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
	PrivateKey          string        `yaml:"privatekey"`
	PrivateKeyPassword  string        `yaml:"privatekey_password"`
	Address             string        `yaml:"address"`
	Port                int           `yaml:"port"`
	Proxy               string        `yaml:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`
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

// mmh context
type Context struct {
	Name       string `yaml:"name"`
	ConfigPath string `yaml:"config_path"`
}

// mmh contexts
type Contexts []Context

func (cs Contexts) Len() int {
	return len(cs)
}
func (cs Contexts) Less(i, j int) bool {
	return cs[i].Name < cs[j].Name
}
func (cs Contexts) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

// main config struct
type MainConfig struct {
	configPath string
	Basic      string   `yaml:"basic"`
	Contexts   Contexts `yaml:"contexts"`
	Current    string   `yaml:"current"`
}

// set config file path
func (cfg *MainConfig) SetConfigPath(configPath string) {
	cfg.configPath = configPath
}

// write config
func (cfg *MainConfig) Write() error {
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
func (cfg *MainConfig) WriteTo(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Write()
}

// load config
func (cfg *MainConfig) Load() error {
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
func (cfg *MainConfig) LoadFrom(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Load()
}

// context config(eg: default.yaml)
type ContextConfig struct {
	configPath string
	Basic      BasicServerConfig `yaml:"basic"`
	Servers    Servers           `yaml:"servers"`
	Tags       Tags              `yaml:"tags"`
}

// set config file path
func (cfg *ContextConfig) SetConfigPath(configPath string) {
	cfg.configPath = configPath
}

// write config
func (cfg *ContextConfig) Write() error {
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
func (cfg *ContextConfig) WriteTo(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Write()
}

// load config
func (cfg *ContextConfig) Load() error {
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
func (cfg *ContextConfig) LoadFrom(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Load()
}

// basic config example
func BasicServerExample() BasicServerConfig {
	home, _ := homedir.Dir()
	return BasicServerConfig{
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

// main config example
func MainConfigExample() *MainConfig {
	return &MainConfig{
		Basic: "default",
		Contexts: []Context{
			{
				Name:       "default",
				ConfigPath: "./default.yaml",
			},
			{
				Name:       "test",
				ConfigPath: "./test.yaml",
			},
		},
	}
}

// context config example
func ContextConfigExample() *ContextConfig {
	return &ContextConfig{
		Basic:   BasicServerExample(),
		Servers: ServersExample(),
		Tags:    TagsExample(),
	}
}
