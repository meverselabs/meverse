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

/*
	contract : curve-contract
	files :
		tests/pools/common/integration/test_curve.py
		tests/pools/common/integration/test_heavily_imbalanced.py
		tests/pools/common/integration/test_virtual_price_increases.py
*/

var _ = Describe("Uni Integration", func() {

	var (
		cdx int
		cn  *chain.Chain
		ctx *types.Context
	)

	It("random", func() {
		// eve : whitelist

		beforeEach()
		is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], uniTokens[0], _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		pair = is[0].(common.Address)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		uniAddInitialLiquidity(ctx, alice)

		uniMint(ctx, alice)
		uniApprove(ctx, alice)
		uniMint(ctx, bob)
		uniApprove(ctx, bob)
		uniMint(ctx, charlie)
		uniApprove(ctx, charlie)
		uniMint(ctx, eve)
		uniApprove(ctx, eve)

		tokenApprove(ctx, pair, alice, routerAddr)
		tokenApprove(ctx, pair, bob, routerAddr)
		tokenApprove(ctx, pair, charlie, routerAddr)
		tokenApprove(ctx, pair, eve, routerAddr)

		for k := 0; k < 1000; k++ {
			var who common.Address
			switch rand.Intn(4) {
			case 0:
				who = alice
			case 1:
				who = bob
			case 2:
				who = charlie
			case 3:
				who = eve
			}

			switch rand.Intn(7) {
			case 0:
				token0Amount, _ := ToBigInt(float64(rand.Uint64()))
				token1Amount, _ := ToBigInt(float64(rand.Uint64()))
				balance0, _ := tokenBalanceOf(ctx, token0, who)
				balance1, _ := tokenBalanceOf(ctx, token1, who)
				if balance0.Cmp(token0Amount) > 0 && balance1.Cmp(token1Amount) > 0 {
					_, err = Exec(ctx, who, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token1Amount, ZeroAmount, ZeroAmount})
					Expect(err).To(Succeed())
				}
			case 1:
				token0Amount, _ := ToBigInt(float64(rand.Uint64()))
				token1Amount, _ := ToBigInt(float64(rand.Uint64()))
				balance0, _ := tokenBalanceOf(ctx, token0, who)
				balance1, _ := tokenBalanceOf(ctx, token1, who)
				switch rand.Intn(2) {
				case 0:
					if balance0.Cmp(token0Amount) > 0 {
						_, err = Exec(ctx, who, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token0, token0Amount, ZeroAmount})
						Expect(err).To(Succeed())
					}
				case 1:
					if balance1.Cmp(token1Amount) > 0 {
						_, err = Exec(ctx, who, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token1, token1Amount, ZeroAmount})
						Expect(err).To(Succeed())
					}
				}
			case 2:
				liquidity, _ := ToBigInt(float64(rand.Uint64()))
				minted := amount.NewAmount(0, 0)
				if who == alice { // owner
					minted, _ = ViewAmount(ctx, pair, "MintedAdminBalance")
				}
				balance, _ := tokenBalanceOf(ctx, pair, who)
				if balance.Sub(minted).Cmp(liquidity) > 0 {
					_, err = Exec(ctx, who, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity, ZeroAmount, ZeroAmount})
					Expect(err).To(Succeed())
				}
			case 3:
				liquidity, _ := ToBigInt(float64(rand.Uint64()))
				minted := amount.NewAmount(0, 0)
				if who == alice { // owner
					minted, _ = ViewAmount(ctx, pair, "MintedAdminBalance")
				}
				balance, _ := tokenBalanceOf(ctx, pair, who)
				if balance.Sub(minted).Cmp(liquidity) > 0 {
					switch rand.Intn(2) {
					case 0:
						_, err = Exec(ctx, who, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, liquidity, token0, ZeroAmount})
						Expect(err).To(Succeed())
					case 1:
						_, err = Exec(ctx, who, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, liquidity, token1, ZeroAmount})
						Expect(err).To(Succeed())
					}
				}
			case 4:
				swapAmount, _ := ToBigInt(float64(rand.Uint64()))
				switch rand.Intn(2) {
				case 0:
					_, err := Exec(ctx, who, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
					Expect(err).To(Succeed())
				case 1:
					_, err := Exec(ctx, who, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}
			case 5:
				outputAmount, _ := ToBigInt(float64(rand.Uint64()))
				switch rand.Intn(2) {
				case 0:
					_, err := Exec(ctx, who, routerAddr, "UniSwapTokensForExactTokens", []interface{}{outputAmount, MaxUint256, []common.Address{token0, token1}})
					Expect(err).To(Succeed())
				case 1:
					_, err := Exec(ctx, who, routerAddr, "UniSwapTokensForExactTokens", []interface{}{outputAmount, MaxUint256, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}
			case 6:
				// owner = alice
				_, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(err).To(Succeed())
			default:
			}
		}

		RemoveChain(cdx)
		afterEach()
	})
})
