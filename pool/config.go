package pool

import (
	"github.com/code-to-go/safepool/core"
)

func Define(c Config) error {
	return sqlSave(c.Name, c.Configs)
}

func GetConfig(name string) (Config, error) {
	configs, err := sqlLoad(name)
	if core.IsErr(err, "cannot load config for pool '%s'", name) {
		return Config{}, err
	}
	return Config{
		Name:    name,
		Configs: configs,
	}, nil
}
