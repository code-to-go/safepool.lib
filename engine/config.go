package engine

import (
	"encoding/base64"
	"weshare/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// GetConfig returns the value for a configuration parameter
func sqlGetConfig(domain string, name string) (s string, i int, b []byte, ok bool) {
	var b64 string
	err := sql.QueryRow("GET_CONFIG", sql.Args{"domain": domain, "name": name}, &s, &i, &b64)
	switch err {
	case sql.ErrNoRows:
		ok = false
	case nil:
		ok = true
		if b64 != "" {
			b = sql.DecodeBase64(b64)
		}
	default:
		logrus.Errorf("cannot get config '%s': %v", name, err)
		ok = false
	}
	return s, i, b, ok
}

// SetConfig stores a configuration parameter in the DB
func sqlSetConfig(domain string, name string, s string, i int, b []byte) error {
	b64 := base64.StdEncoding.EncodeToString(b)
	_, err := sql.Exec("SET_CONFIG", sql.Args{"domain": domain, "name": name, "s": s, "i": i, "b": b64})
	if err != nil {
		logrus.Errorf("cannot exec 'SET_CONFIG': %v", err)
	}
	return err
}
