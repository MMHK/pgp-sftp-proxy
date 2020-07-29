package lib

import (
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	_ "golang.org/x/crypto/ripemd160"
	"io"
	"os"
)

func PGP_Encrypt_Reader(reader io.Reader, PublicKey io.Reader) (*bytes.Buffer, error) {
	entryList, err := openpgp.ReadArmoredKeyRing(PublicKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	buffer := new(bytes.Buffer)

	header := map[string]string{"Creator": "MixMedia"}
	cWriter, err := armor.Encode(buffer, "PGP MESSAGE", header)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer cWriter.Close()

	writer, err := openpgp.Encrypt(cWriter, entryList, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return buffer, nil
}

func PGP_Encrypt(src []byte, PublicKey io.Reader) (EncryptEntry string, err error) {
	reader := bytes.NewReader(src)
	buffer, err := PGP_Encrypt_Reader(reader, PublicKey)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return buffer.String(), nil
}

func PGP_Encrypt_File(src []byte, PublicKey io.Reader, save_path string) (err error) {
	save_file, err := os.Create(save_path)
	if err != nil {
		log.Error(err)
		return
	}
	defer save_file.Close()

	reader := bytes.NewReader(src)

	buffer, err := PGP_Encrypt_Reader(reader, PublicKey)
	if err != nil {
		log.Error(err)
		return err
	}

	_, err = io.Copy(save_file, buffer)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
