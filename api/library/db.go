package library

import (
	"time"

	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/security"
	"github.com/code-to-go/safepool.lib/sql"
)

func sqlSetDocument(pool string, d Document) error {
	author, _ := d.Author.Public().Base64()

	_, err := sql.Exec("SET_DOCUMENT", sql.Args{"pool": pool, "id": d.Id, "name": d.Name,
		"size": d.Size, "modTime": sql.EncodeTime(d.ModTime), "author": author,
		"contentType": d.ContentType,
		"hash":        sql.EncodeBase64(d.Hash), "localPath": d.LocalPath, "hasChanged": d.HasChanged,
		"ts": sql.EncodeTime(time.Now()),
	})
	core.IsErr(err, "cannot set document %d on db: %v", d.Name)
	return err
}

func sqlDocumentByName(pool string, name string) (Document, error) {
	d := Document{Name: name}

	var modTime int64
	var author string
	var hash string
	err := sql.QueryRow("GET_LOCAL_DOCUMENT", sql.Args{"pool": pool, "name": name}, &d.Id, &d.Size, &modTime, &author,
		&d.ContentType, &hash, &d.LocalPath, &d.HasChanged)
	if err == nil {
		d.Author, _ = security.IdentityFromBase64(author)
		d.Hash = sql.DecodeBase64(hash)
	}
	return d, err
}

func sqlGetDocuments(pool string, beforeId uint64, limit int) ([]Document, error) {
	rows, err := sql.Query("GET_DOCUMENTS", sql.Args{"pool": pool, "beforeId": beforeId, "limit": limit})
	if core.IsErr(err, "cannot query documents from db: %v") {
		return nil, err
	}
	var documents []Document
	for rows.Next() {
		var d Document
		var modTime int64
		var hash string
		err = rows.Scan(&d.Id, &d.Name, &d.Size, &modTime, &d.Author, &d.ContentType, &hash, &d.LocalPath, &d.HasChanged)
		if !core.IsErr(err, "cannot scan row in Documents: %v", err) {
			d.ModTime = sql.DecodeTime(modTime)
			d.Hash = sql.DecodeBase64(hash)
			documents = append(documents, d)
		}
	}
	return documents, nil
}

func sqlGetOffset(pool string) time.Time {
	var ts int64
	err := sql.QueryRow("GET_DOCUMENTS_OFFSET", sql.Args{"pool": pool}, &ts)
	if err == nil {
		return sql.DecodeTime(ts)
	} else {
		return time.Time{}
	}
}
