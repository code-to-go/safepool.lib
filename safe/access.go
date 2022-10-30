package safe

import (
	"crypto/aes"
	"path"
	"time"
	"weshare/core"
	"weshare/security"
	"weshare/transport"

	"github.com/patrickmn/go-cache"
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

var knownIdentities = cache.New(time.Hour, 10*time.Hour)

func (s *Safe) ImportAccess() error {
	a, err := s.readAccessFile()
	if core.IsErr(err, "cannot read access file:%v") {
		return err
	}

	s.masterKeyId = a.MasterKeyId
	err = s.importGrants(a)
	if core.IsErr(err, "cannot sync grants: %v") {
		return err
	}
	err = s.importKeystore(a)
	if core.IsErr(err, "cannot sync grants: %v") {
		return err
	}

	return nil
}

func (s *Safe) ExportAccessFile() error {
	identities, err := s.sqlGetIdentities(false)
	if core.IsErr(err, "cannot get identities for safe '%s':%v", s.Name) {
		return err
	}

	var grants []Grant
	for _, i := range identities {
		keystoreKey, err := security.EcEncrypt(i, s.masterKey)
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

	return s.writeAccessFile(a)
}

func (s *Safe) importGrants(a AccessFile) error {
	for _, g := range a.Grants {
		i := g.Identity
		k := s.Name + string(append(i.SignatureKey.Public, i.EncryptionKey.Public...))
		if _, found := knownIdentities.Get(k); !found {
			err := security.SetIdentity(i)
			if !core.IsErr(err, "cannot add identity %s: %v", g.Identity.Nick) {
				s.sqlSetIdentity(i)
			}
		}
		if security.SameIdentity(s.self, i) {
			masterKey, err := security.EcDecrypt(s.self, g.KeystoreKey)
			if !core.IsErr(err, "corrupted master key in access grant: %v", err) {
				s.sqlSetKey(a.MasterKeyId, masterKey)
			}
		}
	}
	return nil
}

func (s *Safe) readAccessFile() (AccessFile, error) {
	var a AccessFile
	var sh security.SignedHash
	lockFile := path.Join(s.Name, ".access.lock")
	signatureFile := path.Join(s.Name, ".access.sign")
	accessFile := path.Join(s.Name, ".access")
	e := s.e

	err := transport.ReadJSON(e, signatureFile, &sh, nil)
	if core.IsErr(err, "cannot read signature file '%s': %v", signatureFile, err) {
		return AccessFile{}, err
	}

	h := security.NewHash()
	err = transport.ReadJSON(e, accessFile, &a, h)
	if core.IsErr(err, "cannot read access file: %s", err) {
		return AccessFile{}, err
	}

	if security.VerifySignedHash(sh, []security.Identity{s.self}, h.Sum(nil)) {
		return a, nil
	}

	trusted, err := s.sqlGetIdentities(true)
	if core.IsErr(err, "cannot get trusted identities: %v") {
		return AccessFile{}, nil
	}
	if !security.VerifySignedHash(sh, trusted, h.Sum(nil)) {
		return AccessFile{}, ErrNotTrusted
	}

	_ = security.AppendToSignedHash(sh, s.self)
	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	if !core.IsErr(err, "cannot lock access on %s: %v", s.Name, err) {
		defer transport.UnlockFile(e, lockFile, lockId)
		if security.AppendToSignedHash(sh, s.self) == nil {
			err = transport.WriteJSON(e, signatureFile, sh, nil)
			core.IsErr(err, "cannot write signature file on %s: %v", s.Name, err)
		}
	}

	return a, nil
}

func (s *Safe) writeAccessFile(a AccessFile) error {
	lockFile := path.Join(s.Name, ".access.lock")
	signatureFile := path.Join(s.Name, ".access.sign")
	accessFile := path.Join(s.Name, ".access")

	e := s.e
	lockId, err := transport.LockFile(e, lockFile, time.Minute)
	if core.IsErr(err, "cannot lock access on %s: %v", s.Name, err) {
		return err
	}
	defer transport.UnlockFile(e, lockFile, lockId)

	h := security.NewHash()
	err = transport.WriteJSON(e, accessFile, a, h)
	if core.IsErr(err, "cannot write access file on %s: %v", s.Name, err) {
		return err
	}

	sh, err := security.NewSignedHash(h.Sum(nil), s.self)
	if core.IsErr(err, "cannot generate signature hash on %s: %v", s.Name, err) {
		return err
	}
	err = transport.WriteJSON(e, signatureFile, sh, nil)
	if core.IsErr(err, "cannot write signature file on %s: %v", s.Name, err) {
		return err
	}

	return nil
}
