package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/mritd/mmh/pkg/sshutils"

	"sync"

	"github.com/mritd/mmh/pkg/common"

	"os"

	"context"

	"os/signal"
	"syscall"
)

// Exec batch execution of commands
func Exec(cmd, tagOrName string, single, ping bool) {
	// use context to manage goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// monitor os signal
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		switch <-sigs {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// exit all goroutine
			cancel()
		}
	}()

	// single server exec
	if single {
		server, err := findServerByName(tagOrName)
		common.CheckAndExit(err)
		err = exec(ctx, cmd, server, single, ping)
		common.PrintErr(err)
	} else {
		// multiple servers
		servers, err := findServersByTag(tagOrName)
		common.CheckAndExit(err)

		// create goroutine
		var execWg sync.WaitGroup
		execWg.Add(len(servers))
		for _, s := range servers {
			// async exec
			// because it takes time for ssh to establish a connection
			go func(s *Server) {
				defer execWg.Done()
				err = exec(ctx, cmd, s, single, false)
				common.PrintErrWithPrefix(s.Name, err)
			}(s)
		}
		execWg.Wait()
	}
}

// single server execution command
// since multiple tasks are executed async, the error is returned using channel
func exec(ctx context.Context, cmd string, s *Server, single, ping bool) error {
	// get ssh client
	sshClient, err := s.sshClient(ping)
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	// get ssh session
	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	// sshutils utils session
	sshSession := sshutils.NewSSHSession(session, s.HookCmd)
	defer func() { _ = sshSession.Close() }()
	go func() {
		select {
		case <-ctx.Done():
			_ = sshSession.Close()
			_ = sshClient.Close()
		}
	}()

	// print to stdout
	go func() {
		// wait session ready
		<-sshSession.Ready()

		// read from sshSession.Stdout and print to os.stdout
		if single {
			_, _ = io.Copy(os.Stdout, sshSession.Stdout)
		} else {
			buf := bufio.NewReader(sshSession.Stdout)
			var output bytes.Buffer
			for {
				line, err := buf.ReadString('\n')
				if err != nil {
					if err == io.EOF || err == io.ErrClosedPipe {
						break
					} else {
						common.PrintErr(err)
						break
					}
				}

				err = common.ColorOutput(&output, common.ColorLine{Prefix: s.Name, Value: line})
				if err != nil {
					common.PrintErr(err)
				} else {
					fmt.Print(output.String())
				}
				output.Reset()
			}
		}
	}()

	return sshSession.PipeExec(cmd)
}
