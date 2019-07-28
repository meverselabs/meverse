package types

import (
	"bytes"
	"reflect"

	"github.com/fletaio/fleta/common/amount"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(AddressAmountMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressAmountMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, am *amount.Amount) bool {
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
		item := NewAddressAmountMap()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			am := amount.NewCoinAmount(0, 0)
			if err := dec.Decode(&am); err != nil {
				return err
			}
			item.Put(addr, am)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressAmountMap struct {
	key   common.Address
	value *amount.Amount
}

func (a *pairAddressAmountMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressAmountMap).key[:]) < 0
	}
}

// AddressAmountMap address and Amount map
type AddressAmountMap struct {
	m *llrb.LLRB
}

// NewAddressAmountMap returns a AddressAmountMap
func NewAddressAmountMap() *AddressAmountMap {
	sm := &AddressAmountMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressAmountMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressAmountMap) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressAmountMap{key: addr})
}

// Get returns data of the key
func (sm *AddressAmountMap) Get(addr common.Address) (*amount.Amount, bool) {
	item := sm.m.Get(&pairAddressAmountMap{key: addr})
	if item == nil {
		return nil, false
	}
	return item.(*pairAddressAmountMap).value, true
}

// Put adds data of the key
func (sm *AddressAmountMap) Put(addr common.Address, am *amount.Amount) {
	sm.m.ReplaceOrInsert(&pairAddressAmountMap{key: addr, value: am})
}

// Delete removes data of the key
func (sm *AddressAmountMap) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressAmountMap{key: addr})
}

// EachAll iterates all elements
func (sm *AddressAmountMap) EachAll(fn func(common.Address, *amount.Amount) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressAmountMap).key, item.(*pairAddressAmountMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *AddressAmountMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := item.(*pairAddressAmountMap).key.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := item.(*pairAddressAmountMap).value.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
