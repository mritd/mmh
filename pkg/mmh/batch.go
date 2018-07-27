package mmh

import (
	"net"

	"io"

	"bufio"
	"fmt"

	"sync"

	"os"

	"github.com/mritd/mmh/pkg/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func Batch(tag, cmd string) {

	initTagsGroup()

	var serverWg sync.WaitGroup

	servers := tagsMap[tag]
	if len(servers) == 0 {
		utils.Exit("Tagged server not found!", 1)
	}

	serverWg.Add(len(servers))

	for _, server := range servers {
		s := server
		go func() {
			defer serverWg.Done()
			exec(s, cmd)
		}()
	}
	serverWg.Wait()
}

func exec(s Server, cmd string) {
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

	termWidth, termHeight, err := terminal.GetSize(fd)
	utils.CheckAndExit(err)

	// only xterm-256color support
	err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
	utils.CheckAndExit(err)

	pr, pw := io.Pipe()
	session.Stdout = pw
	session.Stderr = pw

	var execWg sync.WaitGroup
	execWg.Add(2)

	go func() {

		defer func() {
			pr.Close()
			execWg.Done()
		}()

		buf := bufio.NewReader(pr)
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				} else {
					panic(err)
				}
			}
			fmt.Printf("%s:  %s", s.Name, string(line))
		}
	}()

	go func() {
		defer func() {
			pw.Close()
			execWg.Done()
		}()
		session.Run(cmd)
	}()

	execWg.Wait()
}
