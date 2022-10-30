package safe

import (
	"sync"
	"weshare/transport"
)

var Connections = map[string]transport.Exchanger{}
var ConnectionsMutex = &sync.Mutex{}

const pingName = ".reserved.ping.%d.test"

// func initSafe(safe string, e transport.Exchanger) error {

// 	return nil
// }

// func validateDomain(domain string) error {
// 	e := Connections[domain]
// 	if e == nil {
// 		logrus.Warnf("no connection available to domain '%s'", domain)
// 		return core.ErrNoExchange
// 	}
// 	usersFile := path.Join(domain, core.UsersFilename)
// 	waitForLock(domain, e)
// 	defer removeLock(domain, e)

// 	_, err := e.Stat(usersFile)
// 	if os.IsNotExist(err) {
// 		return initDomain(domain, e)
// 	}
// 	data := bytes.Buffer{}
// 	err = e.Read(usersFile, nil, &data)
// 	if core.IsErr(err, "cannot read users file %s from %s/%s", usersFile, e, domain) {
// 		return err
// 	}

// 	signFile := usersFile + ".sign"
// 	signature := bytes.Buffer{}
// 	err = e.Read(signFile, nil, &signature)
// 	if core.IsErr(err, "cannot read sign file from %s/%s", e, domain) {
// 		return err
// 	}

// 	trusted, err := sql.GetAllTrusted(domain)
// 	if core.IsErr(err, "cannot read admins identities from db: %v") {
// 		return err
// 	}

// 	trusted = append(trusted, model.User{Identity: Self, Active: true})
// 	valid := false
// 	for _, t := range trusted {
// 		if security.Verify(t.Identity, data.Bytes(), signature.Bytes()) {
// 			valid = true
// 			break
// 		}
// 	}

// 	if !valid {
// 		logrus.Warnf("user file is not signed by a trusted user")
// 		return err
// 	}

// 	var df model.UsersFile
// 	err = json.Unmarshal(data.Bytes(), &df)
// 	if core.IsErr(err, "cannot unmarshal domain file") {
// 		return err
// 	}

// 	keys, err := sql.GetEncKeys(domain)
// 	if core.IsErr(err, "cannot read keys from db") {
// 		return err
// 	}

// 	for _, u := range df.Users {
// 		err = sql.SetUser(domain, u)
// 		if core.IsErr(err, "cannot save user %s: %v", u.Identity.Nick) {
// 			return err
// 		}
// 	}

// 	encKey := df.EncKeys[Self.Id]
// 	if encKey == nil {
// 		return core.ErrNotAuthorized
// 	}
// 	encKey, err = security.EcDecrypt(Self, encKey)
// 	if core.IsErr(err, "cannot decrypt AES encryption key with Secp256k1 key") {
// 		return err
// 	}

// 	if _, ok := keys[df.GenerationId]; ok {
// 		err = sql.SetEncKey(domain, df.GenerationId, encKey)
// 		if core.IsErr(err, "cannot store AES encryption key to DB") {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func accessDomain(name string) error {
// 	ConnectionsMutex.Lock()
// 	defer ConnectionsMutex.Unlock()

// 	data := make([]byte, 4192)
// 	rand.Seed(time.Now().Unix())
// 	rand.Read(data)

// 	domain, err := sql.GetAccess(name)
// 	if err != nil {
// 		logrus.Errorf("no valid configuration for domain %s", name)
// 		if Connections[name] != nil {
// 			Connections[name].Close()
// 		}
// 		Connections[name] = nil
// 		return err
// 	}
// 	var found bool
// 	min := time.Duration(math.MaxInt64)
// 	for _, c := range domain.Exchanges {
// 		e, err := transport.NewExchanger(c)
// 		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
// 			continue
// 		}

// 		ping, err := pingExchanger(e, name, data)
// 		if err != nil {
// 			logrus.Warnf("no connection to %v", e)
// 			continue
// 		} else {
// 			logrus.Infof("connection to %s is %v", e, ping)
// 		}
// 		found = true
// 		if ping < min {
// 			min = ping
// 			if Connections[name] != nil {
// 				Connections[name].Close()
// 			}
// 			Connections[name] = e
// 		}
// 	}
// 	if !found {
// 		logrus.Warnf("no available exchange for domain %s", name)
// 		return core.ErrNoExchange
// 	}

// 	logrus.Infof("connected to %s with ping %s", Connections[name], min)
// 	return validateDomain(name)
// }

