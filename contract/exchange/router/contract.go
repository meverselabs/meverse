package router

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

type RouterContract struct {
	addr   common.Address
	master common.Address
}

func (cont *RouterContract) Address() common.Address {
	return cont.addr
}
func (cont *RouterContract) Master() common.Address {
	return cont.master
}
func (cont *RouterContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}
func (cont *RouterContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &RouterContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagFactory}, data.Factory[:])
	return nil
}
func (cont *RouterContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}
func (cont *RouterContract) factory(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagFactory})
	return common.BytesToAddress(bs)
}
func (cont *RouterContract) _uniAddLiquidity(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	amountADesired, amountBDesired, amountAMin, amountBMin *big.Int, mintCalculation bool) (common.Address, *big.Int, *big.Int, *big.Int, uint64, error) {

	if amountADesired.Cmp(Zero) <= 0 {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_A_AMOUNT")
	}
	if amountBDesired.Cmp(Zero) <= 0 {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_B_AMOUNT")
	}

	factory := cont.factory(cc)
	pair, reserveA, reserveB, err := uniGetReserves(cc, factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, nil, nil, nil, 0, err
	}

	var amountA, amountB *big.Int
	if reserveA.Cmp(Zero) == 0 && reserveB.Cmp(Zero) == 0 {
		amountA, amountB = amountADesired, amountBDesired
	} else {
		amountBOptimal, err := trade.UniQuote(amountADesired, reserveA, reserveB)
		if err != nil {
			return ZeroAddress, nil, nil, nil, 0, err
		}
		if amountBOptimal.Cmp(amountBDesired) <= 0 {
			if !(amountBOptimal.Cmp(amountBMin) >= 0) {
				return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_B_AMOUNT")
			}
			amountA, amountB = amountADesired, amountBOptimal
		} else {
			amountAOptimal, err := trade.UniQuote(amountBDesired, reserveB, reserveA)
			if err != nil {
				return ZeroAddress, nil, nil, nil, 0, err
			}
			if !(amountAOptimal.Cmp(amountADesired) <= 0) {
				return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_OUTPUT_AMOUNT")
			}
			if !(amountAOptimal.Cmp(amountAMin) >= 0) {
				return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_A_AMOUNT")
			}
			amountA, amountB = amountAOptimal, amountBDesired
		}
	}

	if mintCalculation {
		liquidity, ratio, err := uniGetMintAmount(cc, pair, amountA, amountB, reserveA, reserveB)
		if err != nil {
			return ZeroAddress, nil, nil, nil, 0, err
		}
		return pair, amountA, amountB, liquidity, ratio, nil
	}

	return pair, amountA, amountB, nil, 0, nil
}
func (cont *RouterContract) uniAddLiquidity(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	amountADesired, amountBDesired, amountAMin, amountBMin *big.Int,
	to common.Address) (*big.Int, *big.Int, *big.Int, common.Address, error) {

	pair, amountA, amountB, _, _, err := cont._uniAddLiquidity(cc, tokenA, tokenB, amountADesired, amountBDesired, amountAMin, amountBMin, false)
	if err != nil {
		return nil, nil, nil, ZeroAddress, err
	}

	if err := SafeTransferFrom(cc, tokenA, cc.From(), pair, amountA); err != nil {
		return nil, nil, nil, ZeroAddress, err
	}
	if err := SafeTransferFrom(cc, tokenB, cc.From(), pair, amountB); err != nil {
		return nil, nil, nil, ZeroAddress, err
	}

	is, err := cc.Exec(cc, pair, "Mint", []interface{}{to})
	if err != nil {
		return nil, nil, nil, ZeroAddress, err
	}
	liquidity := is[0].(*amount.Amount).Int

	return amountA, amountB, liquidity, pair, nil
}

func (cont *RouterContract) SetCubicRootType(cc *types.ContractContext, cubicRootType uint8) {
	cc.SetContractData([]byte{tagCubicRoot}, []byte{cubicRootType})
}

func (cont *RouterContract) CubicRootType(cc *types.ContractContext) (cubicRootType uint8) {
	bs := cc.ContractData([]byte{tagCubicRoot})
	if len(bs) != 0 {
		cubicRootType = uint8(bs[0])
	}
	return
}

