package access

import (
	"weshare/core"
	"weshare/sql"
)

func sqlGetHeads(topic string, after uint64, limit int) ([]Head, error) {
	//GET_HEADS: SELECT id, name, modTime, size, hash FROM TopicHeads ORDER BY id DESC LIMIT :limit
	rows, err := sql.Query("GET_HEADS", sql.Args{"topic": topic, "after": after, "limit": limit})
	if core.IsErr(err, "cannot get topics heads from db: %v") {
		return nil, err
	}

	var heads []Head
	for rows.Next() {
		var h Head
		var modTime int64
		var hash string
		err = rows.Scan(&h.Id, &h.Name, &modTime, &h.Size, &hash)
		if !core.IsErr(err, "cannot read topic heads from db: %v") {
			continue
		}
		h.ModTime = sql.DecodeTime(modTime)
		heads = append(heads, h)
	}
	return heads, nil
}

func sqlAddHead(topic string, h Head) error {
	//ADD_HEAD: INSERT
	_, err := sql.Exec("ADD_HEAD", sql.Args{
		"topic":   topic,
		"id":      h.Id,
		"name":    h.Name,
		"size":    h.Size,
		"modTime": sql.EncodeTime(h.ModTime),
		"hash":    sql.EncodeBase64(h.Hash[:]),
	})
	return err
}

func sqlGetKey(topic string, keyId uint64) []byte {
	return nil
}

func sqlSetKey(topic string, keyId uint64, value []byte) error {
	return nil
}
