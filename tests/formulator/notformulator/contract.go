package notformulator

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

type NotFormulatorContract struct {
	addr   common.Address
	master common.Address
}

func (cont *NotFormulatorContract) Address() common.Address {
	return cont.addr
}
func (cont *NotFormulatorContract) Master() common.Address {
	return cont.master
}
func (cont *NotFormulatorContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}
func (cont *NotFormulatorContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	return nil
}
func (cont *NotFormulatorContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

func (cont *NotFormulatorContract) setGenerator(cc *types.ContractContext) error {
	addr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	return cc.SetGenerator(addr, true)
}
