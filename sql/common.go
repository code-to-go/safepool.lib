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

type Args map[string]any

func named(m Args) []any {
	var args []any
	if m != nil {
		for k, v := range m {
			args = append(args, sql.Named(k, v))
		}
	}
	return args
}

func Exec(key string, m Args) (sql.Result, error) {
	return getStatement(key).Exec(named(m)...)
}

func QueryRow(key string, m Args) *sql.Row {
	return getStatement(key).QueryRow(named(m)...)
}

func Query(key string, m Args) (*sql.Rows, error) {
	return getStatement(key).Query(named(m)...)
}

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(data string) []byte {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	return b
}

func EncodeTime(t time.Time) int64 {
	return t.Unix()
}

func DecodeTime(v int64) time.Time {
	return time.Unix(v, 0)
}
