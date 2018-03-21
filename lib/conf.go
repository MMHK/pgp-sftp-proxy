package lib

import (
	"encoding/json"
	"os"
)

type DeployPath struct {
	Development string `json:"dev"`
	Production  string `json:"pro"`
	Testing     string `json:"test"`
}

type Config struct {
	Listen     string     `json:"listen"`
	TempPath   string     `json:"tmp_path"`
	WebRoot    string     `json:"web_root"`
	SSH        SSHItem    `json:"ssh"`
	Deploy     DeployPath `json:"deploy_path"`
	WebKitBin  string     `json:"webkit_bin"`
	WebKitArgs []string   `json:"webkit_args"`
	save_path  string
}

func NewConfig(filename string) (err error, c *Config) {
	c = &Config{}
	c.save_path = filename
	err = c.load(filename)
	return
}

func (c *Config) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		ErrLogger.Println(err)
	}
	return err
}

func (c *Config) Save() error {
	file, err := os.Create(c.save_path)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer file.Close()
	data, err2 := json.MarshalIndent(c, "", "    ")
	if err2 != nil {
		ErrLogger.Println(err2)
		return err2
	}
	_, err3 := file.Write(data)
	if err3 != nil {
		ErrLogger.Println(err3)
	}
	return err3
}

func (c *Config) GetDeployPath(deploy_type string) string {
	switch deploy_type {
	case "dev":
		return c.Deploy.Development
	case "pro":
		return c.Deploy.Production
	}

	return c.Deploy.Testing
}
