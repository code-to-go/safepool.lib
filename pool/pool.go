package pool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
	"github.com/code-to-go/safepool/transport"

	"github.com/godruoyi/go-snowflake"
	"github.com/sirupsen/logrus"
)

const SafeConfigFile = ".safepool-pool.json"

var ErrNoExchange = errors.New("no Exchange available")
var ErrInvalidSignature = errors.New("signature is invalid")
var ErrNotTrusted = errors.New("the author is not a trusted user")
var ErrNotAuthorized = errors.New("no authorization for this file")

// type SafeConfig struct {
// 	Version float32
// 	Id      uint64
// }

type Consumer interface {
	TimeOffset(s *Pool) time.Time
	Accept(s *Pool, h Head) bool
}

type Pool struct {
	Name      string
	Self      security.Identity
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
	//Since is the keyId used when the identity was added to the Pool access
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

type Config struct {
	Name    string
	Configs []transport.Config
}

func Define(c Config) error {
	return sqlSave(c.Name, c.Configs)
}

func List() []string {
	names, _ := sqlList()
	return names
}

func Create(self security.Identity, name string) (*Pool, error) {
	configs, err := sqlLoad(name)
	if core.IsErr(err, "unknown pool %s: %v", name) {
		return nil, err
	}

	s := &Pool{
		Name:        name,
		Self:        self,
		lastReplica: time.Now(),
	}
	err = s.connectSafe(name, configs)
	if err != nil {
		return nil, err
	}

	s.masterKeyId = snowflake.ID()
	s.masterKey = security.GenerateBytesKey(32)
	err = s.sqlSetKey(s.masterKeyId, s.masterKey)
	if core.IsErr(err, "çannot store master encryption key to db: %v") {
		return nil, err
	}
	err = security.SetIdentity(self)
	if core.IsErr(err, "çannot save identity to db: %v") {
		return nil, err
	}

	err = s.sqlSetIdentity(self, s.masterKeyId)
	if core.IsErr(err, "çannot link identity to save: %v") {
		return nil, err
	}

	if !ForceCreation {
		_, err = s.e.Stat(path.Join(s.Name, ".access"))
		if err == nil {
			return nil, ErrNotAuthorized
		}
	}

	err = s.ExportAccessFile(s.e)
	if core.IsErr(err, "cannot export access file: %v") {
		return nil, err
	}

	return s, err
}

// Init initialized a domain on the specified exchangers
func Open(self security.Identity, name string) (*Pool, error) {
	configs, err := sqlLoad(name)
	if core.IsErr(err, "unknown pool %s: %v", name) {
		return nil, err
	}
	s := &Pool{
		Name: name,
		Self: self,
	}
	err = s.connectSafe(name, configs)
	if err != nil {
		return nil, err
	}

	_, err = s.ImportAccess(s.e)
	return s, err
}

func (p *Pool) List(afterId uint64, afterTs time.Time) []Head {
	hs, _ := p.list(afterId, afterTs)
	return hs
}

func (p *Pool) Post(name string, r io.Reader) (Head, error) {
	id := snowflake.ID()
	n := path.Join(p.Name, fmt.Sprintf("%d.body", id))
	hr, err := p.writeFile(n, r)
	if core.IsErr(err, "cannot post file %s to %s: %v", name, p.e) {
		return Head{}, err
	}

	hash := hr.Hash()
	signature, err := security.Sign(p.Self, hash)
	if core.IsErr(err, "cannot sign file %s.body in %s: %v", name, p.e) {
		return Head{}, err
	}
	h := Head{
		Id:        id,
		Name:      name,
		Size:      hr.Size(),
		Hash:      hash,
		ModTime:   time.Now(),
		Author:    p.Self.Public(),
		Signature: signature,
	}
	data, err := json.Marshal(h)
	if core.IsErr(err, "cannot marshal header to json: %v") {
		return Head{}, err
	}

	n = path.Join(p.Name, fmt.Sprintf("%d.head", id))
	_, err = p.writeFile(n, bytes.NewBuffer(data))
	core.IsErr(err, "cannot write header %s.head in %s: %v", name, p.e)

	return h, nil
}

func (p *Pool) Get(id uint64, w io.Writer) error {
	headName := path.Join(p.Name, fmt.Sprintf("%d.head", id))
	bodyName := path.Join(p.Name, fmt.Sprintf("%d.body", id))

	h, err := p.readHead(headName)
	if core.IsErr(err, "cannot read header '%s': %v") {
		return err
	}

	hr, err := p.readFile(bodyName, w)
	if core.IsErr(err, "cannot read body '%s': %v", bodyName) {
		return err
	}
	hash := hr.Hash()
	if !bytes.Equal(hash, h.Hash) {
		return ErrInvalidSignature
	}
	return nil
}

func (p *Pool) Close() {
	for _, e := range p.exchangers {
		_ = e.Close()
	}
}

func (p *Pool) Delete() error {
	for _, e := range p.exchangers {
		err := e.Delete(p.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pool) Identities() ([]Identity, error) {
	identities, err := p.sqlGetIdentities(false)
	return identities, err
}

func (p *Pool) Poll() {
	logrus.Infof("poll request on %s", p.Name)

	timeOffset := time.Now()
	offsets := map[Consumer]time.Time{}
	for _, c := range p.Consumers {
		o := c.TimeOffset(p)
		offsets[c] = o
		if timeOffset.After(o) {
			timeOffset = o
		}
	}

	for _, h := range p.List(0, timeOffset) {
		for _, c := range p.Consumers {
			if c.Accept(p, h) {
				break
			}
		}
	}

	if time.Since(p.lastReplica) > ReplicaPeriod {
		p.replica()
	}
}
