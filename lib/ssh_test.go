package lib

import "testing"

func TestSSHClient_Connect(t *testing.T) {

	conf, err := loadConfig()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	ssh := NewSSHClient(&conf.SSH)

	session, err := ssh.Connect()
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}

	defer session.Close()
}
