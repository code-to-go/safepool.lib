package access

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"weshare/core"
	"weshare/security"
)

func (t *Topic) refresh() error {
	heads, err := sqlGetHeads(t.Name, 0, 1)
	if core.IsErr(err, "cannot read Topic heads: %v") {
		return err
	}

	var after uint64
	if len(heads) > 0 {
		after = heads[0].Id
	}

	fs, err := t.primary.ReadDir(t.Name, 0)
	if core.IsErr(err, "cannot read content in topic %s/%s", t.primary, t.Name) {
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

		h, err := t.readHead(name)
		if core.IsErr(err, "cannot read file %s from %s: %v", name, t.primary) {
			continue
		}
		_ = sqlAddHead(t.Name, h)
	}
	return nil
}

func (t *Topic) readHead(name string) (Head, error) {
	var b bytes.Buffer
	err := t.primary.Read(name, nil, &b)
	if core.IsErr(err, "cannot read header of %s in %s: %v", name, t.primary) {
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
