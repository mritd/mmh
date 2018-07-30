package mmh

import (
	"fmt"
	"strings"

	"os"

	"path"

	"io"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func (s Server) sftpClient1() *sftp.Client {
	sshConfig := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.authMethod(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	connection, err := ssh.Dial("tcp", fmt.Sprint(s.Address, ":", s.Port), sshConfig)
	utils.CheckAndExit(err)
	defer connection.Close()

	c, err := sftp.NewClient(connection)
	utils.CheckAndExit(err)
	return c
}

func (s Server) sftpWrite(localFilePath, remotePath string) {

	localFile, err := os.Open(localFilePath)
	utils.CheckAndExit(err)
	defer localFile.Close()
	info, err := localFile.Stat()
	utils.CheckAndExit(err)

	if !info.IsDir() {
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

	} else {

	}

}

func Copy(path1, path2 string, singleServer bool) {
	tmpSp1 := strings.Split(path1, ":")
	tmpSp2 := strings.Split(path2, ":")
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
