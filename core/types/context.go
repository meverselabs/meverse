package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

// Context is an intermediate in-memory state using the context data stack between blocks
type Context struct {
	loader          internalLoader
	genTargetHeight uint32
	genLastHash     hash.Hash256
	cache           *contextCache
	stack           []*ContextData
	isLatestHash    bool
	dataHash        hash.Hash256
}

// NewContext returns a Context
func NewContext(loader internalLoader) *Context {
	ctx := &Context{
		loader:          loader,
		genTargetHeight: loader.TargetHeight(),
		genLastHash:     loader.LastHash(),
		stack:           []*ContextData{NewContextData(loader, nil)},
	}
	ctx.cache = newContextCache(ctx)
	return ctx
}

// Name returns the name of the chain
func (ctx *Context) Name() string {
	return ctx.loader.Name()
}

// Version returns the version of the chain
func (ctx *Context) Version() uint16 {
	return ctx.loader.Version()
}

// NextContext returns the next Context of the Context
func (ctx *Context) NextContext(NextHash hash.Hash256) *Context {
	nctx := NewContext(ctx)
	nctx.genTargetHeight = ctx.genTargetHeight + 1
	nctx.genLastHash = NextHash
	return nctx
}

// Hash returns the hash value of it
func (ctx *Context) Hash() hash.Hash256 {
	if !ctx.isLatestHash {
		ctx.dataHash = hash.Hashes(ctx.genLastHash, ctx.Top().Hash())
		ctx.isLatestHash = true
	}
	return ctx.dataHash
}

// TargetHeight returns the recorded target height when context generation
func (ctx *Context) TargetHeight() uint32 {
	return ctx.genTargetHeight
}

// LastHash returns the recorded prev hash when context generation
func (ctx *Context) LastHash() hash.Hash256 {
	return ctx.genLastHash
}

// LastTimestamp returns the last timestamp of the chain
func (ctx *Context) LastTimestamp() uint64 {
	return ctx.loader.LastTimestamp()
}

// Top returns the top snapshot
func (ctx *Context) Top() *ContextData {
	return ctx.stack[len(ctx.stack)-1]
}

// Seq returns the sequence of the target account
func (ctx *Context) Seq(addr common.Address) uint64 {
	return ctx.Top().Seq(addr)
}

// AddSeq update the sequence of the target account
func (ctx *Context) AddSeq(addr common.Address) {
	ctx.isLatestHash = false
	ctx.Top().AddSeq(addr)
}

// Account returns the account instance of the address
func (ctx *Context) Account(addr common.Address) (Account, error) {
	ctx.isLatestHash = false
	return ctx.Top().Account(addr)
}

// AddressByName returns the account address of the name
func (ctx *Context) AddressByName(Name string) (common.Address, error) {
	return ctx.Top().AddressByName(Name)
}

// IsExistAccount checks that the account of the address is exist or not
func (ctx *Context) IsExistAccount(addr common.Address) (bool, error) {
	return ctx.Top().IsExistAccount(addr)
}

// IsExistAccountName checks that the account of the name is exist or not
func (ctx *Context) IsExistAccountName(Name string) (bool, error) {
	return ctx.Top().IsExistAccountName(Name)
}

// CreateAccount inserts the account to the top snapshot
func (ctx *Context) CreateAccount(acc Account) error {
	ctx.isLatestHash = false
	return ctx.Top().CreateAccount(acc)
}

// DeleteAccount deletes the account from the top snapshot
func (ctx *Context) DeleteAccount(acc Account) error {
	ctx.isLatestHash = false
	return ctx.Top().DeleteAccount(acc)
}

// AccountDataKeys returns all data keys of the account in the context
func (ctx *Context) AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error) {
	return ctx.Top().AccountDataKeys(addr, Prefix)
}

// AccountData returns the account data from the top snapshot
func (ctx *Context) AccountData(addr common.Address, name []byte) []byte {
	return ctx.Top().AccountData(addr, name)
}

// SetAccountData inserts the account data to the top snapshot
func (ctx *Context) SetAccountData(addr common.Address, name []byte, value []byte) {
	ctx.isLatestHash = false
	ctx.Top().SetAccountData(addr, name, value)
}

// IsExistUTXO checks that the utxo of the id is exist or not
func (ctx *Context) IsExistUTXO(id uint64) (bool, error) {
	return ctx.Top().IsExistUTXO(id)
}

// UTXO returns the UTXO from the top snapshot
func (ctx *Context) UTXO(id uint64) (*UTXO, error) {
	return ctx.Top().UTXO(id)
}

// CreateUTXO inserts the UTXO to the top snapshot
func (ctx *Context) CreateUTXO(id uint64, vout *TxOut) error {
	ctx.isLatestHash = false
	return ctx.Top().CreateUTXO(id, vout)
}

