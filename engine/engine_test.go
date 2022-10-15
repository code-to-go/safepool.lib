package engine

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"weshare/model"
	"weshare/sql"
	"weshare/transport"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLocalSync(t *testing.T) {
	sql.DbName = "weshare.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	Start()

	data := []byte("This is just some random text for test purposes")
	os.RemoveAll(filepath.Join(WesharePath, "test.weshare.zone"))
	os.MkdirAll(filepath.Join(WesharePath, "test.weshare.zone"), 0755)
	for i := 0; i < len(data); i++ {
		ioutil.WriteFile(filepath.Join(WesharePath, "test.weshare.zone", fmt.Sprintf("T%d", i)), data[i:], 0755)
	}

	Add(filepath.Join(WesharePath, "test.weshare.zone", "T0"), true)
	files, err := State("test.weshare.zone")
	assert.NoErrorf(t, err, "cannot get status: %v", err)

	if files[0].State&model.Staged == 0 {
		t.Errorf("Expected staged")
		t.Fail()
	}
	for i := 0; i < len(data); i++ {
		os.Remove(filepath.Join(WesharePath, "test.weshare.zone", fmt.Sprintf("T%d", i)))
	}

}

func TestRemoteSync(t *testing.T) {
	sql.DbName = "weshare.test.db"
	sql.DeleteDB()
	sql.LoadSQLFromFile("../sql/sqlite.sql")
	Start()

	var config transport.Config
	data, _ := ioutil.ReadFile("../../credentials/s3-2.yaml")
	yaml.Unmarshal(data, &config)

	Join(model.Transport{
		Domain:    "test.weshare.zone",
		Exchanges: []transport.Config{config},
	})

}
