package access

import (
	"io"
	"time"
	"weshare/security"
	"weshare/transport"
)

type Access struct {
	identity  security.Identity
	domain    string
	exchanger transport.Exchanger
}

type FileInfo struct {
	Name    string
	Size    int64
	Hash    security.Hash256
	Id      uint64
	ModTime time.Time
}

func Connect(identity security.Identity, domain string) (Access, error) {
	a := Access{
		identity:  identity,
		domain:    domain,
		exchanger: e,
	}

}

func Join(domain string, exchangers []transport.Exchanger, self security.Identity, trusts []security.Identity) error {

}

func ReadDir(a Access, name string) []FileInfo {

}

func Read(a Access, name string, thread string, w io.Writer) error {
	return nil
}

func Write(a Access, name string, thread string, r io.Reader) error {
	return nil
}

func Trust(thread string, members map[security.Identity]bool) string {

}

func readThreads(a Access) []string {

}

func refreshKeys(a Access, thread string) error {
	return nil
}
