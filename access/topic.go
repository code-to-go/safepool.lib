package access

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

const TopicConfigFile = ".weshare-topic.json"

var ErrNoExchange = errors.New("no Exchange available")
var ErrInvalidSignature = errors.New("signature is invalid")
var ErrNotTrusted = errors.New("the author is not a trusted user")
var ErrNotAuthorized = errors.New("no authorization for this file")

type TopicConfig struct {
	Version float32
	Id      uint64
}

type Topic struct {
	Self       security.Identity
	Name       string
	Id         uint64
	primary    transport.Exchanger
	exchangers []transport.Exchanger
	keyId      uint64
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

// Init initialized a domain on the specified exchangers
func CreateTopic(self security.Identity, name string, configs []transport.Config) (Topic, error) {
	t, err := OpenTopic(self, name, snowflake.ID(), configs)
	if err != nil {
		return t, err
	}

	_, err = t.addAccess([]security.Identity{self})
	return t, err
}

// Open  connects to a domain
func OpenTopic(self security.Identity, name string, id uint64, configs []transport.Config) (Topic, error) {
	t := Topic{
		Self: self,
		Name: name,
		Id:   id,
	}
	return t, t.connectTopic(configs)
}

func (t Topic) List(after uint64) []Head {
	t.refresh()

	return nil
}

func (t Topic) Post(name string, r io.Reader) (Head, error) {
	id := snowflake.ID()
	n := path.Join(t.Name, fmt.Sprintf("%d.body", id))
	hr, err := t.writeFile(n, r)
	if core.IsErr(err, "cannot post file %s to %s: %v", name, t.primary) {
		return Head{}, err
	}

	signature, err := security.Sign(t.Self, hr.Hash())
	if core.IsErr(err, "cannot sign file %s.body in %s: %v", name, t.primary) {
		return Head{}, err
	}
	h := Head{
		Id:        id,
		Name:      name,
		Size:      hr.Size(),
		Hash:      hr.Hash(),
		ModTime:   time.Now(),
		Author:    t.Self,
		Signature: signature,
	}
	data, err := json.Marshal(h)
	if core.IsErr(err, "cannot marshal header to json: %v") {
		return Head{}, err
	}

	n = path.Join(t.Name, fmt.Sprintf("%d.head", id))
	_, err = t.writeFile(n, bytes.NewBuffer(data))
	core.IsErr(err, "cannot write header %s.head in %s: %v", name, t.primary)

	return h, nil
}

func (t Topic) Get(id uint64, w io.Writer) error {
	return nil
}

func (t Topic) Close() {
	for _, e := range t.exchangers {
		_ = e.Close()
	}
}

func (t Topic) Delete() error {
	for _, e := range t.exchangers {
		err := e.Delete(t.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func Trust(thread string, add []security.Identity, remove []security.Identity) error {
	return nil
}
