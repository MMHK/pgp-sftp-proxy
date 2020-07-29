package lib

import (
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

type SSHItem struct {
	Host       string `json:"host"`
	Username   string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"key"`
}

type Storage interface{
	Put(local string, remote string) (error)
	PutStream(local string, reader io.Reader) (error)
	Get(local string, remote string) (error)
	Remove(remote string) (error)
	GetFiles(remoteDir string) ([]string, error)
}

type SSHClient struct {
	config     *SSHItem
	ssh_client *ssh.Client
}

type RemoteNode struct {
	Info os.FileInfo
	FullPath string
}

func NewStorage(conf *SSHItem) (Storage) {
	return NewSSHClient(conf)
}

func (this *SSHClient) getSftpClient(callback func(client *sftp.Client) (error)) (err error) {
	session, err := this.Connect()
	if err != nil {
		log.Error(err)
		return
	}
	defer session.Close()

	sftpClient, err := sftp.NewClient(this.ssh_client)
	if err != nil {
		log.Error(err)
		return
	}
	defer sftpClient.Close()

	return callback(sftpClient)
}

func NewSSHClient(conf *SSHItem) *SSHClient {
	return &SSHClient{
		config: conf,
	}
}

func (this *SSHClient) GetFiles(remoteDir string) (list []string, err error) {
	list = make([]string, 0)
	err = this.getSftpClient(func(client *sftp.Client) error {
		dir := filepath.ToSlash(remoteDir)

		info, err := client.Stat(dir)
		if err != nil {
			log.Error(err)
			return err
		}
		if !info.IsDir() {
			return errors.New("remote path is not a directory")
		}

		sub := readDirNodes(client, dir)
		for _, item := range sub {
			if !item.Info.IsDir() {
				list = append(list, item.FullPath)
			}
		}

		return nil
	})

	return list, err
}

func (this *SSHClient) Get(local string, remote string) (err error) {
	return this.getSftpClient(func(client *sftp.Client) error {

		info, err := client.Stat(remote)
		if err != nil {
			log.Error(err)
			return err
		} else if info.IsDir() {
			return errors.New(fmt.Sprintf("remote path:[%s] is directory", remote))
		}

		remoteFile, err := client.Open(remote)
		if err != nil {
			log.Error(err)
			return err
		}
		defer remoteFile.Close()

		if _, err := os.Stat(local); os.IsExist(err) {
			err = os.Remove(local)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		localFile, err := os.Create(local)
		if err != nil {
			log.Error(err)
			return err
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, remoteFile)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})

}

func readDirNodes(client *sftp.Client, remoteDir string) (list []*RemoteNode)  {
	list = make([]*RemoteNode, 0)
	dir := filepath.ToSlash(remoteDir)

	fileList, err := client.ReadDir(dir)
	if err != nil {
		log.Error(err)
		return nil
	}

	for _, info := range fileList {
		fullPath := filepath.ToSlash(filepath.Join(dir, info.Name()))
		if info.IsDir() {
			sub := readDirNodes(client, fullPath)
			if err != nil {
				log.Error(err)
				continue
			}
			list = append(list, sub...)
		}

		list = append(list, &RemoteNode{
			Info:     info,
			FullPath: fullPath,
		})
	}

	return list
}

func (this *SSHClient) Remove(remote string) (error)  {
	return this.getSftpClient(func(client *sftp.Client) error {
		info, err := client.Stat(remote)
		if err != nil {
			log.Error(err)
			return err
		}

		if info.IsDir() {
			list := readDirNodes(client, remote)
			for _, item := range list {
				if !item.Info.IsDir() {
					client.Remove(item.FullPath)
				}
			}
			for _, item := range list {
				if item.Info.IsDir() {
					client.RemoveDirectory(item.FullPath)
				}
			}
		}

		err = client.Remove(remote)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
}

func (c *SSHClient) getKeyFile(filename string) (key ssh.Signer, err error) {
	buffer, err1 := ioutil.ReadFile(filename)
	if err1 != nil {
		err = err1
		log.Error(err1)
		return
	}
	key, err = ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Error(err)
	}
	return
}

func (this *SSHClient) Session(callback func(*ssh.Session) error) error {
	session, err := this.Connect()
	if err != nil {
		return err
	}
	defer session.Close()

	return callback(session)
}

func (c *SSHClient) Connect() (session *(ssh.Session), err error) {
	if c.ssh_client == nil {
		authMethods := make([]ssh.AuthMethod, 0)
		if len(c.config.PrivateKey) > 0 {
			key, err := c.getKeyFile(c.config.PrivateKey)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			authMethods = append(authMethods, ssh.PublicKeys(key))
		}
		authMethods = append(authMethods, ssh.Password(c.config.Password))
		config := &ssh.ClientConfig{
			User: c.config.Username,
			Auth: authMethods,
			//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

		client, err := ssh.Dial("tcp", c.config.Host, config)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		c.ssh_client = client
	}

	session, err = c.ssh_client.NewSession()
	if err != nil {
		log.Error(err)
		return
	}
	return session, nil
}

func (this *SSHClient) Put(localFilePath string, remoteFilePath string) error {
	return this.getSftpClient(func(sftpClient *sftp.Client) error {
		remoteDir := filepath.ToSlash(filepath.Dir(remoteFilePath));
		if _, err := sftpClient.Stat(remoteDir); err != nil && os.IsNotExist(err) {
			log.Info(remoteDir)

			err = sftpClient.MkdirAll(remoteDir)
			if err != nil {
				log.Error(err)
				return err
			}
		}
		remoteFile, err := sftpClient.Create(filepath.ToSlash(remoteFilePath))
		if err != nil {
			log.Error(err)
			return err
		}
		defer remoteFile.Close()


		localFile, err := os.Open(localFilePath)
		if err != nil {
			log.Error(err)
			return err
		}
		defer localFile.Close()

		_, err = io.Copy(remoteFile, localFile)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
}

func (this *SSHClient) PutStream(remoteFilePath string, reader io.Reader) (err error) {
	return this.getSftpClient(func(sftpClient *sftp.Client) error {
		remoteDir := filepath.ToSlash(filepath.Dir(remoteFilePath));
		if _, err := sftpClient.Stat(remoteDir); err != nil && os.IsNotExist(err) {
			log.Info(remoteDir)

			err = sftpClient.MkdirAll(remoteDir)
			if err != nil {
				log.Error(err)
				return err
			}
		}
		remoteFile, err := sftpClient.Create(filepath.ToSlash(remoteFilePath))
		if err != nil {
			log.Error(err)
			return err
		}
		defer remoteFile.Close()

		_, err = io.Copy(remoteFile, reader)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
}
