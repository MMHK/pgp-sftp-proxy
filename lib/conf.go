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

	//log.Debug(buf.String())
	timezone, err := time.LoadLocation("Asia/Hong_Kong")
	if err != nil {
		log.Error(err)
		timezone = time.FixedZone("GMT", 8)
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", buf.String(), timezone)
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
	return this.GetTime(this.EndRaw)
}

func CanUpload(settings []*TimeRange) bool {
	now := time.Now()

	for _, setting := range settings {
		//log.Debug(now, setting.GetBegin(), setting.GetEnd())
		if now.After(setting.GetBegin()) && now.Before(setting.GetEnd()) {
			return false
		}
	}

	return true
}

type SftpOptions struct {
	DownloadDir string `json:"download-dir"`
	UploadDir   string `json:"upload-dir"`
}

type PGPOption struct {
	PublicKeyPath  string `json:"public-key"`
	PrivateKeyPath string `json:"private-key"`
}

type AWSOption struct {
	Region    string `json:"region"`
	AccessKey string `json:"access-key"`
	SecretKey string `json:"secret-key"`
	Bucket    string `json:"bucket"`
}

type Config struct {
	Listen        string       `json:"listen"`
	WebRoot       string       `json:"web_root"`
	TempDir       string       `json:"tmp_path"`
	SSH           SSHItem      `json:"ssh"`
	SFTP          *SftpOptions `json:"sftp"`
	PGP           *PGPOption   `json:"pgp"`
	AWS           *AWSOption   `json:"aws"`
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
