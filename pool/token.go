package pool

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/security"
)

type Token struct {
	Config Config
	Host   security.Identity
}

func EncodeToken(t Token, guest *security.Identity) (string, error) {
	tk, err := json.Marshal(Token{
		Config: t.Config,
		Host:   t.Host.Public(),
	})
	if core.IsErr(err, "cannot marshal config to token: %v") {
		return "", err
	}

	gk := ""
	if guest != nil {
		gk, err = guest.Public().Base64()
		if core.IsErr(err, "invalid guest key: %v") {
			return "", err
		}
		tk, err = security.EcEncrypt(*guest, tk)
		if core.IsErr(err, "cannot encrypt with guest key: %v") {
			return "", err
		}
	}

	sig, err := security.Sign(t.Host, tk)
	if core.IsErr(err, "cannot sign with host key: %v") {
		return "", err
	}

	return fmt.Sprintf("%s:%s:%s", gk,
		base64.StdEncoding.EncodeToString(tk),
		base64.StdEncoding.EncodeToString(sig)), nil

}

func DecodeToken(guest *security.Identity, token string) (Token, error) {
	var t Token
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return t, ErrInvalidToken
	}

	gk, tk64, sig64 := parts[0], parts[1], parts[2]
	tk, _ := base64.StdEncoding.DecodeString(tk64)
	sig, _ := base64.StdEncoding.DecodeString(sig64)
	if gk != "" {
		g, err := security.IdentityFromBase64(gk)
		if core.IsErr(err, "cannot decode guest token from base64: %s") {
			return t, ErrInvalidToken
		}

		if guest == nil || !bytes.Equal(g.EncryptionKey.Public, guest.EncryptionKey.Public) {
			core.IsErr(ErrNotAuthorized, "mismatch between guests keys: %v")
			return t, ErrNotAuthorized
		}

		tk, err = security.EcDecrypt(*guest, tk)
		if core.IsErr(err, "cannot decrypt guest token with own key: %s") {
			return t, ErrInvalidToken
		}
	}

	err := json.Unmarshal(tk, &t)
	if core.IsErr(err, "cannot unmarshal token: %s") {
		return t, ErrInvalidToken
	}

	if !security.Verify(t.Host, tk, sig) {
		core.IsErr(ErrInvalidSignature, "token has invalid signature: %v")
		return t, ErrInvalidSignature
	}

	return t, nil
}
