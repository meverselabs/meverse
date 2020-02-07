package types

import (
	"reflect"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(Uint32StringBoolMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Uint32StringBoolMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key uint32, StringBoolMap *StringBoolMap) bool {
			if err := enc.EncodeUint32(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(StringBoolMap); err != nil {
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
		item := NewUint32StringBoolMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeUint32()
			if err != nil {
				return err
			}
			var StringBoolMap *StringBoolMap
			if err := dec.Decode(&StringBoolMap); err != nil {
				return err
			}
			item.Put(key, StringBoolMap)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairUint32StringBoolMap struct {
	key   uint32
	value *StringBoolMap
}

func (a *pairUint32StringBoolMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return a.key < b.(*pairUint32StringBoolMap).key
	}
}

// Uint32StringBoolMap StringBoolMapess and *StringBoolMap map
type Uint32StringBoolMap struct {
	m *llrb.LLRB
}

// NewUint32StringBoolMap returns a Uint32StringBoolMap
func NewUint32StringBoolMap() *Uint32StringBoolMap {
	sm := &Uint32StringBoolMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *Uint32StringBoolMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *Uint32StringBoolMap) Has(key uint32) bool {
	return sm.m.Has(&pairUint32StringBoolMap{key: key})
}

// Get returns data of the key
func (sm *Uint32StringBoolMap) Get(key uint32) (*StringBoolMap, bool) {
	item := sm.m.Get(&pairUint32StringBoolMap{key: key})
	if item == nil {
		return nil, false
	}
	return item.(*pairUint32StringBoolMap).value, true
}

// Put adds data of the key
func (sm *Uint32StringBoolMap) Put(key uint32, StringBoolMap *StringBoolMap) {
	sm.m.ReplaceOrInsert(&pairUint32StringBoolMap{key: key, value: StringBoolMap})
}

// Delete removes data of the key
func (sm *Uint32StringBoolMap) Delete(key uint32) {
	sm.m.Delete(&pairUint32StringBoolMap{key: key})
}

// EachAll iterates all elements
func (sm *Uint32StringBoolMap) EachAll(fn func(uint32, *StringBoolMap) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairUint32StringBoolMap).key, item.(*pairUint32StringBoolMap).value)
	})
}
