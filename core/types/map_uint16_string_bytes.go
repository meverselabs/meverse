package types

import (
	"reflect"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(Uint8StringBytesMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Uint8StringBytesMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key uint8, StringBytesMap *StringBytesMap) bool {
			if err := enc.EncodeUint8(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(StringBytesMap); err != nil {
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
		item := NewUint8StringBytesMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeUint8()
			if err != nil {
				return err
			}
			var StringBytesMap *StringBytesMap
			if err := dec.Decode(&StringBytesMap); err != nil {
				return err
			}
			item.Put(key, StringBytesMap)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairUint8StringBytesMap struct {
	key   uint8
	value *StringBytesMap
}

func (a *pairUint8StringBytesMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return a.key < b.(*pairUint8StringBytesMap).key
	}
}

// Uint8StringBytesMap StringBytesMapess and *StringBytesMap map
type Uint8StringBytesMap struct {
	m *llrb.LLRB
}

// NewUint8StringBytesMap returns a Uint8StringBytesMap
func NewUint8StringBytesMap() *Uint8StringBytesMap {
	sm := &Uint8StringBytesMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *Uint8StringBytesMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *Uint8StringBytesMap) Has(key uint8) bool {
	return sm.m.Has(&pairUint8StringBytesMap{key: key})
}

// Get returns data of the key
func (sm *Uint8StringBytesMap) Get(key uint8) (*StringBytesMap, bool) {
	item := sm.m.Get(&pairUint8StringBytesMap{key: key})
	if item == nil {
		return nil, false
	}
	return item.(*pairUint8StringBytesMap).value, true
}

// Put adds data of the key
func (sm *Uint8StringBytesMap) Put(key uint8, StringBytesMap *StringBytesMap) {
	sm.m.ReplaceOrInsert(&pairUint8StringBytesMap{key: key, value: StringBytesMap})
}

// Delete removes data of the key
func (sm *Uint8StringBytesMap) Delete(key uint8) {
	sm.m.Delete(&pairUint8StringBytesMap{key: key})
}

// EachAll iterates all elements
func (sm *Uint8StringBytesMap) EachAll(fn func(uint8, *StringBytesMap) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairUint8StringBytesMap).key, item.(*pairUint8StringBytesMap).value)
	})
}
