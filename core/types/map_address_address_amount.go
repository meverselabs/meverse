package types

import (
	"bytes"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(AddressAddressAmountMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressAddressAmountMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, am *AddressAmountMap) bool {
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
		item := NewAddressAddressAmountMap()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			am := NewAddressAmountMap()
			if err := dec.Decode(&am); err != nil {
				return err
			}
			item.Put(addr, am)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressAddressAmountMap struct {
	key   common.Address
	value *AddressAmountMap
}

func (a *pairAddressAddressAmountMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressAddressAmountMap).key[:]) < 0
	}
}

// AddressAddressAmountMap address and AddressAmountMap map
type AddressAddressAmountMap struct {
	m *llrb.LLRB
}

// NewAddressAddressAmountMap returns a AddressAddressAmountMap
func NewAddressAddressAmountMap() *AddressAddressAmountMap {
	sm := &AddressAddressAmountMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressAddressAmountMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressAddressAmountMap) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressAddressAmountMap{key: addr})
}

// Get returns data of the key
func (sm *AddressAddressAmountMap) Get(addr common.Address) (*AddressAmountMap, bool) {
	item := sm.m.Get(&pairAddressAddressAmountMap{key: addr})
	if item == nil {
		return nil, false
	}
	return item.(*pairAddressAddressAmountMap).value, true
}

// Put adds data of the key
func (sm *AddressAddressAmountMap) Put(addr common.Address, am *AddressAmountMap) {
	sm.m.ReplaceOrInsert(&pairAddressAddressAmountMap{key: addr, value: am})
}

// Delete removes data of the key
func (sm *AddressAddressAmountMap) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressAddressAmountMap{key: addr})
}

// EachAll iterates all elements
func (sm *AddressAddressAmountMap) EachAll(fn func(common.Address, *AddressAmountMap) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressAddressAmountMap).key, item.(*pairAddressAddressAmountMap).value)
	})
}
