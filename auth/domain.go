package auth

import (
	"math/rand"
	"time"
	"weshare/core"

	"github.com/godruoyi/go-snowflake"
)

type Domain struct {
	Name        string
	Snowflakeid uint64
	Users       []User
	Key         []byte
	LegacyKeys  map[int64][]byte
}

type User struct {
	Public core.Public
	Name   string
	Name2  string
}

const aesKeySize = 32

func NewDomain(Name string, user User) Domain {
	snowflakeId := snowflake.ID()

	token := make([]byte, aesKeySize)
	rand.Seed(time.Now().UnixNano())
	rand.Read(token)

	return Domain{
		Name:        Name,
		Snowflakeid: snowflakeId,
		Users:       []User{user},
		Key:         token,
		LegacyKeys:  map[int64][]byte{},
	}
}
