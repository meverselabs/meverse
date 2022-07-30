package gateway

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *GatewayContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *GatewayContract
}

func (f *front) AddPlatform(cc *types.ContractContext, Platform string, Amount *amount.Amount) error {
	return f.cont.AddPlatform(cc, Platform, Amount)
}
func (f *front) SetFeeOwner(cc *types.ContractContext, feeOwner common.Address) error {
	return f.cont.SetFeeOwner(cc, feeOwner)
}
func (f *front) Transfer(cc *types.ContractContext, to common.Address, Amount *amount.Amount) error {
	return f.cont.Transfer(cc, to, Amount)
}
func (f *front) TokenIn(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	return f.cont.TokenIn(cc, Platform, ercHash, to, Amount)
}
func (f *front) TokenIndexIn(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	return f.cont.TokenIndexIn(cc, Platform, ercHash, to, Amount)
}
func (f *front) TokenInRevert(cc *types.ContractContext, Platform, ercHash string, txid1, txid2 []byte, to common.Address, Amount *amount.Amount) error {
	return f.cont.TokenInRevert(cc, Platform, ercHash, txid1, txid2, to, Amount)
}
func (f *front) TokenOut(cc *types.ContractContext, Platform string, withdrawAddress common.Address, Amount *amount.Amount) error {
	return f.cont.TokenOut(cc, Platform, withdrawAddress, Amount)
}
func (f *front) TokenLeave(cc *types.ContractContext, CoinTXID string, ERC20TXID string, Platform string) error {
	return f.cont.TokenLeave(cc, CoinTXID, ERC20TXID, Platform)
}
func (f *front) SetSender(cc *types.ContractContext, To common.Address, Is bool) error {
	return f.cont.SetSender(cc, To, Is)
}
func (f *front) TokenAddress(cc *types.ContractContext) common.Address {
	return f.cont.TokenAddress(cc)
}
func (f *front) IsSender(cc types.ContractLoader, addr common.Address) bool {
	return f.cont.IsSender(cc, addr)
}
