package data

import (
	"time"
	"weshare/core"
)

type LogEntry struct {
	FileId core.Hash256
}

type Log struct {
	Version uint16
	Time    time.Time
	UserId  core.PublicKey
	Entries []LogEntry
}
