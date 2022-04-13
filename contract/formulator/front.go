package formulator

import (
	"math/big"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/fletaio/fleta_v2/core/types"
)

func (cont *FormulatorContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *FormulatorContract
}

func (f *front) CreateGenesisAlpha(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisAlpha(cc, owner)
}

func (f *front) CreateGenesisSigma(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisSigma(cc, owner)
}

func (f *front) CreateGenesisOmega(cc *types.ContractContext, owner common.Address) (common.Address, error) {
	return f.cont.CreateGenesisOmega(cc, owner)
}

func (f *front) AddGenesisStakingAmount(cc *types.ContractContext, HyperAddress common.Address, StakingAddress common.Address, StakingAmount *amount.Amount) error {
	return f.cont.AddGenesisStakingAmount(cc, HyperAddress, StakingAddress, StakingAmount)
}

func (f *front) CreateAlpha(cc *types.ContractContext) (common.Address, error) {
	return f.cont.CreateAlpha(cc)
}

func (f *front) CreateAlphaBatch(cc *types.ContractContext, count *big.Int) ([]common.Address, error) {
	return f.cont.CreateAlphaBatch(cc, count)
}

func (f *front) CreateSigma(cc *types.ContractContext, TokenIDs []common.Address) error {
	return f.cont.CreateSigma(cc, TokenIDs)
}

func (f *front) CreateOmega(cc *types.ContractContext, TokenIDs []common.Address) error {
	return f.cont.CreateOmega(cc, TokenIDs)
}

func (f *front) Revoke(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.Revoke(cc, TokenID)
}

func (f *front) RevokeBatch(cc *types.ContractContext, TokenIDs []common.Address) ([]common.Address, error) {
	return f.cont.RevokeBatch(cc, TokenIDs)
}

func (f *front) Stake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	return f.cont.Stake(cc, HyperAddress, Amount)
}

func (f *front) Unstake(cc *types.ContractContext, HyperAddress common.Address, Amount *amount.Amount) error {
	return f.cont.Unstake(cc, HyperAddress, Amount)
}

func (f *front) Approve(cc *types.ContractContext, To common.Address, TokenID common.Address) error {
	return f.cont.Approve(cc, To, TokenID)
}

func (f *front) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, TokenID common.Address) error {
	return f.cont.TransferFrom(cc, From, To, TokenID)
}

func (f *front) RegisterSales(cc *types.ContractContext, TokenID common.Address, Amount *amount.Amount) error {
	return f.cont.RegisterSales(cc, TokenID, Amount)
}

func (f *front) CancelSales(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.CancelSales(cc, TokenID)
}

func (f *front) BuyFormulator(cc *types.ContractContext, TokenID common.Address) error {
	return f.cont.BuyFormulator(cc, TokenID)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) Formulator(cc types.ContractLoader, _tokenID common.Address) (*Formulator, error) {
	return f.cont.Formulator(cc, _tokenID)
}

func (f *front) StakingAmount(cc types.ContractLoader, HyperAddress common.Address, StakingAddress common.Address) *amount.Amount {
	return f.cont.StakingAmount(cc, HyperAddress, StakingAddress)
}

func (f *front) StakingAmountMap(cc types.ContractLoader, HyperAddress common.Address) (map[common.Address]*amount.Amount, error) {
	return f.cont.StakingAmountMap(cc, HyperAddress)
}

func (f *front) FormulatorMap(cc types.ContractLoader) (map[common.Address]*Formulator, error) {
	return f.cont.FormulatorMap(cc)
}

func (f *front) BalanceOf(cc types.ContractLoader, _owner common.Address) uint32 {
	return f.cont.BalanceOf(cc, _owner)
}

func (f *front) OwnerOf(cc types.ContractLoader, _tokenID common.Address) (common.Address, error) {
	return f.cont.OwnerOf(cc, _tokenID)
}

func (f *front) GetApproved(cc types.ContractLoader, TokenID common.Address) common.Address {
	return f.cont.GetApproved(cc, TokenID)
}
