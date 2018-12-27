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

// set server default config
func (s *Server) setDefault() {
	if s.User == "" {
		s.User = BasicCfg.User
	}
	if s.Port == 0 {
		s.Port = BasicCfg.Port
	}
	if s.Password == "" {
		s.Password = BasicCfg.Password
		if s.PrivateKey == "" {
			s.PrivateKey = BasicCfg.PrivateKey
		}
	}

	if s.PrivateKeyPassword == "" {
		s.PrivateKeyPassword = BasicCfg.PrivateKeyPassword
	}
	if s.Proxy == "" {
		s.Proxy = BasicCfg.Proxy
	}
}

// return a ssh client intense point
func (s *Server) sshClient() (*ssh.Client, error) {

	// default to basic config
	s.setDefault()

	var client *ssh.Client
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

		// check max proxy
		if s.proxyCount > MaxProxy {
			return nil, errors.New(fmt.Sprintf("too many proxy server, proxy server must be <= %d", MaxProxy))
		} else {
			s.proxyCount++
		}

		// find proxy server
		proxyServer := ServersCfg.FindServerByName(s.Proxy)
		if proxyServer == nil {
			return nil, errors.New(fmt.Sprintf("cloud not found proxy server: %s", s.Proxy))
		} else {
			fmt.Printf("🔑 using proxy [%s], connect to => %s\n", s.Proxy, s.Name)
		}

		// recursive connect
		proxyClient, err := proxyServer.sshClient()
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

// start a ssh terminal
func (s *Server) Terminal() error {
	sshClient, err := s.sshClient()
	if err != nil {
		return err
	}
	defer func() {
		_ = sshClient.Close()
	}()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	sshSession := sshutils.NewSSHSession(session)
	defer func() {
		_ = sshSession.Close()
	}()

	if s.ServerAliveInterval > 0 {
		return sshSession.TerminalWithKeepAlive(s.ServerAliveInterval)
	}
	return sshSession.Terminal()

}

// get auth method
// priority use privateKey method
func (s *Server) authMethod() (ssh.AuthMethod, error) {
	if strings.TrimSpace(s.PrivateKey) != "" {
		return privateKeyFile(s.PrivateKey, s.PrivateKeyPassword)
	} else {
		return password(s.Password), nil
	}
}

// use private key to return ssh auth method
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

// use password to return ssh auth method
func password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}
