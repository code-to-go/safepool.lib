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

func writeTopicConfig(e transport.Exchanger, topic string, c TopicConfig) error {
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

	if bytes.Equal(data, buf.Bytes()) {
		return time.Since(start), nil
	} else {
		return 0, err
	}
}

func (t *Topic) getExchangers(configs []transport.Config) {
	for _, e := range t.exchangers {
		e.Close()
	}
	t.exchangers = nil

	for _, c := range configs {
		e, err := transport.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}

		c, err := readTopicConfig(e, t.Name)
		if err != nil {
			err = writeTopicConfig(e, t.Name, TopicConfig{
				Version: 1.0,
				Id:      t.Id,
			})
			if core.IsErr(err, "cannot write topic config to %s. Skip", e) {
				e.Close()
				continue
			}
		} else if t.Id == 0 {
			t.Id = c.Id
		} else if c.Id != t.Id {
			core.IsErr(err, "exchange %s has unexpected key. Skip it", e)
			e.Close()
			continue
		}

		t.exchangers = append(t.exchangers, e)
	}
}

func (t *Topic) findPrimary() {
	min := time.Duration(math.MaxInt64)

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	t.primary = nil
	for _, e := range t.exchangers {
		ping, err := pingExchanger(e, t.Name, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		}
		if ping < min {
			min = ping
			t.primary = e
		}
	}
}

func (t *Topic) connectTopic(configs []transport.Config) error {
	t.getExchangers(configs)
	t.findPrimary()

	if t.primary == nil {
		logrus.Warnf("no available exchange for domain %s", t.Name)
		return ErrNoExchange
	} else {
		return nil
	}
}
