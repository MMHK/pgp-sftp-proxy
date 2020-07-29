package lib

import (
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

	buffer, err := PGP_Encrypt_Reader(sourceFile, keyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(buffer.String())
}

func Test_PGP_Decrypt_Reader(t *testing.T)  {
	publicKeyFile, err := os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer publicKeyFile.Close()

	raw := "Hello! World!"
	rawReader := strings.NewReader(raw)

	encryptedMsg, err := PGP_Encrypt_Reader(rawReader, publicKeyFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	privateKeyFile, err := os.Open(getLocalPath("../test/test-private-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer privateKeyFile.Close()

	t.Log(encryptedMsg.String());

	rawBuffer, err := PGP_Decrypt_Reader(encryptedMsg, privateKeyFile)
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
