package trade

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"reflect"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

const (
	//exchangeType uint8
	UNI    = uint8(1)
	STABLE = uint8(2)

	//exchange
	FEE_DENOMINATOR = 10000000000
	PRECISION       = amount.FractionalMax
	MAX_FEE         = FEE_DENOMINATOR / 2 //  50%
	MAX_ADMIN_FEE   = FEE_DENOMINATOR     // 100%
	MAX_WINNER_FEE  = FEE_DENOMINATOR     // 100%

	//uniswap
	MINIMUM_LIQUIDITY = 1000

	//stableswap
	MAX_A         = 1000000
	MAX_A_CHANGE  = 10
	MIN_RAMP_TIME = 86400 // 1 day
	A_PRECISION   = 100
)

var (
	//owner
	tagOwner = byte(0x00)

	//token
	tagTokenName        = byte(0x01)
	tagTokenSymbol      = byte(0x02)
	tagTokenTotalSupply = byte(0x03)
	tagTokenAmount      = byte(0x04)
	tagTokenApprove     = byte(0x05)

	//exchange
	tagExType                        = byte(0x21)
	tagFactory                       = byte(0x22)
	tagExNTokens                     = byte(0x23)
	tagExTokens                      = byte(0x24)
	tagExPayToken                    = byte(0x25)
	tagExWinner                      = byte(0x26)
	tagExFee                         = byte(0x27)
	tagExAdminFee                    = byte(0x28)
	tagExWinnerFee                   = byte(0x29)
	tagExWhiteList                   = byte(0x30)
	tagExGroupId                     = byte(0x31)
	tagExAdminActionsDeadline        = byte(0x32)
	tagExFutureFee                   = byte(0x33)
	tagExFutureAdminFee              = byte(0x34)
	tagExFutureWinnerFee             = byte(0x35)
	tagExTransferOwnerWinnerDeadline = byte(0x36)
	tagExFutureOwner                 = byte(0x37)
	tagExFutureWinner                = byte(0x38)
	tagExWhiteListDeadline           = byte(0x39)
	tagExFutureWhiteList             = byte(0x40)
	tagExFutureGroupId               = byte(0x41)
	tagExIsKilled                    = byte(0x42)
	tagBlockTimestampLast            = byte(0x43)

	//UniSwap
	tagUniToken0               = byte(0x61)
	tagUniToken1               = byte(0x62)
	tagUniReserve0             = byte(0x63)
	tagUniReserve1             = byte(0x64)
	tagUniPrice0CumulativeLast = byte(0x65)
	tagUniPrice1CumulativeLast = byte(0x66)
	tagUniKLast                = byte(0x67)
	tagUniAdminBalance         = byte(0x68)

	//StableSwap
	tagStablePrecisionMul   = byte(0x81)
	tagStableRates          = byte(0x82)
	tagStableReserves       = byte(0x83)
	tagStableInitialAmp     = byte(0x84)
	tagStableFutureAmp      = byte(0x85)
	tagStableInitialAmpTime = byte(0x86)
	tagStableFutureAmpTime  = byte(0x87)
)

/////////// key  ////////////
func makeTokenKey(sender common.Address, key byte) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = key
	copy(bs[1:], sender[:])
	return bs
}

/////////// Contract Address  ///////////
// Class ID
func contractClassID(cont types.Contract) uint64 {
	rt := reflect.TypeOf(cont)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		pkgPath = strings.Replace(pkgPath, "meverselabs/meverse", "fletaio/fleta_v2", -1)
		name = pkgPath + "." + name
	}
	h := hash.Hash([]byte(name))
	ClassID := bin.Uint64(h[len(h)-8:])

	return ClassID
}

