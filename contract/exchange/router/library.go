package router

import (
	"math/big"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

// fetches and sorts the reserves for a pair
func uniGetReserves(cc *types.ContractContext, factory, tokenA, tokenB common.Address) (common.Address, *big.Int, *big.Int, error) {
	token0, _, _ := trade.SortTokens(tokenA, tokenB)
	pair, err := trade.PairFor(factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, nil, nil, err
	}

	is, err := cc.Exec(cc, pair, "Reserves", []interface{}{})
	if err != nil {
		return ZeroAddress, nil, nil, err
	}
	reserve0 := is[0].([]*amount.Amount)[0]
	reserve1 := is[0].([]*amount.Amount)[1]

	if tokenA.String() == token0.String() {
		return pair, reserve0.Int, reserve1.Int, nil
	}
	return pair, reserve1.Int, reserve0.Int, nil
}

func pairForUniSwap(cc *types.ContractContext, factory, tokenA, tokenB common.Address) (common.Address, error) {
	pair, err := trade.PairFor(factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, err
	}
	if err := onlyUniSwap(cc, pair); err != nil {
		return ZeroAddress, err
	}

	return pair, nil
}

// fetches and sorts the reserves for a pair
func uniGetFeeAndReserves(cc *types.ContractContext, factory, tokenA, tokenB, from common.Address) (common.Address, uint64, *big.Int, *big.Int, error) {
	pair, reserveA, reserveB, err := uniGetReserves(cc, factory, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, uint64(0), nil, nil, err
	}
	fee, err := FeeAddress(cc, pair, from)
	if err != nil {
		return ZeroAddress, uint64(0), nil, nil, err
	}
	return pair, fee, reserveA, reserveB, err
}

// mintFee 뱐영
func uniGetMintAmount(cc *types.ContractContext, _pair common.Address, amountA, amountB, _reserveA, _reserveB *big.Int) (*big.Int, uint64, error) {
	mintFee, err := GetMintAdminFee(cc, _pair, _reserveA, _reserveB)
	if err != nil {
		return nil, 0, err
	}

	liquidity := big.NewInt(0)
	ratio := uint64(0)
	_totalSupply, err := TokenTotalSupply(cc, _pair)
	if err != nil {
		return nil, 0, err
	}
	_totalSupply = Add(_totalSupply, mintFee)
	if _totalSupply.Cmp(Zero) > 0 {
		liquidity = Min(MulDiv(amountA, _totalSupply, _reserveA), MulDiv(amountB, _totalSupply, _reserveB))
		ratio = uint64(MulDiv(liquidity, big.NewInt(amount.FractionalMax), Add(_totalSupply, liquidity)).Int64()) // 발행량까지 포함
	}
	return liquidity, ratio, nil
}

// performs chained getAmountOut calculations on any number of pairs
func getAmountsOut(cc *types.ContractContext, factory common.Address, amountIn *big.Int, path []common.Address, from common.Address) ([]*big.Int, error) {
	if len(path) < 2 {
		return nil, errors.New("Router: INVALID_PATH")
	}
	if amountIn.Cmp(Zero) <= 0 {
		return nil, errors.New("Router: INSUFFICIENT_IN_AMOUNT")
	}

	amounts := make([]*big.Int, len(path))
	amounts[0] = amountIn
	for i := int(0); i < len(path)-1; i++ {
		pair, err := trade.PairFor(factory, path[i], path[i+1])
		if err != nil {
			return nil, err
		}
		exType, err := ExType(cc, pair)
		if err != nil {
			return nil, err
		}
		if exType == trade.UNI {
			_, fee, reserveIn, reserveOut, err := uniGetFeeAndReserves(cc, factory, path[i], path[i+1], from)
			if err != nil {
				return nil, err
			}
			am, err := trade.UniGetAmountOut(fee, amounts[i], reserveIn, reserveOut)
			if err != nil {
				return nil, err
			}
			amounts[i+1] = am
		} else if exType == trade.STABLE {
			in, err := TokenIndex(cc, pair, path[i])
			if err != nil {
				return nil, err
			}
			out, err := TokenIndex(cc, pair, path[i+1])
			if err != nil {
				return nil, err
			}
			amounts[i+1], err = GetDy(cc, pair, in, out, amounts[i], from)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("Exchange: INVALID_EXTYPE")
		}
	}
	return amounts, nil
}

// performs chained getAmountIn calculations on any number of pairs
func uniGetAmountsIn(cc *types.ContractContext, factory common.Address, amountOut *big.Int, path []common.Address, from common.Address) ([]*big.Int, error) {
	if len(path) < 2 {
		return nil, errors.New("Router: INVALID_PATH")
	}

	if amountOut.Cmp(Zero) <= 0 {
		return nil, errors.New("Router: INSUFFICIENT_OUT_AMOUNT")
	}

	amounts := make([]*big.Int, len(path))
	amounts[len(amounts)-1] = amountOut
	for i := len(path) - 1; i > 0; i-- {
		_, fee, reserveIn, reserveOut, err := uniGetFeeAndReserves(cc, factory, path[i-1], path[i], from)
		if err != nil {
			return nil, err
		}
		am, err := trade.UniGetAmountIn(fee, amounts[i], reserveIn, reserveOut)
		if err != nil {
			return nil, err
		}
		amounts[i-1] = am
	}

	return amounts, nil
}
func onlyUniSwap(cc *types.ContractContext, _pair common.Address) error {
	exType, err := ExType(cc, _pair)
	if err != nil {
		return err
	}
	if exType != trade.UNI {
		return errors.New("Router: ONLY_UNISWAP")
	}
	return nil
}
