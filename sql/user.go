package sql

import (
	"encoding/json"
	"weshare/core"
	"weshare/model"
	"weshare/security"
)

// GetUsersIdentities returns the id (i.e. the public key) of users for a domain
func GetUsersIdentities(domain string, active bool) ([]security.Identity, error) {
	rows, err := query("GET_USERS_ID_BY_DOMAIN", names{"domain": domain, "active": active})
	if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
		return nil, err
	}

	var res []security.Identity
	for rows.Next() {
		var publicKey string
		var identity security.Identity
		err = rows.Scan(&publicKey)
		if !core.IsErr(err, "cannot read user's public key from db: %v") {
			data := base64dec(publicKey)
			err := json.Unmarshal(data, &identity)
			if !core.IsErr(err, "cannot unmarshal identity from db: %v") {
				res = append(res, identity)
			}
		}
	}
	return res, nil
}

func GetUsers(domain string) ([]model.User, error) {
	rows, err := query("GET_USERS_BY_DOMAIN", names{"domain": domain})
	if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
		return nil, err
	}

	var res []model.User
	for rows.Next() {
		var publicKey string
		var active bool
		err = rows.Scan(&publicKey, &active)
		if core.IsErr(err, "cannot read user's public key from db: %v") {
			continue
		}

		var identity security.Identity
		err := json.Unmarshal(base64dec(publicKey), &identity)
		if core.IsErr(err, "cannot unmarshall identity from db: %v") {
			continue
		}

		res = append(res, model.User{
			Identity: identity,
			Active:   active,
		})
	}
	return res, nil
}

func GetUsersByNick(domain string, nick string, active bool) ([]model.User, error) {
	rows, err := query("GET_USERS_BY_NICK", names{"domain": domain, "nick": nick, "active": active})
	if core.IsErr(err, "cannot get users by nick from db for domain '%s': %v", domain) {
		return nil, err
	}

	var res []model.User
	for rows.Next() {
		var data string
		var active bool
		err = rows.Scan(&data, &active)
		if core.IsErr(err, "cannot read user's public key from db: %v") {
			continue
		}

		identity, err := security.IdentityFromBase64(data)
		if core.IsErr(err, "cannot unmarshall identity from db: %v") {
			continue
		}

		res = append(res, model.User{
			Identity: identity,
			Active:   active,
		})
	}
	return res, nil
}

func SetUser(domain string, user model.User) error {
	data, err := user.Identity.Base64()
	if core.IsErr(err, "cannot unmarshall identity of user %s: %v", user.Identity.Nick) {
		return err
	}

	_, err = exec("SET_USER", names{"domain": domain, "identity": data, "nick": user.Identity.Nick, "active": user.Active})
	if core.IsErr(err, "cannot set user %s to db: %v", user.Identity.Nick) {
		return err
	}
	return nil
}

func GetAllTrusted(domain string) ([]model.User, error) {
	rows, err := query("GET_ALL_TRUSTED", names{"domain": domain})
	if core.IsErr(err, "cannot get users by nick from db for domain '%s': %v", domain) {
		return nil, err
	}
	var res []model.User
	for rows.Next() {
		var data string
		err = rows.Scan(&data)
		if core.IsErr(err, "cannot read user's public key from db: %v") {
			continue
		}

		identity, err := security.IdentityFromBase64(data)
		if core.IsErr(err, "cannot unmarshall identity from db: %v") {
			continue
		}

		res = append(res, model.User{
			Identity: identity,
			Active:   true,
		})
	}
	return res, nil
}

func SetTrusted(domain string, identity security.Identity, trusted bool) error {
	data, err := identity.Base64()
	if core.IsErr(err, "cannot unmarshall identity of user %s: %v", identity.Nick) {
		return err
	}

	_, err = exec("SET_TRUSTED", names{"domain": domain, "identity": data, "trusted": trusted})
	if core.IsErr(err, "cannot set trusted for user %s to db: %v", identity.Nick) {
		return err
	}
	return nil
}
