package pool

import (
	"bytes"
	"crypto/aes"
	"hash"
	"path"
	"time"

	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/security"
	"github.com/code-to-go/safepool.lib/transport"

	"github.com/godruoyi/go-snowflake"
)

type Grant struct {
	Identity    security.Identity
	Since       uint64
	KeystoreKey []byte
}

type AccessFile struct {
	Version     float32
	Grants      []Grant
	Nonce       []byte
	MasterKeyId uint64
	Keystore    []byte
}

func (p *Pool) ImportAccess(e transport.Exchanger) (hash.Hash, error) {
	l, err := p.lockAccessFile(e)
	if core.IsErr(err, "cannot lock access on %s: %v", p.e) {
		return nil, err
	}
	defer p.unlockAccessFile(e, l)

	a, h, err := p.readAccessFile(e)
	if core.IsErr(err, "cannot read access file:%v") {
		return nil, err
	}
	if bytes.Equal(h.Sum(nil), p.accessHash) {
		return h, nil
	}

	if p.masterKeyId != a.MasterKeyId {
		err = p.importMasterKey(a)
		if core.IsErr(err, "cannot import master key: %v") {
			return nil, err
		}
	}

	ks, err := p.importKeystore(a)
	if core.IsErr(err, "cannot sync grants: %v") {
		return nil, err
	}

	err = p.importGrants(a, ks)
	if core.IsErr(err, "cannot sync grants: %v") {
		return nil, err
	}

	p.accessHash = h.Sum(nil)
	return h, nil
}

func (p *Pool) ExportAccessFile(e transport.Exchanger) error {
	identities, err := p.sqlGetIdentities(false)
	if core.IsErr(err, "cannot get identities for pool '%s':%v", p.Name) {
		return err
	}

	var grants []Grant
	for _, i := range identities {
		keystoreKey, err := security.EcEncrypt(i.Identity, p.masterKey)
		if core.IsErr(err, "cannot encrypt master key for identity %s: %v", i.Nick) {
			return err
		}
		grants = append(grants, Grant{
			Identity:    i.Public(),
			Since:       p.masterKeyId,
			KeystoreKey: keystoreKey,
		})
	}

	ks, err := p.sqlGetKeystore()
	if core.IsErr(err, "cannot read keystore from db for pool '%s': %v", p.Name) {
		return err
	}

	noonce := security.GenerateBytesKey(aes.BlockSize)
	cipherks, err := p.marshalKeystore(p.masterKey, noonce, ks)
	if core.IsErr(err, "cannot marshal keystore for pool '%s': %v", p.Name) {
		return err
	}

	a := AccessFile{
		Version:     0.0,
		Nonce:       noonce,
		MasterKeyId: p.masterKeyId,
		Grants:      grants,
		Keystore:    cipherks,
	}

	l, err := p.lockAccessFile(p.e)
	if core.IsErr(err, "cannot lock access on %s: %v", p.e) {
		return err
	}
	defer p.unlockAccessFile(e, l)
	_, err = p.writeAccessFile(e, a)
	return err
}

func (p *Pool) importMasterKey(a AccessFile) error {
	for _, g := range a.Grants {
		if security.SameIdentity(p.Self, g.Identity) {
			masterKey, err := security.EcDecrypt(p.Self, g.KeystoreKey)
			if !core.IsErr(err, "corrupted master key in access grant: %v", err) {
				err = p.sqlSetKey(a.MasterKeyId, masterKey)
				if core.IsErr(err, "cannot write master key to db: %v", err) {
					return err
				}
			}
		}
	}
	p.masterKeyId = a.MasterKeyId
	return nil
}

