package model

import (
	"time"
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
	ExchangeCreated  State = 0x10
	ExchangeModified State = 0x20
	ExchangeDeleted  State = 0x40
	ExchangeRenamed  State = 0x80
	Staged           State = 0x100
	Watched          State = 0x200
	Alternate        State = 0x400
	Conflict         State = 0x800
)

// File represents a file and its state both on different locations including
type File struct {
	Domain  string
	Name    string
	Id      uint64
	FirstId uint64
	Author  security.Identity
	ModTime time.Time
	Hash    []byte
	State   State
}

type LockFile struct {
	ExpectedDuration time.Duration
	Id               uint64
}
