package lib

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_PGP_Encrypt(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/Sample.jpg"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	keyFile, err :=  os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer keyFile.Close()


	pgp, err := PGP_Encrypt(sourceFile, keyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(pgp)
}

func Test_PGP_Encrypt_File(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/Sample.jpg"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	keyFile, err := os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer keyFile.Close()

	distFile := getLocalPath("../test/Sample.jpg.pgp")

	err = PGP_Encrypt_File(sourceFile, keyFile, distFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer func() {
		os.Remove(distFile)
	}()
}

func Test_PGP_Encrypt_File_Binary(t *testing.T) {
	sourceFile, err := os.Open(getLocalPath("../test/temp/A12345_MO_20200609.xml"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer sourceFile.Close()

	keyFile, err := os.Open(getLocalPath("../test/temp/dahsing-public-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer keyFile.Close()

	distFile, err := os.Create(getLocalPath("../test/temp/A12345_MO_20200609.xml.pgp"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer distFile.Close()

	buffer, err := PGP_Encrypt_Binary_Reader(sourceFile, keyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	_, err = io.Copy(distFile, buffer)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

}

func Test_PGP_Encrypt_Reader(t *testing.T) {
	sourceFile, err := os.Open(getLocalPath("../test/Sample.jpg"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer sourceFile.Close()

	keyFile, err := os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer keyFile.Close()

	buffer, err := PGP_Encrypt_Ascii_Armor_Reader(sourceFile, keyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(buffer.String())
}

func Test_PGP_Decrypt_Reader(t *testing.T)  {
	raw := "Hello! World!"
	rawReader := strings.NewReader(raw)

	privateKey := getLocalPath("../test/test-private-key.pem")
	publicKey := getLocalPath("../test/test-key.pem")
	PGP, err := NewPGPHelper(privateKey, publicKey)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	encryptedMsg, err := PGP.EncryptBinary(rawReader)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(encryptedMsg.String());

	encryptedReader := bytes.NewReader(encryptedMsg.Bytes())

	rawBuffer, err := PGP.Decrypt(encryptedReader)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	decryptRaw := new(strings.Builder)
	_, err = io.Copy(decryptRaw, rawBuffer)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(decryptRaw.String())

	if !strings.EqualFold(decryptRaw.String(), raw) {
		t.Error("not the same")
		t.Fail()
		return
	}

}

func Test_PGP_Decrypt_File(t *testing.T) {
	privateKeyFile := getLocalPath("../test/temp/driver-private-key.pem")

	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	conf.PGP.PrivateKeyPath = privateKeyFile
	conf.PGP.PublicKeyPath = getLocalPath("../" + conf.PGP.PublicKeyPath)

	PGP, err := NewPGPHelper(conf.PGP.PrivateKeyPath, conf.PGP.PublicKeyPath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	encryptedFilePath := getLocalPath("../test/temp/20191023_204179_01_rb_gs8010_ask_fmt01_#001.pdf.pgp")
	encryptedFile, err := os.Open(encryptedFilePath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer encryptedFile.Close()

	rawBuffer, err := PGP.Decrypt(encryptedFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	decryptedFilePath := strings.Replace(encryptedFilePath, ".pgp", "", 1)
	decryptedFile, err := os.Create(decryptedFilePath)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	_, err = io.Copy(decryptedFile, rawBuffer)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}