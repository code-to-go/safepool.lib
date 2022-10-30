package model

import (
	"weshare/transport"
)

type Identity []byte

type Access struct {
	Domain    string
	Granted   bool
	Exchanges []transport.Config
}

type AccessToken struct {
	Access    Access
	Transport Transport
	Identity  Identity
}
