package transport

import (
	"bytes"
)

func ReadFile(e Exchanger, name string) ([]byte, error) {
	var b bytes.Buffer
	err := e.Read(name, nil, &b)
	return b.Bytes(), err
}

func WriteFile(e Exchanger, name string, data []byte) error {
	b := bytes.NewBuffer(data)
	return e.Write(name, b)
}
