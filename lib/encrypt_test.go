package lib

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_PGP_Encrypt(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/M_Article_Zurich_ca.gif"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	//
	//keyFile, err := ioutil.ReadFile(getLocalPath("../test/test-key.pem"))
	//if err != nil {
	//	t.Log(err)
	//	t.Fail()
	//	return
	//}
	reader, err := os.Open(getLocalPath("../test/test-key.pem"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	defer reader.Close()


	pgp, err := PGP_Encrypt(sourceFile, reader)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	t.Log(pgp)
}

func Test_PGP_Encrypt_File(t *testing.T) {
	sourceFile, err := ioutil.ReadFile(getLocalPath("../test/M_Article_Zurich_ca.gif"))
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

	distFile := getLocalPath("../test/M_Article_Zurich_ca.gif.pgp")
	defer func() {
		os.Remove(distFile)
	}()

	err = PGP_Encrypt_File(sourceFile, keyFile, distFile)
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
}
