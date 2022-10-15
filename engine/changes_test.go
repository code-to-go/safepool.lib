package engine

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"weshare/algo"
	"weshare/model"
	"weshare/security"
	"weshare/sql"

	"github.com/stretchr/testify/assert"
)

func TestChanges(t *testing.T) {
	sql.DbName = "weshare.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	Start()

	domain := "test.weshare.zone"
	sql.SetEncKey(domain, 1, []byte("sample"))
	sql.SetUser(domain, model.User{
		Identity: Self,
		Active:   true,
	})

	data := make([]byte, 128*1024)
	rand.Read(data)
	os.RemoveAll(filepath.Join(WesharePath, domain))
	os.MkdirAll(filepath.Join(WesharePath, domain), 0755)
	filename := filepath.Join(WesharePath, domain, "sample.txt")
	ioutil.WriteFile(filename, data, 0644)

	data[271] = '$'

	blocks, err := algo.HashSplit(bytes.NewBuffer(data), 13, nil)
	assert.NoErrorf(t, err, "Cannot hash split: %v", err)

	f := model.File{
		Domain: domain,
		Name:   "sample.txt",
	}
	changeFile, f, err := GenerateChangeFile(f, blocks)
	assert.NoErrorf(t, err, "Cannot generate file: %v", err)

	h, err := StatChangeFile(changeFile)
	assert.NoErrorf(t, err, "Cannot stat change file: %v", err)
	assert.Equal(t, domain, h.Domain)
	assert.Equal(t, "sample.txt", h.Name)
	assert.Equal(t, Self.Keys[security.Ed25519].Public, h.Author)

	os.RemoveAll(filepath.Join(WesharePath, domain))

}
