package model

import (
	"time"
	"weshare/security"
)

// ChangeFile stores information about recent change files in different transport
type ChangeFile struct {
	Domain      string
	Name        string
	Id          uint64
	ExchangeUrl string
	ModTime     time.Time
}

type ChangeFileHeader struct {
	Version float32
	Flag    uint32
	Domain  string
	Name    string
	Ids     []uint64
	FirstId uint64
	Author  security.Identity
}
