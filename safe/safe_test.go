package safe

import (
	"bytes"
	"testing"
	"weshare/security"
	"weshare/sql"
	"weshare/transport"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSafeCreation(t *testing.T) {
	sql.DbName = "weshare.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	err := sql.OpenDB()
	assert.NoErrorf(t, err, "cannot open db")

	self, err := security.NewIdentity("test")
	assert.NoErrorf(t, err, "cannot create identity")

	c, err := transport.ReadConfig("../../credentials/s3-2.yaml")
	assert.NoErrorf(t, err, "Cannot load S3 config: %v", err)

	ForceCreation = true
	tp, err := CreateSafe(self, "test.weshare.net/public", []transport.Config{c})
	assert.NoErrorf(t, err, "Cannot create safe: %v", err)
	tp.Close()

	tp, err = OpenSafe(self, "test.weshare.net/public", []transport.Config{c})
	assert.NoErrorf(t, err, "Cannot open safe: %v", err)
	defer tp.Close()

	_, err = tp.Post("test.txt", bytes.NewBufferString("just a simple test"))
	assert.NoErrorf(t, err, "Cannot create post: %v", err)

}
