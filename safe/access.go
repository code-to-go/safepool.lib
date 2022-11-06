package safe

import (
	"bytes"
	"crypto/aes"
	"hash"
	"path"
	"time"
	"weshare/core"
	"weshare/security"
	"weshare/transport"

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

func (s *Safe) ImportAccess(e transport.Exchanger) (hash.Hash, error) {
	l, err := s.lockAccessFile(e)
	if core.IsErr(err, "cannot lock access on %s: %v", s.e) {
		return nil, err
	}
	defer s.unlockAccessFile(e, l)

	a, h, err := s.readAccessFile(e)
	if core.IsErr(err, "cannot read access file:%v") {
		return nil, err
	}
	if bytes.Equal(h.Sum(nil), s.accessHash) {
		return h, nil
	}

	if s.masterKeyId != a.MasterKeyId {
		err = s.importMasterKey(a)
		if core.IsErr(err, "cannot import master key: %v") {
			return nil, err
		}
	}

	ks, err := s.importKeystore(a)
	if core.IsErr(err, "cannot sync grants: %v") {
		return nil, err
	}

	err = s.importGrants(a, ks)
	if core.IsErr(err, "cannot sync grants: %v") {
		return nil, err
	}

	s.accessHash = h.Sum(nil)
	return h, nil
}

func (s *Safe) ExportAccessFile(e transport.Exchanger) error {
	identities, err := s.sqlGetIdentities(false)
	if core.IsErr(err, "cannot get identities for safe '%s':%v", s.Name) {
		return err
	}

	var grants []Grant
	for _, i := range identities {
		keystoreKey, err := security.EcEncrypt(i.Identity, s.masterKey)
		if core.IsErr(err, "cannot encrypt master key for identity %s: %v", i.Nick) {
			return err
		}
		grants = append(grants, Grant{
			Identity:    i.Public(),
			Since:       s.masterKeyId,
			KeystoreKey: keystoreKey,
		})
	}

	ks, err := s.sqlGetKeystore()
	if core.IsErr(err, "cannot read keystore from db for safe '%s': %v", s.Name) {
		return err
	}

	noonce := security.GenerateBytesKey(aes.BlockSize)
	cipherks, err := s.marshalKeystore(s.masterKey, noonce, ks)
	if core.IsErr(err, "cannot marshal keystore for safe '%s': %v", s.Name) {
		return err
	}

	a := AccessFile{
		Version:     0.0,
		Nonce:       noonce,
		MasterKeyId: s.masterKeyId,
		Grants:      grants,
		Keystore:    cipherks,
	}

	l, err := s.lockAccessFile(s.e)
	if core.IsErr(err, "cannot lock access on %s: %v", s.e) {
		return err
	}
	defer s.unlockAccessFile(e, l)
	_, err = s.writeAccessFile(e, a)
	return err
}

func (s *Safe) importMasterKey(a AccessFile) error {
	for _, g := range a.Grants {
		if security.SameIdentity(s.Self, g.Identity) {
			masterKey, err := security.EcDecrypt(s.Self, g.KeystoreKey)
			if !core.IsErr(err, "corrupted master key in access grant: %v", err) {
				err = s.sqlSetKey(a.MasterKeyId, masterKey)
				if core.IsErr(err, "cannot write master key to db: %v", err) {
					return err
				}
			}
		}
	}
	s.masterKeyId = a.MasterKeyId
	return nil
}

func (s *Safe) importGrants(a AccessFile, ks Keystore) error {
	identities, err := s.sqlGetIdentities(false)
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
				s.sqlSetIdentity(i, g.Since)
			}
			delete(is, k)
		}
	}

	needNewMasterKey := false
	for _, i := range is {
		if _, ok := ks[i.Since]; ok {
			s.sqlDeleteIdentity(i)
			needNewMasterKey = true
		}
	}
	if needNewMasterKey {
		s.masterKeyId = snowflake.ID()
		s.masterKey = security.GenerateBytesKey(32)
		err = s.sqlSetKey(s.masterKeyId, s.masterKey)
		if core.IsErr(err, "çannot store master encryption key to db: %v") {
			return err
		}
	}

	return nil
}

func (s *Safe) lockAccessFile(e transport.Exchanger) (uint64, error) {
	lockFile := path.Join(s.Name, ".access.lock")
	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	core.IsErr(err, "cannot lock access on %s: %v", s.Name, err)
	return lockId, err
}

func (s *Safe) unlockAccessFile(e transport.Exchanger, lockId uint64) {
	lockFile := path.Join(s.Name, ".access.lock")
	transport.UnlockFile(e, lockFile, lockId)
}

func (s *Safe) readAccessFile(e transport.Exchanger) (AccessFile, hash.Hash, error) {
	var a AccessFile
	var sh security.SignedHash
	signatureFile := path.Join(s.Name, ".access.sign")
	accessFile := path.Join(s.Name, ".access")

	err := transport.ReadJSON(e, signatureFile, &sh, nil)
	if core.IsErr(err, "cannot read signature file '%s': %v", signatureFile, err) {
		return AccessFile{}, nil, err
	}

	h := security.NewHash()
	err = transport.ReadJSON(e, accessFile, &a, h)
	if core.IsErr(err, "cannot read access file: %s", err) {
		return AccessFile{}, nil, err
	}

	if security.VerifySignedHash(sh, []security.Identity{s.Self}, h.Sum(nil)) {
		return a, h, nil
	}

	trusted, err := s.sqlGetIdentities(true)
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

	_ = security.AppendToSignedHash(sh, s.Self)
	if !core.IsErr(err, "cannot lock access on %s: %v", s.Name, err) {
		if security.AppendToSignedHash(sh, s.Self) == nil {
			err = transport.WriteJSON(e, signatureFile, sh, nil)
			core.IsErr(err, "cannot write signature file on %s: %v", s.Name, err)
		}
	}

	return a, h, nil
}

func (s *Safe) writeAccessFile(e transport.Exchanger, a AccessFile) (hash.Hash, error) {
	lockFile := path.Join(s.Name, ".access.lock")
	signatureFile := path.Join(s.Name, ".access.sign")
	accessFile := path.Join(s.Name, ".access")

	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	if core.IsErr(err, "cannot lock access on %s: %v", s.Name, err) {
		return nil, err
	}
	defer transport.UnlockFile(e, lockFile, lockId)

	h := security.NewHash()
	err = transport.WriteJSON(e, accessFile, a, h)
	if core.IsErr(err, "cannot write access file on %s: %v", s.Name, err) {
		return nil, err
	}

	sh, err := security.NewSignedHash(h.Sum(nil), s.Self)
	if core.IsErr(err, "cannot generate signature hash on %s: %v", s.Name, err) {
		return nil, err
	}
	err = transport.WriteJSON(e, signatureFile, sh, nil)
	if core.IsErr(err, "cannot write signature file on %s: %v", s.Name, err) {
		return nil, err
	}

	return h, nil
}
