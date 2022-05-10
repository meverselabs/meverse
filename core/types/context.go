package types

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// Context is an intermediate in-memory state using the context data stack between blocks
type Context struct {
	loader          internalLoader
	genTargetHeight uint32
	genPrevHash     hash.Hash256
	genTimestamp    uint64
	cache           *contextCache
	stack           []*ContextData
	isLatestHash    bool
	dataHash        hash.Hash256
	isProcessReward bool
}

// NewContext returns a Context
func NewContext(loader internalLoader) *Context {
	ctx := &Context{
		loader:          loader,
		genTargetHeight: loader.TargetHeight(),
		genPrevHash:     loader.PrevHash(),
		genTimestamp:    loader.LastTimestamp(),
	}
	ctx.cache = newContextCache(ctx)
	ctx.stack = []*ContextData{NewContextData(ctx.cache, nil)}
	return ctx
}

// NewEmptyContext returns a EmptyContext
func NewEmptyContext() *Context {
	return NewContext(newEmptyLoader())
}

// ChainID returns the id of the chain
func (ctx *Context) ChainID() *big.Int {
	return ctx.loader.ChainID()
}

// Version returns the version of the chain
func (ctx *Context) Version() uint16 {
	return ctx.loader.Version()
}

// TargetHeight returns the recorded target height when context generation
func (ctx *Context) TargetHeight() uint32 {
	return ctx.genTargetHeight
}

// PrevHash returns the recorded prev hash when context generation
func (ctx *Context) PrevHash() hash.Hash256 {
	return ctx.genPrevHash
}

// LastTimestamp returns the prev timestamp of the chain
func (ctx *Context) LastTimestamp() uint64 {
	return ctx.genTimestamp
}

// IsAdmin returns the account is Admin or not
func (ctx *Context) IsAdmin(addr common.Address) bool {
	return ctx.Top().IsAdmin(addr)
}

// IsProcessReward reeturns the reward is processed or not
func (ctx *Context) IsProcessReward() bool {
	return ctx.isProcessReward
}

// SetAdmin adds the account as a Admin or not
func (ctx *Context) SetAdmin(addr common.Address, is bool) error {
	ctx.isLatestHash = false
	return ctx.Top().SetAdmin(addr, is)
}

// IsGenerator returns the account is generator or not
func (ctx *Context) IsGenerator(addr common.Address) bool {
	return ctx.Top().IsGenerator(addr)
}

// SetGenerator adds the account as a generator or not
func (ctx *Context) SetGenerator(addr common.Address, is bool) error {
	ctx.isLatestHash = false
	return ctx.Top().SetGenerator(addr, is)
}
func (ctx *Context) MainToken() *common.Address {
	return ctx.Top().MainToken()
}
func (ctx *Context) SetMainToken(addr common.Address) {
	ctx.isLatestHash = false
	ctx.Top().SetMainToken(addr)
}

// IsUsedTimeSlot returns timeslot is used or not
func (ctx *Context) IsUsedTimeSlot(slot uint32, key string) bool {
	return ctx.Top().IsUsedTimeSlot(slot, key)
}

// UseTimeSlot consumes timeslot
func (ctx *Context) UseTimeSlot(slot uint32, key string) error {
	ctx.isLatestHash = false
	return ctx.Top().UseTimeSlot(slot, key)
}

// AddrSeq returns the sequence of the target account
func (ctx *Context) AddrSeq(addr common.Address) uint64 {
	return ctx.Top().AddrSeq(addr)
}

// AddAddrSeq update the sequence of the target account
func (ctx *Context) AddAddrSeq(addr common.Address) {
	ctx.isLatestHash = false
	ctx.Top().AddAddrSeq(addr)
}

// BasicFee returns the basic fee
func (ctx *Context) BasicFee() *amount.Amount {
	return ctx.Top().BasicFee()
}

// SetBasicFee update the basic fee
func (ctx *Context) SetBasicFee(fee *amount.Amount) {
	ctx.isLatestHash = false
	ctx.Top().SetBasicFee(fee)
}

// Contract returns the contract data
func (ctx *Context) IsContract(addr common.Address) bool {
	return ctx.Top().IsContract(addr)
}

// Contract returns the contract data
func (ctx *Context) Contract(addr common.Address) (Contract, error) {
	return ctx.Top().Contract(addr)
}

// ProcessReward returns processing the reward
func (ctx *Context) ProcessReward(inctx *Context, b *Block) (map[common.Address][]byte, error) {
	if inctx.StackSize() > 1 {
		return nil, errors.WithStack(ErrDirtyContext)
	}
	inctx.isLatestHash = false
	rewardMap, err := ctx.loader.ProcessReward(inctx, b)
	if err != nil {
		return nil, err
	}
	inctx.isProcessReward = true
	return rewardMap, nil
}

