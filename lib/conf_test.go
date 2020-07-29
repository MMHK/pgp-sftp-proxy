package lib

import (
	"testing"
)

func Test_SaveConfig(t *testing.T) {

	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	err = conf.Save()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("PASS")
}

func Test_CanUpload(t *testing.T)  {
	conf, err := loadConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}

	enable := CanUpload(conf.AvailableTime)
	t.Log(enable);
}
