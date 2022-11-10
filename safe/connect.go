package safe

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"path"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/transport"
	"time"

	"github.com/sirupsen/logrus"
)

func pingExchanger(e transport.Exchanger, safe string, data []byte) (time.Duration, error) {
	start := time.Now()
	name := path.Join(safe, fmt.Sprintf(pingName, start.UnixMilli()))
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

func (s *Safe) createExchangers(configs []transport.Config) {
	for _, e := range s.exchangers {
		e.Close()
	}
	s.exchangers = nil

	for _, c := range configs {
		e, err := transport.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}
		s.exchangers = append(s.exchangers, e)
	}
}

func (s *Safe) findPrimary() {
	min := time.Duration(math.MaxInt64)

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	s.e = nil
	for _, e := range s.exchangers {
		ping, err := pingExchanger(e, s.Name, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		}
		if ping < min {
			min = ping
			s.e = e
		}
	}
}

func (s *Safe) connectSafe(name string, configs []transport.Config) error {
	s.createExchangers(configs)
	s.findPrimary()
	if s.e == nil {
		logrus.Warnf("no available exchange for domain %s", s.Name)
		return ErrNoExchange
	} else {
		return nil
	}
}
