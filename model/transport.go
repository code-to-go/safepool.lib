package model

import (
	"weshare/transport"
)

type Transport struct {
	Domain    string
	Granted   bool
	Exchanges []transport.Config
}
