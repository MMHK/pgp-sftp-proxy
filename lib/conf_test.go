package lib

import (
	"testing"
)

func Test_SaveConfig(t *testing.T) {

	err, conf := NewConfig("config.json")
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
