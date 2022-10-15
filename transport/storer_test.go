package transport

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testCreateFile(t *testing.T, s Exchanger) {
	name := uuid.New().String()
	r := bytes.NewReader(make([]byte, 1024))
	assert.NoErrorf(t, s.Write(name, r), "cannot write file %s", name)
	assert.NoErrorf(t, s.Delete(name), "cannot delete file %s", name)
}

func TestCreateFile(t *testing.T) {
	c, err := ReadConfig("../../credentials/s3-2.yaml")
	assert.NoErrorf(t, err, "Cannot load SFTP config: %v", err)

	fs, err := NewExchanger(c)
	assert.NoErrorf(t, err, "Cannot load SFTP config: %v", err)
	testCreateFile(t, fs)

}
