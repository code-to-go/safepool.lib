package sql

import (
	"time"
	"weshare/core"
)

// GetLogs returns all the item in the log filtering out those smaller than start in lexical order
func GetLogs(domain string, start string) ([]core.Log, error) {
	rows, err := query("GET_LOGS", names{"domain": domain, "start": start})
	if core.IsErr(err, "cannot get logs from db: %v") {
		return nil, err
	}

	var logs []core.Log
	for rows.Next() {
		var log core.Log
		var timestamp int64
		err = rows.Scan(&log.Name, &timestamp)
		if !core.IsErr(err, "cannot read log row from db: %v") {
			log.Timestamp = time.Unix(timestamp, 0)
			logs = append(logs, log)
		}
	}
	return logs, nil
}

func AddLog(domain string, name string) error {
	_, err := exec("ADD_LOG", names{"domain": domain, "name": name, "timestamp": time.Now().Unix()})
	return err
}
