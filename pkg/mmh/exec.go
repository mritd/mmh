package mmh

import (
	"io"

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

	if singleServer {
		s := findServerByName(tagOrName)
		if s == nil {
			utils.Exit("Server not found", 1)
		} else {
			exec(ctx, *s, cmd)
		}
	} else {

		// init group data
		initTagsGroup()
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
				exec(ctx, s, cmd)
			}()
		}
		serverWg.Wait()
	}
}

func exec(ctx context.Context, s Server, cmd string) {

	sshClient := s.sshClient()
	defer sshClient.Close()

	session, err := sshClient.NewSession()
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

	// write to pw
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

		f := utils.MapRandomKeyGet(ColorsFuncMap)
		t, err := template.New("").Funcs(ColorsFuncMap).Parse(fmt.Sprintf(`{{ .Name | %s}}{{ ":" | %s}}  {{ .Value }}`, f, f))
		utils.CheckAndExit(err)

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

			fmt.Print(string(utils.Render(t, struct {
				Name  string
				Value string
			}{
				Name:  s.Name,
				Value: string(line),
			})))
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
