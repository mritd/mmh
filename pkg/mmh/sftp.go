package mmh

import (
	"strings"

	"os"

	"path"

	"io"

	"path/filepath"

	"errors"

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
			utils.Exit("File already exist", 1)
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
	if err != nil && err != os.ErrNotExist {
		utils.CheckAndExit(err)
	} else if err == nil && !remoteFileInfo.IsDir() {
		utils.Exit("Remote path is not directory", 1)
	}

	remoteTmpPath := remotePath
	err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			_, err := sftpClient.Stat(filepath.Join(remoteTmpPath, path))
			if err != nil {
				if err == os.ErrNotExist {
					remoteTmpPath = path
					return sftpClient.Mkdir(filepath.Join(remotePath, path))
				}
			} else {
				return errors.New("Remote directory already exist")
			}
		} else {

		}
		return nil
	})
}

func (s Server) sftpWrite(localPath, remotePath string) {
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

		if singleServer {
			s := findServerByName(tmpSp2[0])
			if s == nil {
				utils.Exit("Server not found", 1)
			} else {
				s.sftpWrite(path1, tmpSp2[1])
			}
		} else {
			initTagsGroup()
			servers := tagsMap[tmpSp1[0]]
			if len(servers) == 0 {
				utils.Exit("Tagged server not found", 1)
			}
			for _, s := range servers {
				s.sftpWrite(path1, tmpSp2[1])
			}
		}

	}
}
