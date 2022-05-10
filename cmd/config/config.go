package config

import (
	"bytes"
	"io"
	"os"

	toml "github.com/meverselabs/meverse/cmd/config/go-toml"
)

// LoadFile parse the config from the file of the path
func LoadFile(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return LoadReader(file, v)
}

// LoadString parse the config from the string
func LoadString(data string, v interface{}) error {
	return LoadReader(bytes.NewReader([]byte(data)), v)
}

// LoadReader parse the config from the file of the reader
func LoadReader(r io.Reader, v interface{}) error {
	dec := toml.NewDecoder(r)
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}
