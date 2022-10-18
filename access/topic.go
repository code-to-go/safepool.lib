package access

import (
	"io"
	"time"
	"weshare/security"
	"weshare/transport"
)

type Topic struct {
	name      string
	transport []transport.Exchanger
	primary   transport.Exchanger
}

type Head struct {
	Id      uint64
	Name    string
	Size    int64
	Hash    security.Hash256
	ModTime time.Time
}

// Init initialized a domain on the specified exchangers
func CreateTopic(topic string, transport []transport.Exchanger) (Topic, error) {

}

// Open  connects to a domain
func OpenTopic(topic string, transport []transport.Exchanger) (Topic, error) {
}

func (t Topic) ReadTopic(after uint64) []Head {

}

func Add(name string, r io.Reader) (Head, error) {
	return nil
}

func (t Topic) Get(id uint64, w io.Writer) error {
	return nil
}

func Trust(thread string, add []security.Identity, remove []security.Identity) string {

}
