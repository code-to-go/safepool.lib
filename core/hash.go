package core

import (
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

func HashFromFile(name string) (Hash256, error) {
	f, err := os.Open(name)
	if err != nil {
		return Hash256{}, err
	}
	return HashReader(f)
}

func HashReader(r io.Reader) (Hash256, error) {
	b, err := blake2b.New256(nil)
	if err != nil {
		return Hash256{}, err
	}
	_, err = io.Copy(b, r)
	if err != nil {
		return Hash256{}, err
	}
	return blake2b.Sum256(nil), nil
}
