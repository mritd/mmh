package mmh

import (
	"fmt"

	"os"

	"path"

	"io"

	"github.com/mritd/mmh/pkg/utils"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func (s Server) sftpClient() *sftp.Client {
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
		client := s.sftpClient()
		defer client.Close()
		remoteFileInfo, err := client.Stat(remotePath)
		utils.CheckAndExit(err)
		if remoteFileInfo.IsDir() {
			remoteFilePath := path.Join(remotePath, filename)
			remoteFile, err := client.Create(remoteFilePath)
			utils.CheckAndExit(err)
			defer remoteFile.Close()
			io.Copy(remoteFile, localFile)
		} else {

		}
	} else {

	}

}
