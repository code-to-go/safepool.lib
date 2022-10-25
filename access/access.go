package access

import (
	"io"
	"weshare/security"
	"weshare/transport"
)

type Access interface {
	ReadFile(name string, w io.Writer) error
	WriteFile(name string, r io.Reader) error
	RefreshKeys() error
}

type AccessData struct {
	Identity  security.Identity
	Exchanger transport.Exchanger
	Domain    string
}

func NewAccess(identity security.Identity, e transport.Exchanger, domain string) Access {
	return &AccessData{
		Identity:  identity,
		Exchanger: e,
		Domain:    domain,
	}
}

func (a *AccessData) ReadFile(name string, w io.Writer) error {
	return nil
}

func (a *AccessData) WriteFile(name string, r io.Reader) error {
	return nil
}

func (a *AccessData) RefreshKeys() error {
	return nil
}
