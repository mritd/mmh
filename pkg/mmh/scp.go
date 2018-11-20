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
	"errors"
	"strings"
	"sync"

	"github.com/mritd/mmh/pkg/utils"

	"github.com/fatih/color"

	"github.com/mritd/sshutils"
)

func Copy(args []string, singleServer bool) {
	utils.CheckAndExit(runCopy(args, singleServer))
}

func runCopy(args []string, singleServer bool) error {

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
		s := findServerByName(serverName)
		if s == nil {
			return errors.New("server not found")
		} else {
			client, err := s.sshClient()
			if err != nil {
				return err
			}
			defer client.Close()
			scpClient, err := sshutils.NewSCPClient(client)
			if err != nil {
				return err
			}
			return scpClient.CopyRemote2Local(remotePath, localPath)
		}

		// upload, eg: mcp localFile1 localFile2 localDir test:~
	} else if len(strings.Split(args[len(args)-1], ":")) == 2 {

		serverOrTag := strings.Split(args[len(args)-1], ":")[0]
		remotePath := strings.Split(args[len(args)-1], ":")[1]

		if singleServer {
			s := findServerByName(serverOrTag)
			if s == nil {
				return errors.New("server not found")
			} else {
				client, err := s.sshClient()
				if err != nil {
					return err
				}
				defer client.Close()
				scpClient, err := sshutils.NewSCPClient(client)
				if err != nil {
					return err
				}
				allArg := args[:len(args)-1]
				allArg = append(allArg, remotePath)
				return scpClient.CopyLocal2Remote(allArg...)
			}
		} else {
			servers := tagsMap[serverOrTag]
			if len(servers) == 0 {
				return errors.New("tagged server not found")
			}

			var wg sync.WaitGroup
			wg.Add(len(servers))

			for _, s := range servers {
				tmpServer := s
				go func() {
					defer wg.Done()
					client, err := tmpServer.sshClient()
					if err != nil {
						color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", tmpServer.Name, err)
						return
					}
					defer client.Close()
					scpClient, err := sshutils.NewSCPClient(client)
					if err != nil {
						color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", tmpServer.Name, err)
						return
					}

					allArg := args[:len(args)-1]
					allArg = append(allArg, remotePath)
					err = scpClient.CopyLocal2Remote(allArg...)
					if err != nil {
						color.New(color.BgRed, color.FgHiWhite).Printf("%s:  %s", tmpServer.Name, err)
						return
					}
				}()
			}

			wg.Wait()
		}

	} else {
		return errors.New("unsupported mode")
	}

	return nil
}
