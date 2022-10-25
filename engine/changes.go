package engine

const changeFileSubfolder = ".changes"

// func GenerateChangeFile(file model.File, sourceBlocks []algo.HashBlock) (changeFile string, updatedFile model.File, err error) {
// 	filename := filepath.Join(WesharePath, file.Domain, file.Name)

// 	r, err := os.Open(filename)
// 	if core.IsErr(err, "cannot open file %s: %v", filename) {
// 		return "", file, err
// 	}
// 	defer r.Close()

// 	destBlocks, err := algo.HashSplit(r, 13, nil)
// 	if core.IsErr(err, "cannot hash split file %s: %v", filename) {
// 		return "", file, err
// 	}

// 	edits := algo.HashDiff(sourceBlocks, destBlocks)

// 	newId := snowflake.ID()
// 	changeFile = filepath.Join(WesharePath, file.Domain, changeFileSubfolder, fmt.Sprintf("C.%d", newId))
// 	os.MkdirAll(filepath.Dir(changeFile), 0755)
// 	f, err := os.Create(changeFile)
// 	if core.IsErr(err, "cannot create change file %s: %v", changeFile) {
// 		return "", file, err
// 	}
// 	defer f.Close()

// 	keyId, keyValue, err := sql.GetLastEncKey(file.Domain)
// 	if core.IsErr(err, "cannot get encryption key for domain %s: %v", file.Domain) {
// 		return "", file, err
// 	}
// 	// w, err := security.EncryptedWriter(keyId, keyValue, f)
// 	// if core.IsErr(err, "cannot add encryption layer: %v") {
// 	// 	return "", file, err
// 	// }
// 	// defer w.Close()

// 	header := &model.ChangeFileHeader{
// 		Version: 1,
// 		Flag:    0,
// 		Domain:  file.Domain,
// 		Name:    file.Name,
// 		Ids:     []uint64{newId, file.Id},
// 		FirstId: file.FirstId,
// 		Author:  Self.Public(),
// 	}
// 	data, err := json.Marshal(header)
// 	if core.IsErr(err, "cannot serialize change file header: %v") {
// 		return "", file, err
// 	}

// 	err = security.SignAndWrite(Self, data, w)
// 	if core.IsErr(err, "cannot sign and write the file header: %v") {
// 		return "", file, err
// 	}

// 	var withOffset uint32
// 	for _, e := range edits {
// 		e.With.Start = withOffset
// 	}
// 	data2, err := json.Marshal(&edits)
// 	if core.IsErr(err, "cannot serialize change file edit list: %v") {
// 		return "", file, err
// 	}

// 	err = security.SignAndWrite(Self, data2, w)
// 	if core.IsErr(err, "cannot sign and write the file edit list: %v") {
// 		return "", file, err
// 	}

// 	for _, edit := range edits {
// 		start := int64(edit.With.Start)
// 		length := int64(edit.With.Length)

// 		_, err = r.Seek(start, 0)
// 		if core.IsErr(err, "cannot seek in file %s: %v", filename) {
// 			return "", file, err
// 		}

// 		n, err := io.CopyN(w, r, length)
// 		if n != length && core.IsErr(err, "cannot write to change file %s: %v", changeFile) {
// 			return "", file, err
// 		}
// 	}

// 	updatedFile = file
// 	updatedFile.Id = newId
// 	return
// }

// func StatChangeFile(changeFile string) (model.ChangeFileHeader, error) {
// 	subPath, err := filepath.Rel(WesharePath, changeFile)
// 	if core.IsErr(err, "change file not in WeShare repo") {
// 		return model.ChangeFileHeader{}, core.ErrInvalidChangeFilePath
// 	}
// 	splits := strings.Split(subPath, string(os.PathSeparator))
// 	if len(splits) != 3 || splits[1] != changeFileSubfolder {
// 		return model.ChangeFileHeader{}, core.ErrInvalidChangeFilePath
// 	}
// 	domain := splits[0]
// 	f, err := os.Open(changeFile)
// 	if core.IsErr(err, "cannot open change file '%s':%v", changeFile) {
// 		return model.ChangeFileHeader{}, err
// 	}
// 	defer f.Close()
// 	return StatChangeFileStream(domain, changeFile, f)
// }

// func StatChangeFileStream(domain string, name string, f io.Reader) (model.ChangeFileHeader, error) {
// 	var header model.ChangeFileHeader
// 	encKeys, err := sql.GetEncKeys(domain)
// 	if core.IsErr(err, "cannot get encryption keys:%v") {
// 		return header, err
// 	}

// 	publics, err := sql.GetUsersIdentities(domain, true)
// 	if core.IsErr(err, "cannot get user ids:%v") {
// 		return header, err
// 	}

// 	r, err := security.EncryptedReader(encKeys, f)
// 	if core.IsErr(err, "cannot decrypt file '%s':%v", name) {
// 		return header, err
// 	}

// 	data, _, err := security.ReadAndVerify(publics, r)
// 	if core.IsErr(err, "invalid header in '%s':%v", name) {
// 		return header, err
// 	}

// 	err = json.Unmarshal(data, &header)
// 	if core.IsErr(err, "cannot unmarshal header in '%s':%v", name) {
// 		return header, err
// 	}

// 	return header, nil
// }
