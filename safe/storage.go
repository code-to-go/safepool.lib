package safe

import "weshare/transport"

func Save(name string, configs []transport.Config) error {
	return sqlSave(name, configs)
}

func Load(name string) ([]transport.Config, error) {
	return sqlLoad(name)
}
