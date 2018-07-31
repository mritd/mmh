package mmh

import (
	"strings"

	"os"

	"path"

	"io"

	"path/filepath"

	"sync"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/pkg/sftp"
)

func (s Server) fileWrite(localFilePath, remotePath string) {

	localFile, err := os.Open(localFilePath)
	utils.CheckAndExit(err)
	defer localFile.Close()
	_, err = localFile.Stat()
	utils.CheckAndExit(err)

	filename := path.Base(localFilePath)

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
	localFile, err := os.Open(localPath)
	utils.CheckAndExit(err)
	defer localFile.Close()
	localFileInfo, err := localFile.Stat()
	utils.CheckAndExit(err)

	if localFileInfo.IsDir() {
		s.directoryWrite(localPath, remotePath)
	} else {
		s.fileWrite(localPath, remotePath)
	}

}

func Copy(path1, path2 string, singleServer bool) {
	tmpSp1 := strings.Split(path1, ":")
	tmpSp2 := strings.Split(path2, ":")

	// download file or dir
	if len(tmpSp1) == 2 && len(tmpSp2) == 1 {
		if singleServer {
			s := findServerByName(tmpSp1[0])
			if s == nil {
				utils.Exit("Server not found", 1)
			} else {
				//remotePath := tmpSp1[1]

			}
		} else {
			initTagsGroup()
			servers := tagsMap[tmpSp1[0]]
			if len(servers) == 0 {
				utils.Exit("Tagged server not found", 1)
			}
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

	}
}
