package token

import (
	"math/big"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/fletaio/fleta_v2/core/types"
)

func (cont *TokenContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *TokenContract
}

func (f *front) SubCollectedFee(cc *types.ContractContext, am *amount.Amount) error {
	return f.cont.SubCollectedFee(cc, am)
}

func (f *front) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.Transfer(cc, To, Amount)
}

func (f *front) Burn(cc *types.ContractContext, am *amount.Amount) error {
	return f.cont.Burn(cc, am)
}

func (f *front) Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.Mint(cc, To, Amount)
}

func (f *front) MintBatch(cc *types.ContractContext, Tos []common.Address, Amounts []*amount.Amount) error {
	return f.cont.MintBatch(cc, Tos, Amounts)
}

func (f *front) SetMinter(cc *types.ContractContext, To common.Address, Is bool) error {
	return f.cont.SetMinter(cc, To, Is)
}

func (f *front) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) {
	f.cont.Approve(cc, To, Amount)
}

func (f *front) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	return f.cont.TransferFrom(cc, From, To, Amount)
}
func (f *front) SetGateway(cc *types.ContractContext, Gateway common.Address, Is bool) error {
	return f.cont.SetGateway(cc, Gateway, Is)
}
func (f *front) TokenInRevert(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	return f.cont.TokenInRevert(cc, Platform, ercHash, to, Amount)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) Name(cc types.ContractLoader) string {
	return f.cont.Name(cc)
}

func (f *front) Symbol(cc types.ContractLoader) string {
	return f.cont.Symbol(cc)
}

func (f *front) TotalSupply(cc types.ContractLoader) *amount.Amount {
	return f.cont.TotalSupply(cc)
}

func (f *front) Decimals(cc types.ContractLoader) *big.Int {
	return f.cont.Decimals(cc)
}

func (f *front) BalanceOf(cc types.ContractLoader, from common.Address) *amount.Amount {
	return f.cont.BalanceOf(cc, from)
}

func (f *front) IsMinter(cc types.ContractLoader, addr common.Address) bool {
	return f.cont.IsMinter(cc, addr)
}

func (f *front) CollectedFee(cc types.ContractLoader) *amount.Amount {
	return f.cont.CollectedFee(cc)
}

func (f *front) Allowance(cc types.ContractLoader, _owner common.Address, _spender common.Address) *amount.Amount {
	return f.cont.Allowance(cc, _owner, _spender)
}
