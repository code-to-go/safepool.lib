package transport

import (
	"fmt"
	"io"
	"io/fs"

	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/code-to-go/safepool/core"

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
		key, err := os.ReadFile(config.KeyPath)
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

func (s *SFTP) Read(name string, rang *Range, dest io.Writer) error {
	f, err := s.c.Open(path.Join(s.base, name))
	if core.IsErr(err, "cannot open file on sftp server %v:%v", s) {
		return err
	}

	if rang == nil {
		_, err = io.Copy(dest, f)
	} else {
		left := rang.To - rang.From
		f.Seek(rang.From, 0)
		var b [4096]byte

		for left > 0 && err == nil {
			var sz int64
			if rang.From-rang.To > 4096 {
				sz = 4096
			} else {
				sz = rang.From - rang.To
			}
			_, err = f.Read(b[0:sz])
			dest.Write(b[0:sz])
			left -= sz
		}
	}
	if core.IsErr(err, "cannot read from %s/%s:%v", s, name) {
		return err
	}

	return nil
}

func (s *SFTP) Write(name string, source io.Reader) error {
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
	return s.c.Stat(path.Join(s.base, name))
}

func (s *SFTP) Rename(old, new string) error {
	return s.c.Rename(path.Join(s.base, old), path.Join(s.base, new))
}

func (s *SFTP) Delete(name string) error {
	return s.c.Remove(path.Join(s.base, name))
}

func (s *SFTP) Close() error {
	return s.c.Close()
}

func (s *SFTP) String() string {
	return s.url
}