// func validateDomain(domain string) error {
// 	e := Connections[domain]
// 	if e == nil {
// 		logrus.Warnf("no connection available to domain '%s'", domain)
// 		return core.ErrNoExchange
// 	}
// 	usersFile := path.Join(domain, core.UsersFilename)
// 	waitForLock(domain, e)
// 	defer removeLock(domain, e)

// 	_, err := e.Stat(usersFile)
// 	if os.IsNotExist(err) {
// 		return initDomain(domain, e)
// 	}
// 	data := bytes.Buffer{}
// 	err = e.Read(usersFile, nil, &data)
// 	if core.IsErr(err, "cannot read users file %s from %s/%s", usersFile, e, domain) {
// 		return err
// 	}

// 	signFile := usersFile + ".sign"
// 	signature := bytes.Buffer{}
// 	err = e.Read(signFile, nil, &signature)
// 	if core.IsErr(err, "cannot read sign file from %s/%s", e, domain) {
// 		return err
// 	}

// 	trusted, err := sql.GetAllTrusted(domain)
// 	if core.IsErr(err, "cannot read admins identities from db: %v") {
// 		return err
// 	}

// 	trusted = append(trusted, model.User{Identity: Self, Active: true})
// 	valid := false
// 	for _, t := range trusted {
// 		if security.Verify(t.Identity, data.Bytes(), signature.Bytes()) {
// 			valid = true
// 			break
// 		}
// 	}

// 	if !valid {
// 		logrus.Warnf("user file is not signed by a trusted user")
// 		return err
// 	}

// 	var df model.UsersFile
// 	err = json.Unmarshal(data.Bytes(), &df)
// 	if core.IsErr(err, "cannot unmarshal domain file") {
// 		return err
// 	}

// 	keys, err := sql.GetEncKeys(domain)
// 	if core.IsErr(err, "cannot read keys from db") {
// 		return err
// 	}

// 	for _, u := range df.Users {
// 		err = sql.SetUser(domain, u)
// 		if core.IsErr(err, "cannot save user %s: %v", u.Identity.Nick) {
// 			return err
// 		}
// 	}

// 	encKey := df.EncKeys[Self.Id]
// 	if encKey == nil {
// 		return core.ErrNotAuthorized
// 	}
// 	encKey, err = security.EcDecrypt(Self, encKey)
// 	if core.IsErr(err, "cannot decrypt AES encryption key with Secp256k1 key") {
// 		return err
// 	}

// 	if _, ok := keys[df.GenerationId]; ok {
// 		err = sql.SetEncKey(domain, df.GenerationId, encKey)
// 		if core.IsErr(err, "cannot store AES encryption key to DB") {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func accessDomain(name string) error {
// 	ConnectionsMutex.Lock()
// 	defer ConnectionsMutex.Unlock()

// 	data := make([]byte, 4192)
// 	rand.Seed(time.Now().Unix())
// 	rand.Read(data)

// 	domain, err := sql.GetAccess(name)
// 	if err != nil {
// 		logrus.Errorf("no valid configuration for domain %s", name)
// 		if Connections[name] != nil {
// 			Connections[name].Close()
// 		}
// 		Connections[name] = nil
// 		return err
// 	}
// 	var found bool
// 	min := time.Duration(math.MaxInt64)
// 	for _, c := range domain.Exchanges {
// 		e, err := transport.NewExchanger(c)
// 		if core.IsErr(err, "cannot connect to exchange %s: %v", c) {
// 			continue
// 		}

// 		ping, err := pingExchanger(e, name, data)
// 		if err != nil {
// 			logrus.Warnf("no connection to %v", e)
// 			continue
// 		} else {
// 			logrus.Infof("connection to %s is %v", e, ping)
// 		}
// 		found = true
// 		if ping < min {
// 			min = ping
// 			if Connections[name] != nil {
// 				Connections[name].Close()
// 			}
// 			Connections[name] = e
// 		}
// 	}
// 	if !found {
// 		logrus.Warnf("no available exchange for domain %s", name)
// 		return core.ErrNoExchange
// 	}

// 	logrus.Infof("connected to %s with ping %s", Connections[name], min)
// 	return validateDomain(name)
// }

// func accessDomains() error {
// 	domains, err := sql.GetDomains()
// 	if core.IsErr(err, "cannot load the domains from db: %v") {
// 		return err
// 	}

// 	for _, domain := range domains {
// 		err := accessDomain(domain)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func accessDomains() error {
// 	domains, err := sql.GetDomains()
// 	if core.IsErr(err, "cannot load the domains from db: %v") {
// 		return err
// 	}

// 	for _, domain := range domains {
// 		err := accessDomain(domain)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
