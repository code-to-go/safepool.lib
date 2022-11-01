package safe

import (
	"path"
	"time"
	"weshare/core"
	"weshare/transport"
)

func (s *Safe) replica() {
	time.Sleep(ReplicaPeriod)
	for range s.ticker.C {
		for _, e := range s.exchangers {
			if e != s.e {
				err := s.syncAccess(e)
				if !core.IsErr(err, "cannot sync access between %s and %s: %v", s.e, e) {
					s.syncContent(e)
				}
			}
		}
	}
}

func (s *Safe) syncAccess(e transport.Exchanger) error {
	name := path.Join(s.Name, ".access")
	_, err := e.Stat(name)
	if err == nil {
		_, err := s.ImportAccess(e)
		if core.IsErr(err, "cannot import access file from %s: %v", e) {
			return err
		}
	}

	err = s.ExportAccessFile(e)
	if core.IsErr(err, "cannot export access file to %s: %v", e) {
		return err
	}
	return err
}

func (s *Safe) syncContent(e transport.Exchanger) error {
	ls, err := s.e.ReadDir(s.Name, 0)
	if core.IsErr(err, "cannot read file list from %s: %v", s.e) {
		return err
	}

	m := map[string]bool{}
	for _, l := range ls {
		n := l.Name()
		if n[0] != '.' {
			m[l.Name()] = true
		}
	}

	ls, _ = e.ReadDir(s.Name, 0)
	for _, l := range ls {
		n := l.Name()
		if n[0] != '.' && !m[n] {
			n = path.Join(s.Name, n)
			_ = transport.CopyFile(s.e, n, e, n)
		}
		delete(m, n)
	}

	for n := range m {
		n = path.Join(s.Name, n)
		err = transport.CopyFile(e, n, s.e, n)
		core.IsErr(err, "cannot clone '%s': %v", n)
	}

	return nil
}
