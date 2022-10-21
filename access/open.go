package access

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"path"
	"time"
	"weshare/core"
	"weshare/transport"

	"github.com/sirupsen/logrus"
)

func readTopicConfig(e transport.Exchanger, topic string) (TopicConfig, error) {
	var c TopicConfig
	name := path.Join(topic, TopicConfigFile)

	data, err := transport.ReadFile(e, name)
	if err == nil {
		err = json.Unmarshal(data, &c)
	}
	return c, err
}

func setTopicConfig(e transport.Exchanger, topic string, c TopicConfig) error {)
	name := path.Join(topic, TopicConfigFile)
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return transport.WriteFile(e, name, data)
}

func pingExchanger(e transport.Exchanger, topic string, data []byte) (time.Duration, error) {
	start := time.Now()
	name := path.Join(topic, fmt.Sprintf(pingName, start.UnixMilli()))
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

	if bytes.Compare(data, buf.Bytes()) == 0 {
		return time.Now().Sub(start), nil
	} else {
		return 0, err
	}
}

func (t *Topic) getExchangers(topic string, configs []transport.Config) error {
	min := time.Duration(math.MaxInt64)

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	var exchangers []transport.Exchanger
	primaryExchanger := -1
	for _, c := range configs {
		e, err := transport.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}

		ping, err := pingExchanger(e, topic, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		} else {
			logrus.Infof("connection to %s is %v", e, ping)
		}
		exchangers = append(exchangers, e)

		_, err = readTopicConfig(e, topic)
		if err == nil && ping < min {
			min = ping
			primaryExchanger = e
		}
	}
	if primaryExchanger == -1 {
		logrus.Warnf("no available exchange for domain %s", name)
		return ErrNoExchange
	}

}
