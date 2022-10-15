package transport

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SFTP  *SFTPConfig  `json:"sftp,omitempty" yaml:"sftp,omitempty"`
	S3    *S3Config    `json:"s3,omitempty" yaml:"s3,omitempty"`
	Local *LocalConfig `json:"local,omitempty" yaml:"local,omitempty"`
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

var SampleConfig = Config{
	SFTP: &SFTPConfig{
		Addr:     "hostname",
		Username: "username",
		Password: "password when not using private key authentication",
		KeyPath:  "path when using private key authentication",
		Base:     "local path on the sftp server",
	},
	S3: &S3Config{
		Region:    "AWS region, i.e. eu-central-1",
		Endpoint:  "S3 endpoint, i.e. s3.eu-central-1.amazonaws.com",
		Bucket:    "S3 bucket name",
		AccessKey: "S3 access key",
		Secret:    "S3 secret key",
	},
}
