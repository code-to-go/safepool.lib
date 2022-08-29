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

// EncryptedWriter wraps w with an OFB cipher stream.
func EncryptedWriter(keyId uint32, keyValue []byte, w io.Writer) (*cipher.StreamWriter, error) {

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
	binary.LittleEndian.PutUint32(keyIdBuf, keyId)
	n, err = w.Write(keyIdBuf)
	if err != nil || n != len(keyIdBuf) {
		return nil, errors.New("could not write key id")
	}

	block, err := newBlock(keyValue)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewOFB(block, iv)
	return &cipher.StreamWriter{S: stream, W: w}, nil
}

// EncryptedReader wraps r with an OFB cipher stream.
func EncryptedReader(keys map[uint32][]byte, r io.Reader) (*cipher.StreamReader, error) {

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
	keyId := binary.LittleEndian.Uint32(keyIdB)
	value, ok := keys[keyId]
	if !ok {
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
