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
	encoding.Register(AddressUint32Map{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(AddressUint32Map)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(addr common.Address, value uint32) bool {
			if err := enc.Encode(addr); err != nil {
				inErr = err
				return false
			}
			if err := enc.EncodeUint32(value); err != nil {
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
		item := NewAddressUint32Map()
		for i := 0; i < Len; i++ {
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			value, err := dec.DecodeUint32()
			if err != nil {
				return err
			}
			item.Put(addr, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairAddressUint32Map struct {
	key   common.Address
	value uint32
}

func (a *pairAddressUint32Map) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairAddressUint32Map).key[:]) < 0
	}
}

// AddressUint32Map address and uint32 map
type AddressUint32Map struct {
	m *llrb.LLRB
}

// NewAddressUint32Map returns a AddressUint32Map
func NewAddressUint32Map() *AddressUint32Map {
	sm := &AddressUint32Map{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *AddressUint32Map) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *AddressUint32Map) Has(addr common.Address) bool {
	return sm.m.Has(&pairAddressUint32Map{key: addr})
}

// Get returns data of the key
func (sm *AddressUint32Map) Get(addr common.Address) (uint32, bool) {
	item := sm.m.Get(&pairAddressUint32Map{key: addr})
	if item == nil {
		return 0, false
	}
	return item.(*pairAddressUint32Map).value, true
}

// Put adds data of the key
func (sm *AddressUint32Map) Put(addr common.Address, value uint32) {
	sm.m.ReplaceOrInsert(&pairAddressUint32Map{key: addr, value: value})
}

// Delete removes data of the key
func (sm *AddressUint32Map) Delete(addr common.Address) {
	sm.m.Delete(&pairAddressUint32Map{key: addr})
}

// EachAll iterates all elements
func (sm *AddressUint32Map) EachAll(fn func(common.Address, uint32) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairAddressUint32Map).key, item.(*pairAddressUint32Map).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *AddressUint32Map) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := item.(*pairAddressUint32Map).key.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := json.Marshal(item.(*pairAddressUint32Map).value); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
