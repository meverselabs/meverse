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
	encoding.Register(AddressUint64Map{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressUint64Map)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, value uint64) bool {
			if err := enc.Encode(addr); err != nil {
				inErr = err
				return false
			}
			if err := enc.EncodeUint64(value); err != nil {
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
		item := NewAddressUint64Map()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			value, err := dec.DecodeUint64()
			if err != nil {
				return err
			}
			item.Put(addr, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressUint64Map struct {
	key   common.Address
	value uint64
}

func (a *pairAddressUint64Map) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressUint64Map).key[:]) < 0
	}
}

// AddressUint64Map address and uint64 map
type AddressUint64Map struct {
	m *llrb.LLRB
}

// NewAddressUint64Map returns a AddressUint64Map
func NewAddressUint64Map() *AddressUint64Map {
	sm := &AddressUint64Map{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressUint64Map) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressUint64Map) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressUint64Map{key: addr})
}

// Get returns data of the key
func (sm *AddressUint64Map) Get(addr common.Address) (uint64, bool) {
	item := sm.m.Get(&pairAddressUint64Map{key: addr})
	if item == nil {
		return 0, false
	}
	return item.(*pairAddressUint64Map).value, true
}

// Put adds data of the key
func (sm *AddressUint64Map) Put(addr common.Address, value uint64) {
	sm.m.ReplaceOrInsert(&pairAddressUint64Map{key: addr, value: value})
}

// Delete removes data of the key
func (sm *AddressUint64Map) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressUint64Map{key: addr})
}

// EachAll iterates all elements
func (sm *AddressUint64Map) EachAll(fn func(common.Address, uint64) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressUint64Map).key, item.(*pairAddressUint64Map).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *AddressUint64Map) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := item.(*pairAddressUint64Map).key.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairAddressUint64Map).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
