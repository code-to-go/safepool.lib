package safe

import (
	"weshare/core"
	"weshare/security"
	"weshare/sql"
)

func sqlGetHeads(safe string, after uint64) ([]Head, error) {
	rows, err := sql.Query("GET_HEADS", sql.Args{"safe": safe, "after": after})
	if core.IsErr(err, "cannot get safes heads from db: %v") {
		return nil, err
	}
	defer rows.Close()

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
	_, err := sql.Exec("SET_HEAD", sql.Args{
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
	defer rows.Close()

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
	defer rows.Close()

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

func (s *Safe) sqlGetIdentities(onlyTrusted bool) (identities []security.Identity, sinces []uint64, err error) {
	var q string
	if onlyTrusted {
		q = "GET_TRUSTED_ON_SAFE"
	} else {
		q = "GET_IDENTITY_ON_SAFE"
	}

	rows, err := sql.Query(q, sql.Args{"safe": s.Name})
	if core.IsErr(err, "cannot get trusted identities from db: %v") {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var encryptionKey, signatureKey, nick string
		var since uint64
		err = rows.Scan(&signatureKey, &encryptionKey, &nick, &since)
		if core.IsErr(err, "cannot read identity from db: %v") {
			continue
		}

		ek, _ := security.KeyFromBase64(encryptionKey)
		sk, _ := security.KeyFromBase64(signatureKey)
		identities = append(identities, security.Identity{
			Nick:          nick,
			SignatureKey:  sk,
			EncryptionKey: ek,
		})
		sinces = append(sinces, since)
	}
	return identities, sinces, nil
}

func (s *Safe) sqlSetIdentity(i security.Identity, since uint64) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("SET_IDENTITY_ON_SAFE", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"safe":          s.Name,
		"since":         since,
	})
	return err
}

func (s *Safe) sqlDeleteIdentity(i security.Identity) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("DEL_IDENTITY_ON_SAFE", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"safe":          s.Name,
	})
	return err
}
