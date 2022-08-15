package exchanges

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPConfig struct {
	Addr     string `json:"addr" yaml:"addr"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	KeyPath  string `json:"keyPath" yaml:"keyPath"`
	Base     string `json:"base" yaml:"base"`
}

type SFTP struct {
	c    *sftp.Client
	base string
	url  string
}

func NewSFTP(config SFTPConfig) (Exchanger, error) {
	addr := config.Addr
	if !strings.ContainsRune(addr, ':') {
		addr = fmt.Sprintf("%s:22", addr)
	}

	var url string
	var auth []ssh.AuthMethod
	if config.Password != "" {
		auth = append(auth, ssh.Password(config.Password))
		url = fmt.Sprintf("sftp://%s@%s/%s", config.Username, config.Addr, config.Base)
	}
	if config.KeyPath != "" {
		key, err := ioutil.ReadFile(config.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("cannot load key file %s: %v", config.KeyPath, err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("invalid key file %s: %v", config.KeyPath, err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
		url = fmt.Sprintf("sftp://!%s@%s/%s", filepath.Base(config.KeyPath), config.Addr, config.Base)
	}
	if len(auth) == 0 {
		return nil, fmt.Errorf("no auth method provided for sftp connection to %s", config.Addr)
	}

	cc := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, cc)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s: %v", addr, err)
	}
	c, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("cannot create a sftp client for %s: %v", addr, err)
	}

	base := config.Base
	if base == "" {
		base = "/"
	}
	return &SFTP{c, base, url}, nil
}

func (s *SFTP) Read(name string, dest io.Writer) error {
	return nil
}

func (s *SFTP) Write(name string, source io.Reader) error {
	return nil
}

func (s *SFTP) Concat(name string, source []Source) error {
	return nil
}

func (s *SFTP) ReadDir(prefix string, opts ListOption) ([]fs.FileInfo, error) {
	dir, prefix := path.Split(prefix)
	result, err := s.c.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var infos []fs.FileInfo
	for _, item := range result {
		if strings.HasPrefix(item.Name(), prefix) {
			infos = append(infos, item)
		}
	}

	return infos, nil
}

func (s *SFTP) Stat(name string) (os.FileInfo, error) {
	return nil, nil
}

func (s *SFTP) Delete(name string) error {
	return nil
}

func (s *SFTP) Close() error {
	return nil
}

func (s *SFTP) String() string {
	return ""
}
