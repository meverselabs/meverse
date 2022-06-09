package trade

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

func (cont *UniSwap) Front() interface{} {
	return &UniSwapFront{
		cont: cont,
	}
}

type UniSwapFront struct {
	cont *UniSwap
}

//////////////////////////////////////////////////
// Token
//////////////////////////////////////////////////
func (f *UniSwapFront) Name(cc types.ContractLoader) string {
	return f.cont.name(cc)
}
func (f *UniSwapFront) SetName(cc *types.ContractContext, name string) error {
	return f.cont.setName(cc, name)
}
func (f *UniSwapFront) Symbol(cc types.ContractLoader) string {
	return f.cont.symbol(cc)
}
func (f *UniSwapFront) SetSymbol(cc *types.ContractContext, symbol string) error {
	return f.cont.setSymbol(cc, symbol)
}
func (f *UniSwapFront) TotalSupply(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.totalSupply(cc))
}
func (f *UniSwapFront) Decimals(cc types.ContractLoader) *big.Int {
	return f.cont.decimals(cc)
}
func (f *UniSwapFront) BalanceOf(cc types.ContractLoader, from common.Address) *amount.Amount {
	return ToAmount(f.cont.balanceOf(cc, from))
}
func (f *UniSwapFront) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.transfer(cc, To, Amount.Int)
}
func (f *UniSwapFront) Allowance(cc types.ContractLoader, owner, spender common.Address) *amount.Amount {
	return ToAmount(f.cont.allowance(cc, owner, spender))
}
func (f *UniSwapFront) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.approve(cc, To, Amount.Int)
}
func (f *UniSwapFront) IncreaseAllowance(cc *types.ContractContext, spender common.Address, addAmount *amount.Amount) error {
	return f.cont.increaseAllowance(cc, spender, addAmount.Int)
}
func (f *UniSwapFront) DecreaseAllowance(cc *types.ContractContext, spender common.Address, subtractAmount *amount.Amount) error {
	return f.cont.decreaseAllowance(cc, spender, subtractAmount.Int)
}
func (f *UniSwapFront) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	return f.cont.transferFrom(cc, From, To, Amount.Int)
}

