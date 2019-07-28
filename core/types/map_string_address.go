package types

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(StringAddressMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(StringAddressMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key string, addr common.Address) bool {
			if err := enc.EncodeString(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(addr); err != nil {
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
		item := NewStringAddressMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeString()
			if err != nil {
				return err
			}
			var addr common.Address
			if err := dec.Decode(&addr); err != nil {
				return err
			}
			item.Put(key, addr)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairStringAddressMap struct {
	key   string
	value common.Address
}

func (a *pairStringAddressMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return strings.Compare(a.key, b.(*pairStringAddressMap).key) < 0
	}
}

// StringAddressMap address and common.Address map
type StringAddressMap struct {
	m *llrb.LLRB
}

// NewStringAddressMap returns a StringAddressMap
func NewStringAddressMap() *StringAddressMap {
	sm := &StringAddressMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *StringAddressMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *StringAddressMap) Has(key string) bool {
	return sm.m.Has(&pairStringAddressMap{key: key})
}

// Get returns data of the key
func (sm *StringAddressMap) Get(key string) (common.Address, bool) {
	item := sm.m.Get(&pairStringAddressMap{key: key})
	if item == nil {
		return common.Address{}, false
	}
	return item.(*pairStringAddressMap).value, true
}

// Put adds data of the key
func (sm *StringAddressMap) Put(key string, addr common.Address) {
	sm.m.ReplaceOrInsert(&pairStringAddressMap{key: key, value: addr})
}

// Delete removes data of the key
func (sm *StringAddressMap) Delete(key string) {
	sm.m.Delete(&pairStringAddressMap{key: key})
}

// EachAll iterates all elements
func (sm *StringAddressMap) EachAll(fn func(string, common.Address) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairStringAddressMap).key, item.(*pairStringAddressMap).value)
	})
}

// EachPrefix iterates elements that has the given prefix
func (sm *StringAddressMap) EachPrefix(prefix string, fn func(string, common.Address) bool) {
	sm.m.AscendRange(&pairStringAddressMap{key: prefix}, &pairStringAddressMap{key: prefix + string([]byte{255})}, func(item llrb.Item) bool {
		return fn(item.(*pairStringAddressMap).key, item.(*pairStringAddressMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *StringAddressMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(item.(*pairStringAddressMap).key); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := item.(*pairStringAddressMap).value.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
