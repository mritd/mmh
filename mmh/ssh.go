package mmh

import (
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"

	"github.com/mritd/mmh/utils"

	"github.com/mritd/sshutils"

	"fmt"

	"golang.org/x/crypto/ssh"
)

// return a ssh client intense point
// if secondLast is true, return the second last server
func (s *ServerConfig) sshClient(secondLast bool) (*ssh.Client, error) {

	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		Auth:            s.authMethod(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	if s.Proxy != "" {

		// check max proxy
		if s.proxyCount > CurrentContext.MaxProxy {
			return nil, errors.New(fmt.Sprintf("too many proxy server, proxy server must be <= %d", CurrentContext.MaxProxy))
		} else {
			s.proxyCount++
		}

		// find proxy server
		proxyServer, err := findServerByName(s.Proxy)
		utils.CheckAndExit(err)

		fmt.Printf("🔑 using proxy [%s], connect to => %s\n", s.Proxy, s.Name)

		// recursive connect
		proxyClient, err := proxyServer.sshClient(false)
		if err != nil {
			return nil, err
		}

		if secondLast {
			return proxyClient, nil
		}

		conn, err := proxyClient.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port))
		if err != nil {
			return nil, err
		}
		ncc, channel, request, err := ssh.NewClientConn(conn, fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		if err != nil {
			return nil, err
		}
		return ssh.NewClient(ncc, channel, request), nil

	} else {

		if secondLast {
			return nil, nil
		} else {
			return ssh.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port), sshConfig)
		}
	}
}

// authMethod return ssh auth method
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

// privateKeyFile return private key auth method
func privateKeyFile(file, password string) (ssh.AuthMethod, error) {
	if strings.HasPrefix(file, "~") {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		file = strings.Replace(file, "~", home, 1)
	}
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

// password return password auth method
func password(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

// Terminal start a ssh terminal
func (s *ServerConfig) Terminal() error {
	sshClient, err := s.sshClient(false)
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
