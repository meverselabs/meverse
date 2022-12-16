package erc20wrapper

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *Erc20WrapperContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *Erc20WrapperContract
}

func (f *front) Erc20Token(cc types.ContractLoader) common.Address {
	return f.cont.Erc20Token(cc)
}

// func (f *front) SetErc20Token(cc *types.ContractContext, erc20 common.Address) error {
// 	return f.cont.SetErc20Token(cc, erc20)
// }

func (f *front) Name(cc *types.ContractContext) (string, error) {
	return f.cont.Name(cc)
}

func (f *front) Symbol(cc *types.ContractContext) (string, error) {
	return f.cont.Symbol(cc)
}

func (f *front) Decimals(cc *types.ContractContext) (*big.Int, error) {
	return f.cont.Decimals(cc)
}

func (f *front) TotalSupply(cc *types.ContractContext) (*amount.Amount, error) {
	return f.cont.TotalSupply(cc)
}

func (f *front) BalanceOf(cc *types.ContractContext, from common.Address) (*amount.Amount, error) {
	return f.cont.BalanceOf(cc, from)
}

func (f *front) Allowance(cc *types.ContractContext, _owner common.Address, _spender common.Address) (*amount.Amount, error) {
	return f.cont.Allowance(cc, _owner, _spender)
}

func (f *front) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.Approve(cc, To, Amount)
	return err == nil, err
}

func (f *front) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.Transfer(cc, To, Amount)
	return err == nil, err
}

func (f *front) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) (bool, error) {
	err := f.cont.TransferFrom(cc, From, To, Amount)
	return err == nil, err
}

func (f *front) IncreaseAllowance(cc *types.ContractContext, Spender common.Address, AddedValue *amount.Amount) (bool, error) {
	err := f.cont.IncreaseAllowance(cc, Spender, AddedValue)
	return err == nil, err
}

func (f *front) DecreaseAllowance(cc *types.ContractContext, Spender common.Address, SubtractedValue *amount.Amount) (bool, error) {
	err := f.cont.DecreaseAllowance(cc, Spender, SubtractedValue)
	return err == nil, err
}

func (f *front) IsMinter(cc *types.ContractContext, addr common.Address) (bool, error) {
	return f.cont.IsMinter(cc, addr)
}

func (f *front) SetMinter(cc *types.ContractContext, To common.Address, Is bool) error {
	return f.cont.SetMinter(cc, To, Is)
}

func (f *front) Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.Mint(cc, To, Amount)
}

func (f *front) Burn(cc *types.ContractContext, am *amount.Amount) error {
	return f.cont.Burn(cc, am)
}

func (f *front) BurnFrom(cc *types.ContractContext, addr common.Address, am *amount.Amount) error {
	return f.cont.BurnFrom(cc, addr, am)
}
