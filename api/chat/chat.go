package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/safe"
	"github.com/code-to-go/safepool/security"
	"github.com/godruoyi/go-snowflake"
	"github.com/sirupsen/logrus"
)

type Message struct {
	Id          uint64
	Author      security.Identity
	Content     string
	ContentType string
	Attachments [][]byte
	Signature   []byte
}

func getHash(m *Message) []byte {
	h := security.NewHash()
	h.Write([]byte(m.Content))
	h.Write([]byte(m.ContentType))
	h.Write(m.Author.SignatureKey.Public)
	for _, a := range m.Attachments {
		h.Write(a)
	}
	return h.Sum(nil)
}

type Chat struct {
	Safe *safe.Safe
}

func (c *Chat) TimeOffset(s *safe.Safe) time.Time {
	return sqlGetOffset(s.Name)
}

func (c *Chat) Accept(s *safe.Safe, head safe.Head) bool {
	name := head.Name
	if !strings.HasPrefix(name, "/chat/") || !strings.HasSuffix(name, ".chat") || head.Size > 10*1024*1024 {
		return false
	}
	name = path.Base(name)
	id, err := strconv.ParseInt(name[0:len(name)-5], 10, 64)
	if err != nil {
		return false
	}

	buf := bytes.Buffer{}
	err = s.Get(head.Id, &buf)
	if core.IsErr(err, "cannot read %s from %s: %v", head.Name, s.Name) {
		return true
	}

	var m Message
	err = json.Unmarshal(buf.Bytes(), &m)
	if core.IsErr(err, "invalid chat message %s: %v", head.Name) {
		return true
	}

	h := getHash(&m)
	if !security.Verify(m.Author, h, m.Signature) {
		logrus.Error("message %s has invalid signature", head.Name)
		return true
	}

	err = sqlSetMessage(s.Name, uint64(id), m.Author.SignatureKey.Public, m, head.TimeStamp)
	core.IsErr(err, "cannot write message %s to db:%v", head.Name)
	return true
}

func (c *Chat) Post(m Message) error {
	h := getHash(&m)
	signature, err := security.Sign(c.Safe.Self, h)
	if core.IsErr(err, "cannot sign chat message: %v") {
		return err
	}
	m.Signature = signature

	data, err := json.Marshal(m)
	if core.IsErr(err, "cannot sign chat message: %v") {
		return err
	}

	name := fmt.Sprintf("/chat/%d.chat", snowflake.ID())
	_, err = c.Safe.Post(name, bytes.NewBuffer(data))
	core.IsErr(err, "cannot write chat message: %v")

	return nil
}

func (c *Chat) Pull(beforeId uint64, limit int) ([]Message, error) {
	messages := sqlGetMessages(c.Safe.Name, beforeId, limit)
	return messages, nil
}
