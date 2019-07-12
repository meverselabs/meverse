package types

import (
	"bytes"
	"strconv"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/encoding"
)

// ContextData is a state data of the context
type ContextData struct {
	loader                internalLoader
	Parent                *ContextData
	SeqMap                *AddressUint64Map
	AccountMap            *AddressAccountMap
	DeletedAccountMap     *AddressAccountMap
	AccountNameMap        *StringAddressMap
	DeletedAccountNameMap *StringBoolMap
	AccountDataMap        *StringBytesMap
	DeletedAccountDataMap *StringBoolMap
	ProcessDataMap        *StringBytesMap
	DeletedProcessDataMap *StringBoolMap
	UTXOMap               *Uint64UTXOMap
	CreatedUTXOMap        *Uint64TxOutMap
	DeletedUTXOMap        *Uint64UTXOMap
	Events                []Event
	EventN                uint16
	isTop                 bool
}

// NewContextData returns a ContextData
func NewContextData(loader internalLoader, Parent *ContextData) *ContextData {
	var EventN uint16
	if Parent != nil {
		EventN = Parent.EventN
	}
	ctd := &ContextData{
		loader:                loader,
		Parent:                Parent,
		SeqMap:                NewAddressUint64Map(),
		AccountMap:            NewAddressAccountMap(),
		DeletedAccountMap:     NewAddressAccountMap(),
		AccountNameMap:        NewStringAddressMap(),
		DeletedAccountNameMap: NewStringBoolMap(),
		AccountDataMap:        NewStringBytesMap(),
		DeletedAccountDataMap: NewStringBoolMap(),
		ProcessDataMap:        NewStringBytesMap(),
		DeletedProcessDataMap: NewStringBoolMap(),
		UTXOMap:               NewUint64UTXOMap(),
		CreatedUTXOMap:        NewUint64TxOutMap(),
		DeletedUTXOMap:        NewUint64UTXOMap(),
		Events:                []Event{},
		EventN:                EventN,
		isTop:                 true,
	}
	return ctd
}

// Seq returns the sequence of the account
func (ctd *ContextData) Seq(addr common.Address) uint64 {
	if ctd.DeletedAccountMap.Has(addr) {
		return 0
	}
	if seq, has := ctd.SeqMap.Get(addr); has {
		return seq
	} else if ctd.Parent != nil {
		seq := ctd.Parent.Seq(addr)
		if seq > 0 && ctd.isTop {
			ctd.SeqMap.Put(addr, seq)
		}
		return seq
	} else {
		seq := ctd.loader.Seq(addr)
		if seq > 0 && ctd.isTop {
			ctd.SeqMap.Put(addr, seq)
		}
		return seq
	}
}

// AddSeq update the sequence of the target account
func (ctd *ContextData) AddSeq(addr common.Address) {
	if ctd.DeletedAccountMap.Has(addr) {
		return
	}
	ctd.SeqMap.Put(addr, ctd.Seq(addr)+1)
}

// Account returns the account instance of the address
func (ctd *ContextData) Account(addr common.Address) (Account, error) {
	if ctd.DeletedAccountMap.Has(addr) {
		return nil, ErrNotExistAccount
	}
	if acc, has := ctd.AccountMap.Get(addr); has {
		return acc.(Account), nil
	} else if ctd.Parent != nil {
		if acc, err := ctd.Parent.Account(addr); err != nil {
			return nil, err
		} else {
			if ctd.isTop {
				nacc := acc.Clone()
				ctd.AccountMap.Put(addr, nacc)
				return nacc, nil
			} else {
				return acc, nil
			}
		}
	} else {
		if acc, err := ctd.loader.Account(addr); err != nil {
			return nil, err
		} else {
			if ctd.isTop {
				nacc := acc.Clone()
				ctd.AccountMap.Put(addr, nacc)
				return nacc, nil
			} else {
				return acc, nil
			}
		}
	}
}

