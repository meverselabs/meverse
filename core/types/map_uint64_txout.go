package types

import (
	"reflect"

	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(Uint64TxOutMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Uint64TxOutMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(key uint64, vout *TxOut) bool {
			if err := enc.EncodeUint64(key); err != nil {
				inErr = err
				return false
			}
			if err := enc.Encode(vout); err != nil {
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
		item := NewUint64TxOutMap()
		for i := 0; i < Len; i++ {
			key, err := dec.DecodeUint64()
			if err != nil {
				return err
			}
			var vout *TxOut
			if err := dec.Decode(&vout); err != nil {
				return err
			}
			item.Put(key, vout)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairUint64TxOutMap struct {
	key   uint64
	value *TxOut
}

func (a *pairUint64TxOutMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return a.key < b.(*pairUint64TxOutMap).key
	}
}

// Uint64TxOutMap voutess and *TxOut map
type Uint64TxOutMap struct {
	m *llrb.LLRB
}

// NewUint64TxOutMap returns a Uint64TxOutMap
func NewUint64TxOutMap() *Uint64TxOutMap {
	sm := &Uint64TxOutMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *Uint64TxOutMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *Uint64TxOutMap) Has(key uint64) bool {
	return sm.m.Has(&pairUint64TxOutMap{key: key})
}

// Get returns data of the key
func (sm *Uint64TxOutMap) Get(key uint64) (*TxOut, bool) {
	item := sm.m.Get(&pairUint64TxOutMap{key: key})
	if item == nil {
		return nil, false
	}
	return item.(*pairUint64TxOutMap).value, true
}

// Put adds data of the key
func (sm *Uint64TxOutMap) Put(key uint64, vout *TxOut) {
	sm.m.ReplaceOrInsert(&pairUint64TxOutMap{key: key, value: vout})
}

// Delete removes data of the key
func (sm *Uint64TxOutMap) Delete(key uint64) {
	sm.m.Delete(&pairUint64TxOutMap{key: key})
}

// EachAll iterates all elements
func (sm *Uint64TxOutMap) EachAll(fn func(uint64, *TxOut) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairUint64TxOutMap).key, item.(*pairUint64TxOutMap).value)
	})
}
