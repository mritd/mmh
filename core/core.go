package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	osexec "os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/go-homedir"
	"github.com/mritd/mmh/common"
	"github.com/mritd/mmh/sshutils"
	"golang.org/x/crypto/ssh"
)

const (
	envMMH           = "MMH"
	envAPIAddressKey = "MMH_API_ADDR"

	// sshConnectTimeout is the timeout for TCP dialing and the ssh handshake
	sshConnectTimeout = 5 * time.Second
)

// wrapperClient return a standard ssh client with specific parameters set
func (s *Server) wrapperClient(secondLast bool) (*ssh.Client, error) {
	// TODO: Move "Set MaxProxy Default Value" to other func
	if currentConfig.MaxProxy == 0 {
		currentConfig.MaxProxy = 5
	}

	return s.ssh(secondLast, 0)
}

// wrapperSession returns a standard ssh session with specific parameters set
// if there is an error, the ssh session is nil
func (s *Server) wrapperSession(client *ssh.Client) (*ssh.Session, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	_ = session.Setenv(envMMH, "true")
	if s.Environment != nil {
		for k, v := range s.Environment {
			_ = session.Setenv(k, v)
		}
	}
	return session, nil
}

// ssh returns a standard ssh client, if proxy is configured, it will connect recursively
// if secondLast is true, return the second last server client
func (s *Server) ssh(secondLast bool, proxyCount int) (*ssh.Client, error) {
	addr := net.JoinHostPort(s.Address, fmt.Sprint(s.Port))
	sshConfig := &ssh.ClientConfig{
		User:            s.User,
		Auth:            s.authMethod(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         sshConnectTimeout,
	}

	if s.Proxy != "" {
		if proxyCount > currentConfig.MaxProxy {
			return nil, fmt.Errorf("too many proxy server, proxy server must be <= %d", currentConfig.MaxProxy)
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

		conn, err := proxyClient.Dial("tcp", addr)
		if err != nil {
			// 代理连接没有交给调用方，这里必须自行关闭，否则每次失败都会泄漏一条到跳板机的 ssh 连接
			_ = proxyClient.Close()
			return nil, err
		}
		client, err := newSSHClient(conn, addr, sshConfig)
		if err != nil {
			_ = proxyClient.Close()
			return nil, err
		}

		// 目标 client 关闭后一并释放跳板机连接，避免代理链上的连接只增不减
		go func() {
			_ = client.Wait()
			_ = proxyClient.Close()
		}()
		return client, nil

	} else {
		if secondLast {
			return nil, nil
		}

		// dial directly or through the socks5/http proxy configured on the server
		conn, err := s.dialSSH()
		if err != nil {
			return nil, err
		}
		return newSSHClient(conn, addr, sshConfig)
	}
}

// newSSHClient completes the ssh handshake on conn, when the handshake fails the
// connection is closed so callers don't leak it
func newSSHClient(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	if config.Timeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(config.Timeout))
	}
	ncc, channel, request, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.SetDeadline(time.Time{})
	return ssh.NewClient(ncc, channel, request), nil
}

// authMethod return ssh auth method slice
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

// privateKeyFileAuth return private key auth method
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

// passwordAuth return password auth method
func passwordAuth(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

// keyboardAuth return keyboard auth method
func keyboardAuth(authCmd string) ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		cs, args := common.ParseCommand(authCmd)
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
	sshClient, err := s.wrapperClient(false)
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	se, err := s.wrapperSession(sshClient)
	if err != nil {
		return err
	}
	// Inject mmh api address
	_ = se.Setenv(envAPIAddressKey, s.listenAPI(sshClient))

	sshSession := sshutils.NewSSHSession(se, s.HookCmd, s.HookStdout)
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

func (s *Server) listenAPI(client *ssh.Client) string {
	if s.EnableAPI != "true" {
		return "disabled"
	}

	remoteListener, err := client.Listen("tcp", "127.0.0.1:0")
	if !common.CheckErr(err) {
		return err.Error()
	}

	fmt.Printf("⚡ mmh api listen at [%s] http://%s\n", s.Name, remoteListener.Addr().String())
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("API Server Err: %v\n", r)
			}
			_ = remoteListener.Close()
		}()

		router := mux.NewRouter().StrictSlash(true)
		registerAPI(router)

		hs := http.Server{Handler: router}
		_ = hs.Serve(remoteListener)
	}()
	return "http://" + remoteListener.Addr().String()
}
