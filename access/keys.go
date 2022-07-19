package access

import (
	"baobab/def"
	"crypto/ed25519"
	"crypto/rand"
)

type PublicKey ed25519.PublicKey
type PrivateKey ed25519.PrivateKey

const (
	PublicKeySize  = ed25519.PublicKeySize
	PrivateKeySize = ed25519.PrivateKeySize
	SignatureSize  = ed25519.SignatureSize
)

type Identity struct {
	Curve   string     `json:"curve" yaml:"curve"`
	Public  PublicKey  `json:"public" yaml:"public"`
	Private PrivateKey `json:"private" yaml:"private"`
}

type Public struct {
	Id    PublicKey
	Nick  string
	Email string
}

func NewIdentity() (Identity, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if def.IsErr(err, "cannot generate ed25519 identity: %v") {
		return Identity{}, err
	}

	return Identity{
		Curve:   "ed25519",
		Public:  PublicKey(public),
		Private: PrivateKey(private),
	}, nil
}

func Sign(private PrivateKey, data []byte) ([]byte, error) {
	return ed25519.Sign(ed25519.PrivateKey(private), data), nil
}

func Verify(key PublicKey, data []byte, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(key), data, sig)
}
