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
	"errors"
	"os"
	"strings"

	"path"

	"io"

	"path/filepath"

	"sync"

	"fmt"

	"github.com/pkg/sftp"
)

func (s Server) fileWrite(localPath, remotePath string) error {

	fmt.Printf("Copy: %s\n", localPath)

	sshClient, err := s.sshClient()
	if err != nil {
		return err
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	localFileInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	filename := path.Base(localPath)

	// replace "~" to home path
	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		if err != nil {
			return err
		}
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}

	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		// remotePath is file and not exist
		if err == os.ErrNotExist {
			remoteFile, err := sftpClient.Create(remotePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
			err = sftpClient.Chmod(remotePath, localFileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// remotePath is dir
		if remoteFileInfo.IsDir() {
			remoteFilePath := path.Join(remotePath, filename)
			remoteFile, err := sftpClient.Create(remoteFilePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
			err = sftpClient.Chmod(remoteFilePath, localFileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			// remove remote file
			err = sftpClient.Remove(remotePath)
			if err != nil {
				return err
			}
			remoteFile, err := sftpClient.Create(remotePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
			err = sftpClient.Chmod(remotePath, localFileInfo.Mode())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s Server) directoryWrite(localPath, remotePath string) error {

	sshClient, err := s.sshClient()
	if err != nil {
		return err
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		if err != nil {
			return err
		}
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	localFileInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	// check remote path is directory
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		if err == os.ErrNotExist {
			err = sftpClient.Mkdir(remotePath)
			if err != nil {
				return err
			}
			err = sftpClient.Chmod(remotePath, localFileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			// other err
			return err
		}
	} else {
		if remoteFileInfo.IsDir() {
			remotePath = path.Join(remotePath, path.Base(localPath))
			_, err = sftpClient.Stat(remotePath)
			if err == nil {
				return errors.New(remotePath + " already exist")
			}
			err = sftpClient.Mkdir(remotePath)
			if err != nil {
				return err
			}
			err = sftpClient.Chmod(remotePath, localFileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			return errors.New("remote path is not directory")
		}
	}

	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {

		fmt.Printf("Copy: %s\n", path)

		if info == nil {
			return err
		}

		if path == localPath {
			return nil
		}

		remoteCurrentPath := filepath.Join(remotePath, strings.Replace(path, localPath, "", 1))
		_, err = sftpClient.Stat(remoteCurrentPath)

		if info.IsDir() {
			if err != nil {
				if err == os.ErrNotExist {
					err = sftpClient.Mkdir(remoteCurrentPath)
					if err != nil {
						return err
					}
					return sftpClient.Chmod(remoteCurrentPath, info.Mode())
				}
			}
		} else {
			if err != nil {
				if err == os.ErrNotExist {
					// get remote file
					remoteFile, err := sftpClient.Create(remoteCurrentPath)
					if err != nil {
						return err
					}

					// chmod
					err = sftpClient.Chmod(remoteCurrentPath, info.Mode())
					if err != nil {
						return err
					}

					// get local file
					localFile, err := os.Open(path)
					if err != nil {
						return err
					}

					// copy
					io.Copy(remoteFile, localFile)

					remoteFile.Close()
					localFile.Close()
				}
			}
		}
		return nil
	})

	return err
}

func (s Server) sftpWrite(wg *sync.WaitGroup, localPath, remotePath string) error {
	defer wg.Done()
	localFileInfo, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if localFileInfo.IsDir() {
		s.directoryWrite(localPath, remotePath)
	} else {
		s.fileWrite(localPath, remotePath)
	}
	return nil
}

func (s Server) sftpRead(localPath, remotePath string) error {
	sshClient, err := s.sshClient()
	if err != nil {
		return err
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		if err != nil {
			return err
		}
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	remoteFileInfo, err := remoteFile.Stat()
	if err != nil {
		return err
	}

	// remote path is a dir
	if remoteFileInfo.IsDir() {

		localFileInfo, err := os.Stat(localPath)
		if err != nil {
			if err == os.ErrNotExist {
				err = os.Mkdir(path.Join(localPath), remoteFileInfo.Mode())
			}
			if err != nil {
				return err
			}
		} else {
			if localFileInfo.IsDir() {
				localPath = path.Join(localPath, path.Base(remotePath))
				err = os.Mkdir(localPath, remoteFileInfo.Mode())
				if err != nil {
					return err
				}
			} else {
				return errors.New(localPath + " already exist")
			}
		}

		w := sftpClient.Walk(remotePath)
		for w.Step() {

			fmt.Printf("Copy: %s\n", w.Path())

			if w.Path() == remotePath {
				// skip
				continue
			}

			if w.Stat().IsDir() {
				err = os.Mkdir(strings.Replace(w.Path(), remotePath, localPath, -1), remoteFileInfo.Mode())
				if err != nil {
					return err
				}
			} else {
				localFile, err := os.OpenFile(strings.Replace(w.Path(), remotePath, localPath, -1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, w.Stat().Mode())
				if err != nil {
					return err
				}
				remoteTmpFile, err := sftpClient.Open(w.Path())
				if err != nil {
					return err
				}
				io.Copy(localFile, remoteTmpFile)
				localFile.Close()
				remoteTmpFile.Close()
			}
		}
		// remote path is a file
	} else {
		localFileInfo, err := os.Stat(localPath)
		if err != nil {
			if err == os.ErrNotExist {
				localFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, remoteFileInfo.Mode())
				if err != nil {
					return err
				}
				defer localFile.Close()
				io.Copy(localFile, remoteFile)
			} else {
				return err
			}
		} else {
			if localFileInfo.IsDir() {
				localFile, err := os.OpenFile(path.Join(localPath, path.Base(remotePath)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, remoteFileInfo.Mode())
				if err != nil {
					return err
				}
				defer localFile.Close()
				io.Copy(localFile, remoteFile)
			} else {
				return errors.New(localPath + " already exist")
			}
		}

	}
	return nil
}

func Copy(path1, path2 string, singleServer bool) error {

	tmpSp1 := strings.Split(path1, ":")
	tmpSp2 := strings.Split(path2, ":")

	// download file or dir
	// only support single server download
	if len(tmpSp1) == 2 && len(tmpSp2) == 1 {
		s := findServerByName(tmpSp1[0])
		if s == nil {
			return errors.New("server not found")
		} else {
			return s.sftpRead(path2, tmpSp1[1])
		}
		// upload file or dir
	} else if len(tmpSp1) == 1 && len(tmpSp2) == 2 {

		if _, err := os.Stat(path1); err != nil {
			return err
		}

		var wg sync.WaitGroup
		if singleServer {
			s := findServerByName(tmpSp2[0])
			if s == nil {
				return errors.New("server not found")
			} else {
				wg.Add(1)
				go s.sftpWrite(&wg, path1, tmpSp2[1])
			}
			wg.Wait()
		} else {
			servers := tagsMap[tmpSp2[0]]
			if len(servers) == 0 {
				return errors.New("tagged server not found")
			}
			wg.Add(len(servers))
			for _, s := range servers {
				go s.sftpWrite(&wg, path1, tmpSp2[1])
			}
			wg.Wait()
		}
	} else {
		return errors.New("command format error")
	}
	return nil
}
