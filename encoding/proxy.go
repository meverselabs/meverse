package encoding

import (
	"io"
	"reflect"

	"github.com/fletaio/fleta/common/hash"
	"github.com/vmihailenco/msgpack"
)

// Marshal returns the encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal decodes the encoded data and stores the result to v
func Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// Hash returns the hash value of v
func Hash(v interface{}) hash.Hash256 {
	data, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	return hash.Hash(data)
}

// Register registers encoder and decoder functions for a value.
func Register(value interface{}, enc encoderFunc, dec decoderFunc) {
	msgpack.Register(value, func(e *msgpack.Encoder, rv reflect.Value) error {
		return enc(&Encoder{Encoder: e}, rv)
	}, func(d *msgpack.Decoder, rv reflect.Value) error {
		return dec(&Decoder{Decoder: d}, rv)
	})
}

// NewEncoder returns a Encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		Encoder: msgpack.NewEncoder(w),
	}
}

// NewDecoder returns a Encoder
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		Decoder: msgpack.NewDecoder(r),
	}
}

// Encoder provides encoding functions
type Encoder struct {
	*msgpack.Encoder
}

// Decoder provides decoding functions
type Decoder struct {
	*msgpack.Decoder
}

type encoderFunc func(*Encoder, reflect.Value) error
type decoderFunc func(*Decoder, reflect.Value) error
