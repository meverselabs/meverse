package token

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
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

func (f *front) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.Transfer(cc, To, Amount)
	return err == nil, err
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

func (f *front) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.Approve(cc, To, Amount)
	return err == nil, err
}

func (f *front) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.TransferFrom(cc, From, To, Amount)
	return err == nil, err
}
func (f *front) SetGateway(cc *types.ContractContext, Gateway common.Address, Is bool) error {
	return f.cont.SetGateway(cc, Gateway, Is)
}
func (f *front) TokenInRevert(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	return f.cont.TokenInRevert(cc, Platform, ercHash, to, Amount)
}
func (f *front) SetRouter(cc *types.ContractContext, router common.Address, path []common.Address) error {
	return f.cont.SetRouter(cc, router, path)
}
func (f *front) SwapToMainToken(cc *types.ContractContext, amt *amount.Amount) (*amount.Amount, error) {
	return f.cont.SwapToMainToken(cc, amt)
}
func (f *front) SetName(cc *types.ContractContext, name string) {
	f.cont.SetName(cc, name)
}
func (f *front) SetSymbol(cc *types.ContractContext, symbol string) {
	f.cont.SetSymbol(cc, symbol)
}
func (f *front) IsPause(cc *types.ContractContext) bool {
	return f.cont.isPause(cc)
}
func (f *front) Pause(cc *types.ContractContext) error {
	return f.cont.Pause(cc)
}
func (f *front) Unpause(cc *types.ContractContext) error {
	return f.cont.Unpause(cc)
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
