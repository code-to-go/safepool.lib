package model

import (
	"weshare/security"
	"weshare/transport"
)

type Transport struct {
	Domain    string
	Granted   bool
	Exchanges []transport.Config
}

type AccessToken struct {
	Transport Transport
	Identity  security.Identity
}
