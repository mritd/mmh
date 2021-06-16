package sshutils

import (
	"errors"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
)

type scpClient struct {
	sftpClient *sftp.Client
}

func (s *scpClient) CopyLocalFile2Remote(localFilePath, remotePath string) error {
	localFilePath = s.replaceHome(localFilePath, true)
	remotePath = s.replaceHome(remotePath, false)

	localFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = localFile.Close()
	}()

	localFileInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	remoteFileInfo, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		// remotePath is file and not exist
		if err != os.ErrNotExist {
			return err
		}
	} else {
		// remotePath is dir
		if remoteFileInfo.IsDir() {
			// merge path
			filename := path.Base(localFilePath)
			remotePath = path.Join(remotePath, filename)
		} else { // remotePath is file
			// remove remote file
			err = s.sftpClient.Remove(remotePath)
			if err != nil {
				return err
			}
		}
	}

	// create remote file
	remoteFile, err := s.sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = remoteFile.Close()
	}()

	// copy local file to remote
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return err
	}

	// chmod
	err = s.sftpClient.Chmod(remotePath, localFileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func (s *scpClient) CopyLocalDir2Remote(localDirPath, remotePath string) error {

	localDirPath = s.replaceHome(localDirPath, true)
	remotePath = s.replaceHome(remotePath, false)

	localDir, err := os.Open(localDirPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = localDir.Close()
	}()

	localDirInfo, err := localDir.Stat()
	if err != nil {
		return err
	}

	// check remote path is directory
	remoteInfo, err := s.sftpClient.Stat(remotePath)
	if err != nil {
		// remote dir not exist
		if err == os.ErrNotExist {
			// create remote dir
			err = s.sftpClient.Mkdir(remotePath)
			if err != nil {
				return err
			}
			// chmod
			err = s.sftpClient.Chmod(remotePath, localDirInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			// other err
			return err
		}
	} else {
		// ensure remotePath is a directory, because copy a local directory
		// to a remote file makes no sense
		if remoteInfo.IsDir() {
			remotePath = path.Join(remotePath, path.Base(localDirPath))
			// if remotePath already exist, return error
			_, err = s.sftpClient.Stat(remotePath)
			if err == nil {
				return errors.New(remotePath + " already exist")
			}
			// create remotePath
			err = s.sftpClient.Mkdir(remotePath)
			if err != nil {
				return err
			}
			// chmod
			err = s.sftpClient.Chmod(remotePath, localDirInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			return errors.New("remote path is not directory")
		}
	}

	err = filepath.Walk(localDirPath, func(path string, info os.FileInfo, err error) error {

		if info == nil {
			return err
		}

		// skip
		if path == localDirPath {
			return nil
		}

		// remote relative path
		remoteRelativePath := strings.Replace(path, localDirPath, "", 1)
		// remote absolute path
		remoteAbsolutePath := filepath.Join(remotePath, remoteRelativePath)

		// check remote path state
		_, err = s.sftpClient.Stat(remoteAbsolutePath)
		if err != nil {
			// remote path not exist
			if err == os.ErrNotExist {
				// if the local path is dir, we will create a remote directory
				// with the same name as the local path
				if info.IsDir() {
					err = s.sftpClient.Mkdir(remoteAbsolutePath)
					if err != nil {
						return err
					}
					return s.sftpClient.Chmod(remoteAbsolutePath, info.Mode())
				} else {
					// if the local path is file, we will create a remote file
					// with the same name as the local file
					remoteFile, err := s.sftpClient.Create(remoteAbsolutePath)
					if err != nil {
						return err
					}

					// open local file
					localFile, err := os.Open(path)
					if err != nil {
						return err
					}
					defer func() {
						_ = localFile.Close()
					}()

					// copy
					_, err = io.Copy(remoteFile, localFile)
					if err != nil {
						return err
					}

					// chmod
					err = s.sftpClient.Chmod(remoteAbsolutePath, info.Mode())
					if err != nil {
						return err
					}

					_ = remoteFile.Close()
					_ = localFile.Close()

				}
			} else {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *scpClient) CopyLocal2Remote(paths ...string) error {

	if len(paths) < 2 {
		return errors.New("parameter invalid")
	}

	remotePath := paths[len(paths)-1]
	remotePath = s.replaceHome(remotePath, false)

	if len(paths) > 2 {
		remoteFileInfo, err := s.sftpClient.Stat(remotePath)
		if err != nil {
			return err
		}
		if !remoteFileInfo.IsDir() {
			return errors.New("remote path must a directory")
		}
	}

	for _, localPath := range paths {
		if localPath == paths[len(paths)-1] {
			break
		}
		localAbsolutePath := s.replaceHome(localPath, true)
		info, err := os.Stat(localAbsolutePath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = s.CopyLocalDir2Remote(localAbsolutePath, remotePath)
		} else {
			err = s.CopyLocalFile2Remote(localAbsolutePath, remotePath)
		}
		if err != nil {
			return err
		}
	}
	return nil

}

func (s *scpClient) CopyRemote2Local(remotePath, localPath string) error {

	localPath = s.replaceHome(localPath, true)
	remotePath = s.replaceHome(remotePath, false)

	// get local file info
	localFileInfo, localFileErr := os.Stat(localPath)

	// get remote file
	remoteFile, err := s.sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = remoteFile.Close()
	}()

	// get remote file info
	remoteFileInfo, err := remoteFile.Stat()
	if err != nil {
		return err
	}

	// remote path is a dir
	if remoteFileInfo.IsDir() {
		// if local dir not exist, we will create a local dir
		// with the same name as the remote dir
		if localFileErr != nil {
			if os.IsNotExist(localFileErr) {
				err = os.MkdirAll(path.Join(localPath), remoteFileInfo.Mode())
				if err != nil {
					return err
				}
			}
		} else {
			// if local dir exist, we will merge local dir path and remote dir relative path
			if localFileInfo.IsDir() {
				// create local dir
				localPath = path.Join(localPath, path.Base(remotePath))
				err = os.MkdirAll(localPath, remoteFileInfo.Mode())
				if err != nil {
					return err
				}
			} else {
				return errors.New(localPath + " already exist")
			}
		}

		w := s.sftpClient.Walk(remotePath)
		for w.Step() {

			// skip
			if w.Path() == remotePath {
				continue
			}

			// if remote path is a dir, we will create a local dir with the same
			// name as the remote dir
			if w.Stat().IsDir() {
				err = os.Mkdir(strings.Replace(w.Path(), remotePath, localPath, -1), remoteFileInfo.Mode())
				if err != nil {
					return err
				}
			} else {
				// if remote path is a file, copy it
				localFile, err := os.OpenFile(strings.Replace(w.Path(), remotePath, localPath, 1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, w.Stat().Mode())
				if err != nil {
					return err
				}
				remoteTmpFile, err := s.sftpClient.Open(w.Path())
				if err != nil {
					return err
				}
				_, err = io.Copy(localFile, remoteTmpFile)
				if err != nil {
					return err
				}
				_ = localFile.Close()
				_ = remoteTmpFile.Close()
			}
		}

	} else {
		if localFileErr != nil {
			// if remote path is a file and local file not exist, we will create a local
			// file with the same name as the remote file
			if os.IsNotExist(localFileErr) {
				localFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, remoteFileInfo.Mode())
				if err != nil {
					return err
				}
				defer func() {
					_ = localFile.Close()
				}()

				// copy remote file to local file
				_, err = io.Copy(localFile, remoteFile)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {

			// if remote path is a file and local path is a dir, merge remote file name
			// to local path
			if localFileInfo.IsDir() {
				localPath = path.Join(localPath, path.Base(remotePath))
			} else {
				// if local file already exist, remove it
				err = os.Remove(localPath)
				if err != nil {
					return err
				}
			}

			localFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, remoteFileInfo.Mode())
			if err != nil {
				return err
			}
			defer func() {
				_ = localFile.Close()
			}()
			_, err = io.Copy(localFile, remoteFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// replace "~" to home path
func (s *scpClient) replaceHome(path string, isLocal bool) string {

	if strings.HasPrefix(path, "~") {

		var home string
		var err error

		if isLocal {
			home, err = homedir.Dir()
			if err != nil {
				return path
			}
		} else {
			home, err = s.sftpClient.Getwd()
			if err != nil {
				return path
			}
		}

		return strings.Replace(path, "~", home, 1)

	}
	return path
}

func NewSCPClient(client *ssh.Client) (*scpClient, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	return &scpClient{
		sftpClient: sftpClient,
	}, nil
}
