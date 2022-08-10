package bridge

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *BridgeContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *BridgeContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////
func (f *front) SendToGateway(cc *types.ContractContext, token common.Address, amt *amount.Amount, path []common.Address, toChain string, summary []byte) error {
	return f.cont.sendToGateway(cc, token, amt, path, toChain, summary)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
func (f *front) SetTransferFeeInfo(cc *types.ContractContext, chain string, transferFee *amount.Amount) {
	f.cont.setTransferFeeInfo(cc, chain, transferFee)
}
func (f *front) SetTransferTokenFeeInfo(cc *types.ContractContext, chain string, tokenFee uint16) {
	f.cont.setTransferTokenFeeInfo(cc, chain, tokenFee)
}

func (f *front) SetTokenFeeInfo(cc *types.ContractContext, chain string, tokenFee uint16) {
	f.cont.setTokenFeeInfo(cc, chain, tokenFee)
}

func (f *front) TransferBankOwnership(cc *types.ContractContext, newBank common.Address) error {
	return f.cont.transferBankOwnership(cc, newBank)
}

func (f *front) ChangeMeverseAddress(cc *types.ContractContext, newTokenAddress common.Address) error {
	return f.cont.changeMeverseAddress(cc, newTokenAddress)
}

func (f *front) TransferFeeOwnership(cc *types.ContractContext, newFeeOwner common.Address) error {
	return f.cont.transferFeeOwnership(cc, newFeeOwner)
}

func (f *front) TransferTokenFeeOwnership(cc *types.ContractContext, newFeeOwner common.Address) error {
	return f.cont.transferTokenFeeOwnership(cc, newFeeOwner)
}

func (f *front) ReclaimToken(cc *types.ContractContext, token common.Address, amt *amount.Amount) error {
	return f.cont.reclaimToken(cc, token, amt)
}

//////////////////////////////////////////////////
// Public Writer only banker Functions
//////////////////////////////////////////////////
func (f *front) SendFromGateway(cc *types.ContractContext, token common.Address, to common.Address, amt *amount.Amount, path []common.Address, fromChain string, summary []byte) error {
	return f.cont.sendFromGateway(cc, token, to, amt, path, fromChain, summary)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) MeverseToken(cc *types.ContractContext) common.Address {
	return f.cont.meverseToken(cc)
}
func (f *front) Bank(cc *types.ContractContext) common.Address {
	return f.cont.bank(cc)
}
func (f *front) FeeOwner(cc *types.ContractContext) common.Address {
	return f.cont.feeOwner(cc)
}

func (f *front) TransferFeeInfoToChain(cc *types.ContractContext, chain string) *amount.Amount {
	return f.cont.transferFeeInfoToChain(cc, chain)
}

func (f *front) TokenFeeInfoFromChain(cc *types.ContractContext, chain string) uint16 {
	return f.cont.tokenFeeInfoFromChain(cc, chain)
}

func (f *front) GetSequenceFrom(cc *types.ContractContext, user common.Address, chain string) *big.Int {
	return f.cont.getSequenceFrom(cc, user, chain)
}

func (f *front) GetSequenceTo(cc *types.ContractContext, user common.Address, chain string) *big.Int {
	return f.cont.getSequenceTo(cc, user, chain)
}

func (f *front) AllowanceTokenFromGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error) {
	return f.cont.allowanceTokenFromGateway(cc, token, from)
}

func (f *front) BalanceOfToGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error) {
	return f.cont.balanceOfToGateway(cc, token, from)
}

func (f *front) StringToBytes32(source string) []byte {
	return f.cont.stringToBytes32(source)
}

func (f *front) SetSendMaintoken(cc *types.ContractContext, store common.Address, fromChains []string, overthens, amts []*amount.Amount) error {
	return f.cont.setSendMaintoken(cc, store, fromChains, overthens, amts)
}

func (f *front) UnsetSendMaintoken(cc *types.ContractContext) {
	unsetSendMaintoken(cc)
}

type BridgeFront interface {
	//////////////////////////////////////////////////
	// Public Reader Functions
	//////////////////////////////////////////////////
	MeverseToken(cc *types.ContractContext) common.Address
	Bank(cc *types.ContractContext) common.Address
	FeeOwner(cc *types.ContractContext) common.Address
	TransferFeeInfoToChain(cc *types.ContractContext, chain string) *amount.Amount
	TokenFeeInfoFromChain(cc *types.ContractContext, chain string) uint16
	GetSequenceFrom(cc *types.ContractContext, user common.Address, chain string) *big.Int
	GetSequenceTo(cc *types.ContractContext, user common.Address, chain string) *big.Int

	AllowanceTokenFromGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error)
	BalanceOfToGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error)
	StringToBytes32(source string) []byte
}
