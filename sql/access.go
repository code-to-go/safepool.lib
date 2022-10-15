package sql

import (
	"encoding/json"
	"weshare/core"
	"weshare/model"
)

func GetDomains() ([]string, error) {
	rows, err := query("GET_DOMAINS", nil)
	var domains []string
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var domain string
		err = rows.Scan(&domain)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, nil
}

func SetAccess(a model.Transport) error {

	data, err := json.Marshal(a)
	if core.IsErr(err, "cannot serialize access config for domain %s: %v", a.Domain) {
		return err
	}
	_, err = exec("SET_ACCESS", names{"domain": a.Domain, "granted": a.Granted, "config": data})
	return err
}

// GetAccess returns domain specific information (i.e. transport configuration)
func GetAccess(domain string) (model.Transport, error) {
	var data []byte
	var access model.Transport

	row := queryRow("GET_ACCESS", names{"domain": domain})
	err := row.Scan(&data)
	if core.IsErr(err, "cannot get access for domain %s: %v", domain) {
		return access, err
	}

	err = json.Unmarshal(data, &access)
	core.IsErr(err, "cannot deserialize access for domain %s: %v", domain)
	return access, err
}
