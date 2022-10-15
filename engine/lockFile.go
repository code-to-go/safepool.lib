package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"time"
	"weshare/core"
	"weshare/model"
	"weshare/transport"

	"github.com/godruoyi/go-snowflake"
)

func waitForLock(domain string, e transport.Exchanger) error {
	lockFile := path.Join(domain, core.DomainFilelock)
	var start time.Time
	lockId := uint64(0)
	for {
		_, err := e.Stat(lockFile)
		if os.IsNotExist(err) {
			return nil
		}

		data := bytes.Buffer{}
		err = e.Read(lockFile, nil, &data)
		if err != nil {
			e.Delete(lockFile)
			return nil
		}

		var lock model.LockFile
		err = json.Unmarshal(data.Bytes(), &lock)
		if core.IsErr(err, "invalid lock file. Delete it") {
			e.Delete(lockFile)
			continue
		}

		time.Sleep(time.Second)
		if lock.Id != lockId {
			lockId = lock.Id
			start = time.Now()
			continue
		}

		elapsed := time.Now().Sub(start)
		if elapsed > lock.ExpectedDuration || elapsed > time.Minute*30 {
			e.Delete(lockFile)
			return nil
		}
	}
}

func createLock(domain string, e transport.Exchanger, expectedDuration time.Duration) error {
	waitForLock(domain, e)

	data, err := json.Marshal(model.LockFile{
		ExpectedDuration: expectedDuration,
		Id:               snowflake.ID(),
	})
	if core.IsErr(err, "cannot marshal lock file: %v", err) {
		return err
	}

	lockFile := path.Join(domain, core.DomainFilelock)
	err = e.Write(lockFile, bytes.NewBuffer(data))
	if core.IsErr(err, "cannot create lock file: %v", err) {
		return err
	}
	return nil
}

func removeLock(domain string, e transport.Exchanger) error {
	lockFile := path.Join(domain, core.DomainFilelock)
	return e.Delete(lockFile)
}
