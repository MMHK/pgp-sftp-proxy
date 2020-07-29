package lib

import (
	"fmt"
	"github.com/satori/go.uuid"
	"os"
	"path/filepath"
)

type DownLoader struct {
	config *Config
}

func NewDownLoader(conf *Config) *DownLoader {
	return &DownLoader{
		config: conf,
	}
}

func (this *DownLoader) TempDir(callback func(tempDir string)(error)) (error) {
	id := uuid.NewV4()
	tid := fmt.Sprintf("%s", id)
	tempDirPath := filepath.Join(this.config.TempDir, tid)
	if _, err := os.Stat(tempDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(tempDirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	defer os.RemoveAll(tempDirPath)

	return callback(tempDirPath)
}

func (this *DownLoader) DownloadFiles(localDir string) (error) {
	ssh := NewStorage(&this.config.SSH)
	fileList, err := ssh.GetFiles(this.config.SFTP.DownloadDir)
	if err != nil {
		log.Error(err)
		return err
	}

	queue := make(chan bool, 5)
	done := make(chan bool, 0)
	defer close(queue)
	defer close(done)

	counter := 0

	for _, remoteFile := range fileList {
		localPath := filepath.Join(localDir, filepath.Base(remoteFile))

		counter++

		go func(localPath string, remoteFile string) {
			//进入队列
			queue <- true
			defer func() {
				//退出队列
				<-queue
				done <- true
			}()

			err = ssh.Get(localPath, remoteFile)
			if err != nil {
				log.Error(err)
			}
		}(localPath, remoteFile)
	}

	for counter > 0 {
		if <- done {
			counter--
		}
	}

	return nil
}