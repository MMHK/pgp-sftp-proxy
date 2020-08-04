package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/satori/go.uuid"
	"io"
	"os"
	"path/filepath"
	"time"
)

type S3Storage struct {
	Conf    *AWSOption
	session *session.Session
}

type UploadOptions struct {
	ContentType string
}

type IStorage interface {
	Upload(localPath string, Key string) (string, string, error)
	PutContent(content io.Reader, Key string, opt *UploadOptions) (string, string, error)
	Remove(Key string) (error)
}

func NewS3Storage(conf *AWSOption) (IStorage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, ""),
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &S3Storage{
		Conf:    conf,
		session: sess,
	}, nil
}

func (this *S3Storage) Upload(localPath string, Key string) (path string, url string, err error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", "", err
	}

	defer file.Close()

	uploader := s3manager.NewUploader(this.session)
	path = filepath.ToSlash(Key)

	info, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(this.Conf.Bucket),
		Key:    aws.String(path),
		Body:   file,
		ACL:    aws.String("public-read"),
	})

	return path, info.Location, err
}

func (this *S3Storage) PutContent(content io.Reader, Key string, opt *UploadOptions) (path string, url string, err error) {
	uploader := s3manager.NewUploader(this.session)

	contentType := "application/octet-stream"
	if len(opt.ContentType) > 0 {
		contentType = opt.ContentType
	}

	path = filepath.ToSlash(Key)

	info, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(this.Conf.Bucket),
		Key:         aws.String(path),
		Body:        content,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		log.Error(err)
		return path, path, err
	}

	return path, info.Location, err
}

func (this *S3Storage) Remove(Key string) (error) {
	s3Client := s3.New(this.session)
	remotePath := filepath.ToSlash(Key)

	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Key: aws.String(remotePath),
		Bucket: aws.String(this.Conf.Bucket),
	})
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

type OCRService struct {
	Conf    *AWSOption
	session *session.Session
}

func NewOCRService(conf *AWSOption) (*OCRService, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: credentials.NewStaticCredentials(conf.AccessKey, conf.SecretKey, ""),
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &OCRService{
		Conf:    conf,
		session: sess,
	}, nil
}

func (this *OCRService) TempS3File(reader io.Reader, callback func(remotePath string) error) (error) {
	s3Disk, err := NewS3Storage(this.Conf)
	if err != nil {
		log.Error(err)
		return err
	}

	remotePath := fmt.Sprintf("%s.pdf",  uuid.NewV4())
	_, _, err = s3Disk.PutContent(reader, remotePath, &UploadOptions{
		ContentType: "application/pdf",
	})
	if err != nil {
		log.Error(err)
		return err
	}
	defer s3Disk.Remove(remotePath)

	return callback(remotePath)
}

func (this *OCRService) GetPDFText(reader io.Reader) (error) {
	return this.TempS3File(reader, func(remotePath string) error {
		ocr := textract.New(this.session)
		log.Debug("StartDocumentTextDetection")
		resp, err := ocr.StartDocumentTextDetection(&textract.StartDocumentTextDetectionInput{
			DocumentLocation: &textract.DocumentLocation{
				S3Object: &textract.S3Object{
					Bucket: aws.String(this.Conf.Bucket),
					Name:   aws.String(remotePath),
				},
			},
		})
		if err != nil {
			log.Error(err)
			return err
		}

	restart:
		log.Debug("GetDocumentTextDetection")
		doc, err := ocr.GetDocumentTextDetection(&textract.GetDocumentTextDetectionInput{
			JobId: resp.JobId,
		})
		if err != nil {
			log.Error(err)
			return err
		}

		jobStatus := *doc.JobStatus

		if jobStatus != textract.JobStatusSucceeded {
			time.Sleep(time.Second * 3)
			goto restart
		}

		for _, block := range doc.Blocks {
			log.Debugf("%+v", block)
		}

		return nil
	});
}

