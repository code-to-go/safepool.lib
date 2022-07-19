package access

import "golang.org/x/crypto/blake2b"

type Hash256 [blake2b.Size256]byte
