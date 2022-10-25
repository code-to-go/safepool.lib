package model

import (
	"weshare/security"
	"weshare/transport"
)

type Access struct {
	Domain    string
	Granted   bool
	Exchanges []transport.Config
}

type AccessToken struct {
	Access   Access
	Identity security.Identity
}
