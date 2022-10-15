package sql

import (
	"time"
	"weshare/core"
	"weshare/model"
)

// GetChangesForExchange returns all the item in the log filtering out those smaller than start in lexical order
func GetChangesForExchange(domain string, exchange string) ([]model.ChangeFile, error) {
	rows, err := query("GET_CHANGES", names{"domain": domain, "exchange": exchange})
	if core.IsErr(err, "cannot get logs from db: %v") {
		return nil, err
	}

	var changes []model.ChangeFile
	for rows.Next() {
		var change model.ChangeFile
		var timestamp int64
		err = rows.Scan(&change.Name, &timestamp)
		if !core.IsErr(err, "cannot read log row from db: %v") {
			changes = append(changes, change)
		}
	}
	return changes, nil
}

// GetChanges returns all the item in the log filtering out those smaller than start in lexical order
// func GetChanges(domain string, start string) ([]model.ChangeFile, error) {
// 	rows, err := query("GET_CHANGES", names{"domain": domain, "start": start})
// 	if core.IsErr(err, "cannot get logs from db: %v") {
// 		return nil, err
// 	}

// 	var changes []model.ChangeFile
// 	for rows.Next() {
// 		var change model.ChangeFile
// 		var timestamp int64
// 		err = rows.Scan(&change.Name, &timestamp)
// 		if !core.IsErr(err, "cannot read log row from db: %v") {
// 			changes = append(changes, change)
// 		}
// 	}
// 	return changes, nil
// }

// GetChange returns the change
func GetChange(domain string, id uint64) (c model.ChangeFile, ok bool, err error) {
	rows, err := query("GET_CHANGE", names{"domain": domain, "id": id})
	if core.IsErr(err, "cannot get logs from db: %v") {
		return model.ChangeFile{}, false, err
	}

	var changes []model.ChangeFile
	for rows.Next() {
		var change model.ChangeFile
		var timestamp int64
		err = rows.Scan(&change.Name, &timestamp)
		if !core.IsErr(err, "cannot read log row from db: %v") {
			changes = append(changes, change)
		}
	}
	return changes[0], true, nil
}

func AddChange(domain string, name string) error {
	_, err := exec("ADD_CHANGE", names{"domain": domain, "name": name, "timestamp": time.Now().Unix()})
	return err
}
