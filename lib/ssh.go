package lib

import (
	"errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"
)

const KeyboardInteractiveRetryCount = 3

type SSHItem struct {
	Host       string `json:"host"`
	Username   string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"key"`
}

type SSHClient struct {
	config     *SSHItem
	ssh_client *ssh.Client
}

func NewSSHClient(conf *SSHItem) *SSHClient {
	return &SSHClient{
		config: conf,
	}
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
		retryCounter := 0
		//authMethods = append(authMethods, ssh.Password(c.config.Password))
		authMethods = append(authMethods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			answers = make([]string, len(questions))
			// The second parameter is unused
			for n, _ := range questions {
				answers[n] = c.config.Password
			}
			retryCounter++;
			if retryCounter >= KeyboardInteractiveRetryCount {
				return nil, errors.New("too many login attempts, invalid username or password")
			}

			return answers, nil
		}))
		config := &ssh.ClientConfig{
			User: c.config.Username,
			Auth: authMethods,
			//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
			BannerCallback: func(message string) error {
				log.Info(message)
				return nil
			},
			Timeout: time.Second * 15,
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

func (this *SSHClient) Put(remoteFilePath string, fromReader io.Reader) error {
	return this.Session(func(session *ssh.Session) error {
		sftpClient, err := sftp.NewClient(this.ssh_client)
		if err != nil {
			log.Error(err)
			return err
		}
		defer sftpClient.Close()


		remoteDir := filepath.ToSlash(filepath.Dir(remoteFilePath));
		if _, err := sftpClient.Stat(remoteDir); err != nil {
			err = sftpClient.MkdirAll(remoteDir)
			if err != nil {
				log.Error(err)
				return err
			}
		}
		log.Debug(remoteFilePath)
		remoteFile, err := sftpClient.Create(filepath.ToSlash(remoteFilePath))
		if err != nil {
			log.Error(err)
			return err
		}
		defer remoteFile.Close()
		_, err = io.Copy(remoteFile, fromReader)
		if err != nil {
			log.Error(err)
			return err
		}

		return nil
	})
}

func (c *SSHClient) UploadFile(filename string, remote_folder string) (err error) {
	if c.ssh_client == nil {
		var session *ssh.Session
		session, err = c.Connect()
		if err != nil {
			return
		}
		defer session.Close()
	}
	sftpClient, err1 := sftp.NewClient(c.ssh_client)
	if err1 != nil {
		log.Error(err1)
		err = err1
		return
	}
	defer sftpClient.Close()

	basename := filepath.Base(filename)
	localFile, err4 := os.Open(filename)
	if err4 != nil {
		err = err4
		log.Error(err4)
		return
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(sftpClient.Join(remote_folder, basename))
	if err != nil {
		log.Error(err)
		return
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		log.Error(err)
	}

	return
}
