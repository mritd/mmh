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
	"io/ioutil"

	"github.com/mritd/sshterminal"

	"github.com/spf13/viper"

	"strings"

	"fmt"

	"github.com/mritd/mmh/pkg/utils"
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

type Servers []Server

func (servers Servers) Len() int {
	return len(servers)
}
func (servers Servers) Less(i, j int) bool {
	return servers[i].Name < servers[j].Name
}
func (servers Servers) Swap(i, j int) {
	servers[i], servers[j] = servers[j], servers[i]
}

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	utils.CheckAndExit(err)

	key, err := ssh.ParsePrivateKey(buffer)
	utils.CheckAndExit(err)
	return ssh.PublicKeys(key)
}

func password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func (s Server) authMethod() ssh.AuthMethod {
	// Priority use of public key
	if strings.TrimSpace(s.PublicKey) != "" {
		return publicKeyFile(s.PublicKey)
	} else {
		return password(s.Password)
	}
}

func (s Server) sshClient() *ssh.Client {

	var client *ssh.Client
	var err error
	var proxyCount int
	var maxProxy = viper.GetInt("maxProxy")
	if maxProxy == 0 {
		maxProxy = 5
	}

	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.authMethod(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if s.Proxy != "" {

		if proxyCount > maxProxy {
			utils.Exit("Too many proxy node!", 1)
		} else {
			proxyCount++
		}

		proxy := findServerByName(s.Proxy)
		fmt.Printf("Using proxy [%s], connect to %s\n", s.Proxy, s.Name)
		proxyClient := proxy.sshClient()
		conn, err := proxyClient.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port))
		utils.CheckAndExit(err)
		ncc, channel, request, err := ssh.NewClientConn(conn, fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		utils.CheckAndExit(err)
		client = ssh.NewClient(ncc, channel, request)
	} else {
		client, err = ssh.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		utils.CheckAndExit(err)
	}

	return client
}

func (s Server) Connect() {
	sshClient := s.sshClient()
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	utils.CheckAndExit(err)
	defer session.Close()

	utils.CheckAndExit(sshterminal.New(sshClient))
}
