package apps

import (
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/sql"
	"time"

	"github.com/sirupsen/logrus"
)

func sqlSetFeedTime(safe string, feedTime time.Time) error {
	_, err := sql.Exec("SET_FEED_TIME", sql.Args{"safe": safe, "feedTime": feedTime})
	core.IsErr(err, "cannot write feed time for %s: %v", safe)
	return err
}

func sqlGetFeedTime(safe string) (time.Time, error) {
	var feedTime time.Time
	rows, err := sql.Query("GET_FEED_TIME", sql.Args{"safe": safe})
	if core.IsErr(err, "cannot get feed time for %s: %v", safe) {
		return feedTime, err
	}

	for rows.Next() {
		err = rows.Scan(&feedTime)
		if core.IsErr(err, "cannot get feed time for %s: %v", safe) {
			return feedTime, err
		}
	}

	return feedTime, err
}

func sqlGetConfig(safe string, key string) (s string, i int, b []byte, ok bool) {
	var b64 string
	err := sql.QueryRow("GET_CONFIG", sql.Args{"safe": safe, "key": key}, &s, &i, &b64)
	switch err {
	case sql.ErrNoRows:
		ok = false
	case nil:
		ok = true
		if b64 != "" {
			b = sql.DecodeBase64(b64)
		}
	default:
		logrus.Errorf("cannot get config '%s': %v", key, err)
		ok = false
	}
	return s, i, b, ok
}

// SetConfig stores a configuration parameter in the DB
func sqlSetConfig(safe string, key string, s string, i int, b []byte) error {
	b64 := sql.EncodeBase64(b)
	_, err := sql.Exec("SET_CONFIG", sql.Args{"safe": safe, "key": key, "s": s, "i": i, "b": b64})
	if err != nil {
		logrus.Errorf("cannot exec 'SET_CONFIG': %v", err)
	}
	return err
}
