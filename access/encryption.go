package access

import "weshare/security"

type Header struct {
	Version float32          `json:"v`
	Thread  uint64           `json:"t"`
	Key     uint64           `json:"k"`
	Hash    security.Hash256 `json:"h"`
	Name    string           `json:"s"`
}
