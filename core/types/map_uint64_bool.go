package types

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(Uint64BoolMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Uint64BoolMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key uint64, value bool) bool {
			if err := enc.EncodeUint64(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(value); err != nil {
				inErr = err
				return false
			}
			return true
		})
		if inErr != nil {
			return inErr
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		Len, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		item := NewUint64BoolMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeUint64()
			if err != nil {
				return err
			}
			var value bool
			if err := dec.Decode(&value); err != nil {
				return err
			}
			item.Put(key, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairUint64BoolMap struct {
	key   uint64
	value bool
}

func (a *pairUint64BoolMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return a.key < b.(*pairUint64BoolMap).key
	}
}

// Uint64BoolMap valueess and bool map
type Uint64BoolMap struct {
	m *llrb.LLRB
}

// NewUint64BoolMap returns a Uint64BoolMap
func NewUint64BoolMap() *Uint64BoolMap {
	sm := &Uint64BoolMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *Uint64BoolMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *Uint64BoolMap) Has(key uint64) bool {
	return sm.m.Has(&pairUint64BoolMap{key: key})
}

// Get returns data of the key
func (sm *Uint64BoolMap) Get(key uint64) (bool, bool) {
	item := sm.m.Get(&pairUint64BoolMap{key: key})
	if item == nil {
		return false, false
	}
	return item.(*pairUint64BoolMap).value, true
}

// Put adds data of the key
func (sm *Uint64BoolMap) Put(key uint64, value bool) {
	sm.m.ReplaceOrInsert(&pairUint64BoolMap{key: key, value: value})
}

// Delete removes data of the key
func (sm *Uint64BoolMap) Delete(key uint64) {
	sm.m.Delete(&pairUint64BoolMap{key: key})
}

// EachAll iterates all elements
func (sm *Uint64BoolMap) EachAll(fn func(uint64, bool) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairUint64BoolMap).key, item.(*pairUint64BoolMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *Uint64BoolMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(item.(*pairUint64BoolMap).key); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairUint64BoolMap).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
