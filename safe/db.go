package safe

import (
	"encoding/json"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
	"github.com/code-to-go/safepool/sql"
	"github.com/code-to-go/safepool/transport"
	"time"
)

func sqlGetHeads(safe string, afterId uint64, afterTime time.Time) ([]Head, error) {
	rows, err := sql.Query("GET_HEADS", sql.Args{"safe": safe, "afterId": afterId, "afterTime": sql.EncodeTime(afterTime)})
	if core.IsErr(err, "cannot get safes heads from db: %v") {
		return nil, err
	}
	defer rows.Close()

	var heads []Head
	for rows.Next() {
		var h Head
		var modTime int64
		var ts int64
		var hash string
		err = rows.Scan(&h.Id, &h.Name, &modTime, &h.Size, &hash, &ts)
		if !core.IsErr(err, "cannot read safe heads from db: %v") {
			continue
		}
		h.ModTime = sql.DecodeTime(modTime)
		h.TimeStamp = sql.DecodeTime(ts)
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
		"ts":      sql.EncodeTime(time.Now()),
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

func (s *Safe) sqlGetIdentities(onlyTrusted bool) (identities []Identity, err error) {
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
	defer rows.Close()

	for rows.Next() {
		var encryptionKey, signatureKey, nick string
		var since uint64
		var ts int64
		err = rows.Scan(&signatureKey, &encryptionKey, &nick, &since, &ts)
		if core.IsErr(err, "cannot read identity from db: %v") {
			continue
		}

		ek, _ := security.KeyFromBase64(encryptionKey)
		sk, _ := security.KeyFromBase64(signatureKey)
		identities = append(identities, Identity{
			Identity: security.Identity{
				Nick:          nick,
				SignatureKey:  sk,
				EncryptionKey: ek,
			},
			Since:   since,
			AddedOn: sql.DecodeTime(ts),
		})
	}
	return identities, nil
}

func (s *Safe) sqlSetIdentity(i security.Identity, since uint64) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("SET_IDENTITY_ON_SAFE", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"safe":          s.Name,
		"since":         since,
		"ts":            sql.EncodeTime(time.Now()),
	})
	return err
}

func (s *Safe) sqlDeleteIdentity(i Identity) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("DEL_IDENTITY_ON_SAFE", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"safe":          s.Name,
	})
	return err
}

func sqlSave(name string, configs []transport.Config) error {
	data, err := json.Marshal(&configs)
	if core.IsErr(err, "cannot marshal transport configuration of %s: %v", name) {
		return err
	}

	_, err = sql.Exec("SET_SAFE", sql.Args{"name": name, "configs": sql.EncodeBase64(data)})
	core.IsErr(err, "cannot save transport configuration of %s: %v", name)
	return err
}

func sqlLoad(name string) ([]transport.Config, error) {
	var data []byte
	var configs []transport.Config
	err := sql.QueryRow("GET_SAFE", sql.Args{"name": name}, &data)
	if core.IsErr(err, "cannot get safe %s config: %v", name) {
		return nil, err
	}

	err = json.Unmarshal(data, &configs)
	core.IsErr(err, "cannot unmarshal configs of %s: %v", name)
	return configs, err
}
