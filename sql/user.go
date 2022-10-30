package sql

import (
	"weshare/model"
)

type Identity []byte

// GetUsersIdentities returns the id (i.e. the public key) of users for a domain
func GetUsersIdentities(domain string, active bool) ([]Identity, error) {
	// rows, err := Query("GET_USERS_ID_BY_DOMAIN", Args{"domain": domain, "active": active})
	// if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
	// 	return nil, err
	// }

	// var res []security.Identity
	// for rows.Next() {
	// 	var publicKey string
	// 	var identity security.Identity
	// 	err = rows.Scan(&publicKey)
	// 	if !core.IsErr(err, "cannot read user's public key from db: %v") {
	// 		data := DecodeBase64(publicKey)
	// 		err := json.Unmarshal(data, &identity)
	// 		if !core.IsErr(err, "cannot unmarshal identity from db: %v") {
	// 			res = append(res, identity)
	// 		}
	// 	}
	// }
	// return res, nil
	return nil, nil
}

func GetUsers(domain string) ([]model.User, error) {
	// rows, err := Query("GET_USERS_BY_DOMAIN", Args{"domain": domain})
	// if core.IsErr(err, "cannot get users ids from db for domain '%s': %v", domain) {
	// 	return nil, err
	// }

	// var res []model.User
	// for rows.Next() {
	// 	var publicKey string
	// 	var active bool
	// 	err = rows.Scan(&publicKey, &active)
	// 	if core.IsErr(err, "cannot read user's public key from db: %v") {
	// 		continue
	// 	}

	// 	var identity security.Identity
	// 	err := json.Unmarshal(DecodeBase64(publicKey), &identity)
	// 	if core.IsErr(err, "cannot unmarshall identity from db: %v") {
	// 		continue
	// 	}

	// 	res = append(res, model.User{
	// 		Identity: identity,
	// 		Active:   active,
	// 	})
	// }
	// return res, nil
	return nil, nil
}

func GetUsersByNick(domain string, nick string, active bool) ([]model.User, error) {
	// rows, err := Query("GET_USERS_BY_NICK", Args{"domain": domain, "nick": nick, "active": active})
	// if core.IsErr(err, "cannot get users by nick from db for domain '%s': %v", domain) {
	// 	return nil, err
	// }

	// var res []model.User
	// for rows.Next() {
	// 	var data string
	// 	var active bool
	// 	err = rows.Scan(&data, &active)
	// 	if core.IsErr(err, "cannot read user's public key from db: %v") {
	// 		continue
	// 	}

	// 	identity, err := security.IdentityFromBase64(data)
	// 	if core.IsErr(err, "cannot unmarshall identity from db: %v") {
	// 		continue
	// 	}

	// 	res = append(res, model.User{
	// 		Identity: identity,
	// 		Active:   active,
	// 	})
	// }
	// return res, nil
	return nil, nil
}

func SetUser(domain string, user model.User) error {
	// data, err := user.Identity.Base64()
	// if core.IsErr(err, "cannot unmarshall identity of user %s: %v", user.Identity.Nick) {
	// 	return err
	// }

	// _, err = Exec("SET_USER", Args{"domain": domain, "identity": data, "nick": user.Identity.Nick, "active": user.Active})
	// if core.IsErr(err, "cannot set user %s to db: %v", user.Identity.Nick) {
	// 	return err
	// }
	return nil
}

func GetAllTrusted(domain string) ([]model.User, error) {
	// rows, err := Query("GET_ALL_TRUSTED", Args{"domain": domain})
	// if core.IsErr(err, "cannot get users by nick from db for domain '%s': %v", domain) {
	// 	return nil, err
	// }
	// var res []model.User
	// for rows.Next() {
	// 	var data string
	// 	err = rows.Scan(&data)
	// 	if core.IsErr(err, "cannot read user's public key from db: %v") {
	// 		continue
	// 	}

	// 	identity, err := security.IdentityFromBase64(data)
	// 	if core.IsErr(err, "cannot unmarshall identity from db: %v") {
	// 		continue
	// 	}

	// 	res = append(res, model.User{
	// 		Identity: identity,
	// 		Active:   true,
	// 	})
	// }
	return nil, nil
}

func SetTrusted(domain string, identity Identity, trusted bool) error {
	// data, err := identity.Base64()
	// if core.IsErr(err, "cannot unmarshall identity of user %s: %v", identity.Nick) {
	// 	return err
	// }

	// _, err = Exec("SET_TRUSTED", Args{"domain": domain, "identity": data, "trusted": trusted})
	// if core.IsErr(err, "cannot set trusted for user %s to db: %v", identity.Nick) {
	// 	return err
	// }
	return nil
}
