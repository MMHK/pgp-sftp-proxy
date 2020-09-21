package lib

import (
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	_ "golang.org/x/crypto/ripemd160"
	"io"
	"os"
	"path"
)

type PGPHelper struct {
	toKey []*openpgp.Entity
}

func NewPGPHelper(publicKey io.Reader) (*PGPHelper, error) {
	entryList, err := openpgp.ReadArmoredKeyRing(publicKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	
	return &PGPHelper{
		toKey: entryList,
	}, nil
}

func (this *PGPHelper) Encrypt(source io.Reader) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	
	header := map[string]string{"Creator": "MixMedia"}
	body, err := armor.Encode(buffer, "PGP MESSAGE", header)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer body.Close()
	
	writer, err := openpgp.Encrypt(body, this.toKey, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer writer.Close()
	
	_, err = io.Copy(writer, source)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return buffer, nil
}

func PGP_Encrypt(src []byte, PublicKey io.Reader) (EncryptEntry string, err error) {
	helper, err := NewPGPHelper(PublicKey)
	if err != nil {
		return "", err
	}
	srcReader := bytes.NewReader(src)

	buffer, err := helper.Encrypt(srcReader)
	if err != nil {
		return "", err
	}
	
	return buffer.String(), nil
}

func PGP_Encrypt_File(src []byte, PublicKey io.Reader, save_path string) (err error) {
	distPath := path.Dir(save_path)
	if _, err := os.Stat(distPath); err != nil && os.IsNotExist(err) {
		os.MkdirAll(distPath, os.ModePerm)
	}
	distFile, err := os.Create(save_path)
	if err != nil {
		log.Error(err)
		return err
	}
	defer distFile.Close()
	
	helper, err := NewPGPHelper(PublicKey)
	if err != nil {
		return err
	}
	srcReader := bytes.NewReader(src)
	buffer, err := helper.Encrypt(srcReader)
	
	_, err = io.Copy(distFile, buffer)
	if err != nil {
		log.Error(err)
		return err
	}
	
	return nil
}
