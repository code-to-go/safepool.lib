package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"weshare/core"

	eciesgo "github.com/ecies/go/v2"
	"github.com/godruoyi/go-snowflake"
)

const (
	Secp256k1 = "secp256k1"
	Ed25519   = "ed25519"
)

type Key struct {
	Public  []byte `json:"u"`
	Private []byte `json:"r,omitempty"`
}

type Identity struct {
	Id   uint64         `json:"i"`
	Keys map[string]Key `json:"k"`
	Nick string         `json:"n"`
}

func NewIdentity(nick string) (Identity, error) {
	var identity Identity

	identity.Id = snowflake.ID()
	identity.Nick = nick
	privateCrypt, err := eciesgo.GenerateKey()
	if core.IsErr(err, "cannot generate secp256k1 key: %v") {
		return identity, err
	}
	identity.Keys = map[string]Key{}
	identity.Keys[Secp256k1] = Key{
		Public:  privateCrypt.PublicKey.Bytes(true),
		Private: privateCrypt.Bytes(),
	}

	publicSign, privateSign, err := ed25519.GenerateKey(rand.Reader)
	if core.IsErr(err, "cannot generate ed25519 key: %v") {
		return identity, err
	}
	identity.Keys[Ed25519] = Key{
		Public:  publicSign[:],
		Private: privateSign[:],
	}
	return identity, nil
}

func IdentityFromBase64(b64 string) (Identity, error) {
	var identity Identity
	data, err := base64.StdEncoding.DecodeString(b64)
	if core.IsErr(err, "cannot decode Identity string in base64: %v") {
		return identity, err
	}

	err = json.Unmarshal(data, &identity)
	if core.IsErr(err, "cannot decode Identity string from json: %v") {
		return identity, err
	}
	return identity, nil
}

func (i Identity) Public() Identity {
	identity := Identity{Nick: i.Nick, Keys: map[string]Key{}}
	for k, v := range i.Keys {
		identity.Keys[k] = Key{
			Public: v.Public,
		}
	}
	return identity
}

func (i Identity) Primary() Key {
	return i.Keys[Ed25519]
}

func (i Identity) Base64() (string, error) {
	data, err := json.Marshal(i)
	if core.IsErr(err, "cannot marshal identity: %v") {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func SameIdentity(a, b Identity) bool {
	for k, v := range a.Keys {
		if bytes.Compare(b.Keys[k].Public, v.Public) == 0 {
			return true
		}
	}
	return false
}
