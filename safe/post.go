package safe

import (
	"io"
	"weshare/core"
	"weshare/security"
)

func (s *Safe) writeFile(name string, r io.Reader) (*security.HashReader, error) {
	hr, err := security.NewHashReader(r)
	if core.IsErr(err, "cannot create hash reader: %v") {
		return nil, err
	}

	er, err := security.EncryptingReader(s.masterKeyId, s.keyFunc, hr)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = s.e.Write(name, er)
	return &hr, err
}