// AddressByName returns the account address of the name
func (ctd *ContextData) AddressByName(Name string) (common.Address, error) {
	if ctd.DeletedAccountNameMap.Has(Name) {
		return common.Address{}, ErrNotExistAccount
	}
	if addr, has := ctd.AccountNameMap.Get(Name); has {
		return addr, nil
	} else if ctd.Parent != nil {
		if addr, err := ctd.Parent.AddressByName(Name); err != nil {
			return common.Address{}, err
		} else {
			if ctd.isTop {
				naddr := addr.Clone()
				ctd.AccountNameMap.Put(Name, naddr)
				return naddr, nil
			} else {
				return addr, nil
			}
		}
	} else {
		if addr, err := ctd.loader.AddressByName(Name); err != nil {
			return common.Address{}, err
		} else {
			if ctd.isTop {
				naddr := addr.Clone()
				ctd.AccountNameMap.Put(Name, naddr)
				return naddr, nil
			} else {
				return addr, nil
			}
		}
	}
}

// HasAccount checks that the account of the address is exist or not
func (ctd *ContextData) HasAccount(addr common.Address) (bool, error) {
	if ctd.DeletedAccountMap.Has(addr) {
		return false, nil
	}
	if ctd.AccountMap.Has(addr) {
		return true, nil
	} else if ctd.Parent != nil {
		return ctd.Parent.HasAccount(addr)
	} else {
		return ctd.loader.HasAccount(addr)
	}
}

// HasAccountName checks that the account of the address is exist or not
func (ctd *ContextData) HasAccountName(Name string) (bool, error) {
	if ctd.DeletedAccountNameMap.Has(Name) {
		return false, nil
	}
	if ctd.AccountNameMap.Has(Name) {
		return true, nil
	} else if ctd.Parent != nil {
		return ctd.Parent.HasAccountName(Name)
	} else {
		return ctd.loader.HasAccountName(Name)
	}
}

// CreateAccount inserts the account
func (ctd *ContextData) CreateAccount(acc Account) error {
	if acc.Address() == common.NewAddress(0, 0, 0) {
		return ErrNotAllowedZeroAddressAccount
	}
	if len(acc.Name()) < 4 {
		return ErrInvalidAccountName
	}
	if _, err := ctd.Account(acc.Address()); err != nil {
		if err != ErrNotExistAccount {
			return err
		}
	} else {
		return ErrExistAccount
	}
	if _, err := ctd.AddressByName(acc.Name()); err != nil {
		if err != ErrNotExistAccount {
			return err
		}
	} else {
		return ErrExistAccountName
	}
	ctd.AccountMap.Put(acc.Address(), acc)
	ctd.AccountNameMap.Put(acc.Name(), acc.Address())
	return nil
}

// DeleteAccount deletes the account
func (ctd *ContextData) DeleteAccount(acc Account) error {
	if _, err := ctd.Account(acc.Address()); err != nil {
		return err
	}
	ctd.DeletedAccountMap.Put(acc.Address(), acc)
	ctd.DeletedAccountNameMap.Put(acc.Name(), true)
	ctd.AccountMap.Delete(acc.Address())
	ctd.AccountNameMap.Delete(acc.Name())
	return nil
}

// AccountDataKeys returns all data keys of the account in the context
func (ctd *ContextData) AccountDataKeys(addr common.Address, pid uint8, Prefix []byte) ([][]byte, error) {
	keyMap := NewStringBoolMap()
	if ctd.Parent != nil {
		keys, err := ctd.Parent.AccountDataKeys(addr, pid, Prefix)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			keyMap.Put(string(k), true)
		}
	} else {
		keys, err := ctd.loader.AccountDataKeys(addr, pid, Prefix)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			keyMap.Put(string(k), true)
		}
	}
	prefix := string(addr[:]) + string(pid)
	ctd.AccountDataMap.EachPrefix(prefix, func(key string, value []byte) bool {
		keyMap.Put(key[len(prefix):], true)
		return true
	})
	ctd.DeletedAccountDataMap.EachPrefix(prefix, func(key string, value bool) bool {
		keyMap.Delete(key[len(prefix):])
		return true
	})
	keys := [][]byte{}
	keyMap.EachAll(func(key string, value bool) bool {
		keys = append(keys, []byte(key))
		return true
	})
	return keys, nil
}

