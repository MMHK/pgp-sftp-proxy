package lib

import (
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	_ "golang.org/x/crypto/ripemd160"
	"io"
	"os"
)

func PGP_Encrypt(src []byte, PublicKey io.Reader) (EncryptEntry string, err error) {

	enrtylist, err := openpgp.ReadArmoredKeyRing(PublicKey)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	buffer := new(bytes.Buffer)

	header := map[string]string{"Creator": "MixMedia"}
	cWriter, err := armor.Encode(buffer, "PGP MESSAGE", header)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	writer, err := openpgp.Encrypt(cWriter, enrtylist, nil, nil, nil)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	_, err = writer.Write(src)
	if err != nil {
		ErrLogger.Println(err)
		return
	}

	writer.Close()

	cWriter.Close()

	EncryptEntry = buffer.String()
	return
}

func PGP_Encrypt_File(src []byte, PublicKey io.Reader, save_path string) (err error) {
	enrtylist, err := openpgp.ReadArmoredKeyRing(PublicKey)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	save_file, err := os.Create(save_path)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	defer save_file.Close()

	header := map[string]string{"Creator": "MixMedia"}
	cWriter, err := armor.Encode(save_file, "PGP MESSAGE", header)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	writer, err := openpgp.Encrypt(cWriter, enrtylist, nil, nil, nil)
	if err != nil {
		ErrLogger.Println(err)
		return
	}
	_, err = writer.Write(src)
	if err != nil {
		ErrLogger.Println(err)
		return
	}

	writer.Close()

	cWriter.Close()
	return
}
