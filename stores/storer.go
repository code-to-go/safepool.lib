package stores

import (
	"io"
	"io/fs"
	"os"
)

type Source struct {
	Name   string
	Data   []byte
	Reader io.Reader
	Size   int64
}

const SizeAll = -1

type ListOption uint32

const (
	// IncludeHiddenFiles includes hidden files in a list operation
	IncludeHiddenFiles ListOption = 1
)

// Storer is a low level interface to storage services such as S3 or SFTP
type Storer interface {

	// Read reads data from a file into a writer
	Read(name string, dest io.Writer) error

	// Write writes data to a file name. An existing file is overwritten
	Write(name string, source io.Reader) error

	// Concat writes a new file concata
	Concat(name string, source []Source) error

	//ReadDir returns the entries of a folder content
	ReadDir(name string, opts ListOption) ([]fs.FileInfo, error)

	// Stat provides statistics about a file
	Stat(name string) (os.FileInfo, error)

	// Delete deletes a file
	Delete(name string) error

	// Close releases resources
	Close() error

	// String returns a human-readable representation of the storer (e.g. sftp://user@host.cc/path)
	String() string
}
