package chat

import (
	"weshare/safe"
	"weshare/security"
)

type Message struct {
	Author      security.Identity
	Content     string
	ContentType string
	Attachments [][]byte
	Signature   []byte
}

func Start(s safe.Safe) {

}

func Last()
