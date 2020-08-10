package sshutils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mritd/mmh/pkg/common"

	"golang.org/x/crypto/ssh/terminal"

	"golang.org/x/crypto/ssh"
)

type SSHSession struct {
	// sshutils session
	session *ssh.Session
	// for PipeExec, this channel will be read when stdin、stdout ready
	readyCh chan int
	// for interactive shell, this channel will be read when shell ready
	shellDoneCh chan int
	// shell command exit message
	exitMsg    string
	hookCmd    string
	hookStdout bool
	Stdout     io.Reader
	Stdin      io.Writer
	Stderr     io.Reader
}

func (s *SSHSession) Ready() <-chan int {
	return s.readyCh
}

// close the session
func (s *SSHSession) Close() error {
	pw, ok := s.session.Stdout.(*io.PipeWriter)
	if ok {
		err := pw.Close()
		if err != nil {
			fmt.Println(err)
		}
	}

	pr, ok := s.session.Stdin.(*io.PipeReader)
	if ok {
		err := pr.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
	return s.session.Close()
}

// update shell terminal size in background
func (s *SSHSession) updateTerminalSize() {
	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has changed.
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGWINCH)

		fd := int(os.Stdin.Fd())
		termWidth, termHeight, err := terminal.GetSize(fd)
		if err != nil {
			fmt.Println(err)
		}

		for range sigs {
			currTermWidth, currTermHeight, err := terminal.GetSize(fd)
			// Terminal size has not changed, don's do anything.
			if currTermHeight == termHeight && currTermWidth == termWidth {
				continue
			}

			// The client updated the size of the local PTY. This change needs to occur on the server side PTY as well.
			err = s.session.WindowChange(currTermHeight, currTermWidth)
			if err != nil {
				fmt.Printf("Unable to send window-change reqest: %s", err)
				continue
			}
			termWidth, termHeight = currTermWidth, currTermHeight
		}
	}()
}

func (s *SSHSession) ShellDone() <-chan int {
	return s.shellDoneCh
}

// open a interactive shell
func (s *SSHSession) Terminal() error {
	return s.TerminalWithKeepAlive(10 * time.Second)
}

// open a interactive shell with keepalive
func (s *SSHSession) TerminalWithKeepAlive(interval time.Duration) error {
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		_ = terminal.Restore(fd, state)
	}()

	// get terminal size
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	// default to xterm-256color
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	// request pty
	err = s.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	// update shell terminal size in background
	s.updateTerminalSize()

	// get pipe stdin
	s.Stdin, err = s.session.StdinPipe()
	if err != nil {
		return err
	}

	// get pipe stdout
	s.Stdout, err = s.session.StdoutPipe()
	if err != nil {
		return err
	}

	// get pipe stderr
	s.Stderr, err = s.session.StderrPipe()

	// async copy
	go func() {
		_, _ = io.Copy(os.Stderr, s.Stderr)
	}()
	go func() {
		_, _ = io.Copy(os.Stdout, s.Stdout)
	}()
	go func() {
		_, _ = io.Copy(s.Stdin, os.Stdin)
	}()

	// keepalive
	if interval > 0 {
		go func() {
			tick := time.Tick(interval)
			for range tick {
				_, err := s.session.SendRequest("keepalive@linux.com", true, nil)
				common.PrintErr(err)
			}
		}()
	}

	// stdin hook
	if s.hookCmd != "" {
		go func() {
			cs, args := common.CMD(s.hookCmd)
			cmd := exec.Command(cs, args...)
			if s.hookStdout {
				cmd.Stdin = s.Stdout
			}
			cmd.Stdout = s.Stdin
			common.PrintErr(cmd.Run())
		}()
	}

	// open shell
	err = s.session.Shell()
	if err != nil {
		return err
	}
	s.shellDoneCh <- 1

	return s.session.Wait()
}

// pipe exec
func (s *SSHSession) PipeExec(cmd string) error {
	fd := int(os.Stdin.Fd())
	termWidth, termHeight, err := terminal.GetSize(fd)
	if err != nil {
		return err
	}

	// default to xterm-256color
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	// request pty
	err = s.session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	// update shell terminal size in background
	s.updateTerminalSize()

	// write to pw
	pr, pw := io.Pipe()
	defer func() {
		_ = pw.Close()
		_ = pr.Close()
	}()

	s.session.Stdout = pw
	s.session.Stderr = pw
	s.Stdout = pr
	s.Stderr = pr

	s.readyCh <- 1

	return s.session.Run(cmd)
}

// New Session
func NewSSHSession(session *ssh.Session, hookCmd string, hookStdout bool) *SSHSession {
	return &SSHSession{
		session:     session,
		hookCmd:     hookCmd,
		hookStdout:  hookStdout,
		readyCh:     make(chan int, 1),
		shellDoneCh: make(chan int, 1),
	}
}
