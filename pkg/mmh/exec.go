/*
 * Copyright 2018 mritd <mritd1234@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	"github.com/mritd/mmh/pkg/utils"
	"github.com/mritd/sshutils"
)

func Exec(tagOrName, cmd string, singleServer bool) {

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

	if singleServer {
		server := findServerByName(tagOrName)
		if server == nil {
			utils.Exit("Server not found", 1)
		} else {
			var errCh = make(chan error, 1)
			exec(ctx, server, singleServer, cmd, errCh)
			select {
			case err := <-errCh:
				color.New(color.BgRed, color.FgHiWhite).Print(err)
				fmt.Println()
			default:
			}
		}
	} else {
		servers := tagsMap[tagOrName]
		if len(servers) == 0 {
			utils.Exit("Tagged server not found", 1)
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
				exec(ctx, server, singleServer, cmd, errCh)
				select {
				case err := <-errCh:
					color.New(color.BgRed, color.FgHiWhite).Printf("%server:  %server", server.Name, err)
					fmt.Println()
				default:
				}
			}()
		}
		serverWg.Wait()
	}
}

func exec(ctx context.Context, s *Server, singleServer bool, cmd string, errCh chan error) {

	sshClient, err := s.sshClient()
	if err != nil {
		errCh <- err
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		errCh <- err
		return
	}

	sshSession := sshutils.New(session)
	defer sshSession.Close()
	go sshSession.PipeExec(cmd)
	<-sshSession.ReadDone

	// copy error
	go func() {
		select {
		case err := <-sshSession.ErrCh:
			errCh <- err
		case <-ctx.Done():
			sshSession.Close()
		}
	}()

	// read from sshSession.Stdout and print to os.stdout
	if singleServer {
		io.Copy(os.Stdout, sshSession.Stdout)
	} else {
		f := getColorFuncName()
		t, err := template.New("").Funcs(ColorsFuncMap).Parse(fmt.Sprintf(`{{ .Name | %s}}{{ ":" | %s}}  {{ .Value }}`, f, f))
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
				Value: string(line),
			})
			if err != nil {
				errCh <- err
				break
			}
			fmt.Print(output.String())
		}
	}

}
