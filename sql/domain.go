package sql

import (
	"encoding/json"
	"weshare/core"
	"weshare/exchanges"
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

func SetDomain(name string, configs []exchanges.Config) error {
	data, err := json.Marshal(configs)
	if core.IsErr(err, "cannot serialize exchanges config for domain %s: %v", name) {
		return err
	}
	_, err = exec("SET_DOMAIN", names{"name": name, "config": data})
	return err
}

// GetDomain returns domain specific information (i.e. exchanges configuration)
func GetDomain(name string) ([]exchanges.Config, error) {
	var data []byte
	var configs []exchanges.Config

	row := queryRow("GET_DOMAIN", names{"name": name})
	err := row.Scan(&data)
	if core.IsErr(err, "cannot get configuration of domain %s: %v", name) {
		return configs, err
	}

	err = json.Unmarshal(data, &configs)
	core.IsErr(err, "cannot deserialize configuration of domain %s: %v", name)
	return configs, err
}
