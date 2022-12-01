package pool

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"path"
	"time"

	"github.com/code-to-go/safepool.lib/core"
	"github.com/code-to-go/safepool.lib/transport"

	"github.com/sirupsen/logrus"
)

func pingExchanger(e transport.Exchanger, pool string, data []byte) (time.Duration, error) {
	start := time.Now()
	name := path.Join(pool, fmt.Sprintf(pingName, start.UnixMilli()))
	err := e.Write(name, bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}

	var buf bytes.Buffer
	err = e.Read(name, nil, &buf)
	if err != nil {
		return 0, err
	}
	e.Delete(name)

	if bytes.Equal(data, buf.Bytes()) {
		return time.Since(start), nil
	} else {
		return 0, err
	}
}

func (p *Pool) createExchangers(configs []transport.Config) {
	for _, e := range p.exchangers {
		e.Close()
	}
	p.exchangers = nil

	for _, c := range configs {
		e, err := transport.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}
		p.exchangers = append(p.exchangers, e)
	}
}

func (p *Pool) findPrimary() {
	min := time.Duration(math.MaxInt64)

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	p.e = nil
	for _, e := range p.exchangers {
		ping, err := pingExchanger(e, p.Name, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		}
		if ping < min {
			min = ping
			p.e = e
		}
	}
}

func (p *Pool) connectSafe(name string, configs []transport.Config) error {
	p.createExchangers(configs)
	p.findPrimary()
	if p.e == nil {
		logrus.Warnf("no available exchange for domain %s", p.Name)
		return ErrNoExchange
	} else {
		return nil
	}
}
