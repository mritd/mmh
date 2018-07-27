package mmh

import (
	"errors"
	"strings"

	"strconv"

	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/pkg/utils"
	"github.com/mritd/promptx"
	"github.com/spf13/viper"
)

func ConfigExample() []Server {
	return []Server{
		{
			Name:     "prod11",
			User:     "root",
			Group:    "prod",
			Address:  "10.10.4.11",
			Port:     22,
			Password: "password",
		},
		{
			Name:      "prod12",
			User:      "root",
			Group:     "prod",
			Address:   "10.10.4.12",
			Port:      22,
			PublicKey: "/Users/mritd/.ssh/id_rsa",
		},
	}
}

func AddServer() {

	// Name
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Input is empty!")
		} else if len(line) > 25 {
			return errors.New("Input length must < 25!")
		}
		return nil

	}, "Name:")

	name := p.Run()

	// Group
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		// Allow empty
		return nil
	}, "Group:")

	group := p.Run()

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

	server := Server{
		Name:      name,
		Group:     group,
		User:      user,
		Address:   address,
		Port:      port,
		PublicKey: publicKey,
		Password:  password,
	}

	// Save
	var servers []Server
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))
	servers = append(servers, server)
	viper.Set("Servers", servers)
	utils.CheckAndExit(viper.WriteConfig())
}

func DeleteServer(name string) {
	var servers []Server
	utils.CheckAndExit(viper.UnmarshalKey("servers", &servers))

	var delIdx int
	for i, s := range servers {
		if strings.ToLower(s.Name) == strings.ToLower(name) {
			delIdx = i
		}
	}

	servers = append(servers[:delIdx], servers[delIdx+1:]...)
	viper.Set("Servers", servers)
	utils.CheckAndExit(viper.WriteConfig())
}
