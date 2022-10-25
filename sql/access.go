package sql

import (
	"encoding/json"
	"weshare/core"
	"weshare/model"
)

func GetDomains() ([]string, error) {
	rows, err := Query("GET_DOMAINS", nil)
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
	_, err = Exec("SET_ACCESS", Args{"domain": a.Domain, "granted": a.Granted, "config": data})
	return err
}

// GetAccess returns domain specific information (i.e. transport configuration)
func GetAccess(domain string) (model.Transport, error) {
	var data []byte
	var access model.Transport

	row := QueryRow("GET_ACCESS", Args{"domain": domain})
	err := row.Scan(&data)
	if core.IsErr(err, "cannot get access for domain %s: %v", domain) {
		return access, err
	}

	err = json.Unmarshal(data, &access)
	core.IsErr(err, "cannot deserialize access for domain %s: %v", domain)
	return access, err
}
