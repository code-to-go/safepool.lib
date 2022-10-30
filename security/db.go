package security

import (
	"weshare/core"
	"weshare/sql"
)

func sqlInsertIdentity(i Identity) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()

	_, err := sql.Exec("INSERT_IDENTITY", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"nick":          i.Nick,
	})
	return err
}

func sqlGetIdentities(onlyTrusted bool) ([]Identity, error) {
	var q string
	if onlyTrusted {
		q = "GET_TRUSTED"
	} else {
		q = "GET_IDENTITY"
	}

	rows, err := sql.Query(q, sql.Args{})
	if core.IsErr(err, "cannot get trusted identities from db: %v") {
		return nil, err
	}

	var is []Identity
	for rows.Next() {
		var encryptionKey, signatureKey, nick string
		err = rows.Scan(&signatureKey, &encryptionKey, &nick)
		if !core.IsErr(err, "cannot read safe heads from db: %v") {
			continue
		}

		ek, _ := KeyFromBase64(encryptionKey)
		sk, _ := KeyFromBase64(signatureKey)
		is = append(is, Identity{
			Nick:          nick,
			SignatureKey:  sk,
			EncryptionKey: ek,
		})
	}
	return is, nil
}

func sqlSetTrust(i Identity, trusted bool) error {
	sk, _ := i.SignatureKey.Base64()
	ek, _ := i.EncryptionKey.Base64()
	_, err := sql.Exec("SET_TRUSTED", sql.Args{
		"signatureKey":  sk,
		"encryptionKey": ek,
		"trusted":       trusted,
	})
	return err
}
