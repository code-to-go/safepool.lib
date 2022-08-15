package auth

import (
	"weshare/core"
)

type Users struct {
	Name  string
	Users []core.PublicKey
}

func NewGroup(name string, identity core.Identity) Users {
	return Users{
		Name:  name,
		Users: []core.PublicKey{identity.Public},
	}
}
