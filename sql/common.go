package sql

import (
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/sirupsen/logrus"
)

var stmtCache = map[string]*sql.Stmt{}

func prepareStatement(key, s string) {
	if _, ok := stmtCache[key]; ok {
		logrus.Panicf("duplicate SQL statement for key '%s'", s)
		panic(key)
	}

	stmt, err := db.Prepare(s)
	if err != nil {
		logrus.Panicf("cannot compile SQL statement '%s': %v", s, err)
		panic(err)
	}
	stmtCache[key] = stmt
}

func getStatement(key string) *sql.Stmt {
	if v, ok := stmtCache[key]; ok {
		return v
	} else {
		logrus.Panicf("missing SQL statement for key '%s'", key)
		panic(key)
	}
}

type names map[string]any

func named(m names) []any {
	var args []any
	if m != nil {
		for k, v := range m {
			args = append(args, sql.Named(k, v))
		}
	}
	return args
}

func exec(key string, m names) (sql.Result, error) {
	return getStatement(key).Exec(named(m)...)
}

func queryRow(key string, m names) *sql.Row {
	return getStatement(key).QueryRow(named(m)...)
}

func query(key string, m names) (*sql.Rows, error) {
	return getStatement(key).Query(named(m)...)
}

func base64enc(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func base64dec(data string) []byte {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	return b
}

func timeEnc(t time.Time) int64 {
	return t.Unix()
}

func timeDec(v int64) time.Time {
	return time.Unix(v, 0)
}
