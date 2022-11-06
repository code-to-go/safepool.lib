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
	"github.com/sirupsen/logrus"
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

type Consumer interface {
	TimeOffset(s *Safe) time.Time
	Accept(s *Safe, h Head) bool
}

type Safe struct {
	Name      string
	Self      security.Identity
	Poll      chan string
	Consumers []Consumer

	e           transport.Exchanger
	exchangers  []transport.Exchanger
	masterKeyId uint64
	masterKey   []byte
	lastReplica time.Time
	accessHash  []byte
}

type Identity struct {
	security.Identity
	//Since is the keyId used when the identity was added to the Safe access
	Since uint64
	//AddedOn is the timestamp when the identity is stored on the local DB
	AddedOn time.Time
}

type Head struct {
	Id        uint64
	Name      string
	Size      int64
	Hash      []byte
	ModTime   time.Time
	Author    security.Identity
	Signature []byte
	TimeStamp time.Time `json:"-"`
}

const (
	ID_CREATE       = 0x0
	ID_FORCE_CREATE = 0x1
)

var ForceCreation = false
var ReplicaPeriod = time.Hour

func CreateSafe(self security.Identity, name string, configs []transport.Config) (Safe, error) {
	s := Safe{
		Name:        name,
		Self:        self,
		Poll:        make(chan string),
		lastReplica: time.Now(),
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

	err = s.sqlSetIdentity(self, s.masterKeyId)
	if core.IsErr(err, "çannot link identity to save: %v") {
		return s, err
	}

	if !ForceCreation {
		_, err = s.e.Stat(path.Join(s.Name, ".access"))
		if err == nil {
			return Safe{}, ErrNotAuthorized
		}
	}

	err = s.ExportAccessFile(s.e)
	if core.IsErr(err, "cannot export access file: %v") {
		return Safe{}, err
	}

	go s.poll()
	return s, err
}

// Init initialized a domain on the specified exchangers
func OpenSafe(self security.Identity, name string, configs []transport.Config) (Safe, error) {
	s := Safe{
		Name: name,
		Self: self,
	}
	err := s.connectSafe(name, configs)
	if err != nil {
		return Safe{}, err
	}

	_, err = s.ImportAccess(s.e)
	return s, err
}

func (s Safe) List(afterId uint64, afterTs time.Time) []Head {
	hs, _ := s.list(afterId, afterTs)
	return hs
}

func (s Safe) Post(name string, r io.Reader) (Head, error) {
	id := snowflake.ID()
	n := path.Join(s.Name, fmt.Sprintf("%d.body", id))
	hr, err := s.writeFile(n, r)
	if core.IsErr(err, "cannot post file %s to %s: %v", name, s.e) {
		return Head{}, err
	}

	hash := hr.Hash()
	signature, err := security.Sign(s.Self, hash)
	if core.IsErr(err, "cannot sign file %s.body in %s: %v", name, s.e) {
		return Head{}, err
	}
	h := Head{
		Id:        id,
		Name:      name,
		Size:      hr.Size(),
		Hash:      hash,
		ModTime:   time.Now(),
		Author:    s.Self.Public(),
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
	headName := path.Join(s.Name, fmt.Sprintf("%d.head", id))
	bodyName := path.Join(s.Name, fmt.Sprintf("%d.body", id))

	h, err := s.readHead(headName)
	if core.IsErr(err, "cannot read header '%s': %v") {
		return err
	}

	hr, err := s.readFile(bodyName, w)
	if core.IsErr(err, "cannot read body '%s': %v", bodyName) {
		return err
	}
	hash := hr.Hash()
	if !bytes.Equal(hash, h.Hash) {
		return ErrInvalidSignature
	}
	return nil
}

func (s Safe) Close() {
	for _, e := range s.exchangers {
		_ = e.Close()
	}
	close(s.Poll)
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

func (s Safe) Identities() ([]Identity, error) {
	identities, err := s.sqlGetIdentities(false)
	return identities, err
}

func (s Safe) poll() {
	for r := range s.Poll {
		logrus.Infof("poll request from %s", r)

		timeOffset := time.Now()
		offsets := map[Consumer]time.Time{}
		for _, c := range s.Consumers {
			o := c.TimeOffset(&s)
			offsets[c] = o
			if timeOffset.After(o) {
				timeOffset = o
			}
		}

		for _, h := range s.List(0, timeOffset) {
			for _, c := range s.Consumers {
				if c.Accept(&s, h) {
					break
				}
			}
		}

		if time.Since(s.lastReplica) > ReplicaPeriod {
			s.replica()
		}
	}
}
