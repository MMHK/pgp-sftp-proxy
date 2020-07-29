package lib

import (
	"os"
	"testing"
)

func GetStorage() Storage {
	conf, err := loadConfig();
	if err != nil {
		log.Error(err)
		return nil
	}
	return NewStorage(&conf.SSH)
}

func TestSFTPGetFiles(t *testing.T)  {

	disk := GetStorage()

	list, err := disk.GetFiles("/webroot/1594363294086")
	if err != nil {
		t.Error(err)
		return
	}

	for _, path := range list {
		t.Log(path)
	}
}

func TestSFTPSPut(t *testing.T) {

	disk := GetStorage()

	err := disk.Put(getLocalPath("../test/Sample.jpg"), "/webroot/upload/temp/Sample.jpg")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestSFTPSGet(t *testing.T) {

	disk := GetStorage()

	local := getLocalPath("../test/temp.jpg")

	err := disk.Get(local, "/webroot/upload/temp/Sample.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		os.Remove(local)
	}()
}

func TestSFTPRemove(t *testing.T) {

	disk := GetStorage()

	err := disk.Remove("/webroot/upload/temp")
	if err != nil {
		t.Error(err)
		return
	}
}