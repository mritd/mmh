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
	User                string        `yaml:"user"`
	Password            string        `yaml:"password"`
	PrivateKey          string        `yaml:"privatekey"`
	PrivateKeyPassword  string        `yaml:"privatekey_password"`
	Port                int           `yaml:"port"`
	Proxy               string        `yaml:"proxy"`
	ServerAliveInterval time.Duration `yaml:"server_alive_interval"`
}

// server config
type Server struct {
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

// mmh context
type Context struct {
	Name       string `yaml:"name"`
	ConfigPath string `yaml:"config_path"`
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

// main config struct
type MainConfig struct {
	configPath string
	Basic      string    `yaml:"basic"`
	Contexts   []Context `yaml:"contexts"`
	Current    string    `yaml:"current"`
}

// main config example
func MainConfigExample() MainConfig {
	return MainConfig{
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
