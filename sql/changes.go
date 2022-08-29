package sql

import (
	"time"
	"weshare/core"
	"weshare/model"
)

// GetChanges returns all the item in the log filtering out those smaller than start in lexical order
func GetChanges(domain string, start string) ([]model.Change, error) {
	rows, err := query("GET_CHANGES", names{"domain": domain, "start": start})
	if core.IsErr(err, "cannot get logs from db: %v") {
		return nil, err
	}

	var changes []model.Change
	for rows.Next() {
		var change model.Change
		var timestamp int64
		err = rows.Scan(&change.Name, &timestamp)
		if !core.IsErr(err, "cannot read log row from db: %v") {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

func AddChange(domain string, name string) error {
	_, err := exec("ADD_CHANGE", names{"domain": domain, "name": name, "timestamp": time.Now().Unix()})
	return err
}
