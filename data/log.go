package data

import (
	"baobab/access"
	"baobab/def"
	"baobab/stores"
	"bytes"
	"fmt"
	"path"
	"time"
	"unsafe"
)

type LogEntry struct {
	FileId access.Hash256
}

type Log struct {
	Version uint16
	Time    time.Time
	UserId  access.PublicKey
	Entries []LogEntry
}

func (l *Log) Marshal(identity access.Identity) ([]byte, error) {
	buf := &bytes.Buffer{}

	writeInt(buf, l.Version)
	writeInt(buf, l.Time.Unix())
	buf.Write(l.UserId)
	writeInt(buf, int32(len(l.Entries)))
	for _, entry := range l.Entries {
		buf.Write(entry.FileId[:])
	}

	data := buf.Bytes()
	sig, err := access.Sign(identity.Private, data)
	if err != nil {
		return nil, err
	} else {
		return append(data, sig...), nil
	}
}

func (l *Log) Unmarshal(data []byte) (sig []byte, err error) {
	buf := bytes.NewBuffer(data)

	l.Version = readInt[uint16](buf)
	l.Time = time.Unix(readInt[int64](buf), 0)
	buf.Read(l.UserId)
	length := int(readInt[int32](buf))
	for i := 0; i < length; i++ {
		fileId := access.Hash256{}
		buf.Read(fileId[:])
	}

	sig = make([]byte, access.SignatureSize)
	buf.Read(sig)
	off := len(data) - access.SignatureSize
	if access.Verify(l.UserId, data[0:off], sig) {
		return sig, nil
	} else {
		return nil, def.ErrInvalidSignature
	}

}

func (l *Log) WriteToStore(identity access.Identity, group string, s stores.Storer) error {
	data, err := l.Marshal(identity)
	if err != nil {
		return err
	}
	name := path.Join(group, fmt.Sprintf("log.%d", l.Time.Unix()))
	return s.Write(name, bytes.NewBuffer(data))
}

func (l *Log) ReadFromStore(name string, s stores.Storer) (sig []byte, err error) {
	buf := &bytes.Buffer{}
	err = s.Read(name, buf)
	if err != nil {
		return nil, err
	}

	return l.Unmarshal(buf.Bytes())
}

type intType interface {
	int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

func writeInt[T intType](buf *bytes.Buffer, val T) {
	size := int(unsafe.Sizeof(val))
	for i := 0; i < size; i++ {
		b := byte(val)
		buf.WriteByte(b)
		val = val >> T(8)
	}
}

func readInt[T intType](buf *bytes.Buffer) T {
	var val T
	size := int(unsafe.Sizeof(val))
	for i := 0; i < size; i++ {
		b, _ := buf.ReadByte()
		val += T(b) << T(i*8)
	}
	return val
}
