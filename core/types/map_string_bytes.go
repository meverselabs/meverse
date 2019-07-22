package types

import (
	"reflect"
	"strings"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(StringBytesMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(StringBytesMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key string, value []byte) bool {
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
		item := NewStringBytesMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeString()
			if err != nil {
				return err
			}
			var value []byte
			if err := dec.Decode(&value); err != nil {
				return err
			}
			item.Put(key, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairStringBytesMap struct {
	key   string
	value []byte
}

func (a *pairStringBytesMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return strings.Compare(a.key, b.(*pairStringBytesMap).key) < 0
	}
}

// StringBytesMap valueess and []byte map
type StringBytesMap struct {
	m *llrb.LLRB
}

// NewStringBytesMap returns a StringBytesMap
func NewStringBytesMap() *StringBytesMap {
	sm := &StringBytesMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *StringBytesMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *StringBytesMap) Has(key string) bool {
	return sm.m.Has(&pairStringBytesMap{key: key})
}

// Get returns data of the key
func (sm *StringBytesMap) Get(key string) ([]byte, bool) {
	item := sm.m.Get(&pairStringBytesMap{key: key})
	if item == nil {
		return []byte{}, false
	}
	return item.(*pairStringBytesMap).value, true
}

// Put adds data of the key
func (sm *StringBytesMap) Put(key string, value []byte) {
	nvalue := make([]byte, len(value))
	copy(nvalue, value)
	sm.m.ReplaceOrInsert(&pairStringBytesMap{key: key, value: nvalue})
}

// Delete removes data of the key
func (sm *StringBytesMap) Delete(key string) {
	sm.m.Delete(&pairStringBytesMap{key: key})
}

// EachAll iterates all elements
func (sm *StringBytesMap) EachAll(fn func(string, []byte) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairStringBytesMap).key, item.(*pairStringBytesMap).value)
	})
}

// EachPrefix iterates elements that has the given prefix
func (sm *StringBytesMap) EachPrefix(prefix string, fn func(string, []byte) bool) {
	sm.m.AscendRange(&pairStringBytesMap{key: prefix}, &pairStringBytesMap{key: prefix + string([]byte{255})}, func(item llrb.Item) bool {
		return fn(item.(*pairStringBytesMap).key, item.(*pairStringBytesMap).value)
	})
}
