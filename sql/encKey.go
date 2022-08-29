package sql

import (
	"weshare/core"
)

// GetChanges returns all the item in the log filtering out those smaller than start in lexical order
func GetEncKeys(domain string) (map[uint32][]byte, error) {
	rows, err := query("GET_ENC_KEYS_BY_DOMAIN", names{"domain": domain})
	if core.IsErr(err, "cannot get encryption keys from db for domain '%s': %v", domain) {
		return nil, err
	}

	res := map[uint32][]byte{}
	for rows.Next() {
		var encKey uint32
		var value string
		err = rows.Scan(&encKey, &value)
		if !core.IsErr(err, "cannot read enc key from db: %v") {
			res[encKey] = base64dec(value)
		}
	}
	return res, nil
}

func GetLastEncKey(domain string) (keyId uint32, keyValue []byte, err error) {
	row := queryRow("GET_LAST_ENC_KEY_BY_DOMAIN", names{"domain": domain})
	err = row.Err()
	if core.IsErr(err, "cannot get encryption keys from db for domain '%s': %v", domain) {
		return 0, nil, err
	}

	var value string
	err = row.Scan(&keyId, &value)
	return keyId, base64dec(value), err
}

func SetEncKey(domain string, keyId uint32, keyValue []byte) error {
	_, err := exec("SET_ENC_KEY", names{"domain": domain, "keyId": keyId, "keyValue": base64enc(keyValue)})
	return err
}
