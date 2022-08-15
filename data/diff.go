package data

import (
	"io"
	"weshare/core"
)



func GetChanges(r io.ReadSeekCloser, m core.MerkleTree)