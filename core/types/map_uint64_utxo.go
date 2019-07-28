package types

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(Uint64UTXOMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Uint64UTXOMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key uint64, utxo *UTXO) bool {
			if err := enc.EncodeUint64(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(utxo); err != nil {
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
		item := NewUint64UTXOMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeUint64()
			if err != nil {
				return err
			}
			var utxo *UTXO
			if err := dec.Decode(&utxo); err != nil {
				return err
			}
			item.Put(key, utxo)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairUint64UTXOMap struct {
	key   uint64
	value *UTXO
}

func (a *pairUint64UTXOMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return a.key < b.(*pairUint64UTXOMap).key
	}
}

// Uint64UTXOMap utxoess and *UTXO map
type Uint64UTXOMap struct {
	m *llrb.LLRB
}

// NewUint64UTXOMap returns a Uint64UTXOMap
func NewUint64UTXOMap() *Uint64UTXOMap {
	sm := &Uint64UTXOMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *Uint64UTXOMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *Uint64UTXOMap) Has(key uint64) bool {
	return sm.m.Has(&pairUint64UTXOMap{key: key})
}

// Get returns data of the key
func (sm *Uint64UTXOMap) Get(key uint64) (*UTXO, bool) {
	item := sm.m.Get(&pairUint64UTXOMap{key: key})
	if item == nil {
		return nil, false
	}
	return item.(*pairUint64UTXOMap).value, true
}

// Put adds data of the key
func (sm *Uint64UTXOMap) Put(key uint64, utxo *UTXO) {
	sm.m.ReplaceOrInsert(&pairUint64UTXOMap{key: key, value: utxo})
}

// Delete removes data of the key
func (sm *Uint64UTXOMap) Delete(key uint64) {
	sm.m.Delete(&pairUint64UTXOMap{key: key})
}

// EachAll iterates all elements
func (sm *Uint64UTXOMap) EachAll(fn func(uint64, *UTXO) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairUint64UTXOMap).key, item.(*pairUint64UTXOMap).value)
	})
}

// MarshalJSON is a marshaler function
func (sm *Uint64UTXOMap) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	isFirst := true
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		if isFirst {
			isFirst = false
		} else {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(item.(*pairUint64UTXOMap).key); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		buffer.WriteString(`:`)
		if bs, err := item.(*pairUint64UTXOMap).value.MarshalJSON(); err != nil {
			return false
		} else {
			buffer.Write(bs)
		}
		return true
	})
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
