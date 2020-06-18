package lib

import (
	"bytes"
	"encoding/json"
	"os"
	"time"
)

type TimeRange struct {
	BeginRaw string `json:"begin"`
	EndRaw   string `json:"end"`
}

func (this *TimeRange) GetTime(raw string) time.Time {
	buf := new(bytes.Buffer)
	buf.WriteString(time.Now().Format("2006-01-02 "))
	buf.WriteString(raw)

	t, err := time.Parse("2006-01-02 03:04:05", buf.String())
	if err != nil {
		log.Error(err)
		return time.Now().Add(time.Hour * -24);
	}

	return t;
}

func (this *TimeRange) GetBegin() time.Time {
	return this.GetTime(this.BeginRaw)
}

func (this *TimeRange) GetEnd() time.Time {

}

type Config struct {
	Listen        string       `json:"listen"`
	WebRoot       string       `json:"web_root"`
	SSH           SSHItem      `json:"ssh"`
	AvailableTime []*TimeRange `json:"time-range"`
	save_path     string
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
		log.Error(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (c *Config) Save() error {
	file, err := os.Create(c.save_path)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()
	data, err2 := json.MarshalIndent(c, "", "    ")
	if err2 != nil {
		log.Error(err2)
		return err2
	}
	_, err3 := file.Write(data)
	if err3 != nil {
		log.Error(err3)
	}
	return err3
}

func (c *Config) GetDeployPath(deploy_type string) string {
	return "test"
}
