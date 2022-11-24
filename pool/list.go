package pool

import (
	"bytes"
	"encoding/json"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
)

func (p *Pool) list(afterId uint64, afterTime time.Time) ([]Head, error) {
	hs, err := sqlGetHeads(p.Name, afterId, afterTime)
	if core.IsErr(err, "cannot read Pool heads: %v") {
		return nil, err
	}

	heads := map[uint64]Head{}
	for _, h := range hs {
		heads[h.Id] = h
	}

	fs, err := p.e.ReadDir(p.Name, 0)
	if core.IsErr(err, "cannot read content in pool %s/%s", p.e, p.Name) {
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

		h, err := p.readHead(path.Join(p.Name, name))
		if core.IsErr(err, "cannot read file %s from %s: %v", name, p.e) {
			continue
		}
		_ = sqlAddHead(p.Name, h)
		hs = append(hs, h)
	}
	return hs, nil
}

func (p *Pool) readHead(name string) (Head, error) {
	var b bytes.Buffer
	_, err := p.readFile(name, &b)
	if core.IsErr(err, "cannot read header of %s in %s: %v", name, p.e) {
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
