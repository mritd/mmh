package mmh

import (
	"errors"
	"io/ioutil"
	"time"

	"github.com/mritd/sshutils"

	"fmt"

	"golang.org/x/crypto/ssh"
)

// return a ssh client intense point
func (s *ServerConfig) sshClient() (*ssh.Client, error) {

	var client *ssh.Client
	var err error

	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		Auth:            s.authMethod(),
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
		proxyServer := findServerByName(s.Proxy)
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
	defer func() { _ = sshClient.Close() }()

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

	defer func() { _ = sshSession.Close() }()

	// keep alive
	if s.ServerAliveInterval > 0 {
		return sshSession.TerminalWithKeepAlive(s.ServerAliveInterval)
	}
	return sshSession.Terminal()

}

// get auth method
func (s *ServerConfig) authMethod() []ssh.AuthMethod {

	var ams []ssh.AuthMethod

	if s.Password != "" {
		ams = append(ams, password(s.Password))
	}

	if s.PrivateKey != "" {
		pkAuth, err := privateKeyFile(s.PrivateKey, s.PrivateKeyPassword)
		if err != nil {
			fmt.Println(err)
		} else {
			ams = append(ams, pkAuth)
		}
	}

	return ams
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
