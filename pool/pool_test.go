package pool

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/code-to-go/safepool/security"
	"github.com/code-to-go/safepool/sql"
	"github.com/code-to-go/safepool/transport"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSafeCreation(t *testing.T) {
	sql.DbName = "safepool.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	err := sql.OpenDB()
	assert.NoErrorf(t, err, "cannot open db")

	self, err := security.NewIdentity("test")
	assert.NoErrorf(t, err, "cannot create identity")

	c, err := transport.ReadConfig("../../credentials/s3-2.yaml")
	assert.NoErrorf(t, err, "Cannot load S3 config: %v", err)

	err = Define(Config{"test.safepool.net/public", []transport.Config{c}})
	assert.NoErrorf(t, err, "Cannot define pool: %v", err)

	ForceCreation = true
	ReplicaPeriod = 0
	s, err := Create(self, "test.safepool.net/public")
	assert.NoErrorf(t, err, "Cannot create pool: %v", err)
	s.Close()

	s, err = Open(self, "test.safepool.net/public")
	assert.NoErrorf(t, err, "Cannot open pool: %v", err)
	defer s.Close()

	s1 := "just a simple test"
	h, err := s.Post("test.txt", bytes.NewBufferString(s1), nil)
	assert.NoErrorf(t, err, "Cannot create post: %v", err)

	b2 := bytes.Buffer{}
	err = s.Get(h.Id, nil, &b2)
	assert.NoErrorf(t, err, "Cannot get %d: %v", h.Id, err)

	s2 := b2.String()
	assert.EqualValues(t, s1, s2)

	for _, h := range s.List(uint64(0), time.Time{}) {
		fmt.Printf("\t%s\t%d\t%d", h.Name, h.Size, h.Id)
	}
	s.Delete()
}

func BenchmarkSafe(b *testing.B) {
	sql.DbName = "safepool.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	err := sql.OpenDB()
	assert.NoErrorf(b, err, "cannot open db")

	self, err := security.NewIdentity("test")
	assert.NoErrorf(b, err, "cannot create identity")

	//c, err := transport.ReadConfig("../../credentials/s3-2.yaml")
	c, err := transport.ReadConfig("../../credentials/local.yaml")
	assert.NoErrorf(b, err, "Cannot load S3 config: %v", err)

	err = Define(Config{"test.safepool.net/public", []transport.Config{c}})
	assert.NoErrorf(b, err, "Cannot define pool: %v", err)

	ForceCreation = true
	ReplicaPeriod = 0
	s, err := Create(self, "test.safepool.net/public")
	assert.NoErrorf(b, err, "Cannot create pool: %v", err)
	s.Close()

	s, err = Open(self, "test.safepool.net/public")
	assert.NoErrorf(b, err, "Cannot open pool: %v", err)
	defer s.Close()

	s1 := "just a simple test"
	h, err := s.Post("test.txt", bytes.NewBufferString(s1), nil)
	assert.NoErrorf(b, err, "Cannot create post: %v", err)

	b2 := bytes.Buffer{}
	err = s.Get(h.Id, nil, &b2)
	assert.NoErrorf(b, err, "Cannot get %d: %v", h.Id, err)

	s2 := b2.String()
	assert.EqualValues(b, s1, s2)

	for _, h := range s.List(uint64(0), time.Time{}) {
		fmt.Printf("\t%s\t%d\t%d", h.Name, h.Size, h.Id)
	}
	s.Delete()
}

func TestSafeReplica(t *testing.T) {
	sql.DbName = "safepool.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	err := sql.OpenDB()
	assert.NoErrorf(t, err, "cannot open db")

	self, err := security.NewIdentity("test")
	assert.NoErrorf(t, err, "cannot create identity")

	s3, err := transport.ReadConfig("../../credentials/s3-2.yaml")
	assert.NoErrorf(t, err, "Cannot load S3 config: %v", err)
	local, err := transport.ReadConfig("../../credentials/local.yaml")
	assert.NoErrorf(t, err, "Cannot load local config: %v", err)

	err = Define(Config{"test.safepool.net/public", []transport.Config{s3, local}})
	assert.NoErrorf(t, err, "Cannot define pool: %v", err)

	ForceCreation = true
	ReplicaPeriod = time.Second * 5

	now := time.Now()
	s, err := Create(self, "test.safepool.net/public")
	creationTime := time.Since(now)
	assert.NoErrorf(t, err, "Cannot create pool: %v", err)
	defer s.Close()
	defer s.Delete()

	s1 := "just a simple test"
	now = time.Now()
	_, err = s.Post("test.txt", bytes.NewBufferString(s1), nil)
	postTime := time.Since(now)
	assert.NoErrorf(t, err, "Cannot create post: %v", err)

	time.Sleep(5 * time.Minute)

	fmt.Printf("creation: %s, post: %s\n", creationTime, postTime)
}
