package security

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/code-to-go/safepool.lib/core"

	"github.com/patrickmn/go-cache"

	eciesgo "github.com/ecies/go/v2"
)

const (
	Secp256k1 = "secp256k1"
	Ed25519   = "ed25519"
)

var knownIdentities = cache.New(time.Hour, 10*time.Hour)

type Key struct {
	Public  []byte `json:"pu"`
	Private []byte `json:"pr,omitempty"`
}

type Identity struct {
	Nick          string `json:"n"`
	Email         string `json:"m"`
	SignatureKey  Key    `json:"s"`
	EncryptionKey Key    `json:"e"`
}

func NewIdentity(nick string) (Identity, error) {
	var identity Identity

	identity.Nick = nick
	privateCrypt, err := eciesgo.GenerateKey()
	if core.IsErr(err, "cannot generate secp256k1 key: %v") {
		return identity, err
	}
	identity.EncryptionKey = Key{
		Public:  privateCrypt.PublicKey.Bytes(true),
		Private: privateCrypt.Bytes(),
	}

	publicSign, privateSign, err := ed25519.GenerateKey(rand.Reader)
	if core.IsErr(err, "cannot generate ed25519 key: %v") {
		return identity, err
	}
	identity.SignatureKey = Key{
		Public:  publicSign[:],
		Private: privateSign[:],
	}
	return identity, nil
}

func (i Identity) Public() Identity {
	return Identity{
		Nick:  i.Nick,
		Email: i.Email,
		EncryptionKey: Key{
			Public: i.EncryptionKey.Public,
		},
		SignatureKey: Key{
			Public: i.SignatureKey.Public,
		},
	}
}

func IdentityFromBase64(b64 string) (Identity, error) {
	var i Identity
	data, err := base64.StdEncoding.DecodeString(b64)
	if core.IsErr(err, "cannot decode Identity string in base64: %v") {
		return i, err
	}

	err = json.Unmarshal(data, &i)
	if core.IsErr(err, "cannot decode Identity string from json: %v") {
		return i, err
	}
	return i, nil
}

func (i Identity) Base64() (string, error) {
	data, err := json.Marshal(i)
	if core.IsErr(err, "cannot marshal identity: %v") {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (i Identity) Id() string {
	return base64.StdEncoding.EncodeToString(i.SignatureKey.Public)
}

func SameIdentity(a, b Identity) bool {
	return bytes.Equal(a.SignatureKey.Public, b.SignatureKey.Public) &&
		bytes.Equal(a.EncryptionKey.Public, b.EncryptionKey.Public)
}

func SetIdentity(i Identity) error {
	k := string(append(i.SignatureKey.Public, i.EncryptionKey.Public...))
	if _, found := knownIdentities.Get(k); found {
		return nil
	}
	return sqlSetIdentity(i)
}

func Trust(i Identity, trusted bool) error {
	return sqlSetTrust(i, trusted)
}
