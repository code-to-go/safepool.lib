package stores

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SFTP *SFTPConfig `json:"sftp,omitempty" yaml:"sftp,omitempty"`
	S3   *S3Config   `json:"s3,omitempty" yaml:"s3,omitempty"`
}

func ReadConfig(name string) (Config, error) {
	var c Config
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	return c, err
}
