package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

func Generate32BytesKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return key, err
}

type StreamReader struct {
	loc    int
	header []byte
	r      cipher.StreamReader
}

func (sr *StreamReader) Read(p []byte) (n int, err error) {
	if sr.loc < 8+aes.BlockSize {
		n := copy(p[sr.loc:], sr.header)
		sr.loc += n
		return n, nil
	} else {
		return sr.r.Read(p)
	}
}

// EncryptedWriter wraps w with an OFB cipher stream.
func EncryptingReader(keyId uint64, keyFunc func(uint64) []byte, r io.Reader) (*StreamReader, error) {

	header := make([]byte, 8+aes.BlockSize)
	binary.LittleEndian.PutUint64(header, keyId)

	// generate random initial value
	if _, err := io.ReadFull(rand.Reader, header[8:]); err != nil {
		return nil, err
	}

	value := keyFunc(keyId)
	if value == nil {
		return nil, errors.New("unknown encryption key")
	}

	block, err := newBlock(value)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewOFB(block, header[8:])
	return &StreamReader{
		header: header,
		r:      cipher.StreamReader{S: stream, R: r},
	}, nil
}

type StreamWriter struct {
	loc     int
	header  []byte
	keyFunc func(uint64) []byte
	w       *cipher.StreamWriter
}

func (sr *StreamWriter) Write(p []byte) (n int, err error) {
	if sr.w.S == nil {
		n := copy(sr.header[sr.loc:], p)
		sr.loc += n

		if sr.loc == 8+aes.BlockSize {
			keyId := binary.LittleEndian.Uint64(sr.header)
			value := sr.keyFunc(keyId)
			if value == nil {
				return 0, errors.New("unknown encryption key")
			}

			block, err := newBlock(value)
			if err != nil {
				return 0, err
			}

			iv := sr.header[8:]
			sr.w.S = cipher.NewOFB(block, iv)
		}
		return n, nil
	} else {
		return sr.w.Write(p)
	}
}

// EncryptedWriter wraps w with an OFB cipher stream.
func DecryptingWriter(keyFunc func(uint64) []byte, w io.Writer) (*StreamWriter, error) {
	return &StreamWriter{
		keyFunc: keyFunc,
		header:  make([]byte, 8+aes.BlockSize),
		w:       &cipher.StreamWriter{S: nil, W: w},
	}, nil
}

// EncryptedWriter wraps w with an OFB cipher stream.
func EncryptedWriter(keyId uint64, keyFunc func(uint64) []byte, w io.Writer) (*cipher.StreamWriter, error) {

	// generate random initial value
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// write clear IV to allow for decryption
	n, err := w.Write(iv)
	if err != nil || n != len(iv) {
		return nil, errors.New("could not write initial value")
	}

	keyIdBuf := make([]byte, 4)
	binary.LittleEndian.PutUint64(keyIdBuf, keyId)
	n, err = w.Write(keyIdBuf)
	if err != nil || n != len(keyIdBuf) {
		return nil, errors.New("could not write key id")
	}

	value := keyFunc(keyId)
	if value == nil {
		return nil, errors.New("unknown encryption key")
	}

	block, err := newBlock(value)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewOFB(block, iv)
	return &cipher.StreamWriter{S: stream, W: w}, nil
}

// EncryptedReader wraps r with an OFB cipher stream.
func EncryptedReader(keyFunc func(uint64) []byte, r io.Reader) (*cipher.StreamReader, error) {

	// read initial value
	iv := make([]byte, aes.BlockSize)
	n, err := r.Read(iv)
	if err != nil || n != len(iv) {
		return nil, errors.New("could not read initial value")
	}

	keyIdB := make([]byte, 4)
	n, err = r.Read(keyIdB)
	if err != nil || n != len(keyIdB) {
		return nil, errors.New("could not read key id")
	}
	keyId := binary.LittleEndian.Uint64(keyIdB)
	value := keyFunc(keyId)
	if value == nil {
		return nil, errors.New("unknown encryption key")
	}

	block, err := newBlock(value)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewOFB(block, iv)
	return &cipher.StreamReader{S: stream, R: r}, nil
}

func newBlock(key []byte) (cipher.Block, error) {
	sh := sha256.Sum256(key)
	hash := md5.Sum(sh[:])
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	return block, nil
}
