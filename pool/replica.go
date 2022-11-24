package pool

import (
	"path"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/transport"
)

func (p *Pool) replica() {
	for _, e := range p.exchangers {
		if e != p.e {
			err := p.syncAccess(e)
			if !core.IsErr(err, "cannot sync access between %s and %s: %v", p.e, e) {
				p.syncContent(e)
			}
		}
	}
}

func (p *Pool) syncAccess(e transport.Exchanger) error {
	name := path.Join(p.Name, ".access")
	_, err := e.Stat(name)
	if err == nil {
		_, err := p.ImportAccess(e)
		if core.IsErr(err, "cannot import access file from %s: %v", e) {
			return err
		}
	}

	err = p.ExportAccessFile(e)
	if core.IsErr(err, "cannot export access file to %s: %v", e) {
		return err
	}
	return err
}

func (p *Pool) syncContent(e transport.Exchanger) error {
	ls, err := p.e.ReadDir(p.Name, 0)
	if core.IsErr(err, "cannot read file list from %s: %v", p.e) {
		return err
	}

	m := map[string]bool{}
	for _, l := range ls {
		n := l.Name()
		if n[0] != '.' {
			m[l.Name()] = true
		}
	}

	ls, _ = e.ReadDir(p.Name, 0)
	for _, l := range ls {
		n := l.Name()
		if n[0] != '.' && !m[n] {
			n = path.Join(p.Name, n)
			_ = transport.CopyFile(p.e, n, e, n)
		}
		delete(m, n)
	}

	for n := range m {
		n = path.Join(p.Name, n)
		err = transport.CopyFile(e, n, p.e, n)
		core.IsErr(err, "cannot clone '%s': %v", n)
	}

	return nil
}
