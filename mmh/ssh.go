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
func (s *ServerConfig) setDefault() {
	if s.User == "" {
		s.User = ContextCfg.Basic.User
	}
	if s.Port == 0 {
		s.Port = ContextCfg.Basic.Port
	}
	if s.Password == "" {
		s.Password = ContextCfg.Basic.Password
		if s.PrivateKey == "" {
			s.PrivateKey = ContextCfg.Basic.PrivateKey
		}
	}

	if s.PrivateKeyPassword == "" {
		s.PrivateKeyPassword = ContextCfg.Basic.PrivateKeyPassword
	}
	if s.Proxy == "" {
		s.Proxy = ContextCfg.Basic.Proxy
	}
}

// return a ssh client intense point
func (s *ServerConfig) sshClient() (*ssh.Client, error) {

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
		proxyServer := ContextCfg.Servers.FindServerByName(s.Proxy)
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
func (s *ServerConfig) Terminal() error {
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

	var sshSession *sshutils.SSHSession
	if s.SuRoot {
		sshSession = sshutils.NewSSHSessionWithRoot(session, s.UseSudo, s.NoPasswordSudo, s.RootPassword, s.Password)
	} else {
		sshSession = sshutils.NewSSHSession(session)
	}

	defer func() {
		_ = sshSession.Close()
	}()

	// keep alive
	if s.ServerAliveInterval > 0 {
		return sshSession.TerminalWithKeepAlive(s.ServerAliveInterval)
	}
	return sshSession.Terminal()

}

// get auth method
// priority use privateKey method
func (s *ServerConfig) authMethod() (ssh.AuthMethod, error) {
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
