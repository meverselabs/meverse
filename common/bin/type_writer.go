package bin

import (
	"bytes"
	"io"
	"math/big"
	"reflect"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

type TypeWriter struct {
	sum int64
}

func NewTypeWriter() *TypeWriter {
	return &TypeWriter{
		sum: 0,
	}
}

func TypeWriteAll(vs ...interface{}) []byte {
	rw := NewTypeWriter()
	w := bytes.NewBuffer([]byte{})
	rw.WriteAll(w, vs...)
	return w.Bytes()
}

func (tw *TypeWriter) WriteAll(w io.Writer, vs ...interface{}) (int64, error) {
	for _, v := range vs {
		tw.writeThing(w, v)
	}
	return tw.sum, nil
}

func (tw *TypeWriter) writeThing(w io.Writer, v interface{}) (int64, error) {
	var err error
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int:
		_, err = tw.uint32(w, uint32(v.(int)))
	case reflect.Int16:
		_, err = tw.uint16(w, uint16(v.(int16)))
	case reflect.Int32:
		_, err = tw.uint32(w, uint32(v.(int32)))
	case reflect.Int64:
		_, err = tw.uint64(w, uint64(v.(int64)))
	case reflect.Uint:
		_, err = tw.uint32(w, uint32(v.(uint)))
	case reflect.Uint8:
		_, err = tw.uint8(w, v.(uint8))
	case reflect.Uint16:
		_, err = tw.uint16(w, v.(uint16))
	case reflect.Uint32:
		_, err = tw.uint32(w, v.(uint32))
	case reflect.Uint64:
		_, err = tw.uint64(w, v.(uint64))
	case reflect.String:
		_, err = tw.string(w, v.(string))
	case reflect.Bool:
		_, err = tw.bool(w, v.(bool))
	case reflect.Slice:
		switch rv.Type() {
		case reflect.TypeOf([]byte{}):
			_, err = tw.bytes(w, v.([]byte))
		case reflect.TypeOf([]common.Address{}):
			_, err = tw.addrs(w, v.([]common.Address))
		case reflect.TypeOf([]*amount.Amount{}):
			_, err = tw.amounts(w, v.([]*amount.Amount))
		default:
			_, err = tw.slice(w, v)
		}
	default:
		switch rv.Type() {
		case reflect.TypeOf(hash.Hash256{}):
			_, err = tw.hash256(w, v.(hash.Hash256))
		case reflect.TypeOf(common.Signature{}):
			_, err = tw.signature(w, v.(common.Signature))
		case reflect.TypeOf(common.Address{}):
			_, err = tw.address(w, v.(common.Address))
		case reflect.TypeOf(common.PublicKey{}):
			_, err = tw.publicKey(w, v.(common.PublicKey))
		case reflect.TypeOf(&amount.Amount{}):
			_, err = tw.amount(w, v.(*amount.Amount))
		case reflect.TypeOf(&big.Int{}):
			_, err = tw.bigInt(w, v.(*big.Int))
		}
	}
	if err != nil {
		return tw.sum, err
	}
	return tw.sum, nil
}

