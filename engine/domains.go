package engine

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"path"
	"sync"
	"time"
	"weshare/core"
	"weshare/exchanges"
	"weshare/sql"

	"github.com/sirupsen/logrus"
)

var Domains = map[string]exchanges.Exchanger{}
var DomainsMutex = &sync.Mutex{}

const pingName = ".reserved.ping.%d.test"

func pingExchanger(e exchanges.Exchanger, domain string, data []byte) (time.Duration, error) {
	start := time.Now()
	name := path.Join(domain, fmt.Sprintf(pingName, start.UnixMilli()))
	err := e.Write(name, bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}

	var buf bytes.Buffer
	err = e.Read(name, &buf)
	if err != nil {
		return 0, err
	}

	e.Delete(name)

	if bytes.Compare(data, buf.Bytes()) == 0 {
		return time.Now().Sub(start), nil
	} else {
		return 0, err
	}
}

func connectDomain(domain string) error {
	DomainsMutex.Lock()
	defer DomainsMutex.Unlock()

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	configs, err := sql.GetDomain(domain)
	if err != nil {
		logrus.Errorf("no valid configuration for domain %s", domain)
		if Domains[domain] != nil {
			Domains[domain].Close()
		}
		Domains[domain] = nil
		return err
	}
	min := time.Duration(math.MaxInt64)
	for _, c := range configs {
		e, err := exchanges.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}

		ping, err := pingExchanger(e, domain, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		} else {
			logrus.Infof("connection to %s is %v", e, ping)
		}
		if ping < min {
			min = ping
			if Domains[domain] != nil {
				Domains[domain].Close()
			}
			Domains[domain] = e
		}
	}
	logrus.Infof("connected to %s with ping %s", Domains[domain], min)
	return nil
}

func connectDomains() error {
	domains, err := sql.GetDomains()
	if core.IsErr(err, "cannot load the domains from db: %v") {
		return err
	}

	for _, domain := range domains {
		connectDomain(domain)
	}
	return nil
}
