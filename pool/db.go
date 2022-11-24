package pool

import (
	"encoding/json"
	"time"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
	"github.com/code-to-go/safepool/sql"
	"github.com/code-to-go/safepool/transport"
)

func sqlGetHeads(pool string, afterId uint64, afterTime time.Time) ([]Head, error) {
	rows, err := sql.Query("GET_HEADS", sql.Args{"pool": pool, "afterId": afterId, "afterTime": sql.EncodeTime(afterTime)})
	if core.IsErr(err, "cannot get pools heads from db: %v") {
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
		if !core.IsErr(err, "cannot read pool heads from db: %v") {
			continue
		}
		h.ModTime = sql.DecodeTime(modTime)
		h.TimeStamp = sql.DecodeTime(ts)
		heads = append(heads, h)
	}
	return heads, nil
}

func sqlAddHead(pool string, h Head) error {
	_, err := sql.Exec("SET_HEAD", sql.Args{
		"pool":    pool,
		"id":      h.Id,
		"name":    h.Name,
		"size":    h.Size,
		"modTime": sql.EncodeTime(h.ModTime),
		"hash":    sql.EncodeBase64(h.Hash[:]),
		"ts":      sql.EncodeTime(time.Now()),
	})
	return err
}

func (p *Pool) sqlGetKey(keyId uint64) []byte {
	rows, err := sql.Query("GET_KEY", sql.Args{"pool": p.Name, "keyId": keyId})
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

func (p *Pool) sqlSetKey(keyId uint64, value []byte) error {
	_, err := sql.Exec("SET_KEY", sql.Args{"pool": p.Name, "keyId": keyId, "keyValue": sql.EncodeBase64(value)})
	return err
}

func (p *Pool) sqlGetKeystore() (Keystore, error) {
	rows, err := sql.Query("GET_KEYSTORE", sql.Args{"pool": p.Name})
	if core.IsErr(err, "cannot read keystore for pool %s: %v", p.Name) {
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

func (p *Pool) sqlGetIdentities(onlyTrusted bool) (identities []Identity, err error) {
	var q string
	if onlyTrusted {
		q = "GET_TRUSTED_ON_POOL"
	} else {
		q = "GET_IDENTITY_ON_POOL"
	}

	rows, err := sql.Query(q, sql.Args{"pool": p.Name})
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

func (p *Pool) sqlSetIdentity(i security.Identity, since uint64) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("SET_IDENTITY_ON_POOL", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"pool":          p.Name,
		"since":         since,
		"ts":            sql.EncodeTime(time.Now()),
	})
	return err
}

func (p *Pool) sqlDeleteIdentity(i Identity) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("DEL_IDENTITY_ON_POOL", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"pool":          p.Name,
	})
	return err
}

func sqlSave(name string, configs []transport.Config) error {
	data, err := json.Marshal(&configs)
	if core.IsErr(err, "cannot marshal transport configuration of %s: %v", name) {
		return err
	}

	_, err = sql.Exec("SET_POOL", sql.Args{"name": name, "configs": sql.EncodeBase64(data)})
	core.IsErr(err, "cannot save transport configuration of %s: %v", name)
	return err
}

func sqlLoad(name string) ([]transport.Config, error) {
	var data []byte
	var configs []transport.Config
	err := sql.QueryRow("GET_POOL", sql.Args{"name": name}, &data)
	if core.IsErr(err, "cannot get pool %s config: %v", name) {
		return nil, err
	}

	err = json.Unmarshal(data, &configs)
	core.IsErr(err, "cannot unmarshal configs of %s: %v", name)
	return configs, err
}

func sqlList() ([]string, error) {
	var names []string
	rows, err := sql.Query("LIST_POOL", nil)
	if core.IsErr(err, "cannot list pools: %v") {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err == nil {
			names = append(names, n)
		}
	}
	return names, err
}
