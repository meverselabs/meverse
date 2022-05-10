package types

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
)

type contextCache struct {
	ctx             *Context
	AdminMap        map[string]bool
	SeqMap          map[common.Address]uint64
	GeneratorMap    map[string]bool
	mainToken       *common.Address
	ContractMap     map[common.Address]Contract
	CachedContracts []Contract
	DataMap         map[string][]byte
	TimeSlotMap     map[uint32]map[string]bool
	basicFee        *amount.Amount
}

// NewContextCache is used for generating genesis state
func newContextCache(ctx *Context) *contextCache {
	return &contextCache{
		ctx:          ctx,
		AdminMap:     map[string]bool{},
		SeqMap:       map[common.Address]uint64{},
		GeneratorMap: map[string]bool{},
		ContractMap:  map[common.Address]Contract{},
		DataMap:      map[string][]byte{},
		TimeSlotMap:  map[uint32]map[string]bool{},
	}
}

// IsAdmin returns the account is admin or not
func (cc *contextCache) IsAdmin(addr common.Address) bool {
	key := string(addr[:])
	if is, has := cc.AdminMap[key]; has {
		return is
	} else {
		is := cc.ctx.loader.IsAdmin(addr)
		cc.AdminMap[key] = is
		return is
	}
}

// IsGenerator returns the account is generator or not
func (cc *contextCache) IsGenerator(addr common.Address) bool {
	key := string(addr[:])
	if is, has := cc.GeneratorMap[key]; has {
		return is
	} else {
		is := cc.ctx.loader.IsGenerator(addr)
		cc.GeneratorMap[key] = is
		return is
	}
}

// IsGenerator returns the account is generator or not
func (cc *contextCache) MainToken() *common.Address {
	if cc.mainToken == nil {
		cc.mainToken = cc.ctx.loader.MainToken()
	}
	return cc.mainToken
}

// IsContract returns is the contract
func (cc *contextCache) IsContract(addr common.Address) bool {
	if _, has := cc.ContractMap[addr]; has {
		return true
	} else {
		return cc.ctx.loader.IsContract(addr)
	}
}

// Contract returns the contract of the address
func (cc *contextCache) Contract(addr common.Address) (Contract, error) {
	if cont, has := cc.ContractMap[addr]; has {
		return cont, nil
	} else {
		cont, err := cc.ctx.loader.Contract(addr)
		if err != nil {
			return nil, err
		}
		cc.ContractMap[addr] = cont
		return cont, nil
	}
}

// Data returns the data
func (cc *contextCache) Data(cont common.Address, addr common.Address, name []byte) []byte {
	key := string(cont[:]) + string(addr[:]) + string(name)
	if value, has := cc.DataMap[key]; has {
		return value
	} else {
		value := cc.ctx.loader.Data(cont, addr, name)
		cc.DataMap[key] = value
		return value
	}
}

// IsUsedTimeSlot returns timeslot is used or not
func (cc *contextCache) IsUsedTimeSlot(slot uint32, key string) bool {
	mp, has := cc.TimeSlotMap[slot]
	if !has {
		mp = map[string]bool{}
		cc.TimeSlotMap[slot] = mp
	} else if _, has := mp[key]; has {
		return true
	}
	if has := cc.ctx.loader.IsUsedTimeSlot(slot, key); has {
		mp[key] = true
		return true
	}
	return false
}

// IsUsedTimeSlot returns timeslot is used or not
func (cc *contextCache) AddrSeq(addr common.Address) uint64 {
	seq, has := cc.SeqMap[addr]
	if has {
		return seq
	}
	seq = cc.ctx.loader.AddrSeq(addr)
	cc.SeqMap[addr] = seq
	return seq
}

// BasicFee returns the basic fee
func (cc *contextCache) BasicFee() *amount.Amount {
	if cc.basicFee != nil {
		return cc.basicFee
	}
	cc.basicFee = cc.ctx.loader.BasicFee()
	return cc.basicFee
}
