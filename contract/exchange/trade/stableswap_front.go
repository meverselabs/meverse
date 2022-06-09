package trade

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

func (cont *StableSwap) Front() interface{} {
	return &StableSwapFront{
		cont: cont,
	}
}

type StableSwapFront struct {
	cont *StableSwap
}

//////////////////////////////////////////////////
// Token
//////////////////////////////////////////////////
func (f *StableSwapFront) Name(cc types.ContractLoader) string {
	return f.cont.name(cc)
}
func (f *StableSwapFront) SetName(cc *types.ContractContext, name string) error {
	return f.cont.setName(cc, name)
}
func (f *StableSwapFront) Symbol(cc types.ContractLoader) string {
	return f.cont.symbol(cc)
}
func (f *StableSwapFront) SetSymbol(cc *types.ContractContext, symbol string) error {
	return f.cont.setSymbol(cc, symbol)
}
func (f *StableSwapFront) TotalSupply(cc types.ContractLoader) *amount.Amount {
	return ToAmount(f.cont.totalSupply(cc))
}
func (f *StableSwapFront) Decimals(cc types.ContractLoader) *big.Int {
	return f.cont.decimals(cc)
}
func (f *StableSwapFront) BalanceOf(cc types.ContractLoader, from common.Address) *amount.Amount {
	return ToAmount(f.cont.balanceOf(cc, from))
}
func (f *StableSwapFront) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.transfer(cc, To, Amount.Int)
}
func (f *StableSwapFront) Allowance(cc types.ContractLoader, owner, spender common.Address) *amount.Amount {
	return ToAmount(f.cont.allowance(cc, owner, spender))
}
func (f *StableSwapFront) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	return f.cont.approve(cc, To, Amount.Int)
}
func (f *StableSwapFront) IncreaseAllowance(cc *types.ContractContext, spender common.Address, addAmount *amount.Amount) error {
	return f.cont.increaseAllowance(cc, spender, addAmount.Int)
}
func (f *StableSwapFront) DecreaseAllowance(cc *types.ContractContext, spender common.Address, subtractAmount *amount.Amount) error {
	return f.cont.decreaseAllowance(cc, spender, subtractAmount.Int)
}
func (f *StableSwapFront) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	return f.cont.transferFrom(cc, From, To, Amount.Int)
}

//////////////////////////////////////////////////
// Exchange : public reader functions
//////////////////////////////////////////////////

func (f *StableSwapFront) ExType(cc types.ContractLoader) uint8 {
	return f.cont.exType(cc)
}
func (f *StableSwapFront) Factory(cc types.ContractLoader) common.Address {
	return f.cont.factory(cc)
}
func (f *StableSwapFront) Owner(cc types.ContractLoader) common.Address {
	return f.cont.owner(cc)
}
func (f *StableSwapFront) Winner(cc types.ContractLoader) common.Address {
	return f.cont.winner(cc)
}
func (f *StableSwapFront) FutureOwner(cc types.ContractLoader) common.Address {
	return f.cont.futureOwner(cc)
}
func (f *StableSwapFront) FutureWinner(cc types.ContractLoader) common.Address {
	return f.cont.futureWinner(cc)
}
func (f *StableSwapFront) TransferOwnerWinnerDeadline(cc types.ContractLoader) uint64 {
	return f.cont.transferOwnerWinnerDeadline(cc)
}
func (f *StableSwapFront) Fee(cc *types.ContractContext) uint64 {
	return f.cont.fee(cc)
}
func (f *StableSwapFront) FutureFee(cc types.ContractLoader) uint64 {
	return f.cont.futureFee(cc)
}
func (f *StableSwapFront) AdminFee(cc types.ContractLoader) uint64 {
	return f.cont.adminFee(cc)
}
func (f *StableSwapFront) FutureAdminFee(cc types.ContractLoader) uint64 {
	return f.cont.futureAdminFee(cc)
}
func (f *StableSwapFront) WinnerFee(cc types.ContractLoader) uint64 {
	return f.cont.winnerFee(cc)
}
func (f *StableSwapFront) FutureWinnerFee(cc types.ContractLoader) uint64 {
	return f.cont.futureWinnerFee(cc)
}
func (f *StableSwapFront) AdminActionsDeadline(cc types.ContractLoader) uint64 {
	return f.cont.adminActionsDeadline(cc)
}
func (f *StableSwapFront) FeeWhiteList(cc *types.ContractContext, from common.Address) ([]byte, error) {
	return f.cont.feeWhiteList(cc, from)
}
func (f *StableSwapFront) FeeAddress(cc *types.ContractContext, from common.Address) (uint64, error) {
	return f.cont.feeAddress(cc, from)
}
func (f *StableSwapFront) WhiteList(cc types.ContractLoader) common.Address {
	return f.cont.whiteList(cc)
}
func (f *StableSwapFront) GroupId(cc types.ContractLoader) hash.Hash256 {
	return f.cont.groupId(cc)
}
func (f *StableSwapFront) FutureWhiteList(cc types.ContractLoader) common.Address {
	return f.cont.futureWhiteList(cc)
}
func (f *StableSwapFront) FutureGroupId(cc types.ContractLoader) hash.Hash256 {
	return f.cont.futureGroupId(cc)
}
func (f *StableSwapFront) WhiteListDeadline(cc types.ContractLoader) uint64 {
	return f.cont.whiteListDeadline(cc)
}
func (f *StableSwapFront) NTokens(cc *types.ContractContext) uint8 {
	return f.cont.nTokens(cc)
}
func (f *StableSwapFront) Tokens(cc types.ContractLoader) []common.Address {
	return f.cont.tokens(cc)
}
func (f *StableSwapFront) PayToken(cc types.ContractLoader) common.Address {
	return f.cont.payToken(cc)
}
func (f *StableSwapFront) PayTokenIndex(cc types.ContractLoader) (uint8, error) {
	return f.cont.payTokenIndex(cc)
}
func (f *StableSwapFront) IsKilled(cc types.ContractLoader) bool {
	return f.cont.isKilled(cc)
}
func (f *StableSwapFront) BlockTimestampLast(cc types.ContractLoader) uint64 {
	return f.cont.blockTimestampLast(cc)
}

