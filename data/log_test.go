package data

import (
	"baobab/access"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func TestMarshalling(t *testing.T) {
	identity, err := access.NewIdentity()
	assert.NoErrorf(t, err, "Cannot create identity: %v", err)

	hash, _ := blake2b.New256(nil)
	var fileId access.Hash256
	copy(hash.Sum(nil), fileId[:])

	now := time.Now()
	log := Log{
		Version: 1,
		Time:    now,
		UserId:  identity.Public,
		Entries: []LogEntry{
			LogEntry{FileId: fileId},
		},
	}

	data, err := log.Marshal(identity)
	assert.NoErrorf(t, err, "Cannot marshal log: %v", err)

	_, err = log.Unmarshal(data)
	assert.NoErrorf(t, err, "Cannot unmarshal log: %v", err)

}
