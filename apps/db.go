package apps

import (
	"time"
	"weshare/core"
	"weshare/sql"
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
