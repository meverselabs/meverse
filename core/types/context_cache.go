package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

type contextCache struct {
	ctx                *Context
	SeqMap             map[common.Address]uint64
	AccountMap         map[common.Address]Account
	AccountNameMap     map[string]common.Address
	AccountDataMap     map[string][]byte
	AccountDataKeysMap map[common.Address][][]byte
	ProcessDataMap     map[string][]byte
	ProcessDataKeysMap map[uint8][][]byte
	UTXOMap            map[uint64]*UTXO
}

// NewContextCache is used for generating genesis state
func newContextCache(ctx *Context) *contextCache {
	return &contextCache{
		ctx:                ctx,
		SeqMap:             map[common.Address]uint64{},
		AccountMap:         map[common.Address]Account{},
		AccountNameMap:     map[string]common.Address{},
		AccountDataMap:     map[string][]byte{},
		AccountDataKeysMap: map[common.Address][][]byte{},
		ProcessDataMap:     map[string][]byte{},
		ProcessDataKeysMap: map[uint8][][]byte{},
		UTXOMap:            map[uint64]*UTXO{},
	}
}

// Name returns the name of the chain
func (cc *contextCache) Name() string {
	return cc.ctx.Name()
}

// Version returns the version of the chain
func (cc *contextCache) Version() uint16 {
	return cc.ctx.Version()
}

// TargetHeight returns contextCached target height when context generation
func (cc *contextCache) TargetHeight() uint32 {
	return cc.ctx.TargetHeight()
}

// LastHash returns the recorded prev hash when context generation
func (cc *contextCache) LastHash() hash.Hash256 {
	return cc.ctx.LastHash()
}

// LastTimestamp returns the last timestamp of the chain
func (cc *contextCache) LastTimestamp() uint64 {
	return cc.ctx.LastTimestamp()
}

// Seq returns the sequence of the account
func (cc *contextCache) Seq(addr common.Address) uint64 {
	if seq, has := cc.SeqMap[addr]; has {
		return seq
	} else {
		seq := cc.ctx.loader.Seq(addr)
		cc.SeqMap[addr] = seq
		return seq
	}
}

// Account returns the account instance of the address
func (cc *contextCache) Account(addr common.Address) (Account, error) {
	if acc, has := cc.AccountMap[addr]; has {
		return acc, nil
	} else {
		if acc, err := cc.ctx.loader.Account(addr); err != nil {
			return nil, err
		} else {
			cc.AccountMap[addr] = acc
			return acc, nil
		}
	}
}

// AddressByName returns the account address of the name
func (cc *contextCache) AddressByName(Name string) (common.Address, error) {
	if addr, has := cc.AccountNameMap[Name]; has {
		return addr, nil
	} else {
		if addr, err := cc.ctx.loader.AddressByName(Name); err != nil {
			return common.Address{}, err
		} else {
			cc.AccountNameMap[Name] = addr
			return addr, nil
		}
	}
}

// HasAccount checks that the account of the address is exist or not
func (cc *contextCache) HasAccount(addr common.Address) (bool, error) {
	if _, has := cc.AccountMap[addr]; has {
		return true, nil
	} else {
		return cc.ctx.loader.HasAccount(addr)
	}
}

// HasAccountName checks that the account of the name is exist or not
func (cc *contextCache) HasAccountName(Name string) (bool, error) {
	if _, has := cc.AccountNameMap[Name]; has {
		return true, nil
	} else {
		return cc.ctx.loader.HasAccountName(Name)
	}
}

// AccountDataKeys returns all data keys of the account in the context
func (cc *contextCache) AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error) {
	return cc.ctx.loader.AccountDataKeys(addr, Prefix)
}

// AccountData returns the account data
func (cc *contextCache) AccountData(addr common.Address, name []byte) []byte {
	key := string(addr[:]) + string(name)
	if value, has := cc.AccountDataMap[key]; has {
		return value
	} else {
		value := cc.ctx.loader.AccountData(addr, name)
		cc.AccountDataMap[key] = value
		return value
	}
}

// HasUTXO checks that the utxo of the id is exist or not
func (cc *contextCache) HasUTXO(id uint64) (bool, error) {
	if _, has := cc.UTXOMap[id]; has {
		return true, nil
	} else {
		return false, nil
	}
}

// UTXO returns the UTXO
func (cc *contextCache) UTXO(id uint64) (*UTXO, error) {
	if utxo, has := cc.UTXOMap[id]; has {
		return utxo, nil
	} else {
		if utxo, err := cc.ctx.loader.UTXO(id); err != nil {
			return nil, err
		} else {
			cc.UTXOMap[id] = utxo
			return utxo, nil
		}
	}
}

// ProcessDataKeys returns all data keys of the process in the context
func (cc *contextCache) ProcessDataKeys(pid uint8, Prefix []byte) ([][]byte, error) {
	return cc.ctx.loader.ProcessDataKeys(pid, Prefix)
}

// ProcessData returns the process data
func (cc *contextCache) ProcessData(pid uint8, name []byte) []byte {
	key := string(pid) + string(name)
	if value, has := cc.ProcessDataMap[key]; has {
		return value
	} else {
		value := cc.ctx.loader.ProcessData(pid, name)
		cc.ProcessDataMap[key] = value
		return value
	}
}