// DeployContract deploy contract to the chain
func (ctx *Context) DeployContract(owner common.Address, ClassID uint64, Args []byte) (Contract, error) {
	ctx.isLatestHash = false
	return ctx.Top().DeployContract(owner, ClassID, Args)
}

// DeployContract deploy contract to the chain with address
func (ctx *Context) DeployContractWithAddress(owner common.Address, ClassID uint64, addr common.Address, Args []byte) (Contract, error) {
	ctx.isLatestHash = false
	return ctx.Top().DeployContractWithAddress(owner, ClassID, addr, Args)
}

// Data returns the data from the top snapshot
func (ctx *Context) Data(cont common.Address, addr common.Address, name []byte) []byte {
	return ctx.Top().Data(cont, addr, name)
}

// SetData inserts the data to the top snapshot
func (ctx *Context) SetData(cont common.Address, addr common.Address, name []byte, value []byte) {
	ctx.isLatestHash = false
	ctx.Top().SetData(cont, addr, name, value)
}

// NextContext returns the next Context of the Context
func (ctx *Context) NextContext(LastHash hash.Hash256, Timestamp uint64) *Context {
	ctx.Top().isTop = false
	nctx := NewContext(ctx)
	nctx.genTargetHeight = ctx.genTargetHeight + 1
	nctx.genPrevHash = LastHash
	nctx.genTimestamp = Timestamp
	return nctx
}

// ContractContext returns a ContractContext
func (ctx *Context) ContractContext(cont Contract, from common.Address) *ContractContext {
	cc := &ContractContext{
		cont: cont.Address(),
		from: from,
		ctx:  ctx,
	}
	return cc
}

// ContractLoader returns a ContractLoader
func (ctx *Context) ContractLoader(cont common.Address) ContractLoader {
	cc := &ContractContext{
		cont: cont,
		ctx:  ctx,
	}
	return cc
}

// Top returns the top snapshot
func (ctx *Context) Top() *ContextData {
	return ctx.stack[len(ctx.stack)-1]
}

// StackSize returns the size of the context data stack
func (ctx *Context) StackSize() int {
	return len(ctx.stack)
}

// Snapshot push a snapshot and returns the snapshot number of it
func (ctx *Context) Snapshot() int {
	ctx.isLatestHash = false
	prevCtd := ctx.Top()
	ctd := NewContextData(ctx.cache, prevCtd)
	ctx.stack[len(ctx.stack)-1].isTop = false
	ctx.stack = append(ctx.stack, ctd)
	ctd.mainToken = prevCtd.mainToken
	for k, v := range prevCtd.AddrSeqMap {
		ctd.AddrSeqMap[k] = v
	}
	ctd.seq = prevCtd.seq
	return len(ctx.stack)
}

// GetSize returns context data size
func (ctx *Context) GetPCSize() uint64 {
	return 0
	return ctx.Top().GetPCSize()
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
		for key, value := range ctd.AddrSeqMap {
			top.AddrSeqMap[key] = value
		}
		for key, value := range ctd.GeneratorMap {
			top.GeneratorMap[key] = value
			delete(top.DeletedGeneratorMap, key)
		}
		for key, value := range ctd.DeletedGeneratorMap {
			delete(top.GeneratorMap, key)
			top.DeletedGeneratorMap[key] = value
		}
		for key, value := range ctd.ContractDefineMap {
			top.ContractDefineMap[key] = value
		}
		for key, value := range ctd.DataMap {
			delete(top.DeletedDataMap, key)
			top.DataMap[key] = value
		}
		for key, value := range ctd.DeletedDataMap {
			delete(top.DataMap, key)
			top.DeletedDataMap[key] = value
		}
		for key, value := range ctd.TimeSlotMap {
			if tp, has := top.TimeSlotMap[key]; has {
				for k, v := range value {
					tp[k] = v
				}
			} else {
				top.TimeSlotMap[key] = value
			}
		}
		top.mainToken = ctd.mainToken
		top.seq = ctd.seq
	}
	ctx.stack[len(ctx.stack)-1].isTop = true
}

// Hash returns the hash value of it
func (ctx *Context) Hash() hash.Hash256 {
	if !ctx.isLatestHash {
		ctx.dataHash = hash.Hashes(ctx.genPrevHash, ctx.Top().Hash())
		ctx.isLatestHash = true
	}
	return ctx.dataHash
}

// Dump prints the top context data of the context
func (ctx *Context) Dump() string {
	return ctx.Top().Dump()
}
