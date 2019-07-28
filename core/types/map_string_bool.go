package types

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(StringBoolMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(StringBoolMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key string, value bool) bool {
			if err := enc.EncodeString(key); err != nil {
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
		item := NewStringBoolMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeString()
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

type pairStringBoolMap struct {
	key   string
	value bool
}

func (a *pairStringBoolMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return strings.Compare(a.key, b.(*pairStringBoolMap).key) < 0
	}
}

// StringBoolMap valueess and bool map
type StringBoolMap struct {
	m *llrb.LLRB
}

// NewStringBoolMap returns a StringBoolMap
func NewStringBoolMap() *StringBoolMap {
	sm := &StringBoolMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *StringBoolMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *StringBoolMap) Has(key string) bool {
	return sm.m.Has(&pairStringBoolMap{key: key})
}

// Get returns data of the key
func (sm *StringBoolMap) Get(key string) (bool, bool) {
	item := sm.m.Get(&pairStringBoolMap{key: key})
	if item == nil {
		return false, false
	}
	return item.(*pairStringBoolMap).value, true
}

// Put adds data of the key
func (sm *StringBoolMap) Put(key string, value bool) {
	sm.m.ReplaceOrInsert(&pairStringBoolMap{key: key, value: value})
}

// Delete removes data of the key
func (sm *StringBoolMap) Delete(key string) {
	sm.m.Delete(&pairStringBoolMap{key: key})
}

// EachAll iterates all elements
func (sm *StringBoolMap) EachAll(fn func(string, bool) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairStringBoolMap).key, item.(*pairStringBoolMap).value)
	})
}

// EachPrefix iterates elements that has the given prefix
func (sm *StringBoolMap) EachPrefix(prefix string, fn func(string, bool) bool) {
	sm.m.AscendRange(&pairStringBoolMap{key: prefix}, &pairStringBoolMap{key: prefix + string([]byte{255})}, func(item llrb.Item) bool {
		return fn(item.(*pairStringBoolMap).key, item.(*pairStringBoolMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *StringBoolMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(item.(*pairStringBoolMap).key); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairStringBoolMap).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
