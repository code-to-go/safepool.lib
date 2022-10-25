package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"weshare/core"
	"weshare/model"
	"weshare/security"
	"weshare/sql"

	"github.com/adrg/xdg"
	"github.com/sirupsen/logrus"
)

const weshareFolder = "weshare"
const wesharePathKey = "weshare.path"

var WesharePath string
var Self security.Identity

func Init(nick string, wesharePath string) (security.Identity, error) {
	var identity security.Identity

	err := sql.OpenDB()
	if err != nil {
		return identity, err
	}

	if wesharePath != "" {
		WesharePath = wesharePath
	} else {
		WesharePath = filepath.Join(xdg.UserDirs.Documents, weshareFolder)
		os.MkdirAll(WesharePath, 0755)
	}
	os.MkdirAll(WesharePath, 0755)
	_, err = os.Stat(WesharePath)
	if core.IsErr(err, "weshare path '%s' is invalid: %v", wesharePath) {
		return identity, err
	}

	err = sql.SetConfig("", wesharePathKey, WesharePath, 0, nil)
	if core.IsErr(err, "cannot store weshare path to db: %v", WesharePath) {
		return identity, err
	}

	identity, err = security.NewIdentity(nick)
	if core.IsErr(err, "cannot generate identity: %v") {
		return identity, err
	}

	data, err := json.Marshal(identity)
	if core.IsErr(err, "cannot marshal identity: %v") {
		return identity, err
	}

	err = sql.SetConfig("", "identity", "", 0, data)
	if core.IsErr(err, "cannot store identity to DB: %v") {
		return identity, err
	}

	return identity, err
}

// Start connect to the database and load the configuration
func Start() error {
	err := sql.OpenDB()
	if err != nil {
		return err
	}

	err = loadIdentity()
	if core.IsErr(err, "cannot load identity: %v") {
		return err
	}

	// load or set weshare path
	var ok bool
	WesharePath, _, _, ok = sql.GetConfig("", wesharePathKey)
	if !ok {
		return core.ErrNotInitialized
	}

	return err
}

func loadIdentity() error {
	_, _, data, ok := sql.GetConfig("", "identity")
	if !ok {
		return core.ErrNotInitialized
	}

	var identity security.Identity
	err := json.Unmarshal(data, &identity)
	if core.IsErr(err, "cannot unmarshal identity from db: %v") {
		return err
	}
	Self = identity
	return nil
}

func Stop() error {
	return sql.CloseDB()
}

// ListDomains returns the registered domains for the current user
func ListDomains() []string {
	domains, err := sql.GetDomains()
	if err != nil {
		logrus.Errorf("cannot read domain list from db: %v", err)
	}
	return domains
}

// Join add a new domain to the user
func Join(access model.Transport) error {
	err := sql.SetAccess(access)
	if core.IsErr(err, "cannot store access information to db: %v") {
		return err
	}

	// err = accessDomain(access.Domain)
	// if core.IsErr(err, "cannot access to domain %s: %v", access.Domain) {
	// 	return err
	// }
	return nil
}

func Add(filename string, staged bool) error {
	rel, err := filepath.Rel(WesharePath, filename)
	if core.IsErr(err, "cannot find path relative to weshare root for %s", filename) {
		return core.ErrInvalidFilePath
	}

	p := strings.SplitN(rel, string(os.PathSeparator), 2)
	if len(p) != 2 {
		return core.ErrInvalidFilePath
	}

	domain, name := p[0], p[1]
	syncLocalToDB(domain)

	file := model.File{
		Domain: domain,
		Name:   name,
	}

	files, err := sql.GetFilesByName(file.Domain, file.Name)
	if core.IsErr(err, "cannot find file '%s' in db: %v", file.Name) {
		return err
	}

	for _, f := range files {
		if staged {
			f.State |= model.Staged | model.Watched
		} else {
			f.State &= ^(model.Staged | model.Watched)
		}
		err = sql.SetFile(f)
		if err != nil {
			return err
		}
	}

	core.IsErr(err, "cannot set PushRequest status for %s: %v", file.Name)
	return err
}

// State returns the current
func State(domain string) ([]model.File, error) {
	err := syncLocalToDB(domain)
	if core.IsErr(err, "cannot sync local to db: %v") {
		return nil, err
	}

	return sql.GetFilesWithUpdates(domain)
}

func Update(domain string) ([]model.File, error) {
	return nil, nil
}
