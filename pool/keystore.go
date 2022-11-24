package pool

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"

	"github.com/patrickmn/go-cache"
)

type Keystore map[uint64][]byte

var cachedEncKeys = cache.New(time.Hour, 10*time.Hour)

func (p *Pool) importKeystore(a AccessFile) (Keystore, error) {
	masterKey := p.keyFunc(p.masterKeyId)
	if masterKey == nil {
		return nil, ErrNotAuthorized
	}

	ks, err := p.unmarshalKeystore(masterKey, a.Nonce, a.Keystore)
	if core.IsErr(err, "cannot unmarshal keystore for pool '%s': %v", p.Name) {
		return nil, err
	}

	for id, val := range ks {
		err = p.sqlSetKey(id, val)
		if core.IsErr(err, "cannot set key %d to DB for pool '%s': %v", id, p.Name) {
			return nil, err
		}
	}
	return ks, nil
}

func (p *Pool) marshalKeystore(masterKey []byte, nonce []byte, ks Keystore) ([]byte, error) {
	data, err := json.Marshal(ks)
	if core.IsErr(err, "cannot marshal keystore: %v") {
		return nil, err
	}
	return security.EncryptBlock(masterKey, nonce, data)
}

func (p *Pool) unmarshalKeystore(masterKey []byte, nonce []byte, cipherdata []byte) (Keystore, error) {
	data, err := security.DecryptBlock(masterKey, nonce, cipherdata)
	if core.IsErr(err, "invalid key or corrupted keystore: %v") {
		return nil, err
	}

	var ks Keystore
	err = json.Unmarshal(data, &ks)
	return ks, err
}

func (p *Pool) keyFunc(id uint64) []byte {
	if id == 0 {
		return p.masterKey
	}

	k := fmt.Sprintf("%s-%d", p.Name, id)
	if v, found := cachedEncKeys.Get(k); found {
		return v.([]byte)
	}

	v := p.sqlGetKey(id)
	if v != nil {
		cachedEncKeys.Set(k, v, cache.DefaultExpiration)
	}
	return v
}
