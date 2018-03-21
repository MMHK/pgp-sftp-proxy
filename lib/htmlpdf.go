package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"
)

type HTMLPDF struct {
	config *Config
}

func NewHTMLPDF(conf *Config) *HTMLPDF {
	return &HTMLPDF{
		config: conf,
	}
}

func (pdf *HTMLPDF) run(source_path string, pdf_path string) error {
	bin_args := append(pdf.config.WebKitArgs, source_path, pdf_path)
	cmd := exec.Command(pdf.config.WebKitBin, bin_args...)
	var outbuffer bytes.Buffer
	var errbuffer bytes.Buffer
	cmd.Stderr = &outbuffer
	cmd.Stderr = &errbuffer
	err := cmd.Run()
	if err != nil {
		ErrLogger.Println(err)
		ErrLogger.Println(errbuffer.String())
		return err
	}
	InfoLogger.Println(outbuffer.String())
	return nil
}

func (pdf *HTMLPDF) build(html []byte) (err error, PDFBin []byte) {
	tmp_name := fmt.Sprintf("%d.html", time.Now().UnixNano())
	tmp_name = path.Join(pdf.config.TempPath, tmp_name)
	err = ioutil.WriteFile(tmp_name, html, 0777)
	if err != nil {
		return
	}
	defer os.Remove(tmp_name)

	pdf_name := fmt.Sprintf("%d.pdf", time.Now().UnixNano())
	pdf_name = path.Join(pdf.config.TempPath, pdf_name)
	err = pdf.run(tmp_name, pdf_name)
	if err != nil {
		return
	}
	bin, err := ioutil.ReadFile(pdf_name)
	if err != nil {
		return
	}
	PDFBin = bin
	defer os.Remove(pdf_name)
	return
}