//////////////////////////////////////////////////
// Exchange : public reader functions
//////////////////////////////////////////////////
func (f *UniSwapFront) ExType(cc types.ContractLoader) uint8 {
	return f.cont.exType(cc)
}
func (f *UniSwapFront) Factory(cc types.ContractLoader) common.Address {
	return f.cont.factory(cc)
}
func (f *UniSwapFront) Owner(cc types.ContractLoader) common.Address {
	return f.cont.owner(cc)
}
func (f *UniSwapFront) Winner(cc types.ContractLoader) common.Address {
	return f.cont.winner(cc)
}
func (f *UniSwapFront) FutureOwner(cc types.ContractLoader) common.Address {
	return f.cont.futureOwner(cc)
}
func (f *UniSwapFront) FutureWinner(cc types.ContractLoader) common.Address {
	return f.cont.futureWinner(cc)
}
func (f *UniSwapFront) TransferOwnerWinnerDeadline(cc types.ContractLoader) uint64 {
	return f.cont.transferOwnerWinnerDeadline(cc)
}
func (f *UniSwapFront) Fee(cc *types.ContractContext) uint64 {
	return f.cont.fee(cc)
}
func (f *UniSwapFront) FutureFee(cc types.ContractLoader) uint64 {
	return f.cont.futureFee(cc)
}
func (f *UniSwapFront) AdminFee(cc types.ContractLoader) uint64 {
	return f.cont.adminFee(cc)
}
func (f *UniSwapFront) FutureAdminFee(cc types.ContractLoader) uint64 {
	return f.cont.futureAdminFee(cc)
}
func (f *UniSwapFront) WinnerFee(cc types.ContractLoader) uint64 {
	return f.cont.winnerFee(cc)
}
func (f *UniSwapFront) FutureWinnerFee(cc types.ContractLoader) uint64 {
	return f.cont.futureWinnerFee(cc)
}
func (f *UniSwapFront) AdminActionsDeadline(cc types.ContractLoader) uint64 {
	return f.cont.adminActionsDeadline(cc)
}
func (f *UniSwapFront) FeeWhiteList(cc *types.ContractContext, from common.Address) ([]byte, error) {
	return f.cont.feeWhiteList(cc, from)
}
func (f *UniSwapFront) FeeAddress(cc *types.ContractContext, from common.Address) (uint64, error) {
	return f.cont.feeAddress(cc, from)
}
func (f *UniSwapFront) WhiteList(cc types.ContractLoader) common.Address {
	return f.cont.whiteList(cc)
}
func (f *UniSwapFront) GroupId(cc types.ContractLoader) hash.Hash256 {
	return f.cont.groupId(cc)
}
func (f *UniSwapFront) FutureWhiteList(cc types.ContractLoader) common.Address {
	return f.cont.futureWhiteList(cc)
}
func (f *UniSwapFront) FutureGroupId(cc types.ContractLoader) hash.Hash256 {
	return f.cont.futureGroupId(cc)
}
func (f *UniSwapFront) WhiteListDeadline(cc types.ContractLoader) uint64 {
	return f.cont.whiteListDeadline(cc)
}
func (f *UniSwapFront) NTokens(cc types.ContractLoader) uint8 {
	return f.cont.nTokens(cc)
}
func (f *UniSwapFront) Tokens(cc types.ContractLoader) []common.Address {
	return f.cont.tokens(cc)
}
func (f *UniSwapFront) PayToken(cc types.ContractLoader) common.Address {
	return f.cont.payToken(cc)
}
func (f *UniSwapFront) PayTokenIndex(cc types.ContractLoader) (uint8, error) {
	return f.cont.payTokenIndex(cc)
}
func (f *UniSwapFront) IsKilled(cc types.ContractLoader) bool {
	return f.cont.isKilled(cc)
}
func (f *UniSwapFront) BlockTimestampLast(cc types.ContractLoader) uint64 {
	return f.cont.blockTimestampLast(cc)
}

//////////////////////////////////////////////////
// Uniswap : public reader functions
//////////////////////////////////////////////////
func (f *UniSwapFront) Token0(cc types.ContractLoader) common.Address {
	return f.cont.token0(cc)
}
func (f *UniSwapFront) Token1(cc types.ContractLoader) common.Address {
	return f.cont.token1(cc)
}
func (f *UniSwapFront) Reserve0(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.reserve0(cc))
}
func (f *UniSwapFront) Reserve1(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.reserve1(cc))
}
func (f *UniSwapFront) Reserves(cc types.ContractLoader) ([]*amount.Amount, uint64) {
	reserve0, reserve1, blockTimestampLast := f.cont.reserves(cc)
	return []*amount.Amount{ToAmount(reserve0), ToAmount(reserve1)}, blockTimestampLast
}
func (f *UniSwapFront) Price0CumulativeLast(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.price0CumulativeLast(cc))
}
func (f *UniSwapFront) Price1CumulativeLast(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.price1CumulativeLast(cc))
}
func (f *UniSwapFront) KLast(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.kLast(cc))
}
func (f *UniSwapFront) GetMintAdminFee(cc types.ContractLoader, _reserve0, _reserve1 *amount.Amount) *amount.Amount {
	_, _, _, liquidity := f.cont.getMintAdminFee(cc, _reserve0.Int, _reserve1.Int)
	return ToAmount(liquidity)
}
func (f *UniSwapFront) MintedAdminBalance(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.mintedAdminBalance(cc))
}
func (f *UniSwapFront) AdminBalance(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.adminBalance(cc))
}

