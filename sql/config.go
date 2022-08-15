package sql

import (
	"database/sql"
	"encoding/base64"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var DbName = "weshare.db"

// GetConfig returns the value for a configuration parameter
func GetConfig(domain string, name string) (s string, i int, b []byte, ok bool) {
	var b64 *string
	row := queryRow("GET_CONFIG", names{"domain": domain, "name": name})
	switch err := row.Scan(&s, &i, &b64); err {
	case sql.ErrNoRows:
		ok = false
	case nil:
		ok = true
		if b64 != nil {
			b, err = base64.StdEncoding.DecodeString(*b64)
			if err != nil {
				logrus.Errorf("cannot decode []byte in config '%s': %v", name, err)
				ok = false
			}
		}
	default:
		logrus.Errorf("cannot get config '%s': %v", name, err)
		ok = false
	}
	return s, i, b, ok
}

// SetConfig stores a configuration parameter in the DB
func SetConfig(domain string, name string, s string, i int, b []byte) error {
	b64 := base64.StdEncoding.EncodeToString(b)
	_, err := exec("SET_CONFIG", names{"domain": domain, "name": name, "s": s, "i": i, "b": b64})
	if err != nil {
		logrus.Errorf("cannot exec '%s': %v", stmtSetConfig, err)
	}
	return err
}
