package test

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router", func() {
	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
		err error
	)

	Describe("PayToken Null", func() {
		It("factory", func() {
			beforeEachUni()

			is, err := Exec(genesis, alice, routerAddr, "Factory", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(factoryAddr))

			afterEach()
		})

		It("UniAddLiquidity : Initial Supply", func() {
			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE} // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}     // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}    // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniMint(ctx, alice)

						token0Amount := amount.NewAmount(1, 0)
						token1Amount := amount.NewAmount(4, 0)

						expectedLiquidity := amount.NewAmount(2, 0)

						tokenApprove(ctx, token0, alice, routerAddr)
						tokenApprove(ctx, token1, alice, routerAddr)
						is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token1Amount, ZeroAmount, ZeroAmount})
						Expect(err).To(Succeed())
						Expect(is[0].(*amount.Amount)).To(Equal(token0Amount))
						Expect(is[1].(*amount.Amount)).To(Equal(token1Amount))

						// BalanceOf
						Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount))
						Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount))
						Expect(tokenBalanceOf(ctx, pair, common.ZeroAddr)).To(Equal(_ML))
						Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(expectedLiquidity.Sub(_ML)))

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("UniAddLiquidity : Initial Supply", func() {
			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE} // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}     // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}    // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniMint(ctx, alice)

						token0Amount := amount.NewAmount(1, 0)
						token1Amount := amount.NewAmount(4, 0)

						expectedLiquidity := amount.NewAmount(2, 0)

						tokenApprove(ctx, token0, alice, routerAddr)
						tokenApprove(ctx, token1, alice, routerAddr)
						is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token1Amount, ZeroAmount, ZeroAmount})
						Expect(err).To(Succeed())
						Expect(is[0].(*amount.Amount)).To(Equal(token0Amount))
						Expect(is[1].(*amount.Amount)).To(Equal(token1Amount))

						// BalanceOf
						Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount))
						Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount))
						Expect(tokenBalanceOf(ctx, pair, common.ZeroAddr)).To(Equal(_ML))
						Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(expectedLiquidity.Sub(_ML)))

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("UniAddLiquidity : Price", func() {

			beforeEachUni()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
			uniMint(ctx, alice)

			token0Amount := amount.NewAmount(1, 0)
			token1Amount := amount.NewAmount(4, 0)

			tokenApprove(ctx, token0, alice, routerAddr)
			tokenApprove(ctx, token1, alice, routerAddr)
			is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token1Amount, ZeroAmount, ZeroAmount})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(token0Amount))
			Expect(is[1].(*amount.Amount)).To(Equal(token1Amount))

			is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
			reserve0 := is[0].([]*amount.Amount)[0]
			reserve1 := is[0].([]*amount.Amount)[1]

			priceBefore := reserve0.Div(reserve1)

			is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token0Amount, ZeroAmount, ZeroAmount})
			Expect(err).To(Succeed())

			is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
			reserve0 = is[0].([]*amount.Amount)[0]
			reserve1 = is[0].([]*amount.Amount)[1]

			priceAfter := reserve0.Div(reserve1)

			Expect(priceAfter).To(Equal(priceBefore))

			RemoveChain(cdx)
			afterEach()
		})

		It("AddLiquidity, UniGetLPTokenAmount : negative input", func() {
			beforeEachUni()

			_, err := Exec(genesis, bob, routerAddr, "UniGetLPTokenAmount", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), ToAmount(big.NewInt(1))})
			Expect(err).To(MatchError("Router: INSUFFICIENT_A_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniGetLPTokenAmount", []interface{}{token0, token1, ToAmount(big.NewInt(1)), ToAmount(big.NewInt(-1))})
			Expect(err).To(MatchError("Router: INSUFFICIENT_B_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), ToAmount(big.NewInt(1)), ZeroAmount, ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_A_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, ToAmount(big.NewInt(1)), ToAmount(big.NewInt(-1)), ZeroAmount, ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_B_AMOUNT"))

			afterEach()
		})

		It("AddLiquidity, UniGetLPTokenAmount", func() {
			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE} // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}     // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}    // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						uniMint(ctx, bob)
						uniApprove(ctx, bob)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b0, _ := ToBigInt(float64(rand.Uint64()))
							b1, _ := ToBigInt(float64(rand.Uint64()))
							token0Amount := ToAmount(b0)
							token1Amount := ToAmount(b1)

							is, err := Exec(ctx, bob, routerAddr, "UniGetLPTokenAmount", []interface{}{token0, token1, token0Amount, token1Amount})
							expectedLPToken := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = Exec(ctx, bob, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, token0Amount, token1Amount, ZeroAmount, ZeroAmount})
							Expect(err).To(Succeed())
							amount0 := is[0].(*amount.Amount)
							amount1 := is[1].(*amount.Amount)
							liquidity := is[2].(*amount.Amount)
							Expect(liquidity).To(Equal(expectedLPToken))

							bobBalances[0] = bobBalances[0].Sub(amount0)
							bobBalances[1] = bobBalances[1].Sub(amount1)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(ctx, token0, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(ctx, token1, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token out not match error", func() {
			beforeEachUni()

			_, err = Exec(genesis, bob, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{token0, token1, alice, ToAmount(big.NewInt(1))})
			Expect(err).To(MatchError("Router: INPUT_TOKEN_NOT_MATCH"))

			_, err = Exec(genesis, bob, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, alice, ToAmount(big.NewInt(1)), ZeroAmount})
			Expect(err).To(MatchError("Router: INPUT_TOKEN_NOT_MATCH"))

			afterEach()
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : negative input error", func() {
			beforeEachUni()

			_, err := Exec(genesis, bob, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{token0, token1, token0, ToAmount(big.NewInt(-1))})
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{token0, token1, token1, ToAmount(big.NewInt(-1))})
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token0, ToAmount(big.NewInt(-1)), ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = Exec(genesis, bob, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token1, ToAmount(big.NewInt(-1)), ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			afterEach()
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : Initial Supply Error", func() {
			beforeEachUni()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			_, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token0, amount.NewAmount(1, 0), ZeroAmount})
			Expect(err).To(MatchError("Router: BOTH_RESERVE_0"))

			RemoveChain(cdx)
			afterEach()
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : token0", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						uniMint(ctx, bob)
						uniApprove(ctx, bob)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							tokenAmount := ToAmount(b)

							is, err := Exec(ctx, bob, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{token0, token1, token0, tokenAmount})
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = Exec(ctx, bob, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token0, tokenAmount, ZeroAmount})
							Expect(err).To(Succeed())
							amountIn := is[0].(*amount.Amount)
							liquidity := is[1].(*amount.Amount)
							Expect(amountIn).To(Equal(tokenAmount))
							Expect(liquidity).To(Equal(expected))

							bobBalances[0] = bobBalances[0].Sub(amountIn)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(ctx, token0, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(ctx, token1, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}
						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : token1", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						uniMint(ctx, bob)
						uniApprove(ctx, bob)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							tokenAmount := ToAmount(b)

							is, err := Exec(ctx, bob, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{token0, token1, token1, tokenAmount})
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = Exec(ctx, bob, routerAddr, "UniAddLiquidityOneCoin", []interface{}{token0, token1, token1, tokenAmount, ZeroAmount})
							Expect(err).To(Succeed())
							amountIn := is[0].(*amount.Amount)
							liquidity := is[1].(*amount.Amount)
							Expect(amountIn).To(Equal(tokenAmount))
							Expect(liquidity).To(Equal(expected))

							bobBalances[1] = bobBalances[1].Sub(amountIn)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(ctx, token0, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(ctx, token1, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}
						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("RemoveLiquidity", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)

						token0Amount := amount.NewAmount(1, 0)
						token1Amount := amount.NewAmount(4, 0)

						uniMint(ctx, alice)
						uniApprove(ctx, alice)
						uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

						expectedLiquidity := amount.NewAmount(2, 0)

						tokenApprove(ctx, pair, alice, routerAddr)

						_, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, expectedLiquidity.Sub(_ML), ZeroAmount, ZeroAmount})
						//GPrintf("%+v", err)
						Expect(err).To(Succeed())

						// BalanceOf
						Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(ZeroAmount))
						Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(amount.NewAmount(0, 500))))
						Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(amount.NewAmount(0, 2000))))

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("RemoveLiquidity, UniGetWithdrawAmount : negative input error", func() {
			beforeEachUni()

			_, err := Exec(genesis, bob, routerAddr, "UniGetWithdrawAmount", []interface{}{token0, token1, ToAmount(big.NewInt(-1))})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = Exec(genesis, bob, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), ZeroAmount, ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			afterEach()
		})

		It("RemoveLiquidity, UniGetWithdrawAmount", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						tokenApprove(ctx, pair, alice, routerAddr)

						lpBalance, _ := tokenBalanceOf(ctx, pair, alice)
						lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := Exec(ctx, bob, routerAddr, "UniGetWithdrawAmount", []interface{}{token0, token1, liquidity})
							expected0 := is[0].(*amount.Amount)
							expected1 := is[1].(*amount.Amount)

							is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity, ZeroAmount, ZeroAmount})
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							amount1Out := is[1].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected0))
							Expect(amount1Out).To(Equal(expected1))

							// balance
							balances[0] = balances[0].Add(amount0Out)
							balances[1] = balances[1].Add(amount1Out)
							Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(balances[0]))
							Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(balances[1]))

							// lp
							lpBalance = lpBalance.Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Sub(liquidity)
							Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
						}

						RemoveChain(cdx)
						afterEach()
					}
				}
			}

		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token out not match error", func() {
			beforeEachUni()

			_, err := Exec(genesis, bob, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(1)), alice})
			Expect(err).To(MatchError("Router: OUTPUT_TOKEN_NOT_MATCH"))

			_, err = Exec(genesis, bob, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(1)), alice, ZeroAmount})
			Expect(err).To(MatchError("Router: OUTPUT_TOKEN_NOT_MATCH"))

			afterEach()
		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : negative input error", func() {
			beforeEachUni()

			_, err := Exec(genesis, bob, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), token0})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = Exec(genesis, bob, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), token1})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = Exec(genesis, bob, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), token0, ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = Exec(genesis, bob, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, ToAmount(big.NewInt(-1)), token1, ZeroAmount})
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			afterEach()
		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token0", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						tokenApprove(ctx, pair, alice, routerAddr)

						lpBalance, _ := tokenBalanceOf(ctx, pair, alice)
						lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := Exec(ctx, bob, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{token0, token1, liquidity, token0})
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							mintFee := is[1].(*amount.Amount)

							is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, liquidity, token0, ZeroAmount})
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected))

							// balance
							balances[0] = balances[0].Add(amount0Out)
							Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(balances[0]))
							Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(balances[1]))

							// lp : alice(owner), totalSupply 모두에게 +
							lpBalance = lpBalance.Add(mintFee).Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Add(mintFee).Sub(liquidity)
							Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
						}

						RemoveChain(cdx)
						afterEach()
					}
				}
			}

		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token1", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachUni()
						cn, cdx, ctx, _ = initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
						uniAddInitialLiquidity(ctx, alice)
						tokenApprove(ctx, pair, alice, routerAddr)

						lpBalance, _ := tokenBalanceOf(ctx, pair, alice)
						lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := Exec(ctx, bob, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{token0, token1, liquidity, token1})
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							mintFee := is[1].(*amount.Amount)

							is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{token0, token1, liquidity, token1, ZeroAmount})
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected))

							// balance
							balances[1] = balances[1].Add(amount0Out)
							Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(balances[0]))
							Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(balances[1]))

							// lp : alice(owner), totalSupply 모두에게 +
							lpBalance = lpBalance.Add(mintFee).Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Add(mintFee).Sub(liquidity)
							Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
						}

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		Describe("swapExactTokensForTokens, getAmountsOut", func() {

			token0Amount := amount.NewAmount(5, 0)
			token1Amount := amount.NewAmount(10, 0)
			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(1662497915624478906)}

			BeforeEach(func() {
				beforeEachUni()
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)

				uniMint(ctx, alice)
				uniApprove(ctx, alice)
				uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

				tokenApprove(ctx, token0, alice, routerAddr)
			})

			AfterEach(func() {
				RemoveChain(cdx)
				afterEach()
			})

			It("GetAmountsOut : negative input", func() {
				_, err := Exec(ctx, alice, routerAddr, "GetAmountsOut", []interface{}{ToAmount(big.NewInt(-1)), []common.Address{token0, token1}})
				Expect(err).To(MatchError("Router: INSUFFICIENT_IN_AMOUNT"))
			})

			It("SwapExactTokensForTokens : negative input", func() {
				_, err := Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{ToAmount(big.NewInt(-1)), ZeroAmount, []common.Address{token0, token1}})
				Expect(err).To(MatchError("Router: INSUFFICIENT_SWAP_AMOUNT"))
			})

			It("happy path, amounts", func() {
				is, err := Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))

				// BalanceOf
				Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Sub(swapAmount)))
				Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Add(expectedOutputAmount)))
			})

			It("gas", func() {
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        routerAddr,
					Method:    "SwapExactTokensForTokens",
					Args:      bin.TypeWriteAll(swapAmount, ZeroAmount, []common.Address{token0, token1}),
				}
				ctx, err = Sleep(cn, ctx, tx, uint64(1), aliceKey) // 1 sec
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Sub(swapAmount)))
				Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Add(expectedOutputAmount)))

				// gas
				balance, err := tokenBalanceOf(ctx, mainToken, alice)
				Expect(err).To(Succeed())
				gas := amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance)
				Println("gas :", gas) // 100000000000000000 = 10^17 = 0.1 MEV
			})
		})

		Describe("swapTokensForExactTokens", func() {

			It("GetAmountsIn : negative input", func() {
				_, err := Exec(ctx, alice, routerAddr, "UniGetAmountsIn", []interface{}{ToAmount(big.NewInt(-1)), []common.Address{token0, token1}})
				Expect(err).To(MatchError("Router: INSUFFICIENT_OUT_AMOUNT"))
			})

			It("UniSwapTokensForExactTokens : negative input", func() {
				_, err := Exec(ctx, alice, routerAddr, "UniSwapTokensForExactTokens", []interface{}{ToAmount(big.NewInt(-1)), ZeroAmount, []common.Address{token0, token1}})
				Expect(err).To(MatchError("Router: INSUFFICIENT_SWAP_AMOUNT"))
			})

			It("happy path, amounts", func() {
				beforeEachUni()
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

				token0Amount := amount.NewAmount(5, 0)
				token1Amount := amount.NewAmount(10, 0)
				expectedSwapAmount := &amount.Amount{Int: big.NewInt(557227237267357629)}
				outputAmount := amount.NewAmount(1, 0)

				uniMint(ctx, alice)
				uniApprove(ctx, alice)
				uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

				ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)

				tokenApprove(ctx, token0, alice, routerAddr)

				is, err := Exec(ctx, alice, routerAddr, "UniSwapTokensForExactTokens", []interface{}{outputAmount, MaxUint256, []common.Address{token0, token1}})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(expectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))

				Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Sub(expectedSwapAmount)))
				Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Add(outputAmount)))

				RemoveChain(cdx)
				afterEach()
			})
		})
	})
})
