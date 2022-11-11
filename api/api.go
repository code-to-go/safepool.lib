package api

import (
	"github.com/code-to-go/safepool/api/chat"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/safe"
	"github.com/code-to-go/safepool/security"
)

var Self security.Identity

func init() {
	var err error
	s, _, _, ok := sqlGetConfig("", "SELF")
	if ok {
		Self, err = security.IdentityFromBase64(s)
	} else {
		Self, err = security.NewIdentity("Change Me")
		if err == nil {
			s, err = Self.Base64()
			if err == nil {
				err = sqlSetConfig("", "SELF", s, 0, nil)
			}
		}
	}
	if err != nil {
		panic("corrupted identity in DB")
	}
}

func SetNick(nick string) error {
	Self.Nick = nick
	s, err := Self.Base64()
	if core.IsErr(err, "cannot serialize self to db: %v") {
		return err
	}
	err = sqlSetConfig("", "SELF", s, 0, nil)
	core.IsErr(err, "cannot save nick to db: %v")
	return err
}

func CreateSafe(c safe.Config) (*safe.Safe, error) {
	err := safe.Define(c)
	if core.IsErr(err, "cannot define safe %s: %v", c.Name) {
		return nil, err
	}

	return safe.Create(Self, c.Name)
}

func ListSafe() []string {
	return safe.List()
}

func OpenSafe(name string) (*safe.Safe, error) {
	return safe.Open(Self, name)
}

func OpenChat(s *safe.Safe) chat.Chat {
	return chat.Chat{
		Safe: s,
	}
}
