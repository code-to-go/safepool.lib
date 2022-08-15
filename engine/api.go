package engine

import (
	"encoding/json"
	"path/filepath"
	"weshare/core"
	"weshare/exchanges"
	"weshare/sql"

	"github.com/adrg/xdg"
	"github.com/sirupsen/logrus"
)

const weshareFolder = "Weshare"
const wesharePathKey = "weshare.path"

var WesharePath string
var Identity core.Identity

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
		WesharePath = filepath.Join(xdg.UserDirs.Documents, weshareFolder)
		err = sql.SetConfig("", wesharePathKey, WesharePath, 0, nil)
	}

	if core.IsErr(setIdentity(), "cannot set identity: %v") {
		return err
	}

	return err
}

func setIdentity() error {
	// load or create identity
	if _, _, data, ok := sql.GetConfig("", "identity"); ok {
		json.Unmarshal(data, &Identity)
	} else {
		Identity, err := core.NewIdentity()
		if err != nil {
			return err
		}
		data, err = json.Marshal(Identity)
		if err != nil {
			return err
		}
		err = sql.SetConfig("", "identity", "", 0, data)
		if err != nil {
			return err
		}
	}
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

// SetDomain add a new domain to the user
func SetDomain(domain string, configs []exchanges.Config) error {
	err := sql.SetDomain(domain, configs)
	if err == nil {
		connectDomain(domain)
	}
	return err
}

func SetStagedState(file core.File, staged bool) error {
	f, err := sql.GetFile(file.Domain, file.Name, file.Author)
	if core.IsErr(err, "cannot find file '%s' in db: %v", file.Name) {
		return err
	}

	if staged {
		f.State |= core.Staged
	} else {
		f.State &= ^core.Staged
	}
	err = sql.SetFile(f)
	core.IsErr(err, "cannot set PushRequest status for %s: %v", file.Name)
	return err
}

// Status returns the current
func Status(domain string) ([]core.File, error) {
	return sql.GetFilesWithUpdates(domain)
}

func SyncAll() ([]core.File, error) {
	return nil, nil
}
