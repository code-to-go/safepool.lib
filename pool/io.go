package pool

import (
	"io"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
)

func (p *Pool) writeFile(name string, r io.Reader) (*security.HashStream, error) {
	hr, err := security.NewHashStream(r, nil)
	if core.IsErr(err, "cannot create hash reader: %v") {
		return nil, err
	}

	er, err := security.EncryptingReader(p.masterKeyId, p.keyFunc, hr)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = p.e.Write(name, er)
	return hr, err
}

func (p *Pool) readFile(name string, w io.Writer) (*security.HashStream, error) {
	hw, err := security.NewHashStream(nil, w)
	if core.IsErr(err, "cannot create hash stream: %v") {
		return nil, err
	}

	ew, err := security.DecryptingWriter(p.keyFunc, hw)
	if core.IsErr(err, "cannot create encrypting reader: %v") {
		return nil, err
	}

	err = p.e.Read(name, nil, ew)
	return hw, err
}