func (tw *TypeWriter) writeType(w io.Writer, v byte) (int64, error) {
	if n, err := w.Write([]byte{v}); err != nil {
		return tw.sum, errors.WithStack(err)
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) uint8(w io.Writer, v uint8) (int64, error) {
	if _, err := tw.writeType(w, tagUint8); err != nil {
		return tw.sum, err
	}
	if n, err := WriteUint8(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) uint16(w io.Writer, v uint16) (int64, error) {
	if _, err := tw.writeType(w, tagUint16); err != nil {
		return tw.sum, err
	}
	if n, err := WriteUint16(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) uint32(w io.Writer, v uint32) (int64, error) {
	if _, err := tw.writeType(w, tagUint32); err != nil {
		return tw.sum, err
	}
	if n, err := WriteUint32(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) uint64(w io.Writer, v uint64) (int64, error) {
	if _, err := tw.writeType(w, tagUint64); err != nil {
		return tw.sum, err
	}
	if n, err := WriteUint64(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) bytes(w io.Writer, v []byte) (int64, error) {
	if _, err := tw.writeType(w, tagBytes); err != nil {
		return tw.sum, err
	}
	if n, err := WriteBytes(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) string(w io.Writer, v string) (int64, error) {
	if _, err := tw.writeType(w, tagString); err != nil {
		return tw.sum, err
	}
	if n, err := WriteString(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) bool(w io.Writer, v bool) (int64, error) {
	if _, err := tw.writeType(w, tagBool); err != nil {
		return tw.sum, err
	}
	if n, err := WriteBool(w, v); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) hash256(w io.Writer, v hash.Hash256) (int64, error) {
	if _, err := tw.writeType(w, tagHash256); err != nil {
		return tw.sum, err
	}
	if n, err := WriteBytes(w, v.Bytes()); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) signature(w io.Writer, v common.Signature) (int64, error) {
	if _, err := tw.writeType(w, tagSignature); err != nil {
		return tw.sum, err
	}
	if n, err := WriteBytes(w, v[:]); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) address(w io.Writer, v common.Address) (int64, error) {
	if _, err := tw.writeType(w, tagAddress); err != nil {
		return tw.sum, err
	}
	if n, err := w.Write(v[:]); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) publicKey(w io.Writer, v common.PublicKey) (int64, error) {
	if _, err := tw.writeType(w, tagPublicKey); err != nil {
		return tw.sum, err
	}
	if n, err := w.Write(v[:]); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) amount(w io.Writer, v *amount.Amount) (int64, error) {
	if _, err := tw.writeType(w, tagAmount); err != nil {
		return tw.sum, err
	}
	var bs []byte
	if v != nil {
		bs = v.Bytes()
	}
	if n, err := WriteBytes(w, bs); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) bigInt(w io.Writer, v *big.Int) (int64, error) {
	if _, err := tw.writeType(w, tagBigInt); err != nil {
		return tw.sum, err
	}
	if n, err := WriteBytes(w, v.Bytes()); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
		return tw.sum, nil
	}
}

func (tw *TypeWriter) slice(w io.Writer, v interface{}) (int64, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return tw.sum, errors.New("is not slice")
	}
	if _, err := tw.writeType(w, tagSlice); err != nil {
		return tw.sum, err
	}

	if rv.Len() > 255 {
		return tw.sum, errors.New("slice is too big")
	}

	if n, err := WriteUint8(w, uint8(rv.Len())); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
	}

	for i := 0; i < rv.Len(); i++ {
		tw.writeThing(w, rv.Index(i).Interface())
	}
	return tw.sum, nil
}

func (tw *TypeWriter) addrs(w io.Writer, vs []common.Address) (int64, error) {
	if _, err := tw.writeType(w, tagAddressArr); err != nil {
		return tw.sum, err
	}

	if n, err := WriteUint8(w, uint8(len(vs))); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
	}
	for _, v := range vs {
		if n, err := w.Write(v[:]); err != nil {
			return tw.sum, errors.WithStack(err)
		} else {
			tw.sum += int64(n)
		}
	}
	return tw.sum, nil
}

func (tw *TypeWriter) amounts(w io.Writer, vs []*amount.Amount) (int64, error) {

	if _, err := tw.writeType(w, tagAmountArr); err != nil {
		return tw.sum, err
	}

	if n, err := WriteUint8(w, uint8(len(vs))); err != nil {
		return tw.sum, err
	} else {
		tw.sum += int64(n)
	}
	for _, v := range vs {
		var bs []byte
		if v != nil {
			bs = v.Bytes()
		}
		if n, err := WriteBytes(w, bs); err != nil {
			return tw.sum, err
		} else {
			tw.sum += int64(n)
		}
	}
	return tw.sum, nil
}

func (tw *TypeWriter) Sum() int64 {
	return tw.sum
}
