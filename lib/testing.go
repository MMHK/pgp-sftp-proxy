package lib

import (
	"bytes"
	"encoding/json"
	"github.com/op/go-logging"
	"os"
	"path/filepath"
	"github.com/joho/godotenv"
	"runtime"
)

func init() {
	format := logging.MustStringFormatter(
		`PGP-SFTP-PROXY %{color} %{shortfunc} %{level:.4s} %{shortfile}
%{id:03x}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	log := logging.MustGetLogger("PGP-SFTP-PROXY")

	err := godotenv.Load(getLocalPath("../.env"))
	if err != nil {
		log.Error("Error loading environment")
	}
}


func getLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func ToJSON(target interface{}) string {
	str := new(bytes.Buffer)
	encoder := json.NewEncoder(str)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(target)
	if err != nil {
		return err.Error()
	}

	return str.String()
}

func loadConfig() (*Config, error) {
	LOCAL_CONF_PATH := os.Getenv("TEST_CONF_PATH")
	err, conf := NewConfig(LOCAL_CONF_PATH)

	return conf, err
}
