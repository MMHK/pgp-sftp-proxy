package lib

import (
	"github.com/ledongthuc/pdf"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"strings"
)

func GbkToUtf8(source string) (string, error) {
	reader := transform.NewReader(strings.NewReader(source), simplifiedchinese.GBK.NewDecoder())
	bin, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(bin), nil
}

func GetLines(pdfPath string) ([][]string, error) {
	pdfFile, reader, err := pdf.Open(pdfPath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer pdfFile.Close()

	pageCounter := reader.NumPage()

	rows := make([][]string, 0)

	for pageNum := 1; pageNum <= pageCounter; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		_rows, _ := page.GetTextByRow()

		for _, row := range _rows {
			line := make([]string, 0)
			for _, word := range row.Content {
				log.Debugf("%+v", word)
				s, _ := GbkToUtf8(word.S)
				line = append(line, s)
			}

			rows = append(rows, line)
		}
	}

	return rows, nil
}