package sql

import (
	"database/sql"
	"weshare/core"
	"weshare/model"
	"weshare/security"

	"github.com/sirupsen/logrus"
)

func getFiles(rows *sql.Rows) []model.File {
	var files []model.File
	for rows.Next() {
		var file model.File

		var author string
		var hash string
		var modTime int64
		err := rows.Scan(&file.Domain, &file.Name, &file.Id, &author, &modTime, &file.State, &hash)
		if !core.IsErr(err, "cannot read file row from db: %v") {
			file.ModTime = timeDec(modTime)
			file.Author, _ = security.IdentityFromBase64(author)
			file.Hash = base64dec(hash)
			files = append(files, file)
		}
	}
	return files
}

func GetFiles(domain string) ([]model.File, error) {
	// -- GET_FILES
	// SELECT domain, name, id, firstId, author, modTime, state, hash FROM files " +
	// 	"WHERE domain=:domain

	rows, err := query("GET_FILES", names{"domain": domain})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}

	return getFiles(rows), nil
}

func GetFilesWithUpdates(domain string) ([]model.File, error) {
	// -- GET_FILES_WITH_UPDATES
	// SELECT domain,name, id, firstId, author, modTime, state, hash FROM files " +
	// 	"WHERE domain=:domain AND state != 0

	rows, err := query("GET_FILES_WITH_UPDATES", names{"domain": domain})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}

	return getFiles(rows), nil
}

func GetFilesByFirstId(domain string, firstId uint64) ([]model.File, error) {
	// -- GET_FILES_BY_FIRSTID
	// SELECT domain,name, id, firstId, author, modTime, state, hash FROM files " +
	// 	"WHERE domain=:domain AND firstId=:firstId

	rows, err := query("GET_FILES_BY_FIRSTID", names{"domain": domain})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}

	return getFiles(rows), nil
}

func GetFileByName(domain string, name string) (model.File, error) {
	rows, err := query("GET_FILE_BY_NAME", names{"domain": domain, "name": name})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return model.File{}, err
	}

	files := getFiles(rows)
	if len(files) == 0 {
		return model.File{}, sql.ErrNoRows
	}
	return files[0], nil
}

func GetFilesByName(domain string, name string) ([]model.File, error) {
	rows, err := query("GET_FILES_BY_NAME", names{"domain": domain, "name": name})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}
	return getFiles(rows), nil
}

func GetFilesByHash(hash []byte) ([]model.File, error) {
	//SELECT domain, name, author, modTime, state FROM files WHERE hash=:hash
	rows, err := query("GET_FILE_BY_HASH", names{"hash": base64enc(hash)})
	if err != nil {
		return nil, err
	}

	return getFiles(rows), nil
}

func SetFile(file model.File) error {
	author, _ := file.Author.Base64()
	_, err := exec("SET_FILE", names{"domain": file.Domain, "name": file.Name,
		"firstId": file.FirstId, "lastId": file.Id,
		"author": author, "hash": base64enc(file.Hash),
		"modtime": timeEnc(file.ModTime), "state": file.State})
	return err
}

func GetMerkleTree(domain string, name string, author string) (tree []byte, err error) {
	row := queryRow("GET_MERKLE_TREE", names{"domain": domain, "name": name, "author": author})
	err = row.Scan(&tree)
	if err != nil {
		logrus.Errorf("get merkle")
	}
	return tree, err
}

func SetMerkleTree(domain string, name string, author string, tree []byte) error {
	_, err := exec("SET_MERKLE_TREE", names{"name": name, "author": author, "tree": tree})
	return err
}
