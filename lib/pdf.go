package lib

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/signintech/gopdf"
)

func GetPDF(imagePath string) ([]byte, error) {
	imgFile, err := os.Open(imagePath)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	rect := img.Bounds()
	pdf := gopdf.GoPdf{}
	pageSize := gopdf.Rect{
		W: float64(rect.Dx()) * 0.5,
		H: float64(rect.Dy()) * 0.5,
	}
	pageSize.PointsToUnits(gopdf.UnitCM)
	pdfConfig := gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: pageSize}

	pdf.Start(pdfConfig)
	pdf.AddPage()
	//pdf.Image(imagePath, 0, 0, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return pdf.GetBytesPdf(), nil
}

func getPDFBytes(imageFile io.Reader, tmpDir string) (io.Reader, error) {
	tempFileName := fmt.Sprintf("%d", rand.Int())
	tempDir := filepath.Join(tmpDir, tempFileName)
	if _, err := os.Stat(tempDir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(tempDir, os.ModePerm)
	}
	defer os.RemoveAll(tempDir)

	tempFileName = filepath.Join(tempDir, tempFileName)
	tempFile, err := os.Create(tempFileName)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(tempFile, imageFile)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tempFile.Close()

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	img, _, err := image.Decode(tempFile)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	rect := img.Bounds()
	pdf := gopdf.GoPdf{}
	pageSize := gopdf.Rect{
		W: float64(rect.Dx()),
		H: float64(rect.Dy()),
	}
	pageSize.PointsToUnits(gopdf.UnitPT)
	pdfConfig := gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: pageSize}

	pdf.Start(pdfConfig)
	pdf.AddPage()
	pdf.Image(tempFileName, 0, 0, &pageSize)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return bytes.NewReader(pdf.GetBytesPdf()), nil
}

func SavePDF(src_file io.Reader, distFileName string, saveDir string) ([]byte, error) {
	image_file := path.Join(saveDir, distFileName)
	file, err := os.Create(image_file)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	_, err = io.Copy(file, src_file)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer file.Close()

	tmp_file, err := os.Open(image_file)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer tmp_file.Close()
	img, _, err := image.Decode(tmp_file)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	rect := img.Bounds()
	pdf := gopdf.GoPdf{}
	pageSize := gopdf.Rect{
		W: float64(rect.Dx()),
		H: float64(rect.Dy()),
	}
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitPT,
		PageSize: pageSize,
	})
	pdf.AddPage()
	pdf.Image(image_file, 0, 0, &pageSize)
	err = tmp_file.Close()
	if err != nil {
		log.Error(err)
	}
	defer func() {
		time.AfterFunc(time.Second*2, func() {
			err = os.Remove(image_file)
			if err != nil {
				log.Error(err)
			}
		})
	}()
	return pdf.GetBytesPdf(), nil
}
