package safe

import (
	"encoding/json"
	"fmt"
	"time"
	"weshare/core"
	"weshare/security"

	"github.com/patrickmn/go-cache"
)

type Keystore map[uint64][]byte

var cachedEncKeys = cache.New(time.Hour, 10*time.Hour)

func (s *Safe) importKeystore(a AccessFile) (Keystore, error) {
	masterKey := s.keyFunc(s.masterKeyId)
	if masterKey == nil {
		return nil, ErrNotAuthorized
	}

	ks, err := s.unmarshalKeystore(masterKey, a.Nonce, a.Keystore)
	if core.IsErr(err, "cannot unmarshal keystore for safe '%s': %v", s.Name) {
		return nil, err
	}

	for id, val := range ks {
		err = s.sqlSetKey(id, val)
		if core.IsErr(err, "cannot set key %d to DB for safe '%s': %v", id, s.Name) {
			return nil, err
		}
	}
	return ks, nil
}

func (s *Safe) marshalKeystore(masterKey []byte, nonce []byte, ks Keystore) ([]byte, error) {
	data, err := json.Marshal(ks)
	if core.IsErr(err, "cannot marshal keystore: %v") {
		return nil, err
	}
	return security.EncryptBlock(masterKey, nonce, data)
}

func (s *Safe) unmarshalKeystore(masterKey []byte, nonce []byte, cipherdata []byte) (Keystore, error) {
	data, err := security.DecryptBlock(masterKey, nonce, cipherdata)
	if core.IsErr(err, "invalid key or corrupted keystore: %v") {
		return nil, err
	}

	var ks Keystore
	err = json.Unmarshal(data, &ks)
	return ks, err
}

func (s *Safe) keyFunc(id uint64) []byte {
	if id == 0 {
		return s.masterKey
	}

	k := fmt.Sprintf("%s-%d", s.Name, id)
	if v, found := cachedEncKeys.Get(k); found {
		return v.([]byte)
	}

	v := s.sqlGetKey(id)
	if v != nil {
		cachedEncKeys.Set(k, v, cache.DefaultExpiration)
	}
	return v
}
