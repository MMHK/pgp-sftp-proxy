package lib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)


func Test_TempDir(t *testing.T)  {
	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}

	worker := NewDownLoader(conf)

	err = worker.TempDir(func(tempDir string) error {
		t.Log(tempDir)


		time.Sleep(time.Second * 10)

		return nil
	})

	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
}

func Test_DownloadFiles_DecryptFiles_UnZipFiles(t *testing.T) {
	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}

	conf.TempDir = getLocalPath("../" + conf.TempDir)
	conf.PGP.PrivateKeyPath = getLocalPath("../" + conf.PGP.PrivateKeyPath)
	conf.PGP.PublicKeyPath = getLocalPath("../" + conf.PGP.PublicKeyPath)

	t.Log(conf.TempDir)
	t.Log(conf.PGP.PrivateKeyPath)

	worker := NewDownLoader(conf)

	err = worker.TempDir(func(tempDir string) error {
		t.Log(tempDir)


		err := worker.DownloadFiles(tempDir)
		if err != err {
			return err
		}

		err = worker.DecryptFiles(tempDir)
		if err != err {
			return err
		}

		err = worker.UnZipFiles(tempDir)
		if err != err {
			return err
		}

		list, err := worker.GetLocalFiles(tempDir)
		if err != err {
			return err
		}

		for _, file := range list {
			t.Log(file.FullPath)
		}

		return nil
	})

	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
}

func Test_MatchPolicyFiles(t *testing.T) {

	list := []string{
		`AG057892_MO_DOC_20200731/BMC238973_POLICY_SCHEDULE_20200731.pdf`,
		`asdasd/BMC238973asdasdasdasd20200731.png`,
		`AG057892_MO_DOC_20200731\BMC238973_DEBIT_NOTE_FOR_AGENT_20200731.pdf`,
		`asdasd/BMC238973asdasdasdasd20200731.png`,
		`AG057892_MO_DOC_20200731/BMC238973_MOTOR_CERTIFICATE_OF_INSURANCE_20200731.PDF`,
		`asdasd/BMC238973asdasdasdasd20200731.PDF`,
		`AG057892_MO_DOC_20200731\BMC238973_DUPLICATE_POLICY_SCHEDULE_20200731.PDF`,
		`AG057892_MO_DOC_20200731/BMC238973_PAYMENT_CERTIFICATE_20200731.PDF`,
		`asdasd\BMC238973asdasdasd\asd20200731.gif`,
	}

	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
	conf.TempDir = getLocalPath("../" + conf.TempDir)

	worker := NewDownLoader(conf)

	err = worker.TempDir(func(tempDir string) error {

		for _, f := range list {
			fullPath := filepath.ToSlash(filepath.Join(tempDir, f))
			dir := filepath.Dir(fullPath)
			if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
				os.MkdirAll(dir, os.ModePerm)
			}
			err = ioutil.WriteFile(fullPath, []byte("Hello~"), os.ModePerm)
			if err != nil {
				t.Error(err)
			}
		}

		list, err := worker.FilterPolicyDoc(tempDir)
		if err != nil {
			return err
		}

		for _, i := range list {
			t.Logf("%+v", i)
		}


		return nil
	})

	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

}

func Test_DownLoader_GetPolicyDataWithOCR(t *testing.T) {
	localPDFPath := getLocalPath("../test/temp/dahsing-SCHEDULE-sample.pdf")
	localPDFFileInfo, err := os.Stat(localPDFPath)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	pdfList := []*PolicyPDF{
		&PolicyPDF{
			Node: &RemoteNode{
				Info: localPDFFileInfo,
				FullPath: localPDFPath,
			},
			AgentNumber: "123456",
			CreateTime: "20200803",
			PolicyNumber: "",
			PDFType: PDF_TYPE_SCHEDULE,
		},
	}

	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
	conf.TempDir = getLocalPath("../" + conf.TempDir)
	conf.AWS = &AWSOption{
		Region: "ap-southeast-1",
		Bucket: "s3.test.mixmedia.com",
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	worker := NewDownLoader(conf)

	mapping, err := worker.GetPolicyDataWithOCR(pdfList)
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}

	t.Log(mapping)
}