// DeleteUTXO deletes the UTXO from the top snapshot
func (ctx *Context) DeleteUTXO(id uint64) error {
	ctx.isLatestHash = false
	return ctx.Top().DeleteUTXO(id)
}

// EmitEvent creates the event to the top snapshot
func (ctx *Context) EmitEvent(e Event) error {
	ctx.isLatestHash = false
	return ctx.Top().EmitEvent(e)
}

// ProcessDataKeys returns all data keys of the process in the context
func (ctx *Context) ProcessDataKeys(pid uint8, Prefix []byte) ([][]byte, error) {
	return ctx.Top().ProcessDataKeys(pid, Prefix)
}

// ProcessData returns the process data from the top snapshot
func (ctx *Context) ProcessData(pid uint8, name []byte) []byte {
	return ctx.Top().ProcessData(pid, name)
}

// SetProcessData inserts the process data to the top snapshot
func (ctx *Context) SetProcessData(pid uint8, name []byte, value []byte) {
	ctx.Top().SetProcessData(pid, name, value)
}

// Dump prints the top context data of the context
func (ctx *Context) Dump() string {
	return ctx.Top().Dump()
}

// Snapshot push a snapshot and returns the snapshot number of it
func (ctx *Context) Snapshot() int {
	ctx.isLatestHash = false
	ctd := NewContextData(ctx.cache, ctx.Top())
	ctx.stack[len(ctx.stack)-1].isTop = false
	ctx.stack = append(ctx.stack, ctd)
	return len(ctx.stack)
}

// Revert removes snapshots after the snapshot number
func (ctx *Context) Revert(sn int) {
	ctx.isLatestHash = false
	if len(ctx.stack) >= sn {
		ctx.stack = ctx.stack[:sn-1]
	}
	ctx.stack[len(ctx.stack)-1].isTop = true
}

// Commit apply snapshots to the top after the snapshot number
func (ctx *Context) Commit(sn int) {
	ctx.isLatestHash = false
	for len(ctx.stack) >= sn {
		ctd := ctx.Top()
		ctx.stack = ctx.stack[:len(ctx.stack)-1]
		top := ctx.Top()
		ctd.SeqMap.EachAll(func(addr common.Address, seq uint64) bool {
			top.SeqMap.Put(addr, seq)
			return true
		})
		ctd.AccountMap.EachAll(func(addr common.Address, acc Account) bool {
			top.AccountMap.Put(addr, acc)
			return true
		})
		ctd.DeletedAccountMap.EachAll(func(addr common.Address, value bool) bool {
			top.AccountMap.Delete(addr)
			top.DeletedAccountMap.Put(addr, value)
			return true
		})
		ctd.AccountNameMap.EachAll(func(key string, addr common.Address) bool {
			top.AccountNameMap.Put(key, addr)
			return true
		})
		ctd.DeletedAccountNameMap.EachAll(func(key string, value bool) bool {
			top.AccountNameMap.Delete(key)
			top.DeletedAccountNameMap.Put(key, value)
			return true
		})
		ctd.AccountDataMap.EachAll(func(key string, value []byte) bool {
			top.AccountDataMap.Put(key, value)
			return true
		})
		ctd.DeletedAccountDataMap.EachAll(func(key string, value bool) bool {
			top.AccountDataMap.Delete(key)
			top.DeletedAccountDataMap.Put(key, value)
			return true
		})
		ctd.UTXOMap.EachAll(func(id uint64, utxo *UTXO) bool {
			top.UTXOMap.Put(id, utxo)
			return true
		})
		ctd.CreatedUTXOMap.EachAll(func(id uint64, vout *TxOut) bool {
			top.CreatedUTXOMap.Put(id, vout)
			return true
		})
		ctd.DeletedUTXOMap.EachAll(func(id uint64, value bool) bool {
			top.UTXOMap.Delete(id)
			top.CreatedUTXOMap.Delete(id)
			top.DeletedUTXOMap.Put(id, value)
			return true
		})
		for _, v := range ctd.Events {
			top.Events = append(top.Events, v)
		}
		top.EventN = ctd.EventN
		ctd.ProcessDataMap.EachAll(func(key string, value []byte) bool {
			top.ProcessDataMap.Put(key, value)
			return true
		})
		ctd.DeletedProcessDataMap.EachAll(func(key string, value bool) bool {
			top.ProcessDataMap.Delete(key)
			top.DeletedProcessDataMap.Put(key, value)
			return true
		})
	}
}

// StackSize returns the size of the context data stack
func (ctx *Context) StackSize() int {
	return len(ctx.stack)
}
