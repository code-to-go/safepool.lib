package model

import (
	"weshare/security"
)

type User struct {
	Identity security.Identity
	Active   bool
}

type Domain struct {
	Name          string
	Snowflakeid   uint64
	Users         []security.Identity
	Key           []byte
	EncLegacyKeys []byte
}

type UsersFile struct {
	Version      float32
	GenerationId uint32
	Users        []User
	EncKeys      map[uint64][]byte
}
