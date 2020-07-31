package lib

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	basePath, err := filepath.Abs(this.config.TempDir)
	if err != nil {
		return err
	}
	tempDirPath := filepath.Join(basePath, tid)
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

			log.Debugf("begin download %s => %s", remoteFile, localPath)

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

func (this *DownLoader) GetLocalFiles(localDir string) ([]*RemoteNode, error) {
	fileList := make([]*RemoteNode, 0)

	if _, err := os.Stat(localDir); err != nil && os.IsNotExist(err) {
		return fileList, errors.New(fmt.Sprintf("localDir is not exist, %s", localDir))
	}

	filepath.Walk(localDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Error(err)
				return err
			}

			if !info.IsDir() {
				fileList = append(fileList, &RemoteNode{
					Info:     info,
					FullPath: path,
				})
			}

			return nil
		})

	return fileList, nil
}

func (this *DownLoader) DecryptFiles(localDir string) (error) {
	fileList, err := this.GetLocalFiles(localDir)
	if err != nil {
		log.Error(err)
		return err
	}

	queue := make(chan bool, 5)
	done := make(chan bool, 0)
	defer close(queue)
	defer close(done)

	counter := 0

	privateKeyBin, err := ioutil.ReadFile(this.config.PGP.PrivateKeyPath)
	if err != nil {
		log.Error(err)
		return err
	}

	for _, file := range fileList {
		if strings.EqualFold(strings.ToLower(filepath.Ext(file.Info.Name())), ".pgp") {
			counter++

			privateKeyReader := bytes.NewReader(privateKeyBin)
			go func(fullPath string, privateKey io.Reader) (error) {
				//进入队列
				queue <- true
				defer func() {
					//退出队列
					<-queue
					done <- true
				}()

				saveDir := filepath.Dir(fullPath)
				filename := strings.Replace(filepath.Base(fullPath), ".pgp", "", 1)

				log.Debugf("begin decrypt %s => %s", fullPath, filename)

				raw, err := os.Open(fullPath)
				if err != nil {
					log.Error(err)
					return err
				}
				defer raw.Close()

				PGP := &PGPHelper{PrivateKey: privateKey}
				decryptedReader, err := PGP.Decrypt(raw)
				if err != nil {
					log.Error(err)
					return err
				}

				decryptedFile, err := os.Create(filepath.Join(saveDir, filename))
				if err != nil {
					log.Error(err)
					return err
				}
				defer decryptedFile.Close()

				_, err = io.Copy(decryptedFile, decryptedReader)
				if err != nil {
					log.Error(err)
					return err
				}

				raw.Close()
				err = os.Remove(fullPath)
				if err != nil {
					log.Error(err)
				}

				return nil

			}(file.FullPath, privateKeyReader)
		}
	}

	for counter > 0 {
		if <- done {
			counter--
		}
	}

	return nil
}

func (this *DownLoader) UnZipFiles(localDir string) (error) {
	fileList, err := this.GetLocalFiles(localDir)
	if err != nil {
		log.Error(err)
		return err
	}

	queue := make(chan bool, 5)
	done := make(chan bool, 0)
	defer close(queue)
	defer close(done)

	counter := 0

	for _, file := range fileList {
		if strings.EqualFold(strings.ToLower(filepath.Ext(file.Info.Name())), ".zip") {
			counter++

			go func(fullPath string) (error) {
				//进入队列
				queue <- true
				defer func() {
					//退出队列
					<-queue
					done <- true
				}()

				extractDir := strings.Replace(filepath.Base(fullPath), ".zip", "", 1)
				extractDir = filepath.Join(filepath.Dir(fullPath), extractDir)

				log.Debugf("begin unzip %s => %s", fullPath, extractDir)

				_, err = UnZipFile(fullPath, extractDir)
				if err != nil {
					log.Error(err)
					return err
				}

				defer os.Remove(fullPath)

				return nil
			}(file.FullPath)
		}
	}


	for counter > 0 {
		if <- done {
			counter--
		}
	}

	return nil
}

// UnZipFile will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func UnZipFile(zipFilePath string, extractDir string) ([]string, error) {

	var fileList []string

	unzipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fileList, err
	}
	defer unzipReader.Close()

	//check and create extractDir
	if _, err := os.Stat(extractDir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(extractDir, os.ModePerm)
	}

	for _, subFile := range unzipReader.File {

		// Store filename/path for returning and using later on
		subPath := filepath.Join(extractDir, subFile.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(subPath, filepath.Clean(extractDir)+string(os.PathSeparator)) {
			return fileList, fmt.Errorf("%s: illegal file path", subPath)
		}

		fileList = append(fileList, subPath)

		if subFile.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(subPath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(subPath), os.ModePerm); err != nil {
			return fileList, err
		}

		outFile, err := os.OpenFile(subPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, subFile.Mode())
		if err != nil {
			return fileList, err
		}

		rc, err := subFile.Open()
		if err != nil {
			return fileList, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return fileList, err
		}
	}
	return fileList, nil
}