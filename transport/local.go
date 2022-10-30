package transport

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"weshare/core"
)

type LocalConfig struct {
	Base string `json:"base" yaml:"base"`
}

type Local struct {
	base string
	url  string
}

func NewLocal(config LocalConfig) (Exchanger, error) {
	base := config.Base
	if base == "" {
		base = "/"
	}
	return &Local{base, "file://" + base}, nil
}

func (l *Local) Read(name string, rang *Range, dest io.Writer) error {
	f, err := os.Open(path.Join(l.base, name))
	if core.IsErr(err, "cannot open file on %v:%v", l) {
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
	if core.IsErr(err, "cannot read from %s/%s:%v", l, name) {
		return err
	}

	return nil
}

func (l *Local) Write(name string, source io.Reader) error {
	f, err := os.Create(path.Join(l.base, name))
	if core.IsErr(err, "cannot create file on %v:%v", l) {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, source)
	core.IsErr(err, "cannot copy file on %v:%v", l)
	return err
}

func (l *Local) ReadDir(prefix string, opts ListOption) ([]fs.FileInfo, error) {
	dir, prefix := path.Split(path.Join(l.base, prefix))
	result, err := ioutil.ReadDir(dir)
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

func (l *Local) Stat(name string) (os.FileInfo, error) {
	return os.Stat(path.Join(l.base, name))
}

func (l *Local) Rename(old, new string) error {
	return os.Rename(path.Join(l.base, old), path.Join(l.base, new))
}

func (l *Local) Delete(name string) error {
	return os.Remove(path.Join(l.base, name))
}

func (l *Local) Close() error {
	return nil
}

func (l *Local) String() string {
	return l.url
}
