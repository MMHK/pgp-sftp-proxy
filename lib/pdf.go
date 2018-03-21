package lib

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path"
	"time"

	"github.com/signintech/gopdf"
)

func GetPDF(imagePath string) ([]byte, error) {
	imgFile, err := os.Open(imagePath)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	rect := img.Bounds()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: "px", PageSize: gopdf.Rect{W: float64(rect.Dx()),
		H: float64(rect.Dy())}})
	pdf.AddPage()
	pdf.Image(imagePath, pdf.GetX(), pdf.GetY(), nil)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}

	return pdf.GetBytesPdf(), nil
}

func SavePDF(src_file io.Reader, src_filename string, save_path string) ([]byte, error) {
	image_file := path.Join(save_path, src_filename)
	file, err := os.Create(image_file)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	_, err = io.Copy(file, src_file)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	err = file.Close()
	if err != nil {
		ErrLogger.Println(err)
	}
	tmp_file, err := os.Open(image_file)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	defer tmp_file.Close()
	img, _, err := image.Decode(tmp_file)
	if err != nil {
		ErrLogger.Println(err)
		return nil, err
	}
	rect := img.Bounds()
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: "px", PageSize: gopdf.Rect{W: float64(rect.Dx()),
		H: float64(rect.Dy())}})
	pdf.AddPage()
	pdf.Image(image_file, pdf.GetX(), pdf.GetY(), nil)
	err = tmp_file.Close()
	if err != nil {
		ErrLogger.Println(err)
	}
	defer func() {
		time.AfterFunc(time.Second*2, func() {
			err = os.Remove(image_file)
			if err != nil {
				ErrLogger.Println(err)
			}
		})
	}()
	return pdf.GetBytesPdf(), nil
}
