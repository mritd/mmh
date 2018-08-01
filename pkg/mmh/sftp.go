package mmh

import (
	"strings"

	"os"

	"path"

	"io"

	"path/filepath"

	"sync"

	"fmt"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/pkg/sftp"
)

func (s Server) fileWrite(localPath, remotePath string) {

	localFile, err := os.Open(localPath)
	utils.CheckAndExit(err)
	defer localFile.Close()
	_, err = localFile.Stat()
	utils.CheckAndExit(err)

	filename := path.Base(localPath)

	sshClient := s.sshClient()
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	utils.CheckAndExit(err)
	defer sftpClient.Close()

	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		utils.CheckAndExit(err)
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}
	remoteFileInfo, err := sftpClient.Stat(remotePath)

	if err != nil {
		// remotePath is file and not exist
		if err == os.ErrNotExist {
			remoteFile, err := sftpClient.Create(remotePath)
			utils.CheckAndExit(err)
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
		} else {
			// other err
			utils.Exit(err.Error(), 1)
		}
	} else {
		// remotePath is dir
		if remoteFileInfo.IsDir() {
			remoteFilePath := path.Join(remotePath, filename)
			remoteFile, err := sftpClient.Create(remoteFilePath)
			utils.CheckAndExit(err)
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
		} else {
			// remove remote file
			utils.CheckAndExit(sftpClient.Remove(remotePath))
			remoteFile, err := sftpClient.Create(remotePath)
			utils.CheckAndExit(err)
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
		}
	}
}

func (s Server) fileRead(localPath, remotePath string) {

}

func (s Server) directoryWrite(localPath, remotePath string) {
	sshClient := s.sshClient()
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	utils.CheckAndExit(err)
	defer sftpClient.Close()

	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		utils.CheckAndExit(err)
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}

	// check remote path is directory
	remoteFileInfo, err := sftpClient.Stat(remotePath)
	if err != nil {
		if err == os.ErrNotExist {
			utils.CheckAndExit(sftpClient.Mkdir(remotePath))
		} else {
			// other err
			utils.CheckAndExit(err)
		}
	} else if err == nil {
		if remoteFileInfo.IsDir() {
			remotePath = path.Join(remotePath, path.Base(localPath))
			_, err = sftpClient.Stat(remotePath)
			if err == nil {
				utils.Exit(remotePath+" already exist", 1)
			}
			utils.CheckAndExit(sftpClient.Mkdir(remotePath))
		} else {
			utils.Exit("Remote path is not directory", 1)
		}
	}

	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {

		if info == nil {
			return err
		}

		if path == localPath {
			return nil
		}

		remoteCurrentPath := filepath.Join(remotePath, strings.Replace(path, localPath, "", -1))
		if info.IsDir() {
			_, err := sftpClient.Stat(remoteCurrentPath)
			if err != nil {
				if err == os.ErrNotExist {
					return sftpClient.Mkdir(remoteCurrentPath)
				}
			} else {
				return err
			}
		} else {
			_, err := sftpClient.Stat(remoteCurrentPath)
			if err != nil {
				if err == os.ErrNotExist {
					// get remote file
					remoteFile, err := sftpClient.Create(remoteCurrentPath)
					if err != nil {
						return err
					}
					defer remoteFile.Close()

					// get local file
					localFile, err := os.Open(path)
					if err != nil {
						return err
					}

					// copy
					io.Copy(remoteFile, localFile)
				}
			} else {
				return err
			}
		}
		return nil
	})

	utils.CheckAndExit(err)
}

func (s Server) sftpWrite(wg *sync.WaitGroup, localPath, remotePath string) {
	defer wg.Done()
	localFileInfo, err := os.Stat(localPath)
	utils.CheckAndExit(err)

	if localFileInfo.IsDir() {
		s.directoryWrite(localPath, remotePath)
	} else {
		s.fileWrite(localPath, remotePath)
	}
}

func (s Server) sftpRead(localPath, remotePath string) {
	sshClient := s.sshClient()
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	utils.CheckAndExit(err)
	defer sftpClient.Close()

	if strings.HasPrefix(remotePath, "~") {
		pwd, err := sftpClient.Getwd()
		utils.CheckAndExit(err)
		remotePath = strings.Replace(remotePath, "~", pwd, 1)
	}

	remoteFile, err := sftpClient.Open(remotePath)
	utils.CheckAndExit(err)
	defer remoteFile.Close()

	remoteFileInfo, err := remoteFile.Stat()
	utils.CheckAndExit(err)

	// remote path is a dir
	if remoteFileInfo.IsDir() {

		localFileInfo, err := os.Stat(localPath)
		if err != nil {
			if err == os.ErrNotExist {
				err = os.Mkdir(path.Join(localPath), remoteFileInfo.Mode())
			}
			utils.CheckAndExit(err)
		} else {
			if localFileInfo.IsDir() {
				localPath = path.Join(localPath, path.Base(remotePath))
				err = os.Mkdir(localPath, remoteFileInfo.Mode())
				utils.CheckAndExit(err)
			} else {
				utils.Exit(localPath+" already exist", 1)
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
				utils.CheckAndExit(err)
			} else {
				localFile, err := os.OpenFile(strings.Replace(w.Path(), remotePath, localPath, -1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, w.Stat().Mode())
				utils.CheckAndExit(err)
				remoteTmpFile, err := sftpClient.Open(w.Path())
				utils.CheckAndExit(err)
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
				utils.CheckAndExit(err)
				defer localFile.Close()
				io.Copy(localFile, remoteFile)
			} else {
				utils.CheckAndExit(err)
			}
		} else {
			if localFileInfo.IsDir() {
				localFile, err := os.OpenFile(path.Join(localPath, path.Base(remotePath)), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, remoteFileInfo.Mode())
				utils.CheckAndExit(err)
				defer localFile.Close()
				io.Copy(localFile, remoteFile)
			} else {
				utils.Exit(localPath+" already exist", 1)
			}
		}

	}
}

func Copy(path1, path2 string, singleServer bool) {
	tmpSp1 := strings.Split(path1, ":")
	tmpSp2 := strings.Split(path2, ":")

	// download file or dir
	// only support single server download
	if len(tmpSp1) == 2 && len(tmpSp2) == 1 {
		s := findServerByName(tmpSp1[0])
		if s == nil {
			utils.Exit("Server not found", 1)
		} else {
			s.sftpRead(path2, tmpSp1[1])
		}
		// upload file or dir
	} else if len(tmpSp1) == 1 && len(tmpSp2) == 2 {

		_, err := os.Stat(path1)
		utils.CheckAndExit(err)

		var wg sync.WaitGroup
		if singleServer {
			s := findServerByName(tmpSp2[0])
			if s == nil {
				utils.Exit("Server not found", 1)
			} else {
				wg.Add(1)
				go s.sftpWrite(&wg, path1, tmpSp2[1])
			}
			wg.Wait()
		} else {
			initTagsGroup()
			servers := tagsMap[tmpSp2[0]]
			if len(servers) == 0 {
				utils.Exit("Tagged server not found", 1)
			}
			wg.Add(len(servers))
			for _, s := range servers {
				go s.sftpWrite(&wg, path1, tmpSp2[1])
			}
			wg.Wait()
		}
	} else {
		utils.Exit("Command format error", 1)
	}
}