// Pair Contract Address
func pairContractAddress(ClassID uint64, factory, token0, token1 common.Address) common.Address {

	base := make([]byte, 1+common.AddressLength*3+8)
	base[0] = 0xff

	copy(base[1:], factory[:])
	copy(base[1+common.AddressLength:], token0[:])
	copy(base[1+common.AddressLength*2:], token1[:])
	copy(base[1+common.AddressLength*3:], bin.Uint64Bytes(ClassID))
	h := hash.Hash(base)
	addr := common.BytesToAddress(h[12:])
	return addr
}

/////////// Exchange  ///////////
// Sort two Token Address
func SortTokens(tokenA, tokenB common.Address) (common.Address, common.Address, error) {
	if tokenA == tokenB {
		return ZeroAddress, ZeroAddress, errors.New("Exchange: IDENTICAL_ADDRESSES")
	}

	var token0, token1 common.Address
	if tokenA.String() < tokenB.String() {
		token0, token1 = tokenA, tokenB
	} else {
		token0, token1 = tokenB, tokenA
	}

	if token0 == ZeroAddress {
		return ZeroAddress, ZeroAddress, errors.New("Exchange: ZERO_ADDRESS")
	}

	return token0, token1, nil
}

//calculates the address for a pair without making any external calls
func PairFor(factory, tokenA, tokenB common.Address) (common.Address, error) {

	token0, token1, err := SortTokens(tokenA, tokenB)
	if err != nil {
		return ZeroAddress, err
	}
	//아래부분을 고쳐야 한다.
	//types/contract_creator.go의 ClassID만 가져오는 함수를 만들어야 한다.
	//ClassID, err := types.RegisterContractType(&PairContract{})
	ClassID := contractClassID(&UniSwap{})

	return pairContractAddress(ClassID, factory, token0, token1), nil
}

/////////// Uniswap  ///////////
// given some amount of an asset and pair reserves, returns an equivalent amount of the other asset
func UniQuote(amountA, reserveA, reserveB *big.Int) (*big.Int, error) {
	if !(amountA.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_AMOUNT")
	}
	if !(reserveA.Cmp(Zero) > 0) || !(reserveB.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
	}

	return MulDiv(amountA, reserveB, reserveA), nil
}

// given an input amount of an asset and pair reserves, returns the maximum output amount of the other asset
func UniGetAmountOut(fee uint64, amountIn, reserveIn, reserveOut *big.Int) (*big.Int, error) {
	if !(amountIn.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_INPUT_AMOUNT")
	}
	if !(reserveIn.Cmp(Zero) > 0) || !(reserveOut.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
	}
	amountInWithFee := Mul(amountIn, big.NewInt(FEE_DENOMINATOR-int64(fee)))
	numerator := Mul(amountInWithFee, reserveOut)
	denominator := Add(Mul(reserveIn, big.NewInt(FEE_DENOMINATOR)), amountInWithFee)
	amountOut := Div(numerator, denominator)

	return amountOut, nil
}

// UniGetOptimalOneCoin에서 Call
// (1-f)x^3 + 3(1-f)x_0*x^2 + ((2-f)x_0 - (1-f)A)*x_0*x - A*x_0^2
func optimalCubicRoot(cubicRootType uint8, fee uint64, amountIn *big.Int, reserve *big.Int) (float64, error) {
	f := float64(fee) / float64(FEE_DENOMINATOR)
	A := ToFloat64(amountIn)
	x_0 := ToFloat64(reserve)
	a := 1. - f
	b := 3. * (1 - f) * x_0
	c := ((2.-f)*x_0 - (1.-f)*A) * x_0
	d := -A * x_0 * x_0
	var r1 float64
	var err error
	if cubicRootType == 0 {
		r1, err = CubicRoot(a, b, c, d)
	} else if cubicRootType == 1 {
		r1, err = Newtonian(a, b, c, d, x_0)
	} else {
		return .0, errors.New("undefined cubicRootType")
	}
	if err != nil {
		return 0., err
	}
	if r1 < 0. {
		return 0., errors.New("CUBICROOT: TOO_LARGE_COFFICIENT")
	}
	return r1, nil
}

