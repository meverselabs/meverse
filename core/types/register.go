package types

import (
	"github.com/fletaio/fleta/common/factory"
	"github.com/fletaio/fleta/encoding"
)

// Register adds types of the process to the encoding factory
type Register struct {
	pid        uint8
	txFactory  *factory.Factory
	accFactory *factory.Factory
	evFactory  *factory.Factory
}

// NewRegister returns a Register
func NewRegister(pid uint8) *Register {
	reg := &Register{
		pid:        pid,
		txFactory:  encoding.Factory("transaction"),
		accFactory: encoding.Factory("account"),
		evFactory:  encoding.Factory("event"),
	}
	return reg
}

// RegisterTransaction adds the type of the transaction of the process to the encoding factory
func (reg *Register) RegisterTransaction(t uint8, tx Transaction) uint16 {
	v := uint16(reg.pid)<<8 | uint16(t)
	reg.txFactory.Register(v, tx)
	return v
}

// RegisterAccount adds the type of the account of the process to the encoding factory
func (reg *Register) RegisterAccount(t uint8, acc Account) uint16 {
	v := uint16(reg.pid)<<8 | uint16(t)
	reg.accFactory.Register(v, acc)
	return v
}

// RegisterEvent adds the type of the event of the process to the encoding factory
func (reg *Register) RegisterEvent(t uint8, acc Event) uint16 {
	v := uint16(reg.pid)<<8 | uint16(t)
	reg.evFactory.Register(v, acc)
	return v
}
