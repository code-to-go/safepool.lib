package stores

import "baobab/errors"

// NewStorer creates a new store giving a provided configuration
func NewStorer(c Config) (Storer, error) {
	switch {
	case c.SFTP != nil:
		return NewSFTP(*c.SFTP)
	case c.S3 != nil:
		return NewS3(*c.S3)
	}

	return nil, errors.ErrNoDriver
}

// Read reads the full content of a file and returns it as a byte slice
//func Read(s stores.Storer, name string) ([]byte, error) {
//	b := stores.Block{
//		Offset: 0,
//		Size:   math.MaxInt,
//		Data:   nil,
//	}
//	err := s.Read(name, []stores.Block{b})
//	return b.Data, err
//}
