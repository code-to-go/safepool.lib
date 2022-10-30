package safe

import (
	"weshare/core"
	"weshare/security"
	"weshare/sql"
)

func sqlGetHeads(safe string, after uint64, limit int) ([]Head, error) {
	//GET_HEADS: SELECT id, name, modTime, size, hash FROM SafeHeads ORDER BY id DESC LIMIT :limit
	rows, err := sql.Query("GET_HEADS", sql.Args{"safe": safe, "after": after, "limit": limit})
	if core.IsErr(err, "cannot get safes heads from db: %v") {
		return nil, err
	}

	var heads []Head
	for rows.Next() {
		var h Head
		var modTime int64
		var hash string
		err = rows.Scan(&h.Id, &h.Name, &modTime, &h.Size, &hash)
		if !core.IsErr(err, "cannot read safe heads from db: %v") {
			continue
		}
		h.ModTime = sql.DecodeTime(modTime)
		heads = append(heads, h)
	}
	return heads, nil
}

func sqlAddHead(safe string, h Head) error {
	//ADD_HEAD: INSERT
	_, err := sql.Exec("ADD_HEAD", sql.Args{
		"safe":    safe,
		"id":      h.Id,
		"name":    h.Name,
		"size":    h.Size,
		"modTime": sql.EncodeTime(h.ModTime),
		"hash":    sql.EncodeBase64(h.Hash[:]),
	})
	return err
}

func (s *Safe) sqlGetKey(keyId uint64) []byte {
	rows, err := sql.Query("GET_KEY", sql.Args{"safe": s.Name, "keyId": keyId})
	if err != nil {
		return nil
	}
	for rows.Next() {
		var key string
		var err = rows.Scan(&key)
		if !core.IsErr(err, "cannot read key from db: %v") {
			return sql.DecodeBase64(key)
		}
	}
	return nil
}

func (s *Safe) sqlSetKey(keyId uint64, value []byte) error {
	_, err := sql.Exec("SET_KEY", sql.Args{"safe": s.Name, "keyId": keyId, "keyValue": sql.EncodeBase64(value)})
	return err
}

func (s *Safe) sqlGetKeystore() (Keystore, error) {
	rows, err := sql.Query("GET_KEYSTORE", sql.Args{"safe": s.Name})
	if core.IsErr(err, "cannot read keystore for safe %s: %v", s.Name) {
		return nil, err
	}
	ks := Keystore{}
	for rows.Next() {
		var keyId uint64
		var keyValue string
		var err = rows.Scan(&keyId, &keyValue)
		if !core.IsErr(err, "cannot read key from db: %v") {
			ks[keyId] = sql.DecodeBase64(keyValue)
		}
	}
	return ks, nil
}

func (s *Safe) sqlGetIdentities(onlyTrusted bool) ([]security.Identity, error) {
	var q string
	if onlyTrusted {
		q = "GET_TRUSTED_ON_SAFE"
	} else {
		q = "GET_IDENTITY_ON_SAFE"
	}

	rows, err := sql.Query(q, sql.Args{"safe": s.Name})
	if core.IsErr(err, "cannot get trusted identities from db: %v") {
		return nil, err
	}

	var is []security.Identity
	for rows.Next() {
		var encryptionKey, signatureKey, nick string
		err = rows.Scan(&signatureKey, &encryptionKey, &nick)
		if core.IsErr(err, "cannot read identity from db: %v") {
			continue
		}

		ek, _ := security.KeyFromBase64(encryptionKey)
		sk, _ := security.KeyFromBase64(signatureKey)
		is = append(is, security.Identity{
			Nick:          nick,
			SignatureKey:  sk,
			EncryptionKey: ek,
		})
	}
	return is, nil
}

func (s *Safe) sqlSetIdentity(i security.Identity) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("SET_IDENTITY_ON_SAFE", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"safe":          s.Name,
	})
	return err
}
