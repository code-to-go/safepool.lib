package data

import (
	"io"
	"weshare/core"
	"weshare/protocol"
)

func NewChangeFile(r io.ReadSeekCloser, m core.MerkleTree) (protocol.ChangeFile, error) {
	return protocol.ChangeFile{}, nil
}
