package api

import (
	_ "embed"

	"github.com/code-to-go/safepool/api/chat"
	"github.com/code-to-go/safepool/core"
	pool "github.com/code-to-go/safepool/pool"
	"github.com/code-to-go/safepool/security"
	"github.com/code-to-go/safepool/sql"
)

var Self security.Identity

//go:embed sqlite.sql
var sqlliteDDL string

func Start() {
	sql.InitDDL = sqlliteDDL

	err := sql.OpenDB()
	if err != nil {
		panic("cannot open DB")
	}

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

func CreatePool(c pool.Config) (*pool.Pool, error) {
	err := pool.Define(c)
	if core.IsErr(err, "cannot define pool %s: %v", c.Name) {
		return nil, err
	}

	return pool.Create(Self, c.Name)
}

func ListPool() []string {
	return pool.List()
}

func OpenSafe(name string) (*pool.Pool, error) {
	return pool.Open(Self, name)
}

func OpenChat(p *pool.Pool) chat.Chat {
	return chat.Chat{
		Pool: p,
	}
}
