package imo

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *ImoContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *ImoContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (f *front) Deposit(cc *types.ContractContext, am *amount.Amount) error {
	return f.cont.Deposit(cc, am)
}

func (f *front) Harvest(cc *types.ContractContext) error {
	return f.cont.Harvest(cc)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
//temp remove it
func (f *front) SetEndBlock(cc *types.ContractContext, endBlock uint32) error {
	return f.cont.setEndBlock(cc, endBlock)
}
func (f *front) SetOfferingAmount(cc *types.ContractContext, projectOffer *amount.Amount) error {
	return f.cont.setOfferingAmount(cc, projectOffer)
}

func (f *front) SetRaisingAmount(cc *types.ContractContext, projectRaising *amount.Amount) error {
	return f.cont.setRaisingAmount(cc, projectRaising)
}

func (f *front) FinalWithdraw(cc *types.ContractContext, payAmount *amount.Amount, offerAmount *amount.Amount) error {
	return f.cont.finalWithdraw(cc, payAmount, offerAmount)
}

func (f *front) UsdcReclaim(cc *types.ContractContext) error {
	return f.cont.usdcReclaim(cc)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) ProjectOwner(cc *types.ContractContext) common.Address {
	return f.cont.ProjectOwner(cc)
}
func (f *front) PayToken(cc *types.ContractContext) common.Address {
	return f.cont.PayToken(cc)
}
func (f *front) ProjectToken(cc *types.ContractContext) common.Address {
	return f.cont.ProjectToken(cc)
}
func (f *front) ProjectOffering(cc *types.ContractContext) *amount.Amount {
	return f.cont.ProjectOffering(cc)
}
func (f *front) ProjectRaising(cc *types.ContractContext) *amount.Amount {
	return f.cont.ProjectRaising(cc)
}
func (f *front) PayLimit(cc *types.ContractContext) *amount.Amount {
	return f.cont.PayLimit(cc)
}
func (f *front) StartBlock(cc *types.ContractContext) uint32 {
	return f.cont.StartBlock(cc)
}
func (f *front) EndBlock(cc *types.ContractContext) uint32 {
	return f.cont.EndBlock(cc)
}
func (f *front) WhiteListGroupId(cc *types.ContractContext) hash.Hash256 {
	return f.cont.WhiteListGroupId(cc)
}

func (f *front) IsOwner(cc *types.ContractContext) bool {
	return f.cont.IsOwner(cc)
}

func (f *front) HarvestFeeFactor(cc *types.ContractContext) uint16 {
	return f.cont.HarvestFeeFactor(cc)
}
func (f *front) WhiteListAddress(cc *types.ContractContext) common.Address {
	return f.cont.WhiteListAddress(cc)
}
func (f *front) TotalAmount(cc *types.ContractContext) *amount.Amount {
	return f.cont.TotalAmount(cc)
}

func (f *front) UserInfo(cc *types.ContractContext, user common.Address) (*amount.Amount, bool, error) {
	return f.cont.UserInfo(cc, user)
}

func (f *front) AddressList(cc *types.ContractContext) ([]common.Address, error) {
	return f.cont.AddressList(cc)
}

func (f *front) HasHarvest(cc *types.ContractContext, _user common.Address) (bool, error) {
	return f.cont.HasHarvest(cc, _user)
}

func (f *front) GetUserAllocation(cc *types.ContractContext, _user common.Address) *big.Int {
	return f.cont.GetUserAllocation(cc, _user)
}

func (f *front) GetOfferingAmount(cc *types.ContractContext, _user common.Address) (*amount.Amount, error) {
	return f.cont.GetOfferingAmount(cc, _user)
}

func (f *front) GetRefundingAmount(cc *types.ContractContext, _user common.Address) (*amount.Amount, error) {
	return f.cont.GetRefundingAmount(cc, _user)
}

func (f *front) GetAddressListLength(cc *types.ContractContext) (int, error) {
	return f.cont.GetAddressListLength(cc)
}

func (f *front) CheckWhiteList(cc *types.ContractContext, _user common.Address) (bool, error) {
	return f.cont.CheckWhiteList(cc, _user)
}