func (this *OCRService) GetFormData(reader io.Reader) (map[string]string, error) {
	mapping := make(map[string]string, 0)

	err := this.TempS3File(reader, func(remotePath string) error {
		ocr := textract.New(this.session)
		log.Debug("StartDocumentAnalysis")
		resp, err := ocr.StartDocumentAnalysis(&textract.StartDocumentAnalysisInput{
			DocumentLocation: &textract.DocumentLocation{
				S3Object: &textract.S3Object{
					Bucket: aws.String(this.Conf.Bucket),
					Name:   aws.String(remotePath),
				},
			},
			FeatureTypes: []*string{
				aws.String(textract.FeatureTypeForms),
			},
		})
		if err != nil {
			log.Error(err)
			return err
		}

	restart:
		log.Debug("GetDocumentAnalysis")
		doc, err := ocr.GetDocumentAnalysis(&textract.GetDocumentAnalysisInput{
			JobId: resp.JobId,
		})
		if err != nil {
			log.Error(err)
			return err
		}

		jobStatus := *doc.JobStatus

		if jobStatus != textract.JobStatusSucceeded {
			time.Sleep(time.Second * 3)
			goto restart
		}

		mapping, err = this.GetKeyValues(doc)
		if err != nil {
			log.Error(err)
			return err
		}

		//log.Debug("%+v", mapping)

		return nil
	});

	if err != nil {
		return mapping, err
	}

	return mapping, nil
}

func (this *OCRService) getLocalCache(localPath string) (map[string]string, error) {
	cachePath := localPath + ".json"
	if _, err := os.Stat(cachePath); err != nil && os.IsNotExist(err) {
		return nil, err
	}
	cacheFile, err := os.Open(cachePath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cacheFile.Close()

	decoder := json.NewDecoder(cacheFile)
	mapping := make(map[string]string, 0)

	err = decoder.Decode(&mapping)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return mapping, nil
}

func (this *OCRService) saveLocalCache(localPath string, mapping map[string]string) (error) {
	cachePath := localPath + ".json"
	if _, err := os.Stat(cachePath); err == nil {
		os.Remove(cachePath)
	}
	cacheFile, err := os.Create(cachePath)
	if err != nil {
		log.Error(err)
		return err
	}
	defer cacheFile.Close()

	encoder := json.NewEncoder(cacheFile)
	err = encoder.Encode(mapping)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (this *OCRService) GetFormDataFromFile(localPath string) (map[string]string, error) {
	data, err := this.getLocalCache(localPath)
	if err == nil {
		return data, nil
	}

	inputFile, err := os.Open(localPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer inputFile.Close()

	data, err = this.GetFormData(inputFile)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer this.saveLocalCache(localPath, data)

	return data, nil
}

type KeyValueRaw struct {
	Key   []*string
	Value []*string
}

func (this *KeyValueRaw) GetKey(mapping map[string]*string) string {
	buf := new(bytes.Buffer)
	for i, k := range this.Key {
		v, ok := mapping[*k]
		if ok {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(*v)
		}
	}

	return buf.String()
}

func (this *KeyValueRaw) GetValue(mapping map[string]*string, valueMapping map[string][]*string) string {
	buf := new(bytes.Buffer)
	for i, k := range this.Value {
		valList, ok := valueMapping[*k]
		if ok {
			if i > 0 {
				buf.WriteString(" ")
			}
			for j, v := range valList {
				val, exist := mapping[*v]
				if exist {
					if j > 0 {
						buf.WriteString(" ")
					}
					buf.WriteString(*val)
				}

			}
		}
	}

	return buf.String()
}

func (this *OCRService) GetKeyValues(inputDocument *textract.GetDocumentAnalysisOutput) (map[string]string, error) {
	words := make(map[string]*string, 0)
	formList := make([]*KeyValueRaw, 0)
	values := make(map[string][]*string, 0)
	for _, block := range inputDocument.Blocks {
		blockType := *block.BlockType
		if blockType == textract.BlockTypeWord {
			words[*block.Id] = block.Text
		}

		if blockType == textract.BlockTypeKeyValueSet &&
			len(block.EntityTypes) > 0 &&
			(*block.EntityTypes[0]) == textract.EntityTypeValue {
			valueChild := make([]*string, 0)

			for _, item := range block.Relationships {
				if (*item.Type) == textract.RelationshipTypeChild {
					valueChild = append(valueChild, item.Ids...)
				}
			}

			values[*block.Id] = valueChild
		}

		if blockType == textract.BlockTypeKeyValueSet &&
			len(block.EntityTypes) > 0 &&
			(*block.EntityTypes[0]) == textract.EntityTypeKey {

			keyList := make([]*string, 0)
			valueList := make([]*string, 0)

			for _, item := range block.Relationships {
				if (*item.Type) == textract.RelationshipTypeValue {
					valueList = append(valueList, item.Ids...)
				}
				if (*item.Type) == textract.RelationshipTypeChild {
					keyList = append(keyList, item.Ids...)
				}
			}

			formList = append(formList, &KeyValueRaw{
				Key:   keyList,
				Value: valueList,
			})
		}
	}

	mapping := make(map[string]string, 0)

	for _, kv := range formList {
		mapping[kv.GetKey(words)] = kv.GetValue(words, values)
	}

	return mapping, nil
}