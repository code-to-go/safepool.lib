package sql

import (
	_ "embed"
	"testing"
	"time"
	"weshare/core"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestDb(t *testing.T) {
	LoadSQLFromFile("sqlite.sql")
	err := OpenDB()
	assert.NoErrorf(t, err, "cannot open sqllite: %v", err)

	s, _, _, ok := GetConfig("", "identity.public")
	if !ok {
		err = SetConfig("", "identity.public", "test", 0, nil)
		assert.NoErrorf(t, err, "cannot set config: %v", err)
		s, _, _, ok = GetConfig("", "identity.public")
	}
	assert.Equal(t, "test", s, "cannot get config: %v", err)

	err = SetDomain("public.weshare.zone", nil)
	assert.NoErrorf(t, err, "cannot add domain: %v", err)

	domains, err := GetDomains()
	assert.NoErrorf(t, err, "cannot list domains: %v", err)
	assert.Contains(t, domains, "public.weshare.zone", "cannot find expected domain")

	now := time.Unix(time.Now().Unix(), 0)
	err = SetFile(core.File{
		Domain:  "public.weshare.zone",
		Name:    "test.txt",
		Author:  []byte("author"),
		Hash:    []byte("hash"),
		ModTime: now,
		State:   core.LocalCreated,
	})
	assert.NoErrorf(t, err, "cannot set file: %v", err)

	f, err := GetFile("public.weshare.zone", "test.txt", []byte("author"))
	assert.NoErrorf(t, err, "cannot get file: %v", err)
	assert.Equal(t, now, f.ModTime)

	err = CloseDB()
	assert.NoErrorf(t, err, "cannot close sqllite: %v", err)
}
