package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"weshare/algo"
	"weshare/core"
	"weshare/model"
	"weshare/protocol"
	"weshare/security"
	"weshare/sql"

	"github.com/godruoyi/go-snowflake"
	"github.com/golang/protobuf/proto"
)

const changeFileSubfolder = ".changes"

func GenerateChangeFile(file model.File, sourceBlocks []algo.HashBlock) (changeFile string, updatedFile model.File, err error) {
	filename := filepath.Join(WesharePath, file.Domain, file.Alt, file.Name)
	stat, err := os.Stat(filename)
	if core.IsErr(err, "cannot stat file %s: %v", filename) {
		return "", file, err
	}

	r, err := os.Open(filename)
	if core.IsErr(err, "cannot open file %s: %v", filename) {
		return "", file, err
	}
	defer r.Close()

	destBlocks, err := algo.HashSplit(r, 13, nil)
	if core.IsErr(err, "cannot hash split file %s: %v", filename) {
		return "", file, err
	}

	edits := algo.HashDiff(sourceBlocks, destBlocks)

	newId := snowflake.ID()
	changeFile = filepath.Join(WesharePath, file.Domain, changeFileSubfolder, fmt.Sprintf("C.%d", newId))
	os.MkdirAll(filepath.Dir(changeFile), 0755)
	f, err := os.Create(changeFile)
	if core.IsErr(err, "cannot create change file %s: %v", changeFile) {
		return "", file, err
	}
	defer f.Close()

	keyId, keyValue, err := sql.GetLastEncKey(file.Domain)
	if core.IsErr(err, "cannot get encryption key for domain %s: %v", file.Domain) {
		return "", file, err
	}
	w, err := security.EncryptedWriter(keyId, keyValue, f)
	if core.IsErr(err, "cannot add encryption layer: %v") {
		return "", file, err
	}
	defer w.Close()

	header := &protocol.ChangeFileHeader{
		Version:  1,
		Flag:     0,
		Domain:   file.Domain,
		Name:     file.Name,
		ChainIds: []uint64{newId, file.LastId, file.FirstId},
		Author:   Self.Keys[security.Ed25519].Public,
		Size:     uint64(stat.Size()),
	}
	headerB, err := proto.Marshal(header)
	if core.IsErr(err, "cannot serialize change file header: %v") {
		return "", file, err
	}

	err = security.SignAndWrite(Self, headerB, w, nil)
	if core.IsErr(err, "cannot sign and write the file header: %v") {
		return "", file, err
	}

	var edits2 protocol.Edits
	var withOffset uint32
	for _, edit := range edits {
		edits2.Edits = append(edits2.Edits, &protocol.Edit{
			SliceStart: edit.Slice.Start,
			SliceEnd:   edit.Slice.Start + edit.Slice.Length,
			WithStart:  withOffset,
			WithEnd:    withOffset + edit.With.Length,
		})
	}
	edits2B, err := proto.Marshal(&edits2)
	if core.IsErr(err, "cannot serialize change file edit list: %v") {
		return "", file, err
	}

	err = security.SignAndWrite(Self, edits2B, w, nil)
	if core.IsErr(err, "cannot sign and write the file edit list: %v") {
		return "", file, err
	}

	for _, edit := range edits {
		start := int64(edit.With.Start)
		length := int64(edit.With.Length)

		_, err = r.Seek(start, 0)
		if core.IsErr(err, "cannot seek in file %s: %v", filename) {
			return "", file, err
		}

		n, err := io.CopyN(w, r, length)
		if n != length && core.IsErr(err, "cannot write to change file %s: %v", changeFile) {
			return "", file, err
		}
	}

	updatedFile = file
	updatedFile.LastId = newId
	return
}

func StatChangeFile(changeFile string) (model.File, error) {
	subPath, err := filepath.Rel(WesharePath, changeFile)
	if core.IsErr(err, "change file not in WeShare repo") {
		return model.File{}, core.ErrInvalidChangeFilePath
	}
	splits := strings.Split(subPath, string(os.PathSeparator))
	if len(splits) != 3 || splits[1] != changeFileSubfolder {
		return model.File{}, core.ErrInvalidChangeFilePath
	}
	domain := splits[0]

	f, err := os.Open(changeFile)
	if core.IsErr(err, "cannot open change file '%s':%v", changeFile) {
		return model.File{}, err
	}
	defer f.Close()

	encKeys, err := sql.GetEncKeys(domain)
	if core.IsErr(err, "cannot get encryption keys:%v") {
		return model.File{}, err
	}

	publics, err := sql.GetUsersIdentities(domain, true, false)
	if core.IsErr(err, "cannot get user ids:%v") {
		return model.File{}, err
	}

	r, err := security.EncryptedReader(encKeys, f)
	if core.IsErr(err, "cannot decrypt file '%s':%v", changeFile) {
		return model.File{}, err
	}

	data, _, err := security.ReadAndVerify(publics, r)
	if core.IsErr(err, "invalid header in '%s':%v", changeFile) {
		return model.File{}, err
	}

	var header protocol.ChangeFileHeader
	err = proto.Unmarshal(data, &header)
	if core.IsErr(err, "cannot unmarshal header in '%s':%v", changeFile) {
		return model.File{}, err
	}

	return model.File{
		Domain:  header.Domain,
		Name:    header.Name,
		LastId:  header.ChainIds[0],
		Hash:    header.Hash,
		FirstId: header.ChainIds[len(header.ChainIds)-1],
		Author:  header.Author,
	}, nil
}
