package lib

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)


func Test_getPDFBytes(t *testing.T) {

	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	imgFilePath := getLocalPath("../test/Image 13.png")
	img, err := os.Open(imgFilePath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer img.Close()


	reader, err := getPDFBytes(img, conf.TempPath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	pdf, err := os.Create(getLocalPath("../temp/dist.pdf"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer pdf.Close()

	_, err = io.Copy(pdf, reader)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}

func Test_SavePDF(t *testing.T) {
	imgFilePath := getLocalPath("../test/Image 13.png")
	img, err := os.Open(imgFilePath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer img.Close()

	distFile := getLocalPath("../temp/dist.pdf")

	pdf, err := SavePDF(img, filepath.Base(distFile), filepath.Dir(distFile))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	err = ioutil.WriteFile(distFile, pdf, os.ModePerm)

	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}
