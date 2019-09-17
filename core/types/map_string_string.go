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
	encoding.Register(StringStringMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(StringStringMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key string, value string) bool {
			if err := enc.EncodeString(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.EncodeString(value); err != nil {
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
		item := NewStringStringMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeString()
			if err != nil {
				return err
			}
			value, err := dec.DecodeString()
			if err != nil {
				return err
			}
			item.Put(key, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairStringStringMap struct {
	key   string
	value string
}

func (a *pairStringStringMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return strings.Compare(a.key, b.(*pairStringStringMap).key) < 0
	}
}

// StringStringMap string and string map
type StringStringMap struct {
	m *llrb.LLRB
}

// NewStringStringMap returns a StringStringMap
func NewStringStringMap() *StringStringMap {
	sm := &StringStringMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *StringStringMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *StringStringMap) Has(key string) bool {
	return sm.m.Has(&pairStringStringMap{key: key})
}

// Get returns data of the key
func (sm *StringStringMap) Get(key string) (string, bool) {
	item := sm.m.Get(&pairStringStringMap{key: key})
	if item == nil {
		return "", false
	}
	return item.(*pairStringStringMap).value, true
}

// Put adds data of the key
func (sm *StringStringMap) Put(key string, value string) {
	sm.m.ReplaceOrInsert(&pairStringStringMap{key: key, value: value})
}

// Delete removes data of the key
func (sm *StringStringMap) Delete(key string) {
	sm.m.Delete(&pairStringStringMap{key: key})
}

// EachAll iterates all elements
func (sm *StringStringMap) EachAll(fn func(string, string) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairStringStringMap).key, item.(*pairStringStringMap).value)
	})
}

// EachPrefix iterates elements that has the given prefix
func (sm *StringStringMap) EachPrefix(prefix string, fn func(string, string) bool) {
	sm.m.AscendRange(&pairStringStringMap{key: prefix}, &pairStringStringMap{key: prefix + string([]byte{255})}, func(item llrb.Item) bool {
		return fn(item.(*pairStringStringMap).key, item.(*pairStringStringMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *StringStringMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(item.(*pairStringStringMap).key); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairStringStringMap).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
