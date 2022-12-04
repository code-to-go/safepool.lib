package transport

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-to-go/safepool.lib/core"
)

type LocalConfig struct {
	Base string `json:"base" yaml:"base"`
}

type Local struct {
	base  string
	url   string
	touch map[string]time.Time
}

func NewLocal(config LocalConfig) (Exchanger, error) {
	base := config.Base
	if base == "" {
		base = "/"
	}
	return &Local{base, "file://" + base, map[string]time.Time{}}, nil
}

func (l *Local) Touched(name string) bool {
	touchFile := path.Join(l.base, fmt.Sprintf("%s.touch", name))
	stat, err := l.Stat(touchFile)
	touched := err != nil || stat.ModTime().After(l.touch[name])
	if touched {
		if !core.IsErr(l.Write(touchFile, &bytes.Buffer{}), "cannot write touch file: %v") {
			l.touch[name] = stat.ModTime()
		}
	}
	return touched
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

func createDir(n string) error {
	return os.MkdirAll(filepath.Dir(n), 0755)
}

func (l *Local) Write(name string, source io.Reader) error {
	n := filepath.Join(l.base, name)
	err := createDir(n)
	if core.IsErr(err, "cannot create parent of %s: %v", n) {
		return err
	}

	f, err := os.Create(n)
	if core.IsErr(err, "cannot create file on %v:%v", l) {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, source)
	core.IsErr(err, "cannot copy file on %v:%v", l)
	return err
}

func (l *Local) ReadDir(name string, opts ListOption) ([]fs.FileInfo, error) {
	var dir, prefix string
	if opts&IsPrefix > 0 {
		dir, prefix = filepath.Split(filepath.Join(l.base, name))
	} else {
		dir = filepath.Join(l.base, name)
	}
	result, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var infos []fs.FileInfo
	for _, item := range result {
		if strings.HasPrefix(item.Name(), prefix) {
			info, err := item.Info()
			if err == nil {
				infos = append(infos, info)
			}
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
