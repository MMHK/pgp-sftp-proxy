package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ZurichFile struct {
	Name string
	Path string
	Url  string
}

type Zurich struct {
	conf        *Config
	Files       []*ZurichFile
	NotifyUrl   string
	downloadJob chan bool
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
		downloadJob: make(chan bool, 5),
		prefixPath:  fmt.Sprintf("%d", time.Now().UnixNano()),
		pgpKey:      publicKey,
		notifyTries: 2,
		pgpFiles:    make([]*ZurichFile, len(files)),
		deployENV:   ENV,
	}
}

//准备相关文件
func (d *Zurich) prepareFile() {
	InfoLogger.Println("prepare Files")

	for _, item := range d.Files {
		go d.DownloadRemoteFile(item)
	}
}

func (d *Zurich) Process() {
	d.prepareFile()

	d.WhenAllFileReady(func() {
		d.EncryptFiles()

		d.UploadToSFTP()

		if len(d.NotifyUrl) > 0 {
			d.notifyRemote(d.NotifyUrl)
		}

		time.AfterFunc(time.Minute*1, func() {
			d.ClearAllFiles()
		})
	})
}

//检查路径是否存在
func (d *Zurich) FileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

//下载远程文件
func (d *Zurich) DownloadRemoteFile(zFile *ZurichFile) (err error) {
	InfoLogger.Println("begin download file, url:", zFile.Url)

	basePath := filepath.Join(d.conf.TempPath, d.prefixPath)

	if !d.FileExist(basePath) {
		os.MkdirAll(basePath, 0777)
	}
	localPath := filepath.Join(basePath, zFile.Name)

	defer (func() {
		d.downloadJob <- true
	})()

	file, err := os.Create(localPath)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer file.Close()

	resp, err := http.Get(zFile.Url)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}

	zFile.Path = localPath
	return nil
}

//当所有操作完成时回调
func (d *Zurich) WhenAllFileReady(callback func()) {
	defer callback()

	downloadCount := len(d.Files)

	for {
		<-d.downloadJob
		InfoLogger.Println("a download job Done.")
		downloadCount--
		if downloadCount <= 0 {
			break
		}
	}

	close(d.downloadJob)
}

//清理所有文件夹
func (d *Zurich) ClearAllFiles() {
	InfoLogger.Println("begin clear all files")

	os.RemoveAll(filepath.Join(d.conf.TempPath, d.prefixPath))
}

//通知远程
func (d *Zurich) notifyRemote(remoteURL string) {
	if d.notifyTries < 0 {
		return
	}
	d.notifyTries--
	InfoLogger.Println("begin notify")

	resp, err := http.Get(d.NotifyUrl)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	//	InfoLogger.Println("notify response:", string(body))
	//检查通知返回是否正确
	if !strings.EqualFold(string(body), "success") {
		//一分钟后重试通知
		time.AfterFunc(time.Minute*1, func() {
			InfoLogger.Println("retry notify:", remoteURL)
			d.notifyRemote(remoteURL)
		})
	}
}

//加密索引文件及打包文件
func (d *Zurich) EncryptFiles() error {
	InfoLogger.Println("begin encrypt files")

	for index, zFile := range d.Files {
		var src []byte
		var err error
		//检查是否图片
		if d.isImage(zFile.Path) {
			//将图片转换成PDF
			pdfFileName := zFile.Path + ".pdf"
			src, err = GetPDF(zFile.Path)
			if err != nil {
				ErrLogger.Println(err)
				return err
			}
			zFile.Path = pdfFileName
		} else {
			src, err = ioutil.ReadFile(zFile.Path)
			if err != nil {
				ErrLogger.Println(err)
				return err
			}
		}
		keyReader := strings.NewReader(d.pgpKey)

		pgpFile := &ZurichFile{
			Path: zFile.Path + ".pgp",
		}
		err = PGP_Encrypt_File(src, keyReader, pgpFile.Path)
		if err != nil {
			ErrLogger.Println(err)
			return err
		}
		d.pgpFiles[index] = pgpFile
	}

	return nil
}

//检查是否图片文件
func (d *Zurich) isImage(filePath string) bool {
	ext := filepath.Ext(filePath)
	for _, condition := range []string{"png", "gif", "jpg", "bmp", "jpeg"} {
		if strings.Contains(ext, condition) {
			return true
		}
	}

	return false
}

//上传到SFTP
func (d *Zurich) UploadToSFTP() {
	InfoLogger.Println("begin upload 2 sftp")
	ssh := NewSSHClient(&d.conf.SSH)

	prefixFolder := d.conf.GetDeployPath(d.deployENV)

	for _, pgpFile := range d.pgpFiles {
		err := ssh.UploadFile(pgpFile.Path, prefixFolder)
		if err != nil {
			ErrLogger.Println(err)
		}
	}
}
