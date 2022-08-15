package sql

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/sirupsen/logrus"
)

var db *sql.DB
var InitDDL string

func createTables() error {
	parts := strings.Split(InitDDL, "\n\n")

	for _, part := range parts {
		if strings.Trim(part, " ") == "" {
			continue
		}

		if !strings.HasPrefix(part, "-- ") {
			logrus.Errorf("unexpected break without a comment in '%s'", part)
		}

		cr := strings.Index(part, "\n")
		if cr == -1 {
			logrus.Error("invalid comment without CR")
			return os.ErrInvalid
		}
		key, ql := part[3:cr], part[cr+1:]

		if strings.HasPrefix(key, "INIT") {
			_, err := db.Exec(ql)
			if err != nil {
				logrus.Errorf("cannot execute SQL Init stmt: %v", err)
				return err
			}
		} else {
			prepareStatement(key, ql)
		}
	}
	return nil
}

//LoadSQLFromFile loads the sql queries from the provided file path. It panics in case the file cannot be loaded
func LoadSQLFromFile(name string) {
	ddl, err := ioutil.ReadFile(name)
	if err != nil {
		logrus.Panicf("cannot load SQL queries from %s: %v", name, err)
		panic(err)
	}

	InitDDL = string(ddl)
}

func OpenDB() error {
	dbPath := filepath.Join(xdg.ConfigHome, DbName)
	_, err := os.Stat(dbPath)
	if errors.Is(err, os.ErrNotExist) {
		err := ioutil.WriteFile(dbPath, []byte{}, 0644)
		if err != nil {
			logrus.Errorf("cannot create SQLite db in %s: %v", dbPath, err)
			return err
		}

	} else if err != nil {
		logrus.Errorf("cannot access SQLite db file %s: %v", dbPath, err)
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logrus.Errorf("cannot open SQLite db in %s: %v", dbPath, err)
		return err
	}

	return createTables()
}

func CloseDB() error {
	if db == nil {
		return os.ErrClosed
	}
	err := db.Close()
	db = nil
	return err
}

func DeleteDB() error {
	dbPath := filepath.Join(xdg.ConfigHome, DbName)
	return os.Remove(dbPath)
}
