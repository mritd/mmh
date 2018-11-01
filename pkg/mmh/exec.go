/*
 * Copyright 2018 mritd <mritd1234@gmail.com>.
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
	"bytes"
	"io"

	"github.com/fatih/color"

	"bufio"
	"fmt"

	"sync"

	"os"

	"context"

	"os/signal"
	"syscall"

	"text/template"

	"github.com/mritd/mmh/pkg/utils"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
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
		s := findServerByName(tagOrName)
		if s == nil {
			utils.Exit("Server not found", 1)
		} else {
			var errCh = make(chan error, 1)
			exec(ctx, s, cmd, errCh)
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
		for _, server := range servers {
			s := server
			// async exec
			// because it takes time for ssh to establish a connection
			go func() {
				defer serverWg.Done()
				var errCh = make(chan error, 1)
				exec(ctx, s, cmd, errCh)
				select {
				case err := <-errCh:
					color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", s.Name, err)
					fmt.Println()
				default:
				}
			}()
		}
		serverWg.Wait()
	}
}

func exec(ctx context.Context, s *Server, cmd string, errCh chan error) {

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
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	fd := int(os.Stdin.Fd())
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		errCh <- err
		return
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	err = session.RequestPty(termType, termHeight, termWidth, modes)
	if err != nil {
		errCh <- err
		return
	}

	// write to pw
	pr, pw := io.Pipe()
	session.Stdout = pw
	session.Stderr = pw

	var execWg sync.WaitGroup
	execWg.Add(2)

	// if cancel, close all
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

		f := getColorFuncName()
		t, err := template.New("").Funcs(ColorsFuncMap).Parse(fmt.Sprintf(`{{ .Name | %s}}{{ ":" | %s}}  {{ .Value }}`, f, f))
		if err != nil {
			errCh <- err
			return
		}

		buf := bufio.NewReader(pr)
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
	}()

	// exec and write to pw
	go func() {
		defer func() {
			pw.Close()
			execWg.Done()
		}()
		err := session.Run(cmd)
		if err != nil {
			errCh <- err
			return
		}
	}()

	execWg.Wait()
}
