package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"weshare/core"
	"weshare/protocol"

	eciesgo "github.com/ecies/go/v2"
	"github.com/golang/protobuf/proto"
)

const (
	Secp256k1 = "secp256k1"
	Ed25519   = "ed25519"
)

type Key struct {
	Public  []byte
	Private []byte
}

type Identity struct {
	Keys map[string]Key
	Nick string
}

func NewIdentity(nick string) (Identity, error) {
	var identity Identity

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

func MarshalIdentity(identity Identity, private bool) ([]byte, error) {
	i := protocol.Identity{
		Version: 1.0,
		Private: private,
		Keys:    map[string][]byte{},
		Nick:    identity.Nick,
	}

	for id, key := range identity.Keys {
		if private {
			i.Keys[id] = key.Private
		} else {
			i.Keys[id] = key.Public
		}
	}

	data, err := proto.Marshal(&i)
	if core.IsErr(err, "cannot marshal protocol.Identity: %v") {
		return nil, err
	}

	if !private {
		sign, err := Sign(identity, data)
		if core.IsErr(err, "cannot sign serialized identity: %v") {
			return nil, err
		}
		data = append(data, sign...)
	}

	return data, err
}

func derivePublicKey(id string, private []byte) []byte {
	switch id {
	case Secp256k1:
		pk := eciesgo.NewPrivateKeyFromBytes(private)
		return pk.PublicKey.Bytes(true)
	case Ed25519:
		pk := ed25519.PrivateKey(private)
		return pk.Public().(ed25519.PublicKey)
	}
	return nil
}

func UnmarshalIdentity(data []byte) (Identity, error) {
	var identity Identity
	var i protocol.Identity
	var sign []byte

	err := proto.Unmarshal(data, &i)
	if err != nil {
		sign = data[len(data)-64:]
		data = data[0 : len(data)-64]
		err = proto.Unmarshal(data, &i)
	}

	if core.IsErr(err, "cannot unmarshal identity: %v") {
		return identity, err
	}

	identity.Keys = map[string]Key{}
	for id, key := range i.Keys {
		if i.Private {
			identity.Keys[id] = Key{
				Private: key,
				Public:  derivePublicKey(id, key),
			}
		} else {
			identity.Keys[id] = Key{
				Public: key,
			}
		}
	}
	identity.Nick = i.Nick

	if sign != nil {
		if Verify(identity, data, sign) == false {
			return identity, core.ErrInvalidSignature
		}
	}

	return identity, err
}
