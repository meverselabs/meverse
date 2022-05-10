package trade

import (
	"math"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var adminAddress common.Address

// exchange errors
var (
	_ErrIdenticalAddresses = errors.New("Exchange: IDENTICAL_ADDRESSES")
	_ErrZeroAddress        = errors.New("Exchange: ZERO_ADDRESS")
)

var _ = Describe("util", func() {

	It("ZeroAddress == common.Address", func() {
		var addr common.Address
		GPrintln("var addr common.Address == ZeroAddress", addr == ZeroAddress)
	})

	It("common.BytesToAddress(nil)", func() {
		addr := common.BytesToAddress(nil)
		GPrintln("common.BytesToAddress(nil)", addr)

		addr2 := common.BytesToAddress([]byte{})
		GPrintln("common.BytesToAddress(nil)", addr2)

	})

	It("float64 to big.Int", func() {
		big.NewInt(0)
		f := math.MaxFloat64
		GPrintln("math.MaxFloat64", math.MaxFloat64)
		r, err := ToBigInt(f)
		Expect(err).To(Succeed())
		GPrintln("ToBigInt(", f, ")", r)
	})

	It("address To bytes", func() {
		token, _ := tokenContractAddress(adminAddress, uint32(2))
		GPrintln(token[:])
	})

	It("ClassID", func() {
		ClassID, err := types.RegisterContractType(&token.TokenContract{})
		Expect(err).To(Succeed())
		GPrintln("ClassID", ClassID)
	})

	It("create tokenContractAddress", func() {
		for i := 1; i <= 10; i++ {
			_, err := tokenContractAddress(adminAddress, uint32(i))
			Expect(err).To(Succeed())
		}
	})

	It("SortTokens with TokenA Zero", func() {

		tokenA := ZeroAddress
		tokenB, err := tokenContractAddress(adminAddress, uint32(2))
		Expect(err).To(Succeed())

		token0, token1, err := SortTokens(tokenA, tokenB)
		GPrintf("token0 : %v, token1 %v\n", token0, token1)
		Expect(err).To(MatchError(_ErrZeroAddress.Error()))

	})

	It("SortTokens with TokenB Zero", func() {
		tokenA, err := tokenContractAddress(adminAddress, uint32(1))
		Expect(err).To(Succeed())

		tokenB := ZeroAddress

		GPrintf("tokenA : %v, tokenB %v\n", tokenA, tokenB)

		token0, token1, err := SortTokens(tokenA, tokenB)
		GPrintf("token0 : %v, token1 %v\n", token0, token1)

		Expect(err).To(MatchError(_ErrZeroAddress.Error()))
	})

	It("SortTokens with same addresses", func() {
		tokenA, _ := tokenContractAddress(adminAddress, uint32(1))
		tokenB, _ := tokenContractAddress(adminAddress, uint32(1))

		token0, token1, err := SortTokens(tokenA, tokenB)
		GPrintf(`token0 : %v, token1 : %v`, token0, token1)

		Expect(err).To(MatchError(_ErrIdenticalAddresses.Error()))
	})

	It("PairFor", func() {
		factory, _ := tokenContractAddress(adminAddress, uint32(0))
		tokenA, _ := tokenContractAddress(adminAddress, uint32(1))
		tokenB, _ := tokenContractAddress(adminAddress, uint32(2))

		couple, err := PairFor(factory, tokenA, tokenB)

		GPrintln(couple)
		if err != nil || couple.String() != "0xD69d8C6765BDF2E08496b229AE83322D50872B35" {
			Fail("error creating pair contract creation address")
		}
	})

	It("UniGetAmountOut", func() {
		swapAmount := amount.NewAmount(1, 0)
		reserveIn := amount.NewAmount(1000, 0)
		reserveOut := amount.NewAmount(1000, 0)
		am00, _ := UniGetAmountOut(uint64(0), swapAmount.Int, reserveIn.Int, reserveOut.Int)          // 0%
		am30, _ := UniGetAmountOut(uint64(30000000), swapAmount.Int, reserveIn.Int, reserveOut.Int)   // 0.3%
		am100, _ := UniGetAmountOut(uint64(100000000), swapAmount.Int, reserveIn.Int, reserveOut.Int) // 1%

		GPrintln("am00, am30, am100", am00, am30, am100)
	})
})
