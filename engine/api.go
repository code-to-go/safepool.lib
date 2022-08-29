package engine

import (
	"os"
	"path/filepath"
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

	data, err := security.MarshalIdentity(identity, true)
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

	// load or set weshare path
	var ok bool
	WesharePath, _, _, ok = sql.GetConfig("", wesharePathKey)
	if !ok {
		return core.ErrNotInitialized
	}

	if core.IsErr(loadIdentity(), "cannot load identity: %v") {
		return err
	}

	return err
}

func loadIdentity() error {
	_, _, data, ok := sql.GetConfig("", "identity")
	if !ok {
		return core.ErrNotInitialized
	}

	identity, err := security.UnmarshalIdentity(data)
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

// SetAccess add a new domain to the user
func Join(access model.Access) error {
	err := sql.SetAccess(access)
	if core.IsErr(err, "cannot store access information to db: %v") {
		return err
	}

	return err
}

func SetStagedState(file model.File, staged bool) error {
	f, err := sql.GetFile(file.Domain, file.Name, file.Author)
	if core.IsErr(err, "cannot find file '%s' in db: %v", file.Name) {
		return err
	}

	if staged {
		f.State |= model.Staged
	} else {
		f.State &= ^model.Staged
	}
	err = sql.SetFile(f)
	core.IsErr(err, "cannot set PushRequest status for %s: %v", file.Name)
	return err
}

// Status returns the current
func Status(domain string) ([]model.File, error) {
	return sql.GetFilesWithUpdates(domain)
}

func SyncAll() ([]model.File, error) {
	return nil, nil
}
