package access

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"time"
	"weshare/core"
	"weshare/security"
	"weshare/transport"

	"github.com/godruoyi/go-snowflake"
	"github.com/patrickmn/go-cache"
)

const AccessFolder = ".access"

type User struct {
	Identity security.Identity
	EncKey   []byte
}

type AccessFile struct {
	Version float32
	Id      uint64
	Users   []User
}

type AccessSignature struct {
	Identity  security.Identity
	Signature []byte
}

type AccessSignatureFile struct {
	Items []AccessSignature
}

var keysCache = cache.New(5*time.Minute, 10*time.Minute)

func (t *Topic) addAccess(identities []security.Identity) (uint64, error) {
	a := AccessFile{
		Version: 1.0,
		Id:      snowflake.ID(),
	}

	key, err := security.Generate32BytesKey()
	if core.IsErr(err, "cannot generate AES key: %v") {
		return 0, err
	}

	for _, i := range identities {
		encKey, err := security.EcEncrypt(i, key)
		if core.IsErr(err, "cannot encrypt AES key for user %s: %v", i.Nick) {
			continue
		}

		a.Users = append(a.Users, User{
			Identity: i,
			EncKey:   encKey,
		})
	}
	data, err := json.Marshal(a)
	if core.IsErr(err, "cannot marshal access file to JSON: %v") {
		return 0, err
	}
	sign, err := security.Sign(t.Self, data)
	if core.IsErr(err, "cannot sign access file: %v") {
		return 0, err
	}

	accessPath := path.Join(t.Name, AccessFolder, fmt.Sprintf("%d.json", a.Id))
	err = t.primary.Write(accessPath, bytes.NewBuffer(data))
	if core.IsErr(err, "cannot write access file to %s: %v", t.primary) {
		return 0, err
	}
	signPath := path.Join(t.Name, AccessFolder, fmt.Sprintf("%d.sign", a.Id))
	err = transport.WriteJSON(t.primary, signPath, AccessSignatureFile{
		Items: []AccessSignature{{
			Identity:  t.Self.Public(),
			Signature: sign,
		}},
	})
	if core.IsErr(err, "cannot write domain signature: %v") {
		return 0, err
	}

	return a.Id, nil
}

func (t *Topic) getKey(keyId uint64) []byte {
	v, ok := keysCache.Get(fmt.Sprintf("%s-%d", t.Name, keyId))
	if ok {
		return v.([]byte)
	}

	value := sqlGetKey(t.Name, keyId)
	if value != nil {
		return value
	}
	value, _ = t.getKeyFromTopic(keyId)
	return value
}

func (t *Topic) getKeyFromTopic(keyId uint64) ([]byte, error) {
	accessPath := path.Join(t.Name, AccessFolder, fmt.Sprintf("%d.json", keyId))
	b := bytes.Buffer{}
	err := t.primary.Read(accessPath, nil, &b)
	if core.IsErr(err, "cannot open access file %s: %v", accessPath) {
		return nil, err
	}

	var a AccessFile
	err = json.Unmarshal(b.Bytes(), &a)
	if core.IsErr(err, "corrupted access file %s: %v", accessPath) {
		return nil, err
	}

	s := bytes.Buffer{}
	signPath := path.Join(t.Name, AccessFolder, fmt.Sprintf("%d.sign", keyId))
	err = t.primary.Read(signPath, nil, &s)
	if core.IsErr(err, "cannot open access signature file %s: %v", signPath) {
		return nil, err
	}
	security.Verify()

	var value []byte
	for _, u := range a.Users {
		if bytes.Equal(u.Identity.Primary().Public, t.Self.Primary().Public) {
			value, err = security.EcDecrypt(t.Self, u.EncKey)
			if err == nil {
				err = sqlSetKey(t.Name, keyId, value)
				if err == nil {
					return value, nil
				}
			}
		}
	}
	return nil, ErrNotTrusted
}
