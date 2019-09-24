package mmh

import (
	"bufio"
	"bytes"
	"io"
	"text/template"

	"github.com/fatih/color"

	"fmt"

	"sync"

	"os"

	"context"

	"os/signal"
	"syscall"

	"github.com/mritd/mmh/utils"
	"github.com/mritd/sshutils"
)

// batch execution of commands
func Exec(tagOrName, cmd string, singleServer, pingClient bool) {

	// use context to manage goroutine
	ctx, cancel := context.WithCancel(context.Background())

	// monitor os signal
	cancelChannel := make(chan os.Signal)
	signal.Notify(cancelChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		switch <-cancelChannel {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			// exit all goroutine
			cancel()
		}
	}()

	// single server exec
	if singleServer {
		server, err := findServerByName(tagOrName)
		utils.CheckAndExit(err)

		var errCh = make(chan error, 1)
		exec(ctx, server, singleServer, pingClient, cmd, errCh)
		select {
		case err := <-errCh:
			_, _ = color.New(color.BgRed, color.FgHiWhite).Print(err)
			fmt.Println()
		default:
		}
	} else {
		// multiple servers
		servers := findServersByTag(tagOrName)
		if len(servers) == 0 {
			utils.Exit("tagged server not found", 1)
		}

		// create goroutine
		var serverWg sync.WaitGroup
		serverWg.Add(len(servers))
		for _, tmpServer := range servers {
			server := tmpServer
			// async exec
			// because it takes time for ssh to establish a connection
			go func() {
				defer serverWg.Done()
				var errCh = make(chan error, 1)
				exec(ctx, server, singleServer, false, cmd, errCh)
				select {
				case err := <-errCh:
					_, _ = color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", server.Name, err)
					fmt.Println()
				default:
				}
			}()
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
	defer func() {
		_ = sshClient.Close()
	}()

	// get ssh session
	session, err := sshClient.NewSession()
	if err != nil {
		errCh <- err
		return
	}

	// ssh utils session
	sshSession := sshutils.NewSSHSession(session)
	defer func() {
		_ = sshSession.Close()
	}()

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
				f := utils.GetColorFuncName()
				t, err := template.New("").Funcs(utils.ColorsFuncMap).Parse(fmt.Sprintf(`{{ .Name | %s}}{{ ":" | %s}}  {{ .Value }}`, f, f))
				if err != nil {
					errCh <- err
					return
				}

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
					err = t.Execute(&output, struct {
						Name  string
						Value string
					}{
						Name:  s.Name,
						Value: line,
					})
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
	case <-sshSession.Done():
		_ = sshClient.Close()
	}

	errWg.Wait()

}
