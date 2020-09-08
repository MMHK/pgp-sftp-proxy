package lib

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type DownLoader struct {
	config *Config
}

func NewDownLoader(conf *Config) *DownLoader {
	return &DownLoader{
		config: conf,
	}
}

const (
	PDF_TYPE_SCHEDULE           = `POLICY_SCHEDULE`
	PDF_TYPE_DEBIT_NOTE         = `DEBIT_NOTE_FOR_AGENT`
	PDF_TYPE_CI                 = `MOTOR_CERTIFICATE_OF_INSURANCE`
	PDF_TYPE_DUPLICATE_SCHEDULE = `DUPLICATE_POLICY_SCHEDULE`
	PDF_TYPE_IC                 = `PAYMENT_CERTIFICATE`
)

const (
	MOTORS_PDF_TYPE_SCHEDULE 			= `schedule`
	MOTORS_PDF_TYPE_DEBIT_NOTE 			= `dc`
	MOTORS_PDF_TYPE_CI 					= `ci`
	MOTORS_PDF_TYPE_DUPLICATE_SCHEDULE 	= `schedule`
	MOTORS_PDF_TYPE_IC 					= `ic`
	MOTORS_PDF_TYPE_FULL_POLICY			= `bill`
)

const (
	OCR_CHASSIS_NUMBER 		= `Chassis No.`
	OCR_ENGINE_NUMBER 		= `Engine No. or Type`
	OCR_REGISTRATION_NUMBER = `Registration No.`
	OCR_PREMIUM_PAYABLE 	= `Premium Payable`
	OCR_MAKE 				= `Make`
	OCR_MODEL 				= `Model`
	OCR_TYPE_OF_COVER 		= `Type of Cover`
	OCR_BODY_TYPE 			= `Body`
	OCR_POLICY_NUMBER		= `Policy No.`
	OCR_NCB					= `NCB`
	OCR_PERIOD_OF_INSURANCE = `Period of Insurance`
)

const (
	MOTORS_CHASSIS_NUMBER 		= `chasis_no`
	MOTORS_ENGINE_NUMBER 		= `engn_no`
	MOTORS_REGISTRATION_NUMBER  = `rgtn_no`
	MOTORS_PREMIUM_PAYABLE 		= `payable`
	MOTORS_MAKE 				= `brand`
	MOTORS_MODEL 				= `rgtn_mdl`
	MOTORS_TYPE_OF_COVER 		= `trm_of_cvr`
	MOTORS_BODY_TYPE 			= `typ_of_bdy`
	MOTORS_POLICY_NUMBER		= `pcy_no`
	MOTORS_NCB				    = `ncd_prctg`
	MOTORS_EFFECTIVE_DATE	    = `pcy_cmnt_dt`
)

type PolicyGroup struct {
	Files              []*PolicyPDF
	ClassisNumber      string
	EngineNumber       string
	RegistrationNumber string
	PremiumPayable     string
	Make               string
	Model              string
	TypeOfCover        string
	BodyType           string
	PolicyNumber       string
	NCB                string
	EffectiveDate      *time.Time
	ExpireDate         *time.Time
}

type PolicyPDF struct {
	Node         *RemoteNode
	AgentNumber  string
	CreateTime   string
	PolicyNumber string
	PDFType      string
}

