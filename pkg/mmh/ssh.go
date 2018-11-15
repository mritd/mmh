/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
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

	"github.com/mritd/sshterminal"

	"strings"

	"fmt"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	Name      string   `yml:"Name"`
	Tags      []string `yml:"Tags"`
	User      string   `yml:"User"`
	Password  string   `yml:"Password"`
	PublicKey string   `yml:"PublicKey"`
	Address   string   `yml:"Address"`
	Port      int      `yml:"Port"`
	Proxy     string   `yml:"proxy"`
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

func publicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func (s *Server) authMethod() (ssh.AuthMethod, error) {
	// Priority use of public key
	if strings.TrimSpace(s.PublicKey) != "" {
		return publicKeyFile(s.PublicKey)
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

	return sshterminal.New(sshClient)
}
