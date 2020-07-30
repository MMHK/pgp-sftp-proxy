package lib

import (
	"testing"
	"time"
)


func Test_TempDir(t *testing.T)  {
	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}

	worker := NewDownLoader(conf)

	err = worker.TempDir(func(tempDir string) error {
		t.Log(tempDir)


		time.Sleep(time.Second * 10)

		return nil
	})

	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
}

func Test_DownloadFiles_DecryptFiles_UnZipFiles(t *testing.T) {
	conf, err := loadConfig()
	if err != err {
		t.Error(err)
		t.Fail()
		return
	}

	conf.TempDir = getLocalPath("../" + conf.TempDir)
	conf.PGP.PrivateKeyPath = getLocalPath("../" + conf.PGP.PrivateKeyPath)
	conf.PGP.PublicKeyPath = getLocalPath("../" + conf.PGP.PublicKeyPath)

	t.Log(conf.TempDir)
	t.Log(conf.PGP.PrivateKeyPath)

	worker := NewDownLoader(conf)

	err = worker.TempDir(func(tempDir string) error {
		t.Log(tempDir)


		err := worker.DownloadFiles(tempDir)
		if err != err {
			return err
		}

		err = worker.DecryptFiles(tempDir)
		if err != err {
			return err
		}

		err = worker.UnZipFiles(tempDir)
		if err != err {
			return err
		}

		list, err := worker.GetLocalFiles(tempDir)
		if err != err {
			return err
		}

		for _, file := range list {
			t.Log(file.FullPath)
		}

		return nil
	})

	if err != err {
		t.Error(err)
		t.Fail()
		return
	}
}