// AccountData returns the account data
func (ctd *ContextData) AccountData(addr common.Address, pid uint8, name []byte) []byte {
	key := string(addr[:]) + string(pid) + string(name)
	if ctd.DeletedAccountDataMap.Has(key) {
		return nil
	}
	if value, has := ctd.AccountDataMap.Get(key); has {
		return value
	} else if ctd.Parent != nil {
		value := ctd.Parent.AccountData(addr, pid, name)
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				ctd.AccountDataMap.Put(key, nvalue)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	} else {
		value := ctd.loader.AccountData(addr, pid, name)
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				ctd.AccountDataMap.Put(key, nvalue)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	}
}

// SetAccountData inserts the account data
func (ctd *ContextData) SetAccountData(addr common.Address, pid uint8, name []byte, value []byte) {
	key := string(addr[:]) + string(pid) + string(name)
	if len(value) == 0 {
		ctd.AccountDataMap.Delete(key)
		ctd.DeletedAccountDataMap.Put(key, true)
	} else {
		ctd.DeletedAccountDataMap.Delete(key)
		ctd.AccountDataMap.Put(key, value)
	}
}

// HasUTXO checks that the utxo of the id is exist or not
func (ctd *ContextData) HasUTXO(id uint64) (bool, error) {
	if ctd.DeletedUTXOMap.Has(id) {
		return false, nil
	}
	if ctd.UTXOMap.Has(id) {
		return true, nil
	} else if ctd.CreatedUTXOMap.Has(id) {
		return true, nil
	} else if ctd.Parent != nil {
		return ctd.Parent.HasUTXO(id)
	} else {
		return ctd.loader.HasUTXO(id)
	}
}

// UTXO returns the UTXO
func (ctd *ContextData) UTXO(id uint64) (*UTXO, error) {
	if ctd.DeletedUTXOMap.Has(id) {
		return nil, ErrUsedUTXO
	}
	if utxo, has := ctd.UTXOMap.Get(id); has {
		return utxo, nil
	} else if ctd.Parent != nil {
		if utxo, err := ctd.Parent.UTXO(id); err != nil {
			return nil, err
		} else {
			if ctd.isTop {
				nutxo := utxo.Clone()
				ctd.UTXOMap.Put(id, nutxo)
				return nutxo, nil
			} else {
				return utxo, nil
			}
		}
	} else {
		if utxo, err := ctd.loader.UTXO(id); err != nil {
			return nil, err
		} else {
			if ctd.isTop {
				nutxo := utxo.Clone()
				ctd.UTXOMap.Put(id, nutxo)
				return nutxo, nil
			} else {
				return utxo, nil
			}
		}
	}
}

// CreateUTXO inserts the UTXO
func (ctd *ContextData) CreateUTXO(id uint64, vout *TxOut) error {
	if _, err := ctd.UTXO(id); err != nil {
		if err != ErrNotExistUTXO {
			return err
		}
	} else {
		return ErrExistUTXO
	}
	ctd.CreatedUTXOMap.Put(id, vout)
	return nil
}

// DeleteUTXO deletes the UTXO
func (ctd *ContextData) DeleteUTXO(utxo *UTXO) error {
	if _, err := ctd.UTXO(utxo.ID()); err != nil {
		return err
	}
	ctd.DeletedUTXOMap.Put(utxo.ID(), utxo)
	return nil
}

// EmitEvent creates the event to the top snapshot
func (ctd *ContextData) EmitEvent(e Event) error {
	e.SetN(ctd.EventN)
	ctd.EventN++
	ctd.Events = append(ctd.Events, e)
	return nil
}

