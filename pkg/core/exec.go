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
// If execGroup is true, execute the command on a group of servers
// If ping is true and a proxy is set on the server, execute the
// command on the second-to-last server
func Exec(cmd, tagOrName string, execGroup, ping bool) {
	// use context to manage goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// monitor os signal
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		// exit all goroutine
		<-sigs
		cancel()
	}()

	// single server exec
	if !execGroup {
		server, err := findServerByName(tagOrName)
		common.CheckAndExit(err)
		err = exec(ctx, cmd, server, false, ping)
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
				err = exec(ctx, cmd, s, true, false)
				common.PrintErrWithPrefix(s.Name+": 😱 ", err)
			}(s)
		}
		execWg.Wait()
	}
}

// single server execution command
// since multiple tasks are executed async, the error is returned using channel
func exec(ctx context.Context, cmd string, s *Server, colorPrint, ping bool) error {
	// get ssh client
	sshClient, err := s.wrapperClient(ping)
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	// get ssh session
	session, err := s.wrapperSession(sshClient)
	if err != nil {
		return err
	}

	// ssh utils session
	sshSession := sshutils.NewSSHSession(session, "", false)
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

		// read from wrapperSession.Stdout and print to os.stdout
		if !colorPrint {
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
