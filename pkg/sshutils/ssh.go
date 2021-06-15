package sshutils

import (
	"errors"
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
	// se is a standard ssh session
	se *ssh.Session
	// shell command exit message
	exitMsg    string
	hookCmd    string
	hookStdout bool

	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
}

// Close close the session
func (s *SSHSession) Close() error {
	pw, ok := s.se.Stdout.(*io.PipeWriter)
	if ok {
		if err := pw.Close(); err != nil {
			fmt.Println(err)
		}
	}

	pr, ok := s.se.Stdin.(*io.PipeReader)
	if ok {
		if err := pr.Close(); err != nil {
			fmt.Println(err)
		}
	}
	return s.se.Close()
}

// updateTerminalSize update terminal size in background
func (s *SSHSession) updateTerminalSize() {
	go func() {
		// SIGWINCH is sent to the process when the window size of the terminal has changed.
		sigs := make(chan os.Signal)
		signal.Notify(sigs, syscall.SIGWINCH)

		fd := int(os.Stdin.Fd())
		termWidth, termHeight, err := terminal.GetSize(fd)
		if !common.CheckErr(err) {
			return
		}

		for range sigs {
			currTermWidth, currTermHeight, err := terminal.GetSize(fd)
			if !common.CheckErr(err) {
				continue
			}

			// Terminal size has not changed, don's do anything.
			if currTermHeight == termHeight && currTermWidth == termWidth {
				continue
			}

			// The client updated the size of the local PTY. This change needs to occur on the server side PTY as well.
			err = s.se.WindowChange(currTermHeight, currTermWidth)
			if err != nil {
				fmt.Printf("Unable to send window-change reqest: %s", err)
				continue
			}
			termWidth, termHeight = currTermWidth, currTermHeight
		}
	}()
}

// Terminal open a interactive terminal shell
func (s *SSHSession) Terminal() error {
	return s.TerminalWithKeepAlive(5 * time.Second)
}

// TerminalWithKeepAlive open a interactive terminal shell with keepalive
func (s *SSHSession) TerminalWithKeepAlive(interval time.Duration) error {
	if interval < 3*time.Second {
		return errors.New("the interval must be >= 3s")
	}

	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { _ = terminal.Restore(fd, state) }()

	// request pty
	err = s.requestPty(fd)
	if err != nil {
		return err
	}

	// update terminal size in background
	s.updateTerminalSize()

	// get pipe stdin
	s.Stdin, err = s.se.StdinPipe()
	if err != nil {
		return err
	}

	// get pipe stdout
	s.Stdout, err = s.se.StdoutPipe()
	if err != nil {
		return err
	}

	// get pipe stderr
	s.Stderr, err = s.se.StderrPipe()

	// async copy
	go func() { _, _ = io.Copy(os.Stderr, s.Stderr) }()
	go func() { _, _ = io.Copy(os.Stdout, s.Stdout) }()
	go func() { _, _ = io.Copy(s.Stdin, os.Stdin) }()

	// keepalive
	go func() {
		tick := time.Tick(interval)
		for range tick {
			_, err := s.se.SendRequest("ssh-keepalive@linux.com", true, nil)
			common.PrintErr(err)
		}
	}()

	// stdin hook
	if s.hookCmd != "" {
		go func() {
			cs, args := common.ParseCommand(s.hookCmd)
			cmd := exec.Command(cs, args...)
			if s.hookStdout {
				cmd.Stdin = s.Stdout
			}
			cmd.Stdout = s.Stdin
			common.PrintErr(cmd.Run())
		}()
	}

	// open shell
	if err := s.se.Shell(); err != nil {
		return err
	}
	return s.se.Wait()
}

// requestPty calls the RequestPty method of the standard ssh session, and the terminal width
// and other information are automatically set by default.
func (s *SSHSession) requestPty(fd int) error {
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
	return s.se.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
}

// PipeExec exec remote commands and provides pipeline output reading
func (s *SSHSession) PipeExec(cmd string, printFn func(r io.Reader, w io.Writer)) error {
	// request pty
	err := s.requestPty(int(os.Stdin.Fd()))
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

	s.se.Stdout = pw
	s.se.Stderr = pw
	s.Stdout = pr
	s.Stderr = pr

	go func() { printFn(s.Stdout, os.Stdout) }()

	return s.se.Run(cmd)
}

// NewSSHSession return a pointer to SSHSession
func NewSSHSession(se *ssh.Session, hookCmd string, hookStdout bool) *SSHSession {
	return &SSHSession{
		se:         se,
		hookCmd:    hookCmd,
		hookStdout: hookStdout,
	}
}
