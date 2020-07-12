package lib

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

func getLocalConfigPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func getConfig() (error, *Config) {
	err, conf := NewConfig(getLocalConfigPath("../conf.json"))
	if err != nil {
		return err, nil
	}
	return nil, conf
}

func getPGPKey() (error, []byte) {
	pgpKey := getLocalConfigPath("../temp/dahsing_uat_public.pem")
	src, err := ioutil.ReadFile(pgpKey)
	if err != nil {
		return err, nil
	}

	return nil, src
}

func Test_Process(t *testing.T) {
	err, conf := getConfig()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	err, key := getPGPKey()
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	files := []*ZurichFile{
		&ZurichFile{
			Url:  "https://s3-ap-southeast-1.amazonaws.com/s3.jetso.com/asset/images/M_Article_Zurich_2.gif",
			Name: "M_Article_Zurich_2.gif",
		},
		&ZurichFile{
			Url:  "https://s3-ap-southeast-1.amazonaws.com/s3.jetso.com/asset/images/M_Article_Zurich_4.jpg",
			Name: "M_Article_Zurich_4.jpg",
		},
		&ZurichFile{
			Url:  "https://s3-ap-southeast-1.amazonaws.com/s3.jetso.com/asset/images/M_Article_Zurich_ca.gif",
			Name: "M_Article_Zurich_ca.gif",
		},
	}

	z := NewZurich(conf, files, string(key), "dev", "")

	z.Process()
}
