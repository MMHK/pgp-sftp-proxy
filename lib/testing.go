package lib

import (
	"path/filepath"
	"runtime"
)

func getLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func loadConfig() (*Config, error) {
	err, conf := NewConfig(getLocalPath("../conf.json"))
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func loadCustomConfig(path string) (*Config, error) {
	err, conf := NewConfig(path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}