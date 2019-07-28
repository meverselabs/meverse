package types

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(AddressBoolMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressBoolMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, value bool) bool {
			if err := enc.Encode(addr); err != nil {
				inErr = err
				return false
			}
			if err := enc.EncodeBool(value); err != nil {
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
		item := NewAddressBoolMap()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			value, err := dec.DecodeBool()
			if err != nil {
				return err
			}
			item.Put(addr, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressBoolMap struct {
	key   common.Address
	value bool
}

func (a *pairAddressBoolMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressBoolMap).key[:]) < 0
	}
}

// AddressBoolMap address and bool map
type AddressBoolMap struct {
	m *llrb.LLRB
}

// NewAddressBoolMap returns a AddressBoolMap
func NewAddressBoolMap() *AddressBoolMap {
	sm := &AddressBoolMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressBoolMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressBoolMap) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressBoolMap{key: addr})
}

// Get returns data of the key
func (sm *AddressBoolMap) Get(addr common.Address) (bool, bool) {
	item := sm.m.Get(&pairAddressBoolMap{key: addr})
	if item == nil {
		return false, false
	}
	return item.(*pairAddressBoolMap).value, true
}

// Put adds data of the key
func (sm *AddressBoolMap) Put(addr common.Address, value bool) {
	sm.m.ReplaceOrInsert(&pairAddressBoolMap{key: addr, value: value})
}

// Delete removes data of the key
func (sm *AddressBoolMap) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressBoolMap{key: addr})
}

// EachAll iterates all elements
func (sm *AddressBoolMap) EachAll(fn func(common.Address, bool) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressBoolMap).key, item.(*pairAddressBoolMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *AddressBoolMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := item.(*pairAddressBoolMap).key.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairAddressBoolMap).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
