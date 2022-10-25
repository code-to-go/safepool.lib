package transport

import (
	"bytes"
	"encoding/json"
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

func ReadJSON(e Exchanger, name string, v any) error {
	data, err := ReadFile(e, name)
	if err == nil {
		err = json.Unmarshal(data, v)
	}
	return err
}

func WriteJSON(e Exchanger, name string, v any) error {
	b, err := json.Marshal(v)
	if err == nil {
		err = e.Write(name, bytes.NewBuffer(b))
	}
	return err
}
