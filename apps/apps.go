package apps

import (
	"github.com/code-to-go/safepool/apps/chat"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/safe"
	"github.com/code-to-go/safepool/security"
)

var safes = map[string]*safe.Safe{}
var instances = map[string]int{}
var Self security.Identity

func acquireSafe(name string) (*safe.Safe, error) {
	s, ok := safes[name]
	if ok {
		instances[name] += 1
		return s, nil
	}
	configs, err := safe.Load(name)
	if core.IsErr(err, "unknown safe %s: %v", name) {
		return nil, err
	}
	s2, err := safe.OpenSafe(Self, name, configs)
	safes[name] = &s2
	instances[name] = 1
	return &s2, err
}

func releaseSafe(name string) error {
	i, ok := instances[name]
	if !ok {
		return core.ErrNoExchange
	}

	if i == 1 {
		s := safes[name]
		s.Close()
		delete(instances, name)
		delete(safes, name)
	} else {
		instances[name] = i - 1
	}
	return nil
}

func init() {
	var err error
	_, _, b, ok := sqlGetConfig("", "SELF")
	if ok {
		Self, err = security.IdentityFromBase64(b)
	} else {
		Self, err = security.NewIdentity("change.me")
	}
	if err != nil {
		panic("corrupted identity in DB")
	}
}

func NewChat(safe string) (chat.Chat, error) {
	s, err := acquireSafe(safe)
	return chat.Chat{
		Safe: s,
	}, err
}
