package mmh

import (
	"io/ioutil"
	"net"

	"os"

	"strings"

	"fmt"

	"github.com/mritd/mmh/pkg/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Server struct {
	User      string
	Password  string
	PublicKey string
	Address   string
	Port      int
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
	if strings.TrimSpace(s.PublicKey) != "" {
		return publicKeyFile(s.PublicKey)
	} else {
		return password(s.Password)
	}
}

func (s Server) Connect() {
	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.authMethod(),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// ignore host key
			return nil
		},
	}

	connection, err := ssh.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port), sshConfig)
	utils.CheckAndExit(err)

	session, err := connection.NewSession()
	utils.CheckAndExit(err)
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	utils.CheckAndExit(err)
	defer terminal.Restore(fd, state)

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	termWidth, termHeight, err := terminal.GetSize(fd)
	utils.CheckAndExit(err)

	err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
	utils.CheckAndExit(err)

	err = session.Shell()
	utils.CheckAndExit(err)

	session.Wait()
}
