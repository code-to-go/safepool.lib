package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/code-to-go/safepool.lib/core"
	pool "github.com/code-to-go/safepool.lib/pool"
	"github.com/code-to-go/safepool.lib/security"
	"github.com/godruoyi/go-snowflake"
	"github.com/sirupsen/logrus"
)

type Message struct {
	Id          uint64   `json:"id,string"`
	Author      string   `json:"author"`
	Content     string   `json:"content"`
	ContentType string   `json:"contentType"`
	Attachments [][]byte `json:"attachments"`
	Signature   []byte   `json:"signature"`
}

func getHash(m *Message) []byte {
	h := security.NewHash()
	h.Write([]byte(m.Content))
	h.Write([]byte(m.ContentType))
	h.Write([]byte(m.Author))
	for _, a := range m.Attachments {
		h.Write(a)
	}
	return h.Sum(nil)
}

type Chat struct {
	Pool *pool.Pool
}

func Get(p *pool.Pool) Chat {
	return Chat{
		Pool: p,
	}
}

func (c *Chat) TimeOffset(s *pool.Pool) time.Time {
	return sqlGetOffset(s.Name)
}

func (c *Chat) Accept(s *pool.Pool, head pool.Head) bool {
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
	err = s.Get(head.Id, nil, &buf)
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

	err = sqlSetMessage(s.Name, uint64(id), m.Author, m, head.TimeStamp)
	core.IsErr(err, "cannot write message %s to db:%v", head.Name)
	return true
}

func (c *Chat) Post(m Message) (uint64, error) {
	m.Id = snowflake.ID()
	m.Author = c.Pool.Self.Id()
	h := getHash(&m)
	signature, err := security.Sign(c.Pool.Self, h)
	if core.IsErr(err, "cannot sign chat message: %v") {
		return 0, err
	}
	m.Signature = signature

	data, err := json.Marshal(m)
	if core.IsErr(err, "cannot sign chat message: %v") {
		return 0, err
	}

	go func() {
		name := fmt.Sprintf("/chat/%d.chat", m.Id)
		_, err = c.Pool.Post(name, bytes.NewBuffer(data), nil)
		core.IsErr(err, "cannot write chat message: %v")
	}()

	err = sqlSetMessage(c.Pool.Name, m.Id, c.Pool.Self.Id(), m, time.Now())
	if core.IsErr(err, "cannot save message to db: %v") {
		return 0, err
	}

	core.Info("added chat message with id %d", m.Id)
	return m.Id, nil
}

func (c *Chat) Pull(afterId, beforeId uint64, limit int) ([]Message, error) {
	messages := sqlGetMessages(c.Pool.Name, afterId, beforeId, limit)
	return messages, nil
}