// 결국 amountDesired 만큼 add 된다.
func (cont *RouterContract) uniGetLPTokenAmountOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB,
	tokenIn common.Address,
	amountDesired *big.Int, mintCalculation bool, from common.Address) (common.Address, *big.Int, *big.Int, *big.Int, uint64, error) {

	if tokenIn != tokenA && tokenIn != tokenB {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INPUT_TOKEN_NOT_MATCH")
	}

	if amountDesired.Cmp(Zero) <= 0 {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_AMOUNT")
	}

	factory := cont.factory(cc)
	pair, fee, reserveA, reserveB, err := uniGetFeeAndReserves(cc, factory, tokenA, tokenB, from)
	if err != nil {
		return ZeroAddress, nil, nil, nil, 0, err
	}

	if reserveA.Cmp(Zero) == 0 && reserveB.Cmp(Zero) == 0 {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: BOTH_RESERVE_0")
	}

	var amountA, amountB *big.Int
	var amountIn, amountOut *big.Int
	var addAmountA, addAmountB *big.Int

	cubicRootType := cont.CubicRootType(cc)
	// swap 양 계산
	if tokenIn == tokenA {
		amountA, amountB, err = trade.UniGetOptimalOneCoin(cubicRootType, fee, amountDesired, reserveA, reserveB)
		if err != nil {
			return ZeroAddress, nil, nil, nil, 0, err
		}
		amountIn = Clone(amountA)
		amountOut = Clone(amountB)
		addAmountA = Sub(amountDesired, amountA)
		addAmountB = amountB
		reserveA = Add(reserveA, amountIn)
		reserveB = Sub(reserveB, amountB)
	} else if tokenIn == tokenB {
		amountB, amountA, err = trade.UniGetOptimalOneCoin(cubicRootType, fee, amountDesired, reserveB, reserveA)
		if err != nil {
			return ZeroAddress, nil, nil, nil, 0, err
		}
		amountIn = Clone(amountB)
		amountOut = Clone(amountA)
		addAmountA = amountA
		addAmountB = Sub(amountDesired, amountB)
		reserveA = Sub(reserveA, amountA)
		reserveB = Add(reserveB, amountIn)
	} else {
		return ZeroAddress, nil, nil, nil, 0, errors.New("Router: TokenIn")
	}

	if mintCalculation == true { // mintFee 반영
		liquidity, ratio, err := uniGetMintAmount(cc, pair, addAmountA, addAmountB, reserveA, reserveB)
		if err != nil {
			return ZeroAddress, nil, nil, nil, 0, err
		}
		if liquidity.Cmp(Zero) <= 0 {
			return ZeroAddress, nil, nil, nil, 0, errors.New("Router: INSUFFICIENT_LIQUIDITY")
		}
		return pair, amountIn, amountOut, liquidity, ratio, nil
	}

	return pair, amountIn, amountOut, nil, 0, nil
}
func (cont *RouterContract) uniAddLiquidityOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB,
	tokenIn common.Address,
	amountDesired, amountAMin *big.Int,
	to, from common.Address) (*big.Int, *big.Int, common.Address, error) {

	if tokenIn != tokenA && tokenIn != tokenB {
		return nil, nil, ZeroAddress, errors.New("Router: INPUT_TOKEN_NOT_MATCH")
	}

	if amountDesired.Cmp(Zero) <= 0 {
		return nil, nil, ZeroAddress, errors.New("Router: INSUFFICIENT_AMOUNT")
	}

	// 먼저 tokenIn을 router로 가져온다
	if err := SafeTransferFrom(cc, tokenIn, cc.From(), cont.addr, amountDesired); err != nil {
		return nil, nil, ZeroAddress, err
	}

	pair, amountIn, amountOut, _, _, err := cont.uniGetLPTokenAmountOneCoin(cc, tokenA, tokenB, tokenIn, amountDesired, false, from)
	if err != nil {
		return nil, nil, ZeroAddress, err
	}

	var addAmountA, addAmountB *big.Int
	var amount0Out, amount1Out *big.Int
	if tokenIn == tokenA {
		addAmountA = Sub(amountDesired, amountIn)
		addAmountB = amountOut
	} else {
		addAmountA = amountOut
		addAmountB = Sub(amountDesired, amountIn)
	}
	// swap : 전송 : router -> pair
	if err := SafeTransfer(cc, tokenIn, pair, amountIn); err != nil {
		return nil, nil, ZeroAddress, err
	}
	token0, _, err := trade.SortTokens(tokenA, tokenB)
	if err != nil {
		return nil, nil, ZeroAddress, err
	}
	if tokenIn == token0 {
		amount0Out, amount1Out = big.NewInt(0), amountOut
	} else {
		amount0Out, amount1Out = amountOut, big.NewInt(0)
	}
	if _, err = cc.Exec(cc, pair, "Swap", []interface{}{amount0Out, amount1Out, cont.addr, []byte{}, from}); err != nil {
		return nil, nil, ZeroAddress, err
	}

	// addLiqudity
	if err := SafeTransfer(cc, tokenA, pair, addAmountA); err != nil {
		return nil, nil, ZeroAddress, err
	}
	if err := SafeTransfer(cc, tokenB, pair, addAmountB); err != nil {
		return nil, nil, ZeroAddress, err
	}
	is, err := cc.Exec(cc, pair, "Mint", []interface{}{to})
	if err != nil {
		return nil, nil, ZeroAddress, err
	}
	liquidity := is[0].(*amount.Amount).Int

	if liquidity.Cmp(Zero) <= 0 || liquidity.Cmp(amountAMin) < 0 {
		return nil, nil, ZeroAddress, errors.New("Router: INSUFFICIENT_LIQUIDITY")
	}
	return amountDesired, liquidity, pair, nil
}
func (cont *RouterContract) uniGetWithdrawAmount(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *big.Int) (common.Address, common.Address, *big.Int, *big.Int, *big.Int, error) {

	if liquidity.Cmp(Zero) <= 0 {
		return ZeroAddress, ZeroAddress, nil, nil, nil, errors.New("Router: INSUFFICIENT_LIQUIDITY")
	}

	factory := cont.factory(cc)
	pair, err := pairForUniSwap(cc, factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, nil, err
	}

	balanceA, err := TokenBalanceOf(cc, tokenA, pair)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, nil, err
	}
	balanceB, err := TokenBalanceOf(cc, tokenB, pair)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, nil, err
	}

	_totalSupply, err := TokenTotalSupply(cc, pair)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, nil, err
	}

	//_mintFee
	mintFee, err := GetMintAdminFee(cc, pair, balanceA, balanceB)
	GPrintln("mintFee, balanceA, balanceB", mintFee, balanceA, balanceB)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, nil, err
	}
	_totalSupply = Add(_totalSupply, mintFee)

	amountA := MulDiv(liquidity, balanceA, _totalSupply)
	amountB := MulDiv(liquidity, balanceB, _totalSupply)
	if !(amountA.Cmp(Zero) > 0 && amountB.Cmp(Zero) > 0) {
		return ZeroAddress, ZeroAddress, nil, nil, nil, errors.New("Router: INSUFFICIENT_LIQUIDITY_QUERIED")
	}

	return factory, pair, amountA, amountB, mintFee, nil
}
func (cont *RouterContract) uniRemoveLiquidity(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity, amountAMin, amountBMin *big.Int, to common.Address) (common.Address, common.Address, *big.Int, *big.Int, error) {

	if liquidity.Cmp(Zero) <= 0 {
		return ZeroAddress, ZeroAddress, nil, nil, errors.New("Router: INSUFFICIENT_LIQUIDITY")
	}

	factory := cont.factory(cc)
	pair, err := pairForUniSwap(cc, factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, err
	}

	// mintedAdminFee
	owner, err := Owner(cc, pair)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, err
	}
	if cc.From() == owner {
		ownerBalance, err := TokenBalanceOf(cc, pair, owner)
		if err != nil {
			return ZeroAddress, ZeroAddress, nil, nil, err
		}

		minted, err := MintedAdminBalance(cc, pair)
		if err != nil {
			return ZeroAddress, ZeroAddress, nil, nil, err
		}

		if liquidity.Cmp(Sub(ownerBalance, minted)) > 0 {
			return ZeroAddress, ZeroAddress, nil, nil, errors.New("Router: OWNER_LIQUIDITY")
		}
	}

	// cc.From() -> pair
	if err := SafeTransferFrom(cc, pair, cc.From(), pair, liquidity); err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, err
	}

	is, err := cc.Exec(cc, pair, "Burn", []interface{}{to})
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, err
	}
	amount0 := is[0].(*amount.Amount).Int
	amount1 := is[1].(*amount.Amount).Int
	token0, _, err := trade.SortTokens(tokenA, tokenB)
	if err != nil {
		return ZeroAddress, ZeroAddress, nil, nil, err
	}
	var amountA, amountB *big.Int
	if tokenA == token0 {
		amountA, amountB = amount0, amount1
	} else {
		amountA, amountB = amount1, amount0
	}
	if !(amountA.Cmp(amountAMin) >= 0) {
		return ZeroAddress, ZeroAddress, nil, nil, errors.New("Router: INSUFFICIENT_A_AMOUNT")
	}
	if !(amountB.Cmp(amountBMin) >= 0) {
		return ZeroAddress, ZeroAddress, nil, nil, errors.New("Router: INSUFFICIENT_B_AMOUNT")
	}

	return factory, pair, amountA, amountB, nil
}
func (cont *RouterContract) uniGetWithdrawAmountOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *big.Int,
	tokenOut, from common.Address) (*big.Int, *big.Int, error) {

	if tokenOut != tokenA && tokenOut != tokenB {
		return nil, nil, errors.New("Router: OUTPUT_TOKEN_NOT_MATCH")
	}

	if liquidity.Cmp(Zero) <= 0 {
		return nil, nil, errors.New("Router: INSUFFICIENT_LIQUIDITY")
	}

	factory, _, amountA, amountB, mintFee, err := cont.uniGetWithdrawAmount(cc, tokenA, tokenB, liquidity)
	if err != nil {
		return nil, nil, err
	}

	// reserve 변동
	_, fee, reserveA, reserveB, err := uniGetFeeAndReserves(cc, factory, tokenA, tokenB, from)
	if err != nil {
		return nil, nil, err
	}

	reserveA = Sub(reserveA, amountA)
	reserveB = Sub(reserveB, amountB)

	var amountOut *big.Int
	if tokenOut == tokenA {
		amountOut, err = trade.UniGetAmountOut(fee, amountB, reserveB, reserveA)
		if err != nil {
			return nil, nil, err
		}
		amountOut = Add(amountOut, amountA)
	} else {
		amountOut, err = trade.UniGetAmountOut(fee, amountA, reserveA, reserveB)
		if err != nil {
			return nil, nil, err
		}
		amountOut = Add(amountOut, amountB)
	}
	return amountOut, mintFee, nil
}
func (cont *RouterContract) uniRemoveLiquidityOneCoin(
	cc *types.ContractContext,
	tokenA, tokenB common.Address,
	liquidity *big.Int,
	tokenOut common.Address,
	amountMin *big.Int,
	to, from common.Address) (*big.Int, error) {

	if tokenOut != tokenA && tokenOut != tokenB {
		return nil, errors.New("Router: OUTPUT_TOKEN_NOT_MATCH")
	}

	factory, pair, amountA, amountB, err := cont.uniRemoveLiquidity(cc, tokenA, tokenB, liquidity, Zero, Zero, cont.Address())
	if err != nil {
		return nil, err
	}
	var amountOut *big.Int
	if tokenOut == tokenA {
		path := []common.Address{tokenB, tokenA}
		amounts, err := getAmountsOut(cc, factory, amountB, path, from)
		if err != nil {
			return nil, err
		}
		if err := SafeTransfer(cc, tokenB, pair, amountB); err != nil {
			return nil, err
		}
		if err := cont._uniSwap(cc, amounts, path, cont.Address(), from); err != nil {
			return nil, err
		}
		amountOut = Add(amounts[1], amountA)
	} else {
		path := []common.Address{tokenA, tokenB}
		amounts, err := getAmountsOut(cc, factory, amountA, path, from)
		if err != nil {
			return nil, err
		}
		if err := SafeTransfer(cc, tokenA, pair, amountA); err != nil {
			return nil, err
		}
		if err := cont._uniSwap(cc, amounts, path, cont.Address(), from); err != nil {
			return nil, err
		}
		amountOut = Add(amounts[1], amountB)
	}

	if !(amountOut.Cmp(amountMin) >= 0) {
		return nil, errors.New("Router: INSUFFICIENT_OUTPUT_AMOUNT")
	}

	if err := SafeTransfer(cc, tokenOut, to, amountOut); err != nil {
		return nil, err
	}

	return amountOut, nil
}
func (cont *RouterContract) _uniSwap(cc *types.ContractContext, amounts []*big.Int, path []common.Address, _to, _from common.Address) error {

	for i := int(0); i < len(path)-1; i++ {
		factory := cont.factory(cc)
		input, output := path[i], path[i+1]
		var to common.Address
		if i < len(path)-2 {
			var err error
			to, err = pairForUniSwap(cc, factory, output, path[i+2])
			if err != nil {
				return err
			}
		} else {
			to = _to
		}
		pair, err := pairForUniSwap(cc, factory, input, output)
		if err != nil {
			return err
		}

		token0, _, err := trade.SortTokens(input, output)
		if err != nil {
			return err
		}
		amountOut := amounts[i+1]
		var amount0Out, amount1Out *big.Int
		if input == token0 {
			amount0Out, amount1Out = big.NewInt(0), amountOut
		} else {
			amount0Out, amount1Out = amountOut, big.NewInt(0)
		}

		_, err = cc.Exec(cc, pair, "Swap", []interface{}{amount0Out, amount1Out, to, []byte{}, _from})
		if err != nil {
			return err
		}
	}
	return nil
}
func (cont *RouterContract) swapExactTokensForTokens(
	cc *types.ContractContext,
	amountIn, amountOutMin *big.Int,
	path []common.Address,
	to, from common.Address) ([]*big.Int, error) {

	if amountIn.Cmp(Zero) <= 0 {
		return nil, errors.New("Router: INSUFFICIENT_SWAP_AMOUNT")
	}

	factory := cont.factory(cc)
	amounts, err := getAmountsOut(cc, factory, amountIn, path, from)
	if err != nil {
		return nil, err
	}
	if !(amounts[len(amounts)-1].Cmp(amountOutMin) >= 0) {
		return nil, errors.New("Router: INSUFFICIENT_OUTPUT_AMOUNT")
	}

	allUni := true
	for i := 0; i < len(path)-1; i++ {
		pair, err := trade.PairFor(factory, path[i], path[i+1])
		if err != nil {
			return nil, err
		}
		exType, err := ExType(cc, pair)
		if err != nil {
			return nil, err
		}
		if exType == trade.STABLE {
			allUni = false
			break
		}
	}
	if allUni {
		pair, err := trade.PairFor(factory, path[0], path[1])
		if err != nil {
			return nil, err
		}
		if err := SafeTransferFrom(cc, path[0], cc.From(), pair, amounts[0]); err != nil {
			return nil, err
		}
		if err := cont._uniSwap(cc, amounts, path, to, from); err != nil {
			return nil, err
		}
		return amounts, nil
	}

	for i := 0; i < len(path)-1; i++ {
		pair, err := trade.PairFor(factory, path[i], path[i+1])
		if err != nil {
			return nil, err
		}
		exType, err := ExType(cc, pair)
		if err != nil {
			return nil, err
		}
		if exType == trade.UNI {
			if i == 0 {
				if err := SafeTransferFrom(cc, path[i], cc.From(), pair, amounts[i]); err != nil {
					return nil, err
				}
			} else {
				if err := SafeTransfer(cc, path[i], pair, amounts[i]); err != nil {
					return nil, err
				}
			}
			if err := cont._uniSwap(cc, []*big.Int{amounts[i], amounts[i+1]}, []common.Address{path[i], path[i+1]}, cont.addr, from); err != nil {
				return nil, err
			}
		} else {
			in, err := TokenIndex(cc, pair, path[i])
			if err != nil {
				return nil, err
			}
			out, err := TokenIndex(cc, pair, path[i+1])
			if err != nil {
				return nil, err
			}
			if i == 0 {
				if err := SafeTransferFrom(cc, path[i], cc.From(), cont.addr, amounts[i]); err != nil {
					return nil, err
				}
			}
			if err := TokenApprove(cc, path[i], pair, amounts[i]); err != nil {
				return nil, err
			}
			if _, err := cc.Exec(cc, pair, "Exchange", []interface{}{in, out, amounts[i], ZeroAmount, from}); err != nil {
				return nil, err
			}
		}
	}
	k := len(path) - 1 //lastIndex
	if err := SafeTransfer(cc, path[k], to, amounts[k]); err != nil {
		return nil, err
	}

	return amounts, nil
}

func (cont *RouterContract) uniSwapTokensForExactTokens(
	cc *types.ContractContext,
	amountOut, amountInMax *big.Int,
	path []common.Address,
	to, from common.Address) ([]*big.Int, error) {

	if amountOut.Cmp(Zero) <= 0 {
		return nil, errors.New("Router: INSUFFICIENT_SWAP_AMOUNT")
	}

	factory := cont.factory(cc)
	amounts, err := uniGetAmountsIn(cc, factory, amountOut, path, from)
	if err != nil {
		return nil, err
	}
	if !(amounts[0].Cmp(amountInMax) <= 0) {
		return nil, errors.New("Router: EXCESSIVE_INPUT_AMOUNT")
	}

	pair, err := trade.PairFor(factory, path[0], path[1])
	if err != nil {
		return nil, err
	}
	if err := SafeTransferFrom(cc, path[0], cc.From(), pair, amounts[0]); err != nil {
		return nil, err
	}
	if err := cont._uniSwap(cc, amounts, path, to, from); err != nil {
		return nil, err
	}
	return amounts, nil
}
