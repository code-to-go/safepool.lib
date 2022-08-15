package core

import (
	"time"

	"golang.org/x/crypto/blake2b"
)

// State is the current state of a file
type State uint32

// State bits ||Created|
const (
	LocalCreated    State = 0x1
	LocalChanged    State = 0x2
	LocalDeleted    State = 0x4
	LocalRenamed    State = 0x8
	ExchangeCreated State = 0x100
	ExchangeChanged State = 0x200
	ExchangeDeleted State = 0x400
	ExchangeRenamed State = 0x800
	Staged          State = 0x10000
)

//Identity defines the identifier of the user
type Identity struct {
	Curve   string
	Public  []byte
	Private []byte
	Name    string
	Name2   string
}

// File represents a file and its state both on different locations including
type File struct {
	Domain  string
	Name    string
	Author  []byte
	Alt     string
	ModTime time.Time
	Hash    []byte
	State   State
}

// Log tracks the updates from exchanges that have been applied
type Log struct {
	Domain    string
	Name      string
	Timestamp time.Time
}

type Hash256 [blake2b.Size256]byte