func (p *Pool) importGrants(a AccessFile, ks Keystore) error {
	identities, err := p.sqlGetIdentities(false)
	if core.IsErr(err, "cannot read identities during grant import: %v", err) {
		return err
	}
	is := map[string]Identity{}
	for _, i := range identities {
		k := string(append(i.SignatureKey.Public, i.EncryptionKey.Public...))
		is[k] = i
	}

	for _, g := range a.Grants {
		i := g.Identity
		k := string(append(i.SignatureKey.Public, i.EncryptionKey.Public...))
		if _, found := is[k]; !found {
			err := security.SetIdentity(i)
			if !core.IsErr(err, "cannot add identity %s: %v", g.Identity.Nick) {
				p.sqlSetIdentity(i, g.Since)
			}
			delete(is, k)
		}
	}

	needNewMasterKey := false
	for _, i := range is {
		if _, ok := ks[i.Since]; ok {
			p.sqlDeleteIdentity(i)
			needNewMasterKey = true
		}
	}
	if needNewMasterKey {
		p.masterKeyId = snowflake.ID()
		p.masterKey = security.GenerateBytesKey(32)
		err = p.sqlSetKey(p.masterKeyId, p.masterKey)
		if core.IsErr(err, "çannot store master encryption key to db: %v") {
			return err
		}
	}

	return nil
}

func (p *Pool) lockAccessFile(e transport.Exchanger) (uint64, error) {
	lockFile := path.Join(p.Name, ".access.lock")
	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	core.IsErr(err, "cannot lock access on %s: %v", p.Name, err)
	return lockId, err
}

func (p *Pool) unlockAccessFile(e transport.Exchanger, lockId uint64) {
	lockFile := path.Join(p.Name, ".access.lock")
	transport.UnlockFile(e, lockFile, lockId)
}

func (p *Pool) readAccessFile(e transport.Exchanger) (AccessFile, hash.Hash, error) {
	var a AccessFile
	var sh security.SignedHash
	signatureFile := path.Join(p.Name, ".access.sign")
	accessFile := path.Join(p.Name, ".access")

	err := transport.ReadJSON(e, signatureFile, &sh, nil)
	if core.IsErr(err, "cannot read signature file '%s': %v", signatureFile, err) {
		return AccessFile{}, nil, err
	}

	h := security.NewHash()
	err = transport.ReadJSON(e, accessFile, &a, h)
	if core.IsErr(err, "cannot read access file: %s", err) {
		return AccessFile{}, nil, err
	}

	if security.VerifySignedHash(sh, []security.Identity{p.Self}, h.Sum(nil)) {
		return a, h, nil
	}

	trusted, err := p.sqlGetIdentities(true)
	if core.IsErr(err, "cannot get trusted identities: %v") {
		return AccessFile{}, nil, nil
	}

	var is []security.Identity
	for _, t := range trusted {
		is = append(is, t.Identity)
	}

	if !security.VerifySignedHash(sh, is, h.Sum(nil)) {
		return AccessFile{}, nil, ErrNotTrusted
	}

	_ = security.AppendToSignedHash(sh, p.Self)
	if !core.IsErr(err, "cannot lock access on %s: %v", p.Name, err) {
		if security.AppendToSignedHash(sh, p.Self) == nil {
			err = transport.WriteJSON(e, signatureFile, sh, nil)
			core.IsErr(err, "cannot write signature file on %s: %v", p.Name, err)
		}
	}

	return a, h, nil
}

func (p *Pool) writeAccessFile(e transport.Exchanger, a AccessFile) (hash.Hash, error) {
	lockFile := path.Join(p.Name, ".access.lock")
	signatureFile := path.Join(p.Name, ".access.sign")
	accessFile := path.Join(p.Name, ".access")

	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	if core.IsErr(err, "cannot lock access on %s: %v", p.Name, err) {
		return nil, err
	}
	defer transport.UnlockFile(e, lockFile, lockId)

	h := security.NewHash()
	err = transport.WriteJSON(e, accessFile, a, h)
	if core.IsErr(err, "cannot write access file on %s: %v", p.Name, err) {
		return nil, err
	}

	sh, err := security.NewSignedHash(h.Sum(nil), p.Self)
	if core.IsErr(err, "cannot generate signature hash on %s: %v", p.Name, err) {
		return nil, err
	}
	err = transport.WriteJSON(e, signatureFile, sh, nil)
	if core.IsErr(err, "cannot write signature file on %s: %v", p.Name, err) {
		return nil, err
	}

	return h, nil
}
