package security

import (
	"hash"
	"io"
	"weshare/core"

	"golang.org/x/crypto/blake2b"
)

//type Hash256 [blake2b.Size256]byte

type HashReader struct {
	r     io.Reader
	size  int64
	block hash.Hash
}

func NewHashReader(r io.Reader) (HashReader, error) {
	b, err := blake2b.New256(nil)
	if core.IsErr(err, "cannot create black hash: %v") {
		return HashReader{}, err
	}
	return HashReader{
		block: b,
		r:     r,
	}, nil
}

func (r HashReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if err == nil {
		n, err = r.block.Write(p[0:n])
	}
	r.size += int64(n)
	return n, err
}

func (r HashReader) Hash() []byte {
	return r.block.Sum(nil)
}

func (r HashReader) Size() int64 {
	return r.size
}
