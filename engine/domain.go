package engine

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
	"weshare/auth"
	"weshare/core"
	"weshare/exchanges"
	"weshare/model"
	"weshare/protocol"
	"weshare/security"
	"weshare/sql"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

var Connections = map[string]exchanges.Exchanger{}
var ConnectionsMutex = &sync.Mutex{}

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

func initDomain(d model.Domain, e exchanges.Exchanger) error {
	encKey, err := security.Generate32BytesKey()
	if core.IsErr(err, "cannot generate symmetric encryption key: %v") {
		return err
	}

	encPubKey, err := security.EcEncrypt(Self, encKey)
	if core.IsErr(err, "cannot encrypt key with asymmetric encryption: %v") {
		return err
	}

	err = sql.SetEncKey(d.Name, 0, encKey)
	if core.IsErr(err, "cannot store symmetric encryption key to DB: %v") {
		return err
	}

	myIdentity, err := security.MarshalIdentity(Self, false)
	if core.IsErr(err, "cannot marshal self identity: %v") {
		return err
	}

	domainFile := protocol.DomainFile{
		Version: 1.0,
		Name:    d.Name,
		EncId:   0,
		Users: []*protocol.User{
			{
				Identity: myIdentity,
				Active:   true,
				EncKey:   encPubKey,
			},
		},
	}

	data, err := proto.Marshal(&domainFile)
	if core.IsErr(err, "cannot marshal domain file: %v") {
		return err
	}

	buf := bytes.Buffer{}
	err = security.SignAndWrite(Self, data, &buf, nil)
	if core.IsErr(err, "cannot sign domain file: %v") {
		return err
	}

	domainPath := path.Join(d.Name, core.DomainFilename)
	c := Connections[d.Name]

	err = c.Write(domainPath, &buf)
	return nil
}

func validateDomain(d model.Domain, e exchanges.Exchanger) error {
	domainFile := path.Join(d.Name, core.DomainFilename)

	_, err := e.Stat(domainFile)
	if os.IsNotExist(err) {
		return initDomain(d, e)
	}

	buf := bytes.Buffer{}
	err = e.Read(domainFile, &buf)
	if core.IsErr(err, "cannot read domain file from %s/%s", e, d.Name) {
		return err
	}

	admins, err := sql.GetUsersIdentities(d.Name, true, true)
	if core.IsErr(err, "cannot read admins identities from db: %v") {
		return err
	}
	data, _, err := security.ReadAndVerify(admins, &buf)
	if core.IsErr(err, "domain file is not signed by a known admin") {
		return err
	}

	var p protocol.DomainFile
	err = proto.Unmarshal(data, &p)
	if core.IsErr(err, "cannot unmarshal domain file to proto message") {
		return err
	}

	keys, err := sql.GetEncKeys(d.Name)
	if core.IsErr(err, "cannot read keys from db") {
		return err
	}

	for _, u := range p.Users {
		identity, err := security.UnmarshalIdentity(u.Identity)
		if core.IsErr(err, "cannot unmarshall identity") {
			continue
		}

		sql.SetUser(d.Name, auth.User{
			Identity: identity,
			Active:   u.Active,
		})

		if bytes.Compare(identity.Keys[security.Ed25519].Public, Self.Keys[security.Ed25519].Public) != 0 {
			continue
		}

		encKey, err := security.EcDecrypt(identity, u.EncKey)
		if core.IsErr(err, "cannot decrypt AES encryption key with Secp256k1 key") {
			return err
		}

		if _, ok := keys[p.EncId]; ok {
			err = sql.SetEncKey(domainFile, p.EncId, encKey)
			if core.IsErr(err, "cannot store AES encryption key to DB") {
				return err
			}
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
	min := time.Duration(math.MaxInt64)
	for _, c := range domain.Exchanges {
		e, err := exchanges.NewExchanger(c)
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
		if ping < min {
			min = ping
			if Connections[name] != nil {
				Connections[name].Close()
			}
			Connections[name] = e
		}
	}
	logrus.Infof("connected to %s with ping %s", Connections[name], min)
	return nil
}

func accessDomains() error {
	domains, err := sql.GetDomains()
	if core.IsErr(err, "cannot load the domains from db: %v") {
		return err
	}

	for _, domain := range domains {
		accessDomain(domain)
	}
	return nil
}
