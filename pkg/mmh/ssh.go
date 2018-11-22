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
	"time"

	"github.com/mritd/sshutils"

	"strings"

	"fmt"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Name               string   `yaml:"name" mapstructure:"name"`
	Tags               []string `yaml:"tags" mapstructure:"tags"`
	User               string   `yaml:"user" mapstructure:"user"`
	Password           string   `yaml:"password" mapstructure:"password"`
	PrivateKey         string   `yaml:"privatekey" mapstructure:"privatekey"`
	PrivateKeyPassword string   `yaml:"privatekey_password" mapstructure:"privatekey_password"`
	Address            string   `yaml:"address" mapstructure:"address"`
	Port               int      `yaml:"port" mapstructure:"port"`
	Proxy              string   `yaml:"proxy" mapstructure:"proxy"`
}

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

func privateKeyFile(file, password string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var signer ssh.Signer

	if password == "" {
		signer, err = ssh.ParsePrivateKey(buffer)
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buffer, []byte(password))
	}

	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func (s *Server) authMethod() (ssh.AuthMethod, error) {
	if strings.TrimSpace(s.PrivateKey) != "" {
		return privateKeyFile(s.PrivateKey, s.PrivateKeyPassword)
	} else {
		return password(s.Password), nil
	}
}

func (s *Server) sshClient() (*ssh.Client, error) {

	var client *ssh.Client
	var proxyCount int
	if maxProxy == 0 {
		maxProxy = 5
	}

	auth, err := s.authMethod()
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	if s.Proxy != "" {

		if proxyCount > maxProxy {
			return nil, errors.New("too many proxy node")
		} else {
			proxyCount++
		}

		// find proxy server
		proxy := findServerByName(s.Proxy)
		if proxy == nil {
			return nil, errors.New(fmt.Sprintf("cloud not found proxy server: %s", s.Proxy))
		} else {
			fmt.Printf("Using proxy [%s], connect to %s\n", s.Proxy, s.Name)
		}

		// recursive connect
		proxyClient, err := proxy.sshClient()
		if err != nil {
			return nil, err
		}
		conn, err := proxyClient.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port))
		if err != nil {
			return nil, err
		}
		ncc, channel, request, err := ssh.NewClientConn(conn, fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		if err != nil {
			return nil, err
		}
		client = ssh.NewClient(ncc, channel, request)
	} else {
		client, err = ssh.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (s *Server) Connect() error {
	sshClient, err := s.sshClient()
	if err != nil {
		return err
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	sshSession := sshutils.NewSSHSession(session)
	defer sshSession.Close()

	return sshSession.Terminal()
}