//////////////////////////////////////////////////
// StableSwap : public reader functions
//////////////////////////////////////////////////
func (f *StableSwapFront) TokenIndex(cc types.ContractLoader, _token common.Address) (uint8, error) {
	return f.cont.tokenIndex(cc, _token)
}
func (f *StableSwapFront) Rates(cc types.ContractLoader) []*big.Int {
	return f.cont.rates(cc)
}
func (f *StableSwapFront) PrecisionMul(cc types.ContractLoader) []uint64 {
	return f.cont.precisionMul(cc)
}
func (f *StableSwapFront) Reserves(cc types.ContractLoader) ([]*amount.Amount, uint64) {
	reserves := f.cont.reserves(cc)
	blockTimestampLast := f.cont.blockTimestampLast(cc)
	return ToAmounts(reserves), blockTimestampLast
}
func (f *StableSwapFront) AdminBalances(cc *types.ContractContext, idx uint8) (*amount.Amount, error) {
	result, err := f.cont.adminBalances(cc, idx)
	return ToAmount(result), err
}

// //////////////////////////////////////////////////
// // Exchange : public writer Functions
// //////////////////////////////////////////////////
func (f *StableSwapFront) SetPayToken(cc *types.ContractContext, _token common.Address) error {
	return f.cont.setPayToken(cc, _token)
}
func (f *StableSwapFront) CommitNewFee(cc *types.ContractContext, new_fee, new_admin_fee, new_winner_fee, deadline uint64) error {
	return f.cont.commitNewFee(cc, new_fee, new_admin_fee, new_winner_fee, deadline)
}
func (f *StableSwapFront) ApplyNewFee(cc *types.ContractContext) error {
	return f.cont.applyNewFee(cc)
}
func (f *StableSwapFront) RevertNewFee(cc *types.ContractContext) error {
	return f.cont.revertNewFee(cc)
}
func (f *StableSwapFront) CommitNewWhiteList(cc *types.ContractContext, new_whiteList common.Address, new_groupId hash.Hash256, deadline uint64) error {
	return f.cont.commitNewWhiteList(cc, new_whiteList, new_groupId, deadline)
}
func (f *StableSwapFront) ApplyNewWhiteList(cc *types.ContractContext) error {
	return f.cont.applyNewWhiteList(cc)
}
func (f *StableSwapFront) RevertNewWhiteList(cc *types.ContractContext) error {
	return f.cont.revertNewWhiteList(cc)
}
func (f *StableSwapFront) CommitTransferOwnerWinner(cc *types.ContractContext, new_owner, new_winner common.Address, deadline uint64) error {
	return f.cont.commitTransferOwnerWinner(cc, new_owner, new_winner, deadline)
}
func (f *StableSwapFront) ApplyTransferOwnerWinner(cc *types.ContractContext) error {
	_, _, err := f.cont._applyTransferOwnerWinner(cc)
	return err
}
func (f *StableSwapFront) RevertTransferOwnerWinner(cc *types.ContractContext) error {
	return f.cont.revertTransferOwnerWinner(cc)
}
func (f *StableSwapFront) KillMe(cc *types.ContractContext) error {
	return f.cont.killMe(cc)
}
func (f *StableSwapFront) UnkillMe(cc *types.ContractContext) error {
	return f.cont.unkillMe(cc)
}

