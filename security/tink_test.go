package security

import (
	"encoding/base64"
	"fmt"
	"log"
	"testing"

	"github.com/google/tink/go/keyset"
	"github.com/google/tink/go/signature"
)

func TestSignature(t *testing.T) {
	kh, err := keyset.NewHandle(signature.ECDSAP256KeyTemplate()) // Other key templates can also be used.
	if err != nil {
		log.Fatal(err)
	}

	// TODO: save the private keyset to a pool location. DO NOT hardcode it in source code.
	// Consider encrypting it with a remote key in Cloud KMS, AWS KMS or HashiCorp Vault.
	// See https://github.com/google/tink/blob/master/docs/GOLANG-HOWTO.md#storing-and-loading-existing-keysets.

	s, err := signature.NewSigner(kh)
	if err != nil {
		log.Fatal(err)
	}

	msg := []byte("this data needs to be signed")
	sig, err := s.Sign(msg)
	if err != nil {
		log.Fatal(err)
	}

	pubkh, err := kh.Public()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: share the public with the verifier.

	v, err := signature.NewVerifier(pubkh)
	if err != nil {
		log.Fatal(err)
	}

	if err := v.Verify(sig, msg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Message: %s\n", msg)
	fmt.Printf("Signature: %s\n", base64.StdEncoding.EncodeToString(sig))

}
