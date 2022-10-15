package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
	"weshare/core"
	"weshare/model"
	"weshare/security"
	"weshare/sql"
	"weshare/transport"

	"github.com/sirupsen/logrus"
)

var Connections = map[string]transport.Exchanger{}
var ConnectionsMutex = &sync.Mutex{}

const pingName = ".reserved.ping.%d.test"

func pingExchanger(e transport.Exchanger, domain string, data []byte) (time.Duration, error) {
	start := time.Now()
	name := path.Join(domain, fmt.Sprintf(pingName, start.UnixMilli()))
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

func initDomain(domain string, e transport.Exchanger) error {
	domainFile := path.Join(domain, core.UsersFilename)
	signFile := path.Join(domain, core.UsersFilesign)

	err := createLock(domain, e, time.Second*30)
	if core.IsErr(err, "cannot create lock file") {
		return err
	}

	encKey, err := security.Generate32BytesKey()
	if core.IsErr(err, "cannot generate symmetric encryption key: %v") {
		return err
	}

	err = sql.SetTrusted(domain, Self, true)
	if core.IsErr(err, "cannot add self to the trusted users: %v") {
		return err
	}

	encPubKey, err := security.EcEncrypt(Self, encKey)
	if core.IsErr(err, "cannot encrypt key with asymmetric encryption: %v") {
		return err
	}

	err = sql.SetEncKey(domain, 0, encKey)
	if core.IsErr(err, "cannot store symmetric encryption key to DB: %v") {
		return err
	}

	data, err := json.Marshal(model.UsersFile{
		Version:      1.0,
		GenerationId: 0,
		Users: []model.User{
			{
				Identity: Self.Public(),
				Active:   true,
			},
		},
		EncKeys: map[uint64][]byte{
			Self.Id: encPubKey,
		},
	})
	if core.IsErr(err, "cannot marshal domain file: %v") {
		return err
	}

	sign, err := security.Sign(Self, data)
	if core.IsErr(err, "cannot sign domain file: %v") {
		return err
	}
	err = e.Write(signFile, bytes.NewBuffer(sign))
	if core.IsErr(err, "cannot write domain signature: %v") {
		return err
	}
	err = e.Write(domainFile, bytes.NewBuffer(data))
	if core.IsErr(err, "cannot write domain file: %v") {
		return err
	}
	err = removeLock(domain, e)
	if core.IsErr(err, "cannot delete lock file: %v") {
		return err
	}

	return nil
}

func validateDomain(domain string) error {
	e := Connections[domain]
	if e == nil {
		logrus.Warnf("no connection available to domain '%s'", domain)
		return core.ErrNoExchange
	}
	usersFile := path.Join(domain, core.UsersFilename)
	waitForLock(domain, e)
	defer removeLock(domain, e)

	_, err := e.Stat(usersFile)
	if os.IsNotExist(err) {
		return initDomain(domain, e)
	}
	data := bytes.Buffer{}
	err = e.Read(usersFile, nil, &data)
	if core.IsErr(err, "cannot read users file %s from %s/%s", usersFile, e, domain) {
		return err
	}

	signFile := usersFile + ".sign"
	signature := bytes.Buffer{}
	err = e.Read(signFile, nil, &signature)
	if core.IsErr(err, "cannot read sign file from %s/%s", e, domain) {
		return err
	}

	trusted, err := sql.GetAllTrusted(domain)
	if core.IsErr(err, "cannot read admins identities from db: %v") {
		return err
	}

	trusted = append(trusted, model.User{Identity: Self, Active: true})
	valid := false
	for _, t := range trusted {
		if security.Verify(t.Identity, data.Bytes(), signature.Bytes()) {
			valid = true
			break
		}
	}

	if !valid {
		logrus.Warnf("user file is not signed by a trusted user")
		return err
	}

	var df model.UsersFile
	err = json.Unmarshal(data.Bytes(), &df)
	if core.IsErr(err, "cannot unmarshal domain file") {
		return err
	}

	keys, err := sql.GetEncKeys(domain)
	if core.IsErr(err, "cannot read keys from db") {
		return err
	}

	for _, u := range df.Users {
		err = sql.SetUser(domain, u)
		if core.IsErr(err, "cannot save user %s: %v", u.Identity.Nick) {
			return err
		}
	}

	encKey := df.EncKeys[Self.Id]
	if encKey == nil {
		return core.ErrNotAuthorized
	}
	encKey, err = security.EcDecrypt(Self, encKey)
	if core.IsErr(err, "cannot decrypt AES encryption key with Secp256k1 key") {
		return err
	}

	if _, ok := keys[df.GenerationId]; ok {
		err = sql.SetEncKey(domain, df.GenerationId, encKey)
		if core.IsErr(err, "cannot store AES encryption key to DB") {
			return err
		}
	}

	return nil
}

func accessDomain(name string) error {
	ConnectionsMutex.Lock()
	defer ConnectionsMutex.Unlock()

	data := make([]byte, 4192)
	rand.Seed(time.Now().Unix())
	rand.Read(data)

	domain, err := sql.GetAccess(name)
	if err != nil {
		logrus.Errorf("no valid configuration for domain %s", name)
		if Connections[name] != nil {
			Connections[name].Close()
		}
		Connections[name] = nil
		return err
	}
	var found bool
	min := time.Duration(math.MaxInt64)
	for _, c := range domain.Exchanges {
		e, err := transport.NewExchanger(c)
		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
			continue
		}

		ping, err := pingExchanger(e, name, data)
		if err != nil {
			logrus.Warnf("no connection to %v", e)
			continue
		} else {
			logrus.Infof("connection to %s is %v", e, ping)
		}
		found = true
		if ping < min {
			min = ping
			if Connections[name] != nil {
				Connections[name].Close()
			}
			Connections[name] = e
		}
	}
	if !found {
		logrus.Warnf("no available exchange for domain %s", name)
		return core.ErrNoExchange
	}

	logrus.Infof("connected to %s with ping %s", Connections[name], min)
	return validateDomain(name)
}

func accessDomains() error {
	domains, err := sql.GetDomains()
	if core.IsErr(err, "cannot load the domains from db: %v") {
		return err
	}

	for _, domain := range domains {
		err := accessDomain(domain)
		if err != nil {
			return err
		}
	}
	return nil
}
