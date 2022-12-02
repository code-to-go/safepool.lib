package security

import (
	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/sql"
)

func sqlSetIdentity(i Identity) error {
	i64, err := i.Base64()
	if core.IsErr(err, "cannot serialize identity: %v") {
		return err
	}

	_, err = sql.Exec("SET_IDENTITY", sql.Args{
		"id":  i.Id(),
		"i64": i64,
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
	defer rows.Close()

	var identities []Identity
	for rows.Next() {
		var i64 string
		err = rows.Scan(&i64)
		if !core.IsErr(err, "cannot read pool heads from db: %v") {
			continue
		}

		i, err := IdentityFromBase64(i64)
		if !core.IsErr(err, "invalid identity record '%s': %v", i64) {
			identities = append(identities, i)
		}
	}
	return identities, nil
}

func sqlSetTrust(i Identity, trusted bool) error {
	_, err := sql.Exec("SET_TRUSTED", sql.Args{
		"id":      i.Id(),
		"trusted": trusted,
	})
	return err
}
