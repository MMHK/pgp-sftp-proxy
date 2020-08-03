package lib

import (
	"os"
	"testing"
)

func Test_OCRService_GetPDFText(t *testing.T) {
	conf := &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	t.Logf("%+v", conf)

	ocr, err := NewOCRService(conf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	pdf, err := os.Open(getLocalPath("../test/temp/dahsing-CI-sample.pdf"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	defer pdf.Close()


	err = ocr.GetPDFText(pdf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}

func Test_OCRService_GetFormData(t *testing.T) {
	conf := &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	t.Logf("%+v", conf)

	ocr, err := NewOCRService(conf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	pdf, err := os.Open(getLocalPath("../test/temp/dahsing-SCHEDULE-sample.pdf"))
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	defer pdf.Close()


	err = ocr.GetFormData(pdf)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
}