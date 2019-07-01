package types

import (
	"bytes"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/encoding"
	"github.com/petar/GoLLRB/llrb"
)

func init() {
	encoding.Register(PublicHashBoolMap{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(PublicHashBoolMap)
		if err := enc.EncodeArrayLen(item.Len()); err != nil {
			return err
		}
		var inErr error
		item.EachAll(func(pubhash common.PublicHash, value bool) bool {
			if err := enc.Encode(pubhash); err != nil {
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
		item := NewPublicHashBoolMap()
		for i := 0; i < Len; i++ {
			var pubhash common.PublicHash
			if err := dec.Decode(&pubhash); err != nil {
				return err
			}
			value, err := dec.DecodeBool()
			if err != nil {
				return err
			}
			item.Put(pubhash, value)
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

type pairPublicHashBoolMap struct {
	key   common.PublicHash
	value bool
}

func (a *pairPublicHashBoolMap) Less(b llrb.Item) bool {
	if b == ninf {
		return false
	} else if b == pinf {
		return true
	} else {
		return bytes.Compare(a.key[:], b.(*pairPublicHashBoolMap).key[:]) < 0
	}
}

// PublicHashBoolMap PublicHash and bool map
type PublicHashBoolMap struct {
	m *llrb.LLRB
}

// NewPublicHashBoolMap returns a PublicHashBoolMap
func NewPublicHashBoolMap() *PublicHashBoolMap {
	sm := &PublicHashBoolMap{
		m: llrb.New(),
	}
	return sm
}

// Len returns the length of the map
func (sm *PublicHashBoolMap) Len() int {
	return sm.m.Len()
}

// Has returns data of the key is exist or not
func (sm *PublicHashBoolMap) Has(pubhash common.PublicHash) bool {
	return sm.m.Has(&pairPublicHashBoolMap{key: pubhash})
}

// Get returns data of the key
func (sm *PublicHashBoolMap) Get(pubhash common.PublicHash) (bool, bool) {
	item := sm.m.Get(&pairPublicHashBoolMap{key: pubhash})
	if item == nil {
		return false, false
	}
	return item.(*pairPublicHashBoolMap).value, true
}

// Put adds data of the key
func (sm *PublicHashBoolMap) Put(pubhash common.PublicHash, value bool) {
	sm.m.ReplaceOrInsert(&pairPublicHashBoolMap{key: pubhash, value: value})
}

// Delete removes data of the key
func (sm *PublicHashBoolMap) Delete(pubhash common.PublicHash) {
	sm.m.Delete(&pairPublicHashBoolMap{key: pubhash})
}

// EachAll iterates all elements
func (sm *PublicHashBoolMap) EachAll(fn func(common.PublicHash, bool) bool) {
	sm.m.AscendRange(ninf, pinf, func(item llrb.Item) bool {
		return fn(item.(*pairPublicHashBoolMap).key, item.(*pairPublicHashBoolMap).value)
	})
}
