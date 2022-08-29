package model

import (
	"time"
	"weshare/exchanges"
	"weshare/security"
)

// State is the current state of a file
type State uint32

// State bits ||Created|
const (
	LocalCreated     State = 0x1
	LocalModified    State = 0x2
	LocalDeleted     State = 0x4
	LocalRenamed     State = 0x8
	ExchangeCreated  State = 0x100
	ExchangeModified State = 0x200
	ExchangeDeleted  State = 0x400
	ExchangeRenamed  State = 0x800
	Staged           State = 0x10000
)

// File represents a file and its state both on different locations including
type File struct {
	Domain  string
	Name    string
	LastId  uint64
	FirstId uint64
	Author  []byte
	Alt     string
	ModTime time.Time
	Hash    []byte
	State   State
}

// Log tracks the updates from exchanges that have been applied
type Change struct {
	Domain       string
	Name         string
	SnowFlakeId  uint64
	ProgenitorId uint64
}

type Domain struct {
	Name        string
	Snowflakeid uint64
	Users       []security.Identity
	Key         []byte
	LegacyKeys  map[int64][]byte
}

type Access struct {
	Domain    string
	Granted   bool
	Exchanges []exchanges.Config
}

type AccessToken struct {
	Access   Access
	Identity []byte
}