// ProcessDataKeys returns all data keys of the process in the context
func (ctd *ContextData) ProcessDataKeys(pid uint8, Prefix []byte) ([][]byte, error) {
	keyMap := NewStringBoolMap()
	if ctd.Parent != nil {
		keys, err := ctd.Parent.ProcessDataKeys(pid, Prefix)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			keyMap.Put(string(k), true)
		}
	} else {
		keys, err := ctd.loader.ProcessDataKeys(pid, Prefix)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			keyMap.Put(string(k), true)
		}
	}
	pre := string(pid) + string(Prefix)
	ctd.ProcessDataMap.EachPrefix(pre, func(key string, value []byte) bool {
		keyMap.Put(key[1:], true)
		return true
	})
	ctd.DeletedProcessDataMap.EachPrefix(pre, func(key string, value bool) bool {
		keyMap.Delete(key[1:])
		return true
	})
	keys := [][]byte{}
	keyMap.EachAll(func(key string, value bool) bool {
		keys = append(keys, []byte(key))
		return true
	})
	return keys, nil
}

// ProcessData returns the process data
func (ctd *ContextData) ProcessData(pid uint8, name []byte) []byte {
	key := string(pid) + string(name)
	if ctd.DeletedProcessDataMap.Has(key) {
		return nil
	}
	if value, has := ctd.ProcessDataMap.Get(key); has {
		return value
	} else if ctd.Parent != nil {
		value := ctd.Parent.ProcessData(pid, name)
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				ctd.ProcessDataMap.Put(key, nvalue)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	} else {
		value := ctd.loader.ProcessData(pid, name)
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				ctd.ProcessDataMap.Put(key, nvalue)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	}
}

// SetProcessData inserts the process data
func (ctd *ContextData) SetProcessData(pid uint8, name []byte, value []byte) {
	key := string(pid) + string(name)
	if len(value) == 0 {
		ctd.ProcessDataMap.Delete(key)
		ctd.DeletedProcessDataMap.Put(key, true)
	} else {
		ctd.DeletedProcessDataMap.Delete(key)
		ctd.ProcessDataMap.Put(key, value)
	}
}

// Hash returns the hash value of it
func (ctd *ContextData) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	buffer.WriteString("ChainName")
	buffer.WriteString(ctd.loader.Name())
	buffer.WriteString("ChainVersion")
	buffer.Write(util.Uint16ToBytes(ctd.loader.Version()))
	buffer.WriteString("Height")
	buffer.Write(util.Uint32ToBytes(ctd.loader.TargetHeight()))
	buffer.WriteString("PrevHash")
	lastHash := ctd.loader.LastHash()
	buffer.Write(lastHash[:])
	buffer.WriteString("SeqMap")
	buffer.WriteString(encoding.Hash(ctd.SeqMap).String())
	buffer.WriteString("AccountMap")
	buffer.WriteString(encoding.Hash(ctd.AccountMap).String())
	buffer.WriteString("DeletedAccountMap")
	ctd.DeletedAccountMap.EachAll(func(addr common.Address, acc Account) bool {
		buffer.Write(addr[:])
		return true
	})
	buffer.WriteString("AccountNameMap")
	buffer.WriteString(encoding.Hash(ctd.AccountNameMap).String())
	buffer.WriteString("DeletedAccountNameMap")
	ctd.DeletedAccountNameMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(key)
		return true
	})
	buffer.WriteString("AccountDataMap")
	buffer.WriteString(encoding.Hash(ctd.AccountDataMap).String())
	buffer.WriteString("DeletedAccountDataMap")
	ctd.DeletedAccountDataMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(key)
		return true
	})
	buffer.WriteString("UTXOMap")
	buffer.WriteString(encoding.Hash(ctd.UTXOMap).String())
	buffer.WriteString("CreatedUTXOMap")
	buffer.WriteString(encoding.Hash(ctd.CreatedUTXOMap).String())
	buffer.WriteString("DeletedUTXOMap")
	ctd.DeletedUTXOMap.EachAll(func(key uint64, utxo *UTXO) bool {
		buffer.Write(util.Uint64ToBytes(key))
		return true
	})
	buffer.WriteString("Events")
	if len(ctd.Events) > 0 {
		for _, e := range ctd.Events {
			h := encoding.Hash(e)
			buffer.Write(h[:])
		}
	}
	buffer.WriteString("ProcessDataMap")
	buffer.WriteString(encoding.Hash(ctd.ProcessDataMap).String())
	buffer.WriteString("DeletedProcessDataMap")
	ctd.DeletedProcessDataMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(key)
		return true
	})
	return hash.DoubleHash(buffer.Bytes())
}

