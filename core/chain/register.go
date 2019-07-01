package chain

import (
	"github.com/fletaio/fleta/common/factory"
	"github.com/fletaio/fleta/core/types"
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
func (reg *Register) RegisterTransaction(t uint8, tx types.Transaction) {
	reg.txFactory.Register(uint16(reg.pid)<<8|uint16(t), tx)
}

// RegisterAccount adds the type of the account of the process to the encoding factory
func (reg *Register) RegisterAccount(t uint8, acc types.Account) {
	reg.accFactory.Register(uint16(reg.pid)<<8|uint16(t), acc)
}

// RegisterEvent adds the type of the event of the process to the encoding factory
func (reg *Register) RegisterEvent(t uint8, acc types.Event) {
	reg.evFactory.Register(uint16(reg.pid)<<8|uint16(t), acc)
}
