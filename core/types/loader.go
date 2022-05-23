package types

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// Loader defines functions that loads state data from the target chain
type Loader interface {
	ChainID() *big.Int
	Version() uint16
	TargetHeight() uint32
	PrevHash() hash.Hash256
	LastTimestamp() uint64
}

type internalLoader interface {
	Loader
	IsAdmin(addr common.Address) bool
	IsGenerator(addr common.Address) bool
	MainToken() *common.Address
	AddrSeq(addr common.Address) uint64
	IsUsedTimeSlot(slot uint32, key string) bool
	BasicFee() *amount.Amount
	IsContract(addr common.Address) bool
	Contract(addr common.Address) (Contract, error)
	Data(cont common.Address, addr common.Address, name []byte) []byte
	ProcessReward(ctx *Context, b *Block) (map[common.Address][]byte, error)
}

type emptyLoader struct {
}

// newEmptyLoader is used for generating genesis state
func newEmptyLoader() internalLoader {
	return &emptyLoader{}
}

// ChainID returns 0
func (st *emptyLoader) ChainID() *big.Int {
	return big.NewInt(0)
}

// Name returns ""
func (st *emptyLoader) Name() string {
	return ""
}

// Version returns 0
func (st *emptyLoader) Version() uint16 {
	return 0
}

// TargetHeight returns 0
func (st *emptyLoader) TargetHeight() uint32 {
	return 0
}

// PrevHash returns hash.Hash256{}
func (st *emptyLoader) PrevHash() hash.Hash256 {
	return hash.Hash256{}
}

// LastTimestamp returns 0
func (st *emptyLoader) LastTimestamp() uint64 {
	return 0
}

// IsUsedTimeSlot returns false
func (st *emptyLoader) IsUsedTimeSlot(slot uint32, key string) bool {
	return false
}

// IsAdmin returns false
func (st *emptyLoader) IsAdmin(addr common.Address) bool {
	return false
}

// AddrSeq returns 0
func (st *emptyLoader) AddrSeq(addr common.Address) uint64 {
	return 0
}

// BasicFee returns empty fee
func (st *emptyLoader) BasicFee() *amount.Amount {
	return &amount.Amount{}
}

// IsGenerator returns false
func (st *emptyLoader) IsGenerator(addr common.Address) bool {
	return false
}

// MainToken returns nil
func (st *emptyLoader) MainToken() *common.Address {
	return nil
}

// Contract returns nil
func (st *emptyLoader) IsContract(addr common.Address) bool {
	return false
}

// Contract returns nil
func (st *emptyLoader) Contract(addr common.Address) (Contract, error) {
	return nil, errors.WithStack(ErrNotExistContract)
}

// Data returns nil
func (st *emptyLoader) Data(cont common.Address, addr common.Address, name []byte) []byte {
	return nil
}

// ProcessReward returns nil
func (st *emptyLoader) ProcessReward(ctx *Context, b *Block) (map[common.Address][]byte, error) {
	return nil, nil
}
