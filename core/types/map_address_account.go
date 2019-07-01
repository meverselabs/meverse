package types

import (
	"bytes"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	fc := encoding.Factory("account")
	encoding.Register(AddressAccountMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressAccountMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, acc Account) bool {
			if err := enc.Encode(addr); err != nil {
				inErr = err
				return false
			}
			if t, err := fc.TypeOf(acc); err != nil {
				inErr = err
				return false
			} else if err := enc.EncodeUint16(t); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(acc); err != nil {
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
		item := NewAddressAccountMap()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			t, err := dec.DecodeUint16()
			if err != nil {
				return err
			}
			acc, err := fc.Create(t)
			if err != nil {
				return err
			}
			if err := dec.Decode(&acc); err != nil {
				return err
			}
			item.Put(addr, acc.(Account))
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressAccountMap struct {
	key   common.Address
	value Account
}

func (a *pairAddressAccountMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressAccountMap).key[:]) < 0
	}
}

// AddressAccountMap address and account map
type AddressAccountMap struct {
	m *llrb.LLRB
}

// NewAddressAccountMap returns a AddressAccountMap
func NewAddressAccountMap() *AddressAccountMap {
	sm := &AddressAccountMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressAccountMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressAccountMap) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressAccountMap{key: addr})
}

// Get returns data of the key
func (sm *AddressAccountMap) Get(addr common.Address) (Account, bool) {
	item := sm.m.Get(&pairAddressAccountMap{key: addr})
	if item == nil {
		return nil, false
	}
	return item.(*pairAddressAccountMap).value, true
}

// Put adds data of the key
func (sm *AddressAccountMap) Put(addr common.Address, acc Account) {
	sm.m.ReplaceOrInsert(&pairAddressAccountMap{key: addr, value: acc})
}

// Delete removes data of the key
func (sm *AddressAccountMap) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressAccountMap{key: addr})
}

// EachAll iterates all elements
func (sm *AddressAccountMap) EachAll(fn func(common.Address, Account) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressAccountMap).key, item.(*pairAddressAccountMap).value)
	})
}
