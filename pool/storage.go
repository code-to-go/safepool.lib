package pool

import "github.com/code-to-go/safepool.lib/transport"

func Save(name string, configs []transport.Config) error {
	return sqlSave(name, configs)
}

func Load(name string) ([]transport.Config, error) {
	return sqlLoad(name)
}
