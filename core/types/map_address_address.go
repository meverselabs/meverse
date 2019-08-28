package types

import (
	"bytes"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(AddressAddressMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressAddressMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, am common.Address) bool {
			if err := enc.Encode(addr); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(am); err != nil {
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
		item := NewAddressAddressMap()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			var am common.Address
			if err := dec.Decode(&am); err != nil {
				return err
			}
			item.Put(addr, am)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressAddressMap struct {
	key   common.Address
	value common.Address
}

func (a *pairAddressAddressMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressAddressMap).key[:]) < 0
	}
}

// AddressAddressMap address and Amount map
type AddressAddressMap struct {
	m *llrb.LLRB
}

// NewAddressAddressMap returns a AddressAddressMap
func NewAddressAddressMap() *AddressAddressMap {
	sm := &AddressAddressMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressAddressMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressAddressMap) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressAddressMap{key: addr})
}

// Get returns data of the key
func (sm *AddressAddressMap) Get(addr common.Address) (common.Address, bool) {
	item := sm.m.Get(&pairAddressAddressMap{key: addr})
	if item == nil {
		return common.Address{}, false
	}
	return item.(*pairAddressAddressMap).value, true
}

// Put adds data of the key
func (sm *AddressAddressMap) Put(addr common.Address, am common.Address) {
	sm.m.ReplaceOrInsert(&pairAddressAddressMap{key: addr, value: am})
}

// Delete removes data of the key
func (sm *AddressAddressMap) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressAddressMap{key: addr})
}

// EachAll iterates all elements
func (sm *AddressAddressMap) EachAll(fn func(common.Address, common.Address) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressAddressMap).key, item.(*pairAddressAddressMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *AddressAddressMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := item.(*pairAddressAddressMap).key.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := item.(*pairAddressAddressMap).value.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
