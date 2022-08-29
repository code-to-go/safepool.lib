package auth

import (
	"math/rand"
	"time"
	"weshare/model"
	"weshare/security"

	"github.com/godruoyi/go-snowflake"
)

const aesKeySize = 32

func NewDomain(Name string, user User) model.Domain {
	snowflakeId := snowflake.ID()

	token := make([]byte, aesKeySize)
	rand.Seed(time.Now().UnixNano())
	rand.Read(token)

	return model.Domain{
		Name:        Name,
		Snowflakeid: snowflakeId,
		Users:       []security.Identity{},
		Key:         token,
		LegacyKeys:  map[int64][]byte{},
	}
}
