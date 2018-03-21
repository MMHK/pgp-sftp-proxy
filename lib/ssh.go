package lib

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

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
		ErrLogger.Println(err1)
		return
	}
	key, err = ssh.ParsePrivateKey(buffer)
	if err != nil {
		ErrLogger.Println(err)
	}
	return
}

func (c *SSHClient) Connect() (session *(ssh.Session), err error) {
	if c.ssh_client == nil {
		authmethods := make([]ssh.AuthMethod, 0)
		if len(c.config.PrivateKey) > 0 {
			key, err1 := c.getKeyFile(c.config.PrivateKey)
			if err1 == nil {
				authmethods = append(authmethods, ssh.PublicKeys(key))
			} else {
				err = err1
				return
			}
		}
		authmethods = append(authmethods, ssh.Password(c.config.Password))
		config := &ssh.ClientConfig{
			User: c.config.Username,
			Auth: authmethods,
			//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

		client, err2 := ssh.Dial("tcp", c.config.Host, config)
		if err2 != nil {
			ErrLogger.Println(err2)
			err = err2
			return
		}
		c.ssh_client = client
	}
	client := c.ssh_client

	session1, err3 := client.NewSession()
	if err3 != nil {
		ErrLogger.Println(err3)
		err = err3
		return
	}
	session = session1
	return
}

func (c *SSHClient) Run(cmd string) (result string, err error) {
	session, err1 := c.Connect()
	if err1 != nil {
		err = err1
		return
	}
	defer session.Close()

	var Stdout, Stderr bytes.Buffer
	session.Stdout = &Stdout
	session.Stderr = &Stderr

	InfoLogger.Println("ssh run:", cmd)
	err = session.Run(cmd)
	if err != nil {
		if _, ok := err.(*(ssh.ExitError)); ok {
			err = errors.New(Stderr.String())
		}
		return
	}
	result = Stdout.String()
	InfoLogger.Println("remote cmd result:", result)
	return
}

func (c *SSHClient) RemoveFiles(paths []string) (err error) {
	tpl := "rm -Rf \"%s\""
	out := make([]string, len(paths))
	for i, element := range paths {
		out[i] = fmt.Sprintf(tpl, element)
	}

	cmd := strings.Join(out, ";")
	_, err = c.Run(cmd)
	return
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
		ErrLogger.Println(err1)
		err = err1
		return
	}
	defer sftpClient.Close()

	basename := filepath.Base(filename)
	localFile, err4 := os.Open(filename)
	if err4 != nil {
		err = err4
		ErrLogger.Println(err4)
		return
	}
	defer localFile.Close()

	remoteFile, err3 := sftpClient.Create(sftpClient.Join(remote_folder, basename))
	if err3 != nil {
		err = err3
		ErrLogger.Println(err3)
		return
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		ErrLogger.Println(err)
	}

	return
}
