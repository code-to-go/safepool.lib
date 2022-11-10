package safe

import (
	"io"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
)

func (s *Safe) writeFile(name string, r io.Reader) (*security.HashStream, error) {
	hr, err := security.NewHashStream(r, nil)
	if core.IsErr(err, "cannot create hash reader: %v") {
		return nil, err
	}

	er, err := security.EncryptingReader(s.masterKeyId, s.keyFunc, hr)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = s.e.Write(name, er)
	return hr, err
}

func (s *Safe) readFile(name string, w io.Writer) (*security.HashStream, error) {
	hw, err := security.NewHashStream(nil, w)
	if core.IsErr(err, "cannot create hash stream: %v") {
		return nil, err
	}

	ew, err := security.DecryptingWriter(s.keyFunc, hw)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = s.e.Read(name, nil, ew)
	return hw, err
}
