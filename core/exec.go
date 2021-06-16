package core

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mritd/mmh/common"
	"github.com/mritd/mmh/sshutils"
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
		err = exec(ctx, cmd, server, ping)
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
				err = exec(ctx, cmd, s, false)
				common.PrintErrWithPrefix(s.Name+": 😱 ", err)
			}(s)
		}
		execWg.Wait()
	}
}

// exec executes commands on a single server
func exec(ctx context.Context, cmd string, s *Server, ping bool) error {
	// get ssh client
	sshClient, err := s.wrapperClient(ping)
	if err != nil {
		return err
	}
	defer func() { _ = sshClient.Close() }()

	// get ssh session
	se, err := s.wrapperSession(sshClient)
	if err != nil {
		return err
	}

	// ssh utils session
	sshSession := sshutils.NewSSHSession(se, s.HookCmd, s.HookStdout)
	defer func() { _ = sshSession.Close() }()
	go func() {
		select {
		case <-ctx.Done():
			_ = sshSession.Close()
			_ = sshClient.Close()
		}
	}()

	return sshSession.PipeExec(cmd, func(r io.Reader, w io.Writer) {
		common.Converted2Rendered(r, w, s.Name+":")
	})
}