//////////////////////////////////////////////////
// StableSwap Parameter Functions
//////////////////////////////////////////////////
func (f *StableSwapFront) InitialA(cc types.ContractLoader) *big.Int {
	return f.cont.initialA(cc)
}
func (f *StableSwapFront) FutureA(cc types.ContractLoader) *big.Int {
	return f.cont.futureA(cc)
}
func (f *StableSwapFront) InitialATime(cc types.ContractLoader) uint64 {
	return f.cont.initialATime(cc)
}
func (f *StableSwapFront) FutureATime(cc types.ContractLoader) uint64 {
	return f.cont.futureATime(cc)
}
func (f *StableSwapFront) A(cc types.ContractLoader) *big.Int {
	return f.cont.a(cc)
}
func (f *StableSwapFront) APrecise(cc types.ContractLoader) *big.Int {
	return f.cont.aPrecise(cc)
}
func (f *StableSwapFront) GetVirtualPrice(cc types.ContractLoader) (*amount.Amount, error) {
	result, err := f.cont.getVirtualPrice(cc)
	return ToAmount(result), err
}
func (f *StableSwapFront) WithdrawAdminFees(cc *types.ContractContext) ([]*amount.Amount, []*amount.Amount, error) {
	ownerFees, winnerFees, err := f.cont.withdrawAdminFees(cc)
	return ToAmounts(ownerFees), ToAmounts(winnerFees), err
}
func (f *StableSwapFront) DonateAdminFees(cc *types.ContractContext) error {
	return f.cont.donateAdminFees(cc)
}
func (f *StableSwapFront) RampA(cc *types.ContractContext, _future_A *big.Int, _future_time uint64) error {
	return f.cont.rampA(cc, _future_A, _future_time)
}
func (f *StableSwapFront) StopRampA(cc *types.ContractContext) error {
	return f.cont.stopRampA(cc)
}

//////////////////////////////////////////////////
// StableSwap Liquidity & Swap Functions
//////////////////////////////////////////////////
func (f *StableSwapFront) CalcLPTokenAmount(cc *types.ContractContext, _amounts []*amount.Amount, _is_deposit bool) (*amount.Amount, uint64, error) {
	amt, ratio, err := f.cont.calcLPTokenAmount(cc, ToBigInts(_amounts), _is_deposit)
	return ToAmount(amt), ratio, err
}
func (f *StableSwapFront) AddLiquidity(cc *types.ContractContext, _amounts []*amount.Amount, _min_mint_amount *amount.Amount) (*amount.Amount, error) {
	result, err := f.cont.addLiquidity(cc, ToBigInts(_amounts), _min_mint_amount.Int)
	return ToAmount(result), err
}
func (f *StableSwapFront) CalcWithdrawCoins(cc types.ContractLoader, _amount *amount.Amount) ([]*amount.Amount, error) {
	result, err := f.cont.calcWithdrawCoins(cc, _amount.Int)
	return ToAmounts(result), err
}
func (f *StableSwapFront) RemoveLiquidity(cc *types.ContractContext, _amount *amount.Amount, _min_amounts []*amount.Amount) ([]*amount.Amount, error) {
	result, err := f.cont.removeLiquidity(cc, _amount.Int, ToBigInts(_min_amounts))
	return ToAmounts(result), err
}
func (f *StableSwapFront) RemoveLiquidityImbalance(cc *types.ContractContext, _amounts []*amount.Amount, _max_burn_amount *amount.Amount) (*amount.Amount, error) {
	result, err := f.cont.removeLiquidityImbalance(cc, ToBigInts(_amounts), _max_burn_amount.Int)
	return ToAmount(result), err
}
func (f *StableSwapFront) CalcWithdrawOneCoin(cc *types.ContractContext, liquidity *amount.Amount, out uint8) (*amount.Amount, *amount.Amount, *amount.Amount, error) {
	dy, dyFee, tokenSupply, err := f.cont.calcWithdrawOneCoin(cc, liquidity.Int, out)
	return ToAmount(dy), ToAmount(dyFee), ToAmount(tokenSupply), err
}
func (f *StableSwapFront) RemoveLiquidityOneCoin(cc *types.ContractContext, liquidity *amount.Amount, out uint8, _min_amount *amount.Amount) (*amount.Amount, error) {
	result, err := f.cont.removeLiquidityOneCoin(cc, liquidity.Int, out, _min_amount.Int)
	return ToAmount(result), err
}
func (f *StableSwapFront) GetDy(cc *types.ContractContext, in, out uint8, dx *amount.Amount, from common.Address) (*amount.Amount, error) {
	result, _, _, err := f.cont.get_dy(cc, in, out, dx.Int, from)
	return ToAmount(result), err
}
func (f *StableSwapFront) Exchange(cc *types.ContractContext, in, out uint8, dx, _min_dy *amount.Amount, from common.Address) (*amount.Amount, error) {
	result, err := f.cont.exchange(cc, in, out, dx.Int, _min_dy.Int, from)
	return ToAmount(result), err
}
func (f *StableSwapFront) TokenTransfer(cc *types.ContractContext, token, to common.Address, amt *amount.Amount) error {
	return f.cont.tokenTransfer(cc, token, to, amt)
}
