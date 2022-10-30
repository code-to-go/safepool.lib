package safe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"
	"weshare/core"
	"weshare/security"
	"weshare/transport"

	"github.com/godruoyi/go-snowflake"
)

const SafeConfigFile = ".weshare-safe.json"

var ErrNoExchange = errors.New("no Exchange available")
var ErrInvalidSignature = errors.New("signature is invalid")
var ErrNotTrusted = errors.New("the author is not a trusted user")
var ErrNotAuthorized = errors.New("no authorization for this file")

type SafeConfig struct {
	Version float32
	Id      uint64
}

type Safe struct {
	Name string

	self        security.Identity
	e           transport.Exchanger
	exchangers  []transport.Exchanger
	masterKeyId uint64
	masterKey   []byte
}

type Head struct {
	Id        uint64
	Name      string
	Size      int64
	Hash      []byte
	ModTime   time.Time
	Author    security.Identity
	Signature []byte
}

const (
	ID_CREATE       = 0x0
	ID_FORCE_CREATE = 0x1
)

var ForceCreation = false

func CreateSafe(self security.Identity, name string, configs []transport.Config) (Safe, error) {
	s := Safe{
		Name: name,
		self: self,
	}
	err := s.connectSafe(name, configs)
	if err != nil {
		return Safe{}, err
	}

	s.masterKeyId = snowflake.ID()
	s.masterKey = security.GenerateBytesKey(32)
	err = s.sqlSetKey(s.masterKeyId, s.masterKey)
	if core.IsErr(err, "çannot store master encryption key to db: %v") {
		return s, err
	}
	err = security.SetIdentity(self)
	if core.IsErr(err, "çannot save identity to db: %v") {
		return s, err
	}

	err = s.sqlSetIdentity(self)
	if core.IsErr(err, "çannot link identity to save: %v") {
		return s, err
	}

	err = s.ExportAccessFile()
	return s, err
}

// Init initialized a domain on the specified exchangers
func OpenSafe(self security.Identity, name string, configs []transport.Config) (Safe, error) {
	s := Safe{
		Name: name,
		self: self,
	}
	err := s.connectSafe(name, configs)
	if err != nil {
		return Safe{}, err
	}

	err = s.ImportAccess()
	return s, err
}

func (s Safe) List(after uint64) []Head {
	s.refresh()

	return nil
}

func (s Safe) Post(name string, r io.Reader) (Head, error) {
	id := snowflake.ID()
	n := path.Join(s.Name, fmt.Sprintf("%d.body", id))
	hr, err := s.writeFile(n, r)
	if core.IsErr(err, "cannot post file %s to %s: %v", name, s.e) {
		return Head{}, err
	}

	signature, err := security.Sign(s.self, hr.Hash())
	if core.IsErr(err, "cannot sign file %s.body in %s: %v", name, s.e) {
		return Head{}, err
	}
	h := Head{
		Id:        id,
		Name:      name,
		Size:      hr.Size(),
		Hash:      hr.Hash(),
		ModTime:   time.Now(),
		Author:    s.self,
		Signature: signature,
	}
	data, err := json.Marshal(h)
	if core.IsErr(err, "cannot marshal header to json: %v") {
		return Head{}, err
	}

	n = path.Join(s.Name, fmt.Sprintf("%d.head", id))
	_, err = s.writeFile(n, bytes.NewBuffer(data))
	core.IsErr(err, "cannot write header %s.head in %s: %v", name, s.e)

	return h, nil
}

func (s Safe) Get(id uint64, w io.Writer) error {
	return nil
}

func (s Safe) Close() {
	for _, e := range s.exchangers {
		_ = e.Close()
	}
}

func (s Safe) Delete() error {
	for _, e := range s.exchangers {
		err := e.Delete(s.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Safe) Identities() ([]security.Identity, error) {
	return s.sqlGetIdentities(false)
}
