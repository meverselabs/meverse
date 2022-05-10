package test

import (
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwapExactTokensForTokens", func() {

	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
	)

	It("0->1->2", func() {
		beforeEachWithoutTokens()

		tokens := DeployTokens(genesis, classMap["Token"], 3, admin)

		is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{tokens[0], tokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(Succeed())
		pair = is[0].(common.Address)

		is, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{tokens[1], tokens[2], ZeroAddress, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
		Expect(err).To(Succeed())
		swap = is[0].(common.Address)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
		ctx, _ = setFees(cn, ctx, swap, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), bobKey)

		tokenMint(ctx, tokens[0], alice, _SupplyTokens[0])
		tokenMint(ctx, tokens[1], alice, _SupplyTokens[1])
		tokenApprove(ctx, tokens[0], alice, routerAddr)
		tokenApprove(ctx, tokens[1], alice, routerAddr)
		_, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{tokens[0], tokens[1], _SupplyTokens[0], _SupplyTokens[1], ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())

		tokenMint(ctx, tokens[1], bob, _SupplyTokens[0])
		tokenMint(ctx, tokens[2], bob, _SupplyTokens[1])
		tokenApprove(ctx, tokens[1], bob, swap)
		tokenApprove(ctx, tokens[2], bob, swap)
		_, err = Exec(ctx, bob, swap, "AddLiquidity", []interface{}{_SupplyTokens, amount.NewAmount(0, 0)})
		Expect(err).To(Succeed())

		swapAmount := amount.NewAmount(1, 0)
		expectedOutputAmount1 := amount.NewAmount(0, 1993996023971928199)
		expectedOutputAmount2 := amount.NewAmount(0, 1990340394780245684)

		err = tokenMint(ctx, tokens[0], charlie, swapAmount)
		Expect(err).To(Succeed())

		err = tokenApprove(ctx, tokens[0], charlie, routerAddr)
		Expect(err).To(Succeed())
		is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1], tokens[2]}})
		Expect(err).To(Succeed())
		amounts := is[0].([]*amount.Amount)
		Expect(amounts[0]).To(Equal(swapAmount))
		Expect(amounts[1]).To(Equal(expectedOutputAmount1))
		Expect(amounts[2]).To(Equal(expectedOutputAmount2))
		Expect(tokenBalanceOf(ctx, tokens[0], charlie)).To(Equal(ZeroAmount))
		Expect(tokenBalanceOf(ctx, tokens[1], charlie)).To(Equal(ZeroAmount))
		Expect(tokenBalanceOf(ctx, tokens[2], charlie)).To(Equal(expectedOutputAmount2))

		RemoveChain(cdx)
		afterEach()
	})

	It("Random", func() {

		for k := 0; k < 10; k++ {
			for length := 6; length < 7; length++ {
				pairs := make([]common.Address, length-1)

				beforeEachWithoutTokens()
				tokens := DeployTokens(genesis, classMap["Token"], uint8(length), admin)
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

				swapAmount := amount.NewAmount(1, 0)
				tokenMint(ctx, tokens[0], charlie, swapAmount)

				for i := 0; i < length-1; i++ {
					switch rand.Intn(2) {
					//switch i % 2 {
					//switch 1 {
					case 0:
						is, err := Exec(ctx, admin, factoryAddr, "CreatePairUni", []interface{}{tokens[i], tokens[i+1], ZeroAddress, _PairName, _PairSymbol, admin, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
						Expect(err).To(Succeed())
						pairs[i] = is[0].(common.Address)
						tokenMint(ctx, tokens[i], alice, _SupplyTokens[0])
						tokenMint(ctx, tokens[i+1], alice, _SupplyTokens[1])
						tokenApprove(ctx, tokens[i], alice, routerAddr)
						tokenApprove(ctx, tokens[i+1], alice, routerAddr)
						_, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{tokens[i], tokens[i+1], _SupplyTokens[0], _SupplyTokens[1], ZeroAmount, ZeroAmount})
						Expect(err).To(Succeed())
						//GPrintln("U")

					case 1:
						is, err := Exec(ctx, admin, factoryAddr, "CreatePairStable", []interface{}{tokens[i], tokens[i+1], ZeroAddress, _SwapName, _SwapSymbol, admin, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
						Expect(err).To(Succeed())
						pairs[i] = is[0].(common.Address)
						tokenMint(ctx, tokens[i], bob, _SupplyTokens[0])
						tokenMint(ctx, tokens[i+1], bob, _SupplyTokens[1])
						tokenApprove(ctx, tokens[i], bob, pairs[i])
						tokenApprove(ctx, tokens[i+1], bob, pairs[i])
						_, err = Exec(ctx, bob, pairs[i], "AddLiquidity", []interface{}{_SupplyTokens, amount.NewAmount(0, 0)})
						Expect(err).To(Succeed())
						//GPrintln("S")
					}
					ctx, _ = setFees(cn, ctx, pairs[i], _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
				}

				tokenApprove(ctx, tokens[0], charlie, routerAddr)
				is, err := Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, tokens})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				for j := 0; j < length-1; j++ {
					Expect(tokenBalanceOf(ctx, tokens[j], charlie)).To(Equal(ZeroAmount))
				}
				Expect(tokenBalanceOf(ctx, tokens[length-1], charlie)).To(Equal(amounts[length-1]))

				RemoveChain(cdx)
				afterEach()
			}
		}
	})
})
