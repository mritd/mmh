package core

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	osexec "os/exec"

	"github.com/mritd/mmh/pkg/sshutils"

	"golang.org/x/crypto/ssh"

	"github.com/mitchellh/go-homedir"

	"github.com/mritd/mmh/pkg/common"

	"fmt"
)

func (s *Server) sshClient(secondLast bool) (*ssh.Client, error) {
	return s.ssh(secondLast, 0)
}

// return a ssh client intense point
// if secondLast is true, return the second last server
func (s *Server) ssh(secondLast bool, proxyCount int) (*ssh.Client, error) {
	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		Auth:            s.authMethod(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	if s.Proxy != "" {
		if proxyCount > currentConfig.MaxProxy {
			return nil, errors.New(fmt.Sprintf("too many proxy server, proxy server must be <= %d", currentConfig.MaxProxy))
		} else {
			proxyCount++
		}

		// find proxy server
		proxyServer, err := findServerByName(s.Proxy)
		common.CheckAndExit(err)

		fmt.Printf("🔑 using proxy [%s], connect to => %s\n", s.Proxy, s.Name)
		// recursive connect
		proxyClient, err := proxyServer.ssh(false, proxyCount)
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
func (s *Server) authMethod() []ssh.AuthMethod {
	var ams []ssh.AuthMethod

	if s.Password != "" {
		ams = append(ams, passwordAuth(s.Password))
	}

	if s.PrivateKey != "" {
		pkAuth, err := privateKeyFileAuth(s.PrivateKey, s.PrivateKeyPassword)
		if err != nil {
			common.PrintErr(err)
		} else {
			ams = append(ams, pkAuth)
		}
	}

	if s.KeyboardAuthCmd != "" {
		ams = append(ams, keyboardAuth(s.KeyboardAuthCmd))
	}

	return ams
}

// privateKeyFile return private key auth method
func privateKeyFileAuth(file, password string) (ssh.AuthMethod, error) {
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
func passwordAuth(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

func keyboardAuth(authCmd string) ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		cs, args := common.CMD(authCmd)
		cmd := osexec.Command(cs, args...)
		reqBs, err := jsoniter.Marshal(KeyBoardRequest{
			User:        user,
			Instruction: instruction,
			Questions:   questions,
			Echos:       echos,
		})
		if err != nil {
			return nil, err
		}
		cmd.Stdin = bytes.NewReader(reqBs)
		respBs, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}
		return strings.Split(strings.TrimSpace(string(respBs)), "\n"), nil
	})
}

// Terminal start a ssh terminal
func (s *Server) Terminal() error {
	sshClient, err := s.sshClient(false)
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	sshSession := sshutils.NewSSHSession(session, s.HookCmd, s.HookStdout)
	defer func() { _ = sshSession.Close() }()

	// keep alive
	if s.ServerAliveInterval > 0 {
		if s.ServerAliveInterval < 10*time.Second {
			fmt.Println("WARN: ServerAliveInterval set too small heartbeat time, use the default value of 10s(please set a larger value, such as \"30s\", \"5m\")")
			s.ServerAliveInterval = 10 * time.Second
		}
		return sshSession.TerminalWithKeepAlive(s.ServerAliveInterval)
	}

	return sshSession.Terminal()
}
