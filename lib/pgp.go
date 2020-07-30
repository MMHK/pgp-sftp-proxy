package lib

import (
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	_ "golang.org/x/crypto/ripemd160"
	"io"
	"os"
)



const DECRYPTED_MSG_TYPE_BASE64 = "base64"
const DECRYPTED_MSG_TYPE_BINARY = "binary"

func PGP_Encrypt_Ascii_Armor_Reader(reader io.Reader, PublicKey io.Reader) (*bytes.Buffer, error) {
	return PGP_Encrypt_Reader(reader, PublicKey, DECRYPTED_MSG_TYPE_BASE64)
}

func PGP_Encrypt_Binary_Reader(reader io.Reader, PublicKey io.Reader) (*bytes.Buffer, error) {
	return PGP_Encrypt_Reader(reader, PublicKey, DECRYPTED_MSG_TYPE_BINARY)
}

func PGP_Encrypt_Reader(reader io.Reader, PublicKey io.Reader, outType string) (*bytes.Buffer, error) {
	entryList, err := openpgp.ReadArmoredKeyRing(PublicKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	buffer := new(bytes.Buffer)

	var outWriter io.Writer
	if outType == DECRYPTED_MSG_TYPE_BASE64 {
		header := map[string]string{"Creator": "MixMedia"}
		cWriter, err := armor.Encode(buffer, "PGP MESSAGE", header)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		defer cWriter.Close()
		outWriter = cWriter
	} else {
		outWriter = buffer
	}

	writer, err := openpgp.Encrypt(outWriter, entryList, nil, nil, nil)
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

func PGP_Decrypt_Ascii_Armor_Reader(decryptedMsg io.Reader, PrivateKey io.Reader) (io.Reader, error) {
	return PGP_Decrypt_Reader(decryptedMsg, PrivateKey, DECRYPTED_MSG_TYPE_BASE64)
}

func PGP_Decrypt_Binary_Reader(decryptedMsg io.Reader, PrivateKey io.Reader) (io.Reader, error) {
	return PGP_Decrypt_Reader(decryptedMsg, PrivateKey, DECRYPTED_MSG_TYPE_BINARY)
}

func PGP_Decrypt_Reader(decryptedMsg io.Reader, PrivateKey io.Reader, inputType string) (io.Reader, error) {
	entryList, err := openpgp.ReadArmoredKeyRing(PrivateKey)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var decryptedReader io.Reader
	if inputType == DECRYPTED_MSG_TYPE_BASE64 {
		block, err := armor.Decode(decryptedMsg)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		decryptedReader = block.Body
	} else {
		decryptedReader = decryptedMsg
	}

	msgDesc, err := openpgp.ReadMessage(decryptedReader, entryList, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return msgDesc.UnverifiedBody, nil
}

func PGP_Encrypt(src []byte, PublicKey io.Reader) (EncryptEntry string, err error) {
	reader := bytes.NewReader(src)
	buffer, err := PGP_Encrypt_Ascii_Armor_Reader(reader, PublicKey)
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

	buffer, err := PGP_Encrypt_Ascii_Armor_Reader(reader, PublicKey)
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