func (this *PolicyPDF) CreateAt() (time.Time, error) {
	timezone := time.FixedZone("GMT", 8)
	createAt, err := time.ParseInLocation("20060102", this.CreateTime, timezone)
	if err != nil {
		return time.Now(), err
	}

	return createAt, nil
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

func (this *DownLoader) FilterPolicyDoc(localDir string) ([]*PolicyPDF, error) {
	out := make([]*PolicyPDF, 0)

	list, err := this.GetLocalFiles(localDir)
	if err != nil {
		log.Error(err)
		return out, err
	}

	folderRule := `(?i)([^_\\\/]+)_MO_DOC_([0-9]{8})\/`
	pdfRule := `([^_]+)_(POLICY_SCHEDULE|DEBIT_NOTE_FOR_AGENT|MOTOR_CERTIFICATE_OF_INSURANCE|DUPLICATE_POLICY_SCHEDULE|PAYMENT_CERTIFICATE)_([0-9]{8})\.pdf$`

	r, err := regexp.Compile(folderRule + pdfRule)
	if err != nil {
		log.Error(err)
		return out, err
	}

	for _, item := range list {
		fullPath := filepath.ToSlash(item.FullPath)
		if r.MatchString(fullPath) {
			matches := r.FindAllStringSubmatch(fullPath, 1)
			if len(matches) > 0 && len(matches[0]) > 5 {
				out = append(out, &PolicyPDF{
					Node: item,
					PolicyNumber: matches[0][3],
					AgentNumber: matches[0][1],
					CreateTime: matches[0][5],
					PDFType: matches[0][4],
				})
			}
		}
	}

	return out, nil
}

func (this *DownLoader) GroupPolicyWithOCR(pdfList []*PolicyPDF) ([]*PolicyGroup, error) {
	group := make([]*PolicyGroup, 0)

	policyDPFMapping := make(map[string][]*PolicyPDF, 0)
	for _, pdf := range pdfList {
		policyTmp, ok := policyDPFMapping[pdf.PolicyNumber]
		if !ok {
			policyDPFMapping[pdf.PolicyNumber] = []*PolicyPDF{
				pdf,
			}
			continue
		}
		policyDPFMapping[pdf.PolicyNumber] = append(policyTmp, pdf)
	}

	for _, pdfGroup := range policyDPFMapping {
		mapping := make(map[string]string, 0)
		for _, pdf := range pdfGroup {
			if pdf.PDFType == PDF_TYPE_SCHEDULE {
				ocr, err := NewOCRService(this.config.AWS)
				if err != nil {
					log.Error(err)
					continue
				}
				ocr.SetCache(this.config.TempDir)
				mapping, err = ocr.GetFormDataFromFile(pdf.Node.FullPath)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		}

		if len(mapping) > 0 {
			policy := &PolicyGroup{
				Files:              pdfGroup,
				ClassisNumber:      mapping[OCR_CHASSIS_NUMBER],
				EngineNumber:       mapping[OCR_ENGINE_NUMBER],
				RegistrationNumber: mapping[OCR_REGISTRATION_NUMBER],
				PremiumPayable:     mapping[OCR_PREMIUM_PAYABLE],
				Make:               mapping[OCR_MAKE],
				Model:              mapping[OCR_MODEL],
				TypeOfCover:        mapping[OCR_TYPE_OF_COVER],
				BodyType:           mapping[OCR_BODY_TYPE],
				PolicyNumber:       mapping[OCR_POLICY_NUMBER],
				NCB:                mapping[OCR_NCB],
			}

			start, end, err := SplitEffectiveDateAndExpireDate(mapping[OCR_PERIOD_OF_INSURANCE])
			if err != nil {
				log.Error(err)
			} else {
				policy.EffectiveDate = start
				policy.ExpireDate = end
			}

			group = append(group, policy)
		}
	}

	return group, nil
}

func (this *DownLoader) Callback(list []*PolicyGroup) error {
	uploadEndPoint := this.config.WebHook.Upload

	uploadPathMappings := map[string]string {
		PDF_TYPE_SCHEDULE: MOTORS_PDF_TYPE_SCHEDULE,
		PDF_TYPE_DEBIT_NOTE: MOTORS_PDF_TYPE_DEBIT_NOTE,
		PDF_TYPE_CI: MOTORS_PDF_TYPE_CI,
		PDF_TYPE_DUPLICATE_SCHEDULE: MOTORS_PDF_TYPE_DUPLICATE_SCHEDULE,
		PDF_TYPE_IC: MOTORS_PDF_TYPE_IC,
	}


	for _, group := range list {
		for _, pdf := range group.Files {
			url := fmt.Sprintf(uploadEndPoint, uploadPathMappings[pdf.PDFType])

			log.Infof("callback url:%s", url)

			err := UploadFile(url, map[string]string{
				fmt.Sprintf("data[%s]", MOTORS_CHASSIS_NUMBER): group.ClassisNumber,
				fmt.Sprintf("data[%s]", MOTORS_ENGINE_NUMBER): group.EngineNumber,
				fmt.Sprintf("data[%s]", MOTORS_REGISTRATION_NUMBER): group.RegistrationNumber,
				fmt.Sprintf("data[%s]", MOTORS_MODEL): group.Model,
				fmt.Sprintf("data[%s]", MOTORS_BODY_TYPE): group.BodyType,
				fmt.Sprintf("data[%s]", MOTORS_EFFECTIVE_DATE): group.EffectiveDate.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("data[%s]", MOTORS_POLICY_NUMBER): group.PolicyNumber,
			}, "file", pdf.Node.FullPath, this.config.WebHook.APIKey)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	return nil
}

func UploadFile(url string, params map[string]string, fileFieldName string, filePath string, apiKey string) (error) {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close();

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fileFieldName, filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	nowTime := time.Now()
	timeStr := fmt.Sprintf("%d", nowTime.Unix())
	hash := md5.Sum([]byte(timeStr + apiKey))
	signature := hex.EncodeToString(hash[:])

	err = writer.WriteField("time", timeStr)
	if err != nil {
		log.Error(err)
	}
	err = writer.WriteField("signature", signature)
	if err != nil {
		log.Error(err)
	}

	for key, val := range params {
		err = writer.WriteField(key, val)
		if err != nil {
			log.Error(err)
		}
	}

	defer writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: time.Second * 120,
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Info(string(content))

	return nil
}

func SplitEffectiveDateAndExpireDate(src string) (EffectiveDate *time.Time, ExpireDate *time.Time, err error) {
	params := strings.SplitN(src, ` to `, 2)

	if len(params) < 2 {
		return nil, nil, errors.New("parse error")
	}

	timezone, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		log.Error(err)
		timezone = time.FixedZone("GMT", 8)
	}

	start, err := time.ParseInLocation("02 January 2006 15:04", params[0], timezone)
	if err != nil {
		log.Error(err)
		return nil, nil, err;
	}

	end, err := time.ParseInLocation("02 January 2006", params[1], timezone)
	if err != nil {
		log.Error(err)
		return nil, nil, err;
	}

	return &start, &end, nil
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