// //////////////////////////////////////////////////
// // Exchange : public writer Functions
// //////////////////////////////////////////////////
func (f *UniSwapFront) SetPayToken(cc *types.ContractContext, _token common.Address) error {
	return f.cont.setPayToken(cc, _token)
}
func (f *UniSwapFront) CommitNewFee(cc *types.ContractContext, new_fee, new_admin_fee, new_winner_fee, deadline uint64) error {
	return f.cont.commitNewFee(cc, new_fee, new_admin_fee, new_winner_fee, deadline)
}
func (f *UniSwapFront) ApplyNewFee(cc *types.ContractContext) error {
	return f.cont.applyNewFee(cc)
}
func (f *UniSwapFront) RevertNewFee(cc *types.ContractContext) error {
	return f.cont.revertNewFee(cc)
}
func (f *UniSwapFront) CommitNewWhiteList(cc *types.ContractContext, new_whiteList common.Address, new_groupId hash.Hash256, deadline uint64) error {
	return f.cont.commitNewWhiteList(cc, new_whiteList, new_groupId, deadline)
}
func (f *UniSwapFront) ApplyNewWhiteList(cc *types.ContractContext) error {
	return f.cont.applyNewWhiteList(cc)
}
func (f *UniSwapFront) RevertNewWhiteList(cc *types.ContractContext) error {
	return f.cont.revertNewWhiteList(cc)
}
func (f *UniSwapFront) CommitTransferOwnerWinner(cc *types.ContractContext, new_owner, new_winner common.Address, deadline uint64) error {
	return f.cont.commitTransferOwnerWinner(cc, new_owner, new_winner, deadline)
}
func (f *UniSwapFront) ApplyTransferOwnerWinner(cc *types.ContractContext) error {
	return f.cont.applyTransferOwnerWinner(cc)
}
func (f *UniSwapFront) RevertTransferOwnerWinner(cc *types.ContractContext) error {
	return f.cont.revertTransferOwnerWinner(cc)
}
func (f *UniSwapFront) KillMe(cc *types.ContractContext) error {
	return f.cont.killMe(cc)
}
func (f *UniSwapFront) UnkillMe(cc *types.ContractContext) error {
	return f.cont.unkillMe(cc)
}

//////////////////////////////////////////////////
// Uniswap : public writer functions
//////////////////////////////////////////////////
func (f *UniSwapFront) Mint(cc *types.ContractContext, to common.Address) (*amount.Amount, error) {
	liqudity, err := f.cont.mint(cc, to)
	if err != nil {
		return nil, err
	}
	return &amount.Amount{Int: liqudity}, err
}
func (f *UniSwapFront) Burn(cc *types.ContractContext, to common.Address) (*amount.Amount, *amount.Amount, error) {
	amount0, amount1, err := f.cont.burn(cc, to)
	if err != nil {
		return nil, nil, err
	}
	return &amount.Amount{Int: amount0}, &amount.Amount{Int: amount1}, err
}
func (f *UniSwapFront) Swap(cc *types.ContractContext, amount0Out, amount1Out *amount.Amount, to common.Address, data []byte, from common.Address) error {
	return f.cont.swap(cc, amount0Out.Int, amount1Out.Int, to, data, from)
}
func (f *UniSwapFront) WithdrawAdminFees(cc *types.ContractContext) (*amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount, error) {
	burnAmount, adminFee0, adminFee1, winnerFee0, winnerFee1, err := f.cont.withdrawAdminFees(cc)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return ToAmount(burnAmount), ToAmount(adminFee0), ToAmount(adminFee1), ToAmount(winnerFee0), ToAmount(winnerFee1), err
}
func (f *UniSwapFront) WithdrawAdminFees2(cc *types.ContractContext) (*amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount, error) {
	burnAmount, adminFee0, adminFee1, winnerFee0, winnerFee1, err := f.cont.withdrawAdminFees2(cc)
	return ToAmount(burnAmount), ToAmount(adminFee0), ToAmount(adminFee1), ToAmount(winnerFee0), ToAmount(winnerFee1), err
}
func (f *UniSwapFront) Skim(cc *types.ContractContext, to common.Address) error {
	return f.cont.skim(cc, to)
}
func (f *UniSwapFront) Sync(cc *types.ContractContext) error {
	return f.cont.sync(cc)
}
func (f *UniSwapFront) TokenTransfer(cc *types.ContractContext, token, to common.Address, amt *amount.Amount) error {
	return f.cont.tokenTransfer(cc, token, to, amt)
}
