package sql

import (
	"database/sql"
	"weshare/core"
	"weshare/model"

	"github.com/sirupsen/logrus"
)

func getFiles(domain string, rows *sql.Rows) []model.File {
	var files []model.File
	for rows.Next() {
		file := model.File{Domain: domain}

		var author string
		var modTime int64
		err := rows.Scan(&file.Name, &file.Hash, &author, &modTime, &file.State)
		if !core.IsErr(err, "cannot read file row from db: %v") {
			file.ModTime = timeDec(modTime)
			file.Author = base64dec(author)
			files = append(files, file)
		}
	}
	return files
}

func GetFiles(domain string) ([]model.File, error) {
	// name, hash, modTime, state
	rows, err := query("GET_FILES", names{"domain": domain})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}

	return getFiles(domain, rows), nil
}

func GetFilesWithUpdates(domain string) ([]model.File, error) {
	// name, hash, modTime, state
	rows, err := query("GET_FILES_WITH_UPDATES", names{"domain": domain})
	if core.IsErr(err, "cannot read file rows from db: %v") {
		return nil, err
	}

	return getFiles(domain, rows), nil
}

func GetFile(domain string, name string, author []byte) (model.File, error) {
	row := queryRow("GET_FILE", names{"domain": domain, "name": name, "author": base64enc(author)})
	var modTime int64
	var hash string
	file := model.File{Domain: domain, Name: name, Author: author}
	err := row.Scan(&file.FirstId, &file.LastId, &file.Alt, &hash, &modTime, &file.State)
	if err != nil {
		return file, err
	}
	file.ModTime = timeDec(modTime)
	file.Hash = base64dec(hash)
	return file, nil
}

func GetFilesByHash(hash []byte) ([]model.File, error) {
	//SELECT domain, name, author, modTime, state FROM files WHERE hash=:hash
	rows, err := query("GET_FILE_BY_HASH", names{"hash": base64enc(hash)})
	if err != nil {
		return nil, err
	}
	var files []model.File
	for rows.Next() {
		var file model.File
		var author string
		var modTime int64
		err = rows.Scan(&file.Domain, &file.Name, &author, &file.Alt, &modTime, &file.State)
		if err != nil {
			continue
		}
		file.Author = base64dec(author)
		file.ModTime = timeDec(modTime)
		files = append(files, file)
	}

	return files, nil
}

func SetFile(file model.File) error {
	_, err := exec("SET_FILE", names{"domain": file.Domain, "name": file.Name,
		"firstId": file.FirstId, "lastId": file.LastId,
		"author": base64enc(file.Author), "alt": file.Alt, "hash": base64enc(file.Hash),
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
