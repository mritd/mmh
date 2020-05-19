package core

import (
	"bufio"
	"bytes"
	"io"

	"github.com/fatih/color"

	"fmt"

	"sync"

	"os"

	"context"

	"os/signal"
	"syscall"

	"github.com/mritd/sshutils"
)

// Exec batch execution of commands
func Exec(tagOrName, cmd string, singleServer, pingClient bool) {
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
	if singleServer {
		server, err := findServerByName(tagOrName)
		checkAndExit(err)

		var errCh = make(chan error, 1)
		exec(ctx, server, singleServer, pingClient, cmd, errCh)
		select {
		case err, ok := <-errCh:
			if ok {
				_, _ = color.New(color.BgRed, color.FgHiWhite).Print(err.Error() + "\n")
			}
		}
	} else {
		// multiple servers
		servers := findServersByTag(tagOrName)
		if len(servers) == 0 {
			Exit("tagged server not found", 1)
		}

		// create goroutine
		var serverWg sync.WaitGroup
		serverWg.Add(len(servers))
		for _, s := range servers {
			// async exec
			// because it takes time for ssh to establish a connection
			go func(s *ServerConfig) {
				defer serverWg.Done()
				var errCh = make(chan error, 1)
				exec(ctx, s, singleServer, false, cmd, errCh)
				select {
				case err, ok := <-errCh:
					if ok {
						_, _ = color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s\n", s.Name, err)
					}
				}
			}(s)
		}
		serverWg.Wait()
	}
}

// single server execution command
// since multiple tasks are executed async, the error is returned using channel
func exec(ctx context.Context, s *ServerConfig, singleServer, pingClient bool, cmd string, errCh chan error) {
	// get ssh client
	sshClient, err := s.sshClient(pingClient)
	if err != nil {
		errCh <- err
		return
	}
	defer func() { _ = sshClient.Close() }()

	// get ssh session
	session, err := sshClient.NewSession()
	if err != nil {
		errCh <- err
		return
	}

	// ssh utils session
	sshSession := sshutils.NewSSHSession(session)
	defer func() { _ = sshSession.Close() }()

	// exec cmd
	go sshSession.PipeExec(cmd)

	// copy error
	var errWg sync.WaitGroup
	errWg.Add(1)
	go func() {
		// ensure that the error message is successfully output
		defer errWg.Done()
		select {
		case err, ok := <-sshSession.Error():
			if ok {
				errCh <- err
			}
		}
	}()

	// print to stdout
	go func() {
		select {
		case <-sshSession.Ready():
			// read from sshSession.Stdout and print to os.stdout
			if singleServer {
				_, _ = io.Copy(os.Stdout, sshSession.Stdout)
			} else {
				buf := bufio.NewReader(sshSession.Stdout)
				for {
					line, err := buf.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							break
						} else {
							errCh <- err
							break
						}
					}

					var output bytes.Buffer
					err = colorOutExecute(&output, ColorLine{s.Name, line})
					if err != nil {
						errCh <- err
						break
					}
					fmt.Print(output.String())
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		_ = sshClient.Close()
		close(errCh)
	case <-sshSession.Done():
		_ = sshClient.Close()
		close(errCh)
	}

	errWg.Wait()
}
