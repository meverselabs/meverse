package router

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

func (cont *RouterContract) Front() interface{} {
	return &RounterFront{
		cont: cont,
	}
}

type RounterFront struct {
	cont *RouterContract
}

func (f *RounterFront) Factory(cc types.ContractLoader) common.Address {
	return f.cont.factory(cc)
}
func (f *RounterFront) GetAmountsOut(cc *types.ContractContext, amountIn *amount.Amount, path []common.Address) ([]*amount.Amount, error) {
	factory := f.cont.factory(cc)
	amounts, err := getAmountsOut(cc, factory, amountIn.Int, path, cc.From())
	return ToAmounts(amounts), err
}
func (f *RounterFront) UniGetAmountsIn(cc *types.ContractContext, amountOut *amount.Amount, path []common.Address) ([]*amount.Amount, error) {
	factory := f.cont.factory(cc)
	amounts, err := uniGetAmountsIn(cc, factory, amountOut.Int, path, cc.From())
	return ToAmounts(amounts), err
}
func (f *RounterFront) UniGetLPTokenAmount(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	amountADesired, amountBDesired *amount.Amount) (*amount.Amount, uint64, error) {

	_, _, _, liquidity, ratio, err := f.cont._uniAddLiquidity(cc, tokenA, tokenB, amountADesired.Int, amountBDesired.Int, Zero, Zero, true)
	return ToAmount(liquidity), ratio, err
}
func (f *RounterFront) UniAddLiquidity(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	amountADesired, amountBDesired, amountAMin, amountBMin *amount.Amount) (*amount.Amount, *amount.Amount, *amount.Amount, common.Address, error) {

	amountA, amountB, liquidity, pair, err := f.cont.uniAddLiquidity(cc, tokenA, tokenB, amountADesired.Int, amountBDesired.Int, amountAMin.Int, amountBMin.Int, cc.From())
	return ToAmount(amountA), ToAmount(amountB), ToAmount(liquidity), pair, err
}
func (f *RounterFront) UniGetLPTokenAmountOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB, tokenIn common.Address,
	amountDesired *amount.Amount) (*amount.Amount, uint64, error) {

	_, _, _, liquidity, ratio, err := f.cont.uniGetLPTokenAmountOneCoin(cc, tokenA, tokenB, tokenIn, amountDesired.Int, true, cc.From())
	return ToAmount(liquidity), ratio, err
}
func (f *RounterFront) UniAddLiquidityOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB, tokenIn common.Address,
	amountDesired, amountMin *amount.Amount) (*amount.Amount, *amount.Amount, common.Address, error) {

	amt, liquidity, pair, err := f.cont.uniAddLiquidityOneCoin(cc, tokenA, tokenB, tokenIn, amountDesired.Int, amountMin.Int, cc.From(), cc.From())
	return ToAmount(amt), ToAmount(liquidity), pair, err
}
func (f *RounterFront) UniGetWithdrawAmount(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *amount.Amount) (*amount.Amount, *amount.Amount, *amount.Amount, error) {

	_, _, amountA, amountB, minFee, err := f.cont.uniGetWithdrawAmount(cc, tokenA, tokenB, liquidity.Int)
	return ToAmount(amountA), ToAmount(amountB), ToAmount(minFee), err
}
func (f *RounterFront) UniRemoveLiquidity(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity, amountAMin, amountBMin *amount.Amount) (*amount.Amount, *amount.Amount, error) {

	_, _, amountA, amountB, err := f.cont.uniRemoveLiquidity(cc, tokenA, tokenB, liquidity.Int, amountAMin.Int, amountBMin.Int, cc.From())
	return ToAmount(amountA), ToAmount(amountB), err
}
func (f *RounterFront) UniGetWithdrawAmountOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *amount.Amount,
	tokenOut common.Address) (*amount.Amount, *amount.Amount, error) {

	amount, minFee, err := f.cont.uniGetWithdrawAmountOneCoin(cc, tokenA, tokenB, liquidity.Int, tokenOut, cc.From())
	return ToAmount(amount), ToAmount(minFee), err
}
func (f *RounterFront) UniRemoveLiquidityOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *amount.Amount,
	tokenOut common.Address,
	amountMin *amount.Amount) (*amount.Amount, error) {

	amountOut, err := f.cont.uniRemoveLiquidityOneCoin(cc, tokenA, tokenB, liquidity.Int, tokenOut, amountMin.Int, cc.From(), cc.From())
	return ToAmount(amountOut), err
}
func (f *RounterFront) SwapExactTokensForTokens(
	cc *types.ContractContext,
	amountIn, amountOutMin *amount.Amount,
	path []common.Address) ([]*amount.Amount, error) {

	amounts, err := f.cont.swapExactTokensForTokens(cc, amountIn.Int, amountOutMin.Int, path, cc.From(), cc.From())
	return ToAmounts(amounts), err
}
func (f *RounterFront) UniSwapTokensForExactTokens(
	cc *types.ContractContext,
	amountOut, amountInMax *amount.Amount,
	path []common.Address) ([]*amount.Amount, error) {

	amounts, err := f.cont.uniSwapTokensForExactTokens(cc, amountOut.Int, amountInMax.Int, path, cc.From(), cc.From())
	return ToAmounts(amounts), err
}
func (f *RounterFront) CubicRootType(cc *types.ContractContext) uint8 {
	return f.cont.CubicRootType(cc)
}
func (f *RounterFront) SetCubicRootType(cc *types.ContractContext, cubicRootType uint8) {
	f.cont.SetCubicRootType(cc, cubicRootType)
}
