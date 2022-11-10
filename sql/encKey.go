package sql

import (
	"github.com/code-to-go/safepool/core"
)

// GetChanges returns all the item in the log filtering out those smaller than start in lexical order
func GetEncKeys(domain string) (map[uint32][]byte, error) {
	rows, err := Query("GET_ENC_KEYS_BY_DOMAIN", Args{"domain": domain})
	if core.IsErr(err, "cannot get encryption keys from db for domain '%s': %v", domain) {
		return nil, err
	}

	res := map[uint32][]byte{}
	for rows.Next() {
		var encKey uint32
		var value string
		err = rows.Scan(&encKey, &value)
		if !core.IsErr(err, "cannot read enc key from db: %v") {
			res[encKey] = DecodeBase64(value)
		}
	}
	return res, nil
}

func GetLastEncKey(domain string) (keyId uint32, keyValue []byte, err error) {
	var value string
	err = QueryRow("GET_LAST_ENC_KEY_BY_DOMAIN", Args{"domain": domain}, &keyId, &value)
	if core.IsErr(err, "cannot get encryption keys from db for domain '%s': %v", domain) {
		return 0, nil, err
	}
	return keyId, DecodeBase64(value), err
}

func SetEncKey(domain string, keyId uint32, keyValue []byte) error {
	_, err := Exec("SET_ENC_KEY", Args{"domain": domain, "keyId": keyId, "keyValue": EncodeBase64(keyValue)})
	return err
}
