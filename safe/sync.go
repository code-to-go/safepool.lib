package safe

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"weshare/core"
	"weshare/security"
)

func (s *Safe) refresh() error {
	heads, err := sqlGetHeads(s.Name, 0, 1)
	if core.IsErr(err, "cannot read Safe heads: %v") {
		return err
	}

	var after uint64
	if len(heads) > 0 {
		after = heads[0].Id
	}

	fs, err := s.e.ReadDir(s.Name, 0)
	if core.IsErr(err, "cannot read content in safe %s/%s", s.e, s.Name) {
		return err
	}
	for _, f := range fs {
		name := f.Name()
		if !strings.HasSuffix(name, ".head") {
			continue
		}

		id, err := strconv.ParseInt(name, 10, 64)
		if err != nil || uint64(id) <= after {
			continue
		}

		h, err := s.readHead(name)
		if core.IsErr(err, "cannot read file %s from %s: %v", name, s.e) {
			continue
		}
		_ = sqlAddHead(s.Name, h)
	}
	return nil
}

func (s *Safe) readHead(name string) (Head, error) {
	var b bytes.Buffer
	err := s.e.Read(name, nil, &b)
	if core.IsErr(err, "cannot read header of %s in %s: %v", name, s.e) {
		return Head{}, err
	}

	var h Head
	err = json.Unmarshal(b.Bytes(), &h)
	if core.IsErr(err, "corrupted header for file %s", name) {
		return Head{}, err
	}

	if !security.Verify(h.Author, h.Hash, h.Signature) {
		return Head{}, ErrNoExchange
	}

	return Head{}, err
}
