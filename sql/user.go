package sql

import (
	"database/sql"
	"weshare/auth"
	"weshare/core"
	"weshare/security"
)

// GetUsersIdentities returns the id (i.e. the public key) of users for a domain
func GetUsersIdentities(domain string, active bool, mustBeAdmin bool) ([]security.Identity, error) {
	var rows *sql.Rows
	var err error
	if mustBeAdmin {
		rows, err = query("GET_ADMINS_ID_BY_DOMAIN", names{"domain": domain, "active": active})
	} else {
		rows, err = query("GET_USERS_ID_BY_DOMAIN", names{"domain": domain, "active": active})
	}

	if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
		return nil, err
	}

	var res []security.Identity
	for rows.Next() {
		var publicKey string
		err = rows.Scan(&publicKey)
		if !core.IsErr(err, "cannot read user's public key from db: %v") {
			data := base64dec(publicKey)
			identity, err := security.UnmarshalIdentity(data)
			if !core.IsErr(err, "cannot unmarshal identity from db: %v") {
				res = append(res, identity)
			}
		}
	}
	return res, nil
}

func GetUsers(domain string) ([]auth.User, error) {
	rows, err := query("GET_USERS_BY_DOMAIN", names{"domain": domain})
	if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
		return nil, err
	}

	var res []auth.User
	for rows.Next() {
		var publicKey string
		var admin bool
		var active bool
		err = rows.Scan(&publicKey, &admin, &active)
		if core.IsErr(err, "cannot read user's public key from db: %v") {
			continue
		}

		identity, err := security.UnmarshalIdentity(base64dec(publicKey))
		if core.IsErr(err, "cannot unmarshall identity from db: %v") {
			continue
		}

		res = append(res, auth.User{
			Identity: identity,
			Admin:    admin,
			Active:   active,
		})
	}
	return res, nil
}

func SetUser(domain string, user auth.User) error {
	data, err := security.MarshalIdentity(user.Identity, false)
	if core.IsErr(err, "cannot unmarshall identity of user %s: %v", user.Identity.Nick) {
		return err
	}

	_, err = exec("SET_USER", names{"domain": domain, "publicKey": base64enc(data),
		"admin": user.Admin, "active": user.Active})
	if core.IsErr(err, "cannot set user %s to db: %v", user.Identity.Nick) {
		return err
	}
	return nil
}
