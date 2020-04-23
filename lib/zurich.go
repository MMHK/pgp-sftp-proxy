package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// swagger:model
type ZurichFile struct {
	// required: true
	Name string `json:"name"`
	Path string
	// required: true
	Url string `json:"url"`
}

type Zurich struct {
	conf        *Config
	Files       []*ZurichFile
	NotifyUrl   string
	prefixPath  string
	notifyTries int
	pgpKey      string
	pgpFiles    []*ZurichFile
	deployENV   string
}

func NewZurich(conf *Config, files []*ZurichFile, publicKey string, ENV string, notifyUrl string) *Zurich {
	return &Zurich{
		conf:        conf,
		Files:       files,
		NotifyUrl:   notifyUrl,
		prefixPath:  fmt.Sprintf("%d", rand.Int63()),
		pgpKey:      publicKey,
		notifyTries: 2,
		pgpFiles:    make([]*ZurichFile, len(files)),
		deployENV:   ENV,
	}
}

//准备相关文件
func (this *Zurich) prepareFile() (err error) {
	log.Info("prepare Files")
	queue := make(chan bool, 0)
	counter := 0
	defer close(queue)

	for i, item := range this.Files {
		counter++
		
		go func(i int, item *ZurichFile) {
			defer func() {
				queue <- true
			}()
			
			this.Files[i], err = this.DownloadRemoteFile(item)
		}(i, item)
	}
	
	for <-queue {
		counter--
		log.Debug("total queue:", counter)
		
		if counter <= 0 {
			break
		}
	}
	
	return nil
}

func (this *Zurich) Process() {
	defer this.ClearAllFiles()
	
	err := this.prepareFile()
	if err != nil {
		log.Error(err)
		return
	}
	
	err = this.EncryptFiles()
	if err != nil {
		log.Error(err)
		return
	}

	this.UploadToSFTP()

	if len(this.NotifyUrl) > 0 {
		this.notifyRemote(this.NotifyUrl)
	}
}

//下载远程文件
func (this *Zurich) DownloadRemoteFile(zFile *ZurichFile) (localFile *ZurichFile, err error) {
	log.Info("begin download file, url:", zFile.Url)

	basePath := filepath.Join(this.conf.TempPath, this.prefixPath)

	if _, err := os.Stat(basePath); err != nil && os.IsNotExist(err) {
		os.MkdirAll(basePath, os.ModePerm)
	}
	localPath := filepath.Join(basePath, zFile.Name)

	file, err := os.Create(localPath)
	if err != nil {
		log.Error(err)
		return
	}
	defer file.Close()

	resp, err := http.Get(zFile.Url)
	if err != nil {
		log.Error(err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	zFile.Path = localPath
	return zFile, nil
}

//清理所有文件夹
func (this *Zurich) ClearAllFiles() {
	log.Info("begin clear all files")

	os.RemoveAll(filepath.Join(this.conf.TempPath, this.prefixPath))
}

//通知远程
func (this *Zurich) notifyRemote(remoteURL string) {
	if this.notifyTries < 0 {
		return
	}
	this.notifyTries--
	log.Info("begin notify")

	resp, err := http.Get(this.NotifyUrl)
	if err != nil {
		log.Error(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}
	//	log.Info("notify response:", string(body))
	//检查通知返回是否正确
	if !strings.EqualFold(string(body), "success") {
		//一分钟后重试通知
		time.AfterFunc(time.Minute*1, func() {
			log.Info("retry notify:", remoteURL)
			this.notifyRemote(remoteURL)
		})
	}
}

//加密索引文件及打包文件
func (this *Zurich) EncryptFiles() (err error) {
	log.Info("begin encrypt files")
	queue := make(chan bool, 0)
	counter := 0
	defer close(queue)

	for index, zFile := range this.Files {
		counter++
		
		go func(index int, zFile *ZurichFile) {
			defer func() {
				queue <- true
			}()
			
			log.Debug("begin encrypt file:", zFile.Name)
			
			var src []byte
			var err error
			//检查是否图片
			if this.isImage(zFile.Path) {
				//将图片转换成PDF
				pdfFileName := zFile.Path + ".pdf"
				src, err = GetPDF(zFile.Path)
				if err != nil {
					log.Error(err)
				}
				zFile.Path = pdfFileName
			} else {
				src, err = ioutil.ReadFile(zFile.Path)
				if err != nil {
					log.Error(err)
				}
			}
			keyReader := strings.NewReader(this.pgpKey)
			
			pgpFile := &ZurichFile{
				Path: zFile.Path + ".pgp",
			}
			err = PGP_Encrypt_File(src, keyReader, pgpFile.Path)
			if err != nil {
				log.Error(err)
			}
			this.pgpFiles[index] = pgpFile
			
			
		}(index, zFile)
		
	}
	
	for true {
		<- queue
		counter--
		if counter <= 0 {
			break
		}
	}

	return nil
}

//检查是否图片文件
func (this *Zurich) isImage(filePath string) bool {
	ext := filepath.Ext(filePath)
	for _, condition := range []string{"png", "gif", "jpg", "bmp", "jpeg"} {
		if strings.Contains(ext, condition) {
			return true
		}
	}

	return false
}

//上传到SFTP
func (this *Zurich) UploadToSFTP() {
	log.Info("begin upload 2 sftp")
	ssh := NewSSHClient(&this.conf.SSH)
	queue := make(chan bool, 0)
	counter := 0

	prefixFolder := this.conf.GetDeployPath(this.deployENV)

	for _, pgpFile := range this.pgpFiles {
		counter++
		
		go func(pgpFile *ZurichFile) {
			defer func() {
				queue <- true
			}()
			
			log.Info("upload 2 sftp:", pgpFile.Path)
			
			err := ssh.UploadFile(pgpFile.Path, prefixFolder)
			if err != nil {
				log.Error(err)
			}
		}(pgpFile)
		
	}
	
	for true {
		<- queue
		counter--
		if counter <= 0 {
			break
		}
	}
}
