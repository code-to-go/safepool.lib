package access

import (
	"io"
	"weshare/core"
	"weshare/security"
)

func (t *Topic) writeFile(name string, r io.Reader) (*security.HashReader, error) {
	hr, err := security.NewHashReader(r)
	if core.IsErr(err, "cannot create hash reader: %v") {
		return nil, err
	}

	er, err := security.EncryptingReader(t.keyId, nil, hr)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = t.primary.Write(name, er)
	return &hr, err
}
