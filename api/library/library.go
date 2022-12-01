package library

import (
	"bytes"
	"strings"
	"time"

	"github.com/code-to-go/safepool.lib/core"
	pool "github.com/code-to-go/safepool.lib/pool"
	"github.com/code-to-go/safepool.lib/security"
	"github.com/code-to-go/safepool.lib/transport"
	"github.com/wailsapp/mimetype"
)

type State int

const (
	StateSync State = iota
	StateIn
	StateOut
	StateAlt
)

// Document includes information about a file stored on the library. Most information refers on the synchronized state with the exchange
type Document struct {
	Id          uint64
	Name        string
	Size        uint64
	ModTime     time.Time
	Author      security.Identity
	ContentType string
	Hash        []byte

	// LocalPath is the location on the local storage where the document
	LocalPath string
	// HasChanges is true when the location on the local storage is different than the last available on the exchange
	HasChanged bool
}

type Library struct {
	Pool *pool.Pool
}

func Get(p *pool.Pool) Library {
	return Library{
		Pool: p,
	}
}

func (l *Library) List(beforeId uint64, limit int) ([]Document, error) {
	l.sync()
	return sqlGetDocuments(l.Pool.Name, beforeId, limit)
}

func (l *Library) Download(id uint64, localPath string) error {
	return nil
}

func (l *Library) Upload(id uint64) error {
	return nil
}

func (l *Library) sync() {
	l.Pool.Sync()

}

func (l *Library) TimeOffset(p *pool.Pool) time.Time {
	return sqlGetOffset(p.Name)
}

func (l *Library) Accept(p *pool.Pool, head pool.Head) bool {
	name := head.Name
	if !strings.HasPrefix(name, "/library/") {
		return false
	}

	var buf bytes.Buffer
	err := p.Get(head.Id, &transport.Range{From: 0, To: 1024}, &buf)
	if core.IsErr(err, "cannot get file to detect content type: %v") {
		return false
	}

	contentType := mimetype.Detect(buf.Bytes()).String()
	d := Document{
		Id:          head.Id,
		Name:        head.Name,
		ModTime:     head.ModTime,
		Size:        uint64(head.Size),
		Author:      head.Author,
		ContentType: contentType,
		Hash:        head.Hash,
	}

	return !core.IsErr(sqlSetDocument(p.Name, d), "cannot save document to db: %v")
}