// router.UniAddLiquidityOneCoin의 경우 swap 한후, router.UniAddLiquidity에 배분해서 넣어야 함.
func UniGetOptimalOneCoin(cubicRootType uint8, fee uint64, onecoinAmountIn, reserveIn, reserveOut *big.Int) (*big.Int, *big.Int, error) {
	if !(reserveIn.Cmp(Zero) > 0) || !(reserveOut.Cmp(Zero) > 0) {
		return nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
	}
	if fee == FEE_DENOMINATOR {
		return nil, nil, errors.New("Exchange: FEE_100%")
	}
	var amountIn *big.Int
	if amountIn = checkComplexCalc(cubicRootType, fee, onecoinAmountIn, reserveIn); amountIn == nil {
		floatIn, err := optimalCubicRoot(cubicRootType, fee, onecoinAmountIn, reserveIn)
		if err != nil {
			return nil, nil, err
		}
		amountIn, err = ToBigInt(floatIn)
		if err != nil {
			return nil, nil, err
		}
	}
	if amountIn.Cmp(reserveIn) >= 0 {
		return nil, nil, errors.New("Exchange: TOO_MUCH_INPUT")
	}
	amountOut, err := UniGetAmountOut(fee, amountIn, reserveIn, reserveOut)
	if err != nil {
		return nil, nil, err
	}

	return amountIn, amountOut, nil
}

// given an output amount of an asset and pair reserves, returns a required input amount of the other asset
func UniGetAmountIn(fee uint64, amountOut, reserveIn, reserveOut *big.Int) (*big.Int, error) {
	if !(amountOut.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_OUTPUT_AMOUNT")
	}
	if !(reserveIn.Cmp(Zero) > 0) || !(reserveOut.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
	}

	numerator := Mul(Mul(reserveIn, amountOut), big.NewInt(FEE_DENOMINATOR))
	denominator := Mul(Sub(reserveOut, amountOut), big.NewInt(FEE_DENOMINATOR-int64(fee)))
	amountIn := Add(Div(numerator, denominator), big.NewInt(1))
	return amountIn, nil
}

func TestOptimalCubicRoot(fee uint64, amountIn *big.Int, reserve *big.Int) (float64, error) {
	f := float64(fee) / float64(FEE_DENOMINATOR)
	A := ToFloat64(amountIn)
	x_0 := ToFloat64(reserve)
	a := 1. - f
	b := 3. * (1 - f) * x_0
	c := ((2.-f)*x_0 - (1.-f)*A) * x_0
	d := -A * x_0 * x_0
	r1, err := CubicRoot(a, b, c, d)
	if err != nil {
		return 0., err
	}
	if r1 < 0. {
		return r1, errors.New("CUBICROOT: TOO_LARGE_COFFICIENT")
	}
	return r1, nil
}
func TestOptimalCubicRoot2(fee uint64, amountIn *big.Int, reserve *big.Int) (float64, error) {
	f := float64(fee) / float64(FEE_DENOMINATOR)
	A := ToFloat64(amountIn)
	x_0 := ToFloat64(reserve)
	a := 1. - f
	b := 3. * (1 - f) * x_0
	c := ((2.-f)*x_0 - (1.-f)*A) * x_0
	d := -A * x_0 * x_0
	r1, err := Newtonian(a, b, c, d, x_0)
	if err != nil {
		return 0., err
	}
	if r1 < 0. {
		return r1, errors.New("CUBICROOT: TOO_LARGE_COFFICIENT")
	}
	return r1, nil
}

func checkComplexCalc(cubicRootType uint8, fee uint64, onecoinAmountIn, reserveIn *big.Int) *big.Int {
	if cubicRootType != 0 {
		return nil
	}
	key := fmt.Sprintf("%v:%v:%v", fee, onecoinAmountIn.String(), reserveIn.String())
	return mustParseBigInt(cmMap[key])
}

func mustParseBigInt(str string) *big.Int {
	bi, _ := big.NewInt(0).SetString(str, 10)
	return bi
}
