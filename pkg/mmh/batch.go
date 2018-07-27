package mmh

import (
	"net"

	"io"

	"bufio"
	"fmt"

	"sync"

	"os"

	"context"

	"os/signal"
	"syscall"

	"time"

	"github.com/mritd/mmh/pkg/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func Batch(tag, cmd string) {

	// init group data
	initTagsGroup()
	servers := tagsMap[tag]
	if len(servers) == 0 {
		utils.Exit("Tagged server not found!", 1)
	}

	// use context to manage goroutine
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	// monitor os signal
	go func() {
		switch <-c {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// exit all goroutine
			cancel()
		}
	}()

	// create goroutine
	var serverWg sync.WaitGroup
	serverWg.Add(len(servers))
	for _, server := range servers {
		s := server
		// async exec
		// because it takes time for ssh to establish a connection
		go func() {
			defer serverWg.Done()
			exec(ctx, s, cmd)
		}()
	}
	serverWg.Wait()
}

func exec(ctx context.Context, s Server, cmd string) {
	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.authMethod(),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// ignore host key
			return nil
		},
		Timeout: time.Second * 5,
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

	var execDone = make(chan int)
	defer close(execDone)

	go func() {
		select {
		case <-ctx.Done():
			session.Close()
		}
	}()

	// read from pr and print to stdout
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

	// exec and write to pw
	go func() {
		defer func() {
			pw.Close()
			execWg.Done()
		}()
		session.Run(cmd)
	}()

	execWg.Wait()
}
