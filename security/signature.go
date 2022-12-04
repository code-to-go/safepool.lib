package security

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"

	"github.com/code-to-go/safepool.lib/core"
)

type PublicKey ed25519.PublicKey
type PrivateKey ed25519.PrivateKey

const (
	PublicKeySize  = ed25519.PublicKeySize
	PrivateKeySize = ed25519.PrivateKeySize
	SignatureSize  = ed25519.SignatureSize
)

type SignedData struct {
	Signature [SignatureSize]byte
	Signer    PublicKey
}

type Public struct {
	Id    PublicKey
	Nick  string
	Email string
}

func Sign(identity Identity, data []byte) ([]byte, error) {
	private := identity.SignatureKey.Private
	return ed25519.Sign(ed25519.PrivateKey(private), data), nil
}

func Verify(id string, data []byte, sig []byte) bool {
	public, err := base64.StdEncoding.DecodeString(id)
	if core.IsErr(err, "invalid id '%s': %v", id) {
		return false
	}

	for off := 0; off < len(sig); off += SignatureSize {
		if func() bool {
			defer func() { recover() }()
			return ed25519.Verify(ed25519.PublicKey(public), data, sig[off:off+SignatureSize])
		}() {
			return true
		}
	}
	return false
}

type SignedHashEvidence struct {
	Key       []byte `json:"k"`
	Signature []byte `json:"s"`
}

type SignedHash struct {
	Hash      []byte               `json:"h"`
	Evidences []SignedHashEvidence `json:"e"`
}

func NewSignedHash(hash []byte, i Identity) (SignedHash, error) {
	signature, err := Sign(i, hash)
	if core.IsErr(err, "cannot sign with identity %s: %v", base64.StdEncoding.EncodeToString(i.SignatureKey.Public)) {
		return SignedHash{}, err
	}

	return SignedHash{
		Hash: hash,
		Evidences: []SignedHashEvidence{
			{
				Key:       i.SignatureKey.Public,
				Signature: signature,
			},
		},
	}, nil
}

func AppendToSignedHash(s SignedHash, i Identity) error {
	signature, err := Sign(i, s.Hash)
	if core.IsErr(err, "cannot sign with identity %s: %v", base64.StdEncoding.EncodeToString(i.SignatureKey.Public)) {
		return err
	}
	s.Evidences = append(s.Evidences, SignedHashEvidence{
		Key:       i.SignatureKey.Public,
		Signature: signature,
	})
	return nil
}

func VerifySignedHash(s SignedHash, trusts []Identity, hash []byte) bool {
	if !bytes.Equal(s.Hash, hash) {
		return false
	}

	for _, e := range s.Evidences {
		for _, t := range trusts {
			if bytes.Equal(e.Key, t.SignatureKey.Public) {
				if Verify(t.Id(), hash, e.Signature) {
					return true
				}
			}
		}
	}
	return false
}
