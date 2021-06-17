package core

import (
	"errors"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// BasicServerConfig server basic config
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

// Server server config
type Server struct {
	Name                string            `yaml:"name"`
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
	ServerAliveInterval time.Duration     `yaml:"server_alive_interval,omitempty"`
	Tags                []string          `yaml:"tags,omitempty"`

	ConfigPath string `yaml:"config_path,omitempty"`
}

// Tags server tags
type Tags []string

// Servers mmh servers
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

// Config context config(eg: default.yaml)
type Config struct {
	configPath string
	Basic      BasicServerConfig `yaml:"basic,omitempty"`
	MaxProxy   int               `yaml:"max_proxy,omitempty"`
	Servers    Servers           `yaml:"servers"`
	Tags       Tags              `yaml:"tags,omitempty"`
}

// SetConfigPath set config file path
func (cfg *Config) SetConfigPath(configPath string) {
	cfg.configPath = configPath
}

// Write write config
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

// WriteTo write config to yaml file
func (cfg *Config) WriteTo(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Write()
}

// Load load config
func (cfg *Config) Load() error {
	if cfg.configPath == "" {
		return errors.New("config path not set")
	}
	buf, err := ioutil.ReadFile(cfg.configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		return err
	}
	cfg.mergeBasic()
	return nil
}

// LoadFrom load config from yaml file
func (cfg *Config) LoadFrom(filePath string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	cfg.configPath = filePath
	return cfg.Load()
}

func (cfg *Config) mergeBasic() {
	for _, s := range cfg.Servers {
		s.ConfigPath = cfg.configPath
		if s.User == "" {
			s.User = cfg.Basic.User
			if s.User == "" {
				s.User = "root"
			}
		}
		if s.Password == "" {
			s.Password = cfg.Basic.Password
		}
		if s.PrivateKey == "" {
			s.PrivateKey = cfg.Basic.PrivateKey
		}
		if s.PrivateKeyPassword == "" {
			s.PrivateKeyPassword = cfg.Basic.PrivateKeyPassword
		}
		if s.KeyboardAuthCmd == "" {
			s.KeyboardAuthCmd = cfg.Basic.KeyboardAuthCmd
		}
		if s.EnableAPI == "" {
			s.EnableAPI = cfg.Basic.EnableAPI
		}
		if s.Environment == nil {
			s.Environment = cfg.Basic.Environment
			if s.Environment == nil {
				s.Environment = make(map[string]string)
			}
		}
		if s.Port == 0 {
			s.Port = cfg.Basic.Port
			if s.Port == 0 {
				s.Port = 22
			}
		}
		if s.ServerAliveInterval == 0 {
			s.ServerAliveInterval = cfg.Basic.ServerAliveInterval
			if s.ServerAliveInterval == 0 {
				s.ServerAliveInterval = 10 * time.Second
			}
		}
	}
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

type KeyBoardRequest struct {
	User        string   `json:"user"`
	Instruction string   `json:"instruction"`
	Questions   []string `json:"questions"`
	Echos       []bool   `json:"echos"`
}

// ConfigExample context config example
func ConfigExample() *Config {
	return &Config{
		Basic: BasicServerConfig{
			User:     "root",
			Password: "password",
		},
		Servers: Servers{
			{
				Name:    "prod11",
				Address: "10.10.4.11",
				Proxy:   "prod12",
			},
			{
				Name:    "prod12",
				Address: "10.10.4.12",
			},
		},
		MaxProxy: 5,
	}
}