// Dump prints the context data
func (ctd *ContextData) Dump() string {
	var buffer bytes.Buffer
	buffer.WriteString("ChainName\n")
	buffer.WriteString(ctd.loader.Name())
	buffer.WriteString("\n")
	buffer.WriteString("ChainVersion\n")
	buffer.WriteString(strconv.FormatUint(uint64(ctd.loader.Version()), 10))
	buffer.WriteString("\n")
	buffer.WriteString("Height\n")
	buffer.WriteString(strconv.FormatUint(uint64(ctd.loader.TargetHeight()), 10))
	buffer.WriteString("\n")
	buffer.WriteString("PrevHash\n")
	lastHash := ctd.loader.LastHash()
	buffer.WriteString(lastHash.String())
	buffer.WriteString("\n")
	buffer.WriteString("SeqMap\n")
	ctd.SeqMap.EachAll(func(addr common.Address, seq uint64) bool {
		buffer.WriteString(addr.String())
		buffer.WriteString(": ")
		buffer.WriteString(strconv.FormatUint(seq, 10))
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("AccountMap\n")
	ctd.AccountMap.EachAll(func(addr common.Address, acc Account) bool {
		buffer.WriteString(addr.String())
		buffer.WriteString(": ")
		buffer.WriteString(encoding.Hash(acc).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("DeletedAccountMap\n")
	ctd.DeletedAccountMap.EachAll(func(addr common.Address, acc Account) bool {
		buffer.WriteString(addr.String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("AccountNameMap\n")
	ctd.AccountNameMap.EachAll(func(key string, addr common.Address) bool {
		buffer.WriteString(key)
		buffer.WriteString(": ")
		buffer.WriteString(addr.String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("DeletedAccountNameMap\n")
	ctd.DeletedAccountNameMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(key)
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("AccountDataMap\n")
	ctd.AccountDataMap.EachAll(func(key string, value []byte) bool {
		buffer.WriteString(hash.Hash([]byte(key)).String())
		buffer.WriteString(": ")
		buffer.WriteString(hash.Hash(value).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("DeletedAccountDataMap\n")
	ctd.DeletedAccountDataMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(hash.Hash([]byte(key)).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("UTXOMap\n")
	ctd.UTXOMap.EachAll(func(id uint64, utxo *UTXO) bool {
		buffer.WriteString(strconv.FormatUint(id, 10))
		buffer.WriteString(": ")
		buffer.WriteString(encoding.Hash(utxo).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("CreatedUTXOMap\n")
	ctd.CreatedUTXOMap.EachAll(func(id uint64, vout *TxOut) bool {
		buffer.WriteString(strconv.FormatUint(id, 10))
		buffer.WriteString(": ")
		buffer.WriteString(encoding.Hash(vout).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("DeletedUTXOMap\n")
	ctd.DeletedUTXOMap.EachAll(func(id uint64, utxo *UTXO) bool {
		buffer.WriteString(strconv.FormatUint(id, 10))
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("Events\n")
	{
		for _, e := range ctd.Events {
			buffer.WriteString(encoding.Hash(e).String())
			buffer.WriteString("\n")
		}
	}
	buffer.WriteString("\n")
	buffer.WriteString("ProcessDataMap\n")
	ctd.ProcessDataMap.EachAll(func(key string, value []byte) bool {
		buffer.WriteString(hash.Hash([]byte(key)).String())
		buffer.WriteString(": ")
		buffer.WriteString(hash.Hash(value).String())
		buffer.WriteString("\n")
		return true
	})
	buffer.WriteString("\n")
	buffer.WriteString("DeletedProcessDataMap\n")
	ctd.DeletedProcessDataMap.EachAll(func(key string, value bool) bool {
		buffer.WriteString(hash.Hash([]byte(key)).String())
		buffer.WriteString("\n")
		return true
	})
	return buffer.String()
}
