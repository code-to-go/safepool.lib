package engine

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"weshare/core"
	"weshare/model"
	"weshare/security"
	"weshare/sql"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

func getDomainPath(domain string) (string, error) {
	ph := filepath.Join(WesharePath, domain)
	stat, err := os.Stat(ph)
	switch {
	case os.IsNotExist(err):
		err = os.MkdirAll(ph, 0755)
		if err != nil {
			logrus.Errorf("cannot create domain folder %s: %v", ph, err)
			return "", err
		}
	case err != nil:
		logrus.Errorf("cannot access domain folder %s: %v", ph, err)
		return "", err
	case !stat.IsDir():
		logrus.Errorf("%s exists but it is not a folder", ph)
		return "", os.ErrInvalid
	}
	return ph, nil
}

func getFilesFromFS(domain, domainPath string) ([]model.File, error) {
	_ = os.MkdirAll(domainPath, 0755)
	var files []model.File
	altFolder := ".alt" + string(filepath.Separator)
	filepath.Walk(domainPath, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			path = path[len(domainPath)+1:]
			if !strings.HasPrefix(path, altFolder) {
				files = append(files, model.File{
					Domain:  domain,
					Name:    path,
					ModTime: info.ModTime(),
				})
			}

		}
		return nil
	})
	return files, nil
}

func hasRenamed(hash []byte) (file2 model.File, ok bool) {
	files, err := sql.GetFilesByHash(hash)
	if err != nil && len(files) == 1 {
		return files[0], true
	} else {
		return model.File{}, false
	}
}

//TODO: to complete
func fileHasChanged(ev fsnotify.Event) {
	name, err := filepath.Rel(WesharePath, ev.Name)
	if core.IsErr(err, "path '%s' is not in weshare folder: %v", ev.Name) {
		return
	}

	sep := string(filepath.Separator)

	split := strings.SplitN(strings.Trim(name, sep), sep, 2)
	if len(split) == 2 {
		domain, name := split[0], split[1]

		switch ev.Op {
		case fsnotify.Create:
			stat, err := os.Stat(ev.Name)
			if !core.IsErr(err, "cannot get info about created file %s: %v", ev.Name) {
				h, err := security.HashFromFile(ev.Name)
				if err != nil {
					f := model.File{
						Domain:  domain,
						Name:    name,
						Author:  Self.Keys[security.Ed25519].Public,
						ModTime: stat.ModTime(),
						Hash:    h[:],
						State:   model.LocalCreated,
					}
					core.IsErr(sql.SetFile(f), "cannot set in db '%v': %v", f)
				}
			}
		case fsnotify.Write:
			f, err := sql.GetFile(domain, name, Self.Keys[security.Ed25519].Public)
			if err != nil {
				f.State |= model.LocalModified
				core.IsErr(sql.SetFile(f), "cannot set in db '%v': %v", f)
			}

		case fsnotify.Remove:
			f, err := sql.GetFile(domain, name, Self.Keys[security.Ed25519].Public)
			if err != nil {
				f.State |= model.LocalDeleted
				core.IsErr(sql.SetFile(f), "cannot set in db '%v': %v", f)
			}
		}
	}

}

func syncLocalToDB(domain string) error {
	domainPath, err := getDomainPath(domain)
	if err != nil {
		return err
	}

	// get files from file system
	files1, err := getFilesFromFS(domain, domainPath)
	if err != nil {
		return err
	}
	sort.Slice(files1, func(i, j int) bool {
		return files1[i].Name < files1[j].Name
	})

	// get files info from db
	files2, err := sql.GetFiles(domain)
	if err != nil {
		return err
	}
	sort.Slice(files2, func(i, j int) bool {
		return files2[i].Name < files2[j].Name
	})

	var i, j int
	for i < len(files1) || j < len(files2) {
		// fi is on file system but not on DB
		fileOnlyOnFs := i < len(files1) && (j == len(files2) || files1[i].Name < files2[j].Name)
		// fj is on DB but not file system
		fileOnlyOnDb := j < len(files2) && (i == len(files1) || files1[i].Name < files2[j].Name)

		if fileOnlyOnFs {
			f := files1[i]
			i++
			h, err := security.HashFromFile(filepath.Join(domainPath, f.Name))
			if err != nil {
				continue
			}
			if f2, ok := hasRenamed(h[:]); ok {
				f2.Name = f.Name
				f2.ModTime = f.ModTime
				f2.State |= model.LocalRenamed
				core.IsErr(sql.SetFile(f2), "cannot set in db '%v': %v", f2)
			} else {
				f.State = model.LocalCreated
				f.Author = security.Primary(Self).Public
				core.IsErr(sql.SetFile(f), "cannot set in db '%v': %v", f)
			}
		} else if fileOnlyOnDb {
			f := files2[j]
			j++
			f.State |= model.LocalDeleted
			core.IsErr(sql.SetFile(f), "cannot set in db '%v': %v", f)
		} else {
			f1, f2 := files1[i], files2[j]
			i++
			j++

			if f1.ModTime.After(f2.ModTime) {
				f2.State |= model.LocalModified
				core.IsErr(sql.SetFile(f2), "cannot set in db '%v': %v", f2)
			}
		}
	}

	return nil
}
