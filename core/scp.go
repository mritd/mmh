package core

import (
	"errors"
	"strings"
	"sync"

	"github.com/fatih/color"

	"github.com/mritd/sshutils"
)

func Copy(args []string, multiServer bool) {
	checkAndExit(runCopy(args, multiServer))
}

func runCopy(args []string, multiServer bool) error {
	if len(args) < 2 {
		return errors.New("parameter invalid")
	}

	// download, eg: mcp test:~/file localPath
	// only single file/directory download is supported
	if len(strings.Split(args[0], ":")) == 2 && len(args) == 2 {
		// only single server is supported
		serverName := strings.Split(args[0], ":")[0]
		remotePath := strings.Split(args[0], ":")[1]
		localPath := args[1]
		s, err := findServerByName(serverName)
		checkAndExit(err)

		client, err := s.sshClient(false)
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		scpClient, err := sshutils.NewSCPClient(client)
		if err != nil {
			return err
		}
		return scpClient.CopyRemote2Local(remotePath, localPath)

		// upload, eg: mcp localFile1 localFile2 localDir test:~
	} else if len(strings.Split(args[len(args)-1], ":")) == 2 {
		serverOrTag := strings.Split(args[len(args)-1], ":")[0]
		remotePath := strings.Split(args[len(args)-1], ":")[1]

		// multi server copy
		if multiServer {
			// multi server copy
			servers, err := findServersByTag(serverOrTag)
			if err != nil {
				return err
			}

			var wg sync.WaitGroup
			wg.Add(len(servers))
			for _, s := range servers {
				go func(s *Server, args []string) {
					defer wg.Done()

					client, err := s.sshClient(false)
					if err != nil {
						_, _ = color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", s.Name, err)
						return
					}
					defer func() { _ = client.Close() }()

					scpClient, err := sshutils.NewSCPClient(client)
					if err != nil {
						_, _ = color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", s.Name, err)
						return
					}

					args[len(args)-1] = remotePath
					err = scpClient.CopyLocal2Remote(args...)
					if err != nil {
						_, _ = color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", s.Name, err)
						return
					}
				}(s, args)
			}
			wg.Wait()
		} else {
			s, err := findServerByName(serverOrTag)
			checkAndExit(err)
			client, err := s.sshClient(false)
			if err != nil {
				return err
			}
			defer func() { _ = client.Close() }()
			scpClient, err := sshutils.NewSCPClient(client)
			if err != nil {
				return err
			}
			args[len(args)-1] = remotePath
			return scpClient.CopyLocal2Remote(args...)
		}

	} else {
		return errors.New("unsupported mode")
	}

	return nil
}
