package safe

import (
	"bytes"
	"encoding/json"
	"path"
	"strconv"
	"strings"
	"weshare/core"
	"weshare/security"
)

func (s *Safe) list(after uint64) ([]Head, error) {
	hs, err := sqlGetHeads(s.Name, after)
	if core.IsErr(err, "cannot read Safe heads: %v") {
		return nil, err
	}

	heads := map[uint64]Head{}
	for _, h := range hs {
		heads[h.Id] = h
	}

	fs, err := s.e.ReadDir(s.Name, 0)
	if core.IsErr(err, "cannot read content in safe %s/%s", s.e, s.Name) {
		return hs, err
	}
	for _, f := range fs {
		name := f.Name()
		if !strings.HasSuffix(name, ".head") {
			continue
		}

		id, _ := strconv.ParseInt(name, 10, 64)
		if _, found := heads[uint64(id)]; found {
			continue
		}

		h, err := s.readHead(path.Join(s.Name, name))
		if core.IsErr(err, "cannot read file %s from %s: %v", name, s.e) {
			continue
		}
		_ = sqlAddHead(s.Name, h)
		hs = append(hs, h)
	}
	return hs, nil
}

func (s *Safe) readHead(name string) (Head, error) {
	var b bytes.Buffer
	_, err := s.readFile(name, &b)
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

	return h, err
}
