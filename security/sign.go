package security

import (
	"crypto/ed25519"
	"encoding/binary"
	"io"
	"weshare/core"
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

func Primary(identity Identity) Key {
	return identity.Keys[Ed25519]
}

func Sign(identity Identity, data []byte) ([]byte, error) {
	private := identity.Keys[Ed25519].Private
	return ed25519.Sign(ed25519.PrivateKey(private), data), nil
}

func Verify(identity Identity, data []byte, sig []byte) bool {
	for off := 0; off < len(sig); off += SignatureSize {
		public := identity.Keys[Ed25519].Public
		if func() bool {
			defer func() { recover() }()
			return ed25519.Verify(ed25519.PublicKey(public), data, sig[off:off+SignatureSize])
		}() {
			return true
		}
	}
	return false
}

func SignAndWrite(identity Identity, data []byte, w io.Writer) error {
	sign, err := Sign(identity, data)
	if core.IsErr(err, "cannot sign data: %v") {
		return err
	}

	lenB := make([]byte, 4)
	binary.BigEndian.PutUint32(lenB, uint32(SignatureSize+len(data)))
	w.Write(lenB)
	_, err = w.Write(data)
	if core.IsErr(err, "cannot write data to stream: %v") {
		return err
	}
	_, err = w.Write(sign)
	if core.IsErr(err, "cannot write signature to stream: %v") {
		return err
	}
	return nil
}

func ReadAndVerify(admins []Identity, r io.Reader) (data []byte, signature []byte, err error) {
	lenB := make([]byte, 4)
	_, err = r.Read(lenB)
	if core.IsErr(err, "cannot read length of data") {
		return nil, nil, err
	}

	data = make([]byte, binary.BigEndian.Uint32(lenB))
	_, err = r.Read(data)
	if core.IsErr(err, "cannot read data") {
		return nil, nil, err
	}

	data, sign := data[0:len(data)-SignatureSize], data[len(data)-SignatureSize:]
	verified := false
out:
	for _, identity := range admins {
		if Verify(identity, sign, data) {
			verified = true
			break out
		}
	}

	if verified {
		return data, sign, nil
	} else {
		return nil, nil, core.ErrInvalidSignature
	}
}
