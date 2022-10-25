package engine

import (
	"bytes"
	"path"
	"time"
	"weshare/core"
	"weshare/model"
	"weshare/sql"
	"weshare/transport"
)

func readChange(e transport.Exchanger, domain string, name string, changes map[string]model.ChangeFileHeader) (model.ChangeFileHeader, error) {
	if change, ok := changes[name]; ok {
		return change, nil
	}

	dest := bytes.Buffer{}
	err := e.Read(path.Join(domain, name), &transport.Range{From: 0, To: 2048}, &dest)
	if core.IsErr(err, "cannot read change file '%s' from domain '%s'", name, domain) {
		return model.ChangeFileHeader{}, err
	}

	// h, err := StatChangeFileStream(domain, name, &dest)
	// if core.IsErr(err, "cannot read header from change file '%s' from domain '%s'", name, domain) {
	// 	return model.ChangeFileHeader{}, err
	// }

	// if h.Version >= 2.0 {
	// 	logrus.Errorf("Change file has incompatible version %f", h.Version)
	// 	return model.ChangeFileHeader{}, err
	// }

	return model.ChangeFileHeader{}, nil
}

func processChange(e transport.Exchanger, domain string, name string, changes map[string]model.ChangeFileHeader) error {

	h, err := readChange(e, domain, name, changes)
	if core.IsErr(err, "cannot read header from change file '%s' from domain '%s'", name, domain) {
		return err
	}

	files, err := sql.GetFilesByFirstId(domain, h.FirstId)
	if core.IsErr(err, "cannot read files with firstId %d", h.FirstId) {
		return err
	}

	ids := map[uint64]bool{}
	for _, f := range files {
		ids[f.Id] = true
	}

	var mainExist bool
	for _, f := range files {
		if f.Id < h.Ids[0] {
			f.State |= model.ExchangeModified
		}
		if f.State&model.Alternate == 0 {
			mainExist = true
		}
		err = sql.SetFile(f)
		core.IsErr(err, "cannot save file info: %v")
	}

	if !mainExist {
		err = sql.SetFile(model.File{
			Domain:  h.Domain,
			Name:    h.Name,
			Id:      h.Ids[0],
			FirstId: h.FirstId,
			Author:  h.Author,
			ModTime: time.Now(),
			State:   model.ExchangeCreated,
		})
		core.IsErr(err, "cannot save file info: %v")
	}

	return nil
}

func syncExchangesToDB(domain string) error {

	// ConnectionsMutex.Lock()
	// e := Connections[domain]
	// ConnectionsMutex.Unlock()

	// files, err := e.ReadDir(domain, 0)
	// if core.IsErr(err, "cannot read dir from exchange %v: %v", e) {
	// 	return err
	// }

	// sort.Slice(files, func(i, j int) bool {
	// 	return files[i].Name() > files[j].Name()
	// })

	// changes := map[string]model.ChangeFileHeader{}
	// last, _, _, _ := sql.GetConfig(domain, e.String())
	// for _, f := range files {
	// 	name := f.Name()
	// 	if !strings.HasPrefix(name, "C.") || strings.HasSuffix(name, ".sign") || name <= last {
	// 		continue
	// 	}
	// 	if _, ok := changes[name]; ok {
	// 		continue
	// 	}
	// 	processChange(e, name, domain, changes)
	// }

	return nil
}
