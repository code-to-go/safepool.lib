package access

import (
	"io"
	"time"
	"weshare/security"
	"weshare/transport"
)

const TopicConfigFile = ".weshare-topic.json"

type TopicConfig struct {
	Version float32
}

type Topic struct {
	name                string
	primaryExchanger    transport.Exchanger
	secondaryExchangers []transport.Exchanger
}

type Head struct {
	Id      uint64
	Name    string
	Size    int64
	Hash    security.Hash256
	ModTime time.Time
}

// Init initialized a domain on the specified exchangers
func CreateTopic(topic string, configs []transport.Config) (Topic, error) {

}

// Open  connects to a domain
func OpenTopic(topic string, configs []transport.Config) (Topic, error) {
}

func (t Topic) Read(after uint64) []Head {

}

func (t Topic) Add(name string, r io.Reader) (Head, error) {
	return nil
}

func (t Topic) Get(id uint64, w io.Writer) error {
	return nil
}

func (t Topic) Close() error {
	return nil
}

func Trust(thread string, add []security.Identity, remove []security.Identity) string {

}
