package test

import (
	"math/big"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WithdrawAminFee", func() {
	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
	)

	Describe("WithdrawAminFee, AdminBalance", func() {

		Describe("admin = LP", func() {

			BeforeEach(func() {
				beforeEachUni()
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, pair, 30000000, 5000000000, trade.MAX_FEE, uint64(86400), aliceKey)
				uniAddInitialLiquidity(ctx, alice)
				uniMint(ctx, bob)
				tokenApprove(ctx, token0, bob, routerAddr)
				tokenApprove(ctx, token1, bob, routerAddr)

				swapAmount := amount.NewAmount(1, 0)
				Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
			})
			AfterEach(func() {
				RemoveChain(cdx)
				afterEach()

			})
			It("MintedAdminBalance, AdminBalance : Pair", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // It(payToken : ZeroAddress) 참조

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				lpTotalSupply, _ := tokenTotalSupply(ctx, pair)

				liquidity := lpOwnerBalance.Sub(expectedMintAmount) //최대
				Exec(ctx, alice, pair, "Transfer", []interface{}{pair, liquidity})
				Exec(ctx, alice, pair, "Burn", []interface{}{alice})

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity).Add(expectedMintAmount)))
				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity).Add(expectedMintAmount)))

				Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(ZeroAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity)))
				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity)))
			})

			It("MintedAdminBalance, AdminBalance : Router", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // It(payToken : ZeroAddress) 참조

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				lpTotalSupply, _ := tokenTotalSupply(ctx, pair)

				liquidity := lpOwnerBalance.Sub(expectedMintAmount) //최대
				tokenApprove(ctx, pair, alice, routerAddr)
				//Exec(ctx, alice, pair, "Approve", []interface{}{routerAddr, MaxUint256})
				Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity, ZeroAmount, ZeroAmount})

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity).Add(expectedMintAmount)))
				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity).Add(expectedMintAmount)))

				Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(ZeroAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity)))
				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity)))
			})

			It("admin UniRemoveLiquidity more than LPBalance : Pair", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				liquidity1 := amount.NewAmount(1, 0)
				_, err := Exec(ctx, alice, pair, "Transfer", []interface{}{pair, liquidity1})
				Expect(err).To(Succeed())

				_, err = Exec(ctx, alice, pair, "Burn", []interface{}{alice})
				Expect(err).To(Succeed())

				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				liquidity2 := lpOwnerBalance.Sub(expectedMintAmount).Add(amount.NewAmount(0, 1)) // 최대 + 1

				_, err = Exec(ctx, alice, pair, "Transfer", []interface{}{pair, liquidity2})
				Expect(err).To(Succeed())
				_, err = Exec(ctx, alice, pair, "Burn", []interface{}{alice})
				Expect(err).To(MatchError("Exchange: OWNER_LIQUIDITY"))
			})

			It("admin UniRemoveLiquidity more than LPBalance : Router", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				liquidity1 := amount.NewAmount(1, 0)
				tokenApprove(ctx, pair, alice, routerAddr)
				//Exec(ctx, alice, pair, "Approve", []interface{}{routerAddr, MaxUint256})
				_, err := Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity1, ZeroAmount, ZeroAmount})
				Expect(err).To(Succeed())

				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				liquidity2 := lpOwnerBalance.Sub(expectedMintAmount).Add(amount.NewAmount(0, 1)) // 최대 + 1
				_, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity2, ZeroAmount, ZeroAmount})
				Expect(err).To(MatchError("Router: OWNER_LIQUIDITY"))
			})

			It("owner_change_before_mint_admin_fee", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				lpTotalSupply, _ := tokenTotalSupply(ctx, pair)

				delay := uint64(3 * 86400)
				Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})
				ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)
				Exec(ctx, alice, pair, "ApplyTransferOwnerWinner", []interface{}{})

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))

				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
			})

			It("owner_change_after_mint_admin_fee", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, alice)
				lpTotalSupply, _ := tokenTotalSupply(ctx, pair)

				liquidity1 := amount.NewAmount(1, 0)
				tokenApprove(ctx, pair, alice, routerAddr)
				//Exec(ctx, alice, pair, "Approve", []interface{}{routerAddr, MaxUint256})
				_, err := Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{token0, token1, liquidity1, ZeroAmount, ZeroAmount})
				Expect(err).To(Succeed())

				delay := uint64(3 * 86400)
				_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})
				ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)
				_, err = Exec(ctx, alice, pair, "ApplyTransferOwnerWinner", []interface{}{})

				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(expectedMintAmount))

				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity1)))
				Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(expectedMintAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity1).Add(expectedMintAmount)))

				_, err = Exec(ctx, bob, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(ViewAmount(ctx, pair, "MintedAdminBalance")).To(Equal(ZeroAmount))
				Expect(ViewAmount(ctx, pair, "AdminBalance")).To(Equal(ZeroAmount))
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply.Sub(liquidity1)))
				Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity1)))
				Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(ZeroAmount))
			})
		})

		It("payToken : ZeroAddress", func() {
			// 아래 값은 다음 값으로 설정되어 계산됨
			// Fee (30bp,40bp) AminFee 50%, WinnerFee 50% : (30000000,40000000), 5000000000, 5000000000

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
						uniAddInitialLiquidity(ctx, charlie)
						uniMint(ctx, bob)
						tokenApprove(ctx, token0, bob, routerAddr)
						tokenApprove(ctx, token1, bob, routerAddr)

						pairBalance0, _ := tokenBalanceOf(ctx, token0, pair)
						pairBalance1, _ := tokenBalanceOf(ctx, token1, pair)
						K := Sqrt(Mul(pairBalance0.Int, pairBalance1.Int)) // totalSupply

						// lpOwnerBalance = lpTotalSupply - MinimumLiqudity
						lpOwnerBalance, _ := tokenBalanceOf(ctx, pair, charlie) // charlie = lpOwner 1명
						lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
						Expect(lpOwnerBalance).To(Equal(lpTotalSupply.Sub(_ML)))
						Expect(K).To(Equal(lpTotalSupply.Int))

						swapAmount := amount.NewAmount(1, 0)
						expectedOutputAmount, err := trade.UniGetAmountOut(fee, swapAmount.Int, pairBalance0.Int, pairBalance1.Int) // 1993996023971928199, 1991996031943904367

						is, err := Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
						Expect(err).To(Succeed())
						amounts := is[0].([]*amount.Amount)
						Expect(amounts[0]).To(Equal(swapAmount))
						Expect(amounts[1].Int).To(Equal(expectedOutputAmount))

						// Fee =  10^18*(30bp,40bp) = (300000000000000,400000000000000)
						// pairBalance0 += swapAmount
						// pairBalance1 -= expectedOutputAmount
						pairBalance0AfterSwap := pairBalance0.Add(swapAmount)
						pairBalance1AfterSwap := pairBalance1.Sub(amounts[1])
						Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(pairBalance0AfterSwap))
						Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(pairBalance1AfterSwap))
						Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))

						// totalSupply change 를 K값 변화로 계산 - 아직 totalSupply에 반영되지 않음
						cK := Sqrt(Mul(pairBalance0AfterSwap.Int, pairBalance1AfterSwap.Int)) // changedK : cK - K = 2121316110473344, 2828421484873749

						is, err = Exec(ctx, alice, pair, "AdminBalance", []interface{}{})
						Expect(err).To(Succeed())
						lpAmount := is[0].(*amount.Amount) // lpAmount = 1060658053645688, 1414210739608458
						// lpAmount = deltaK * _AdminFeeRatio

						var expectedLPAmount *big.Int
						if fee != 0 && adminFee != 0 {
							expectedLPAmount = Div(Mul(Sub(cK, K), lpTotalSupply.Int), Add(Sub(MulDivCC(cK, trade.FEE_DENOMINATOR, int64(adminFee)), cK), K))
						} else {
							expectedLPAmount = big.NewInt(0)
						}
						if lpAmount.Cmp(Zero) != 0 {
							Expect(lpAmount.Int).To(Equal(expectedLPAmount))
						} else {
							Expect(expectedLPAmount).To(Equal(Zero))
						}

						is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
						Expect(err).To(Succeed())
						Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
						Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
						// adminFee = balance * lpAmount/(lpTotalSupply + lpAmount)
						adminFee0 := ToAmount(MulDiv(pairBalance0AfterSwap.Int, lpAmount.Int, lpTotalSupply.Add(lpAmount).Int))
						adminFee1 := ToAmount(MulDiv(pairBalance1AfterSwap.Int, lpAmount.Int, lpTotalSupply.Add(lpAmount).Int))

						ownerFee0 := is[1].(*amount.Amount)
						ownerFee1 := is[2].(*amount.Amount)
						winnerFee0 := is[3].(*amount.Amount)
						winnerFee1 := is[4].(*amount.Amount)

						Expect(adminFee0).To(Equal(ownerFee0.Add(winnerFee0)))
						Expect(adminFee1).To(Equal(ownerFee1.Add(winnerFee1)))

						// winnerFee = adminFee * _winnerFeeRatio
						Expect(winnerFee0.Int).To(Equal(MulDivCC(adminFee0.Int, int64(winnerFee), trade.FEE_DENOMINATOR)))
						Expect(winnerFee1.Int).To(Equal(MulDivCC(adminFee1.Int, int64(winnerFee), trade.FEE_DENOMINATOR)))

						// ownerFee = adminFee - _winnerFee
						Expect(ownerFee0).To(Equal(adminFee0.Sub(winnerFee0)))
						Expect(ownerFee1).To(Equal(adminFee1.Sub(winnerFee1)))

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("payToken : token0,1", func() {
			_fee := uint64(30000000)
			_adminFee := uint64(10000000000)
			_winnerFee := uint64(0)
			lp := eve

			for k := 0; k < 2; k++ {

				beforeEach()
				payToken := uniTokens[k]
				is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{token0, token1, payToken, _PairName, _PairSymbol, alice, charlie, _fee, _adminFee, _winnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
				Expect(err).To(Succeed())
				pair = is[0].(common.Address)
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

				uniAddInitialLiquidity(ctx, lp)
				uniMint(ctx, bob)
				tokenApprove(ctx, token0, bob, routerAddr)
				tokenApprove(ctx, token1, bob, routerAddr)

				lpTotalSupply, _ := tokenTotalSupply(ctx, pair)
				swapAmount := amount.NewAmount(1, 0)
				for k := 0; k < 10; k++ { // 각각 10회
					_, err := Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
					Expect(err).To(Succeed())

					_, err = Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}

				pairBalace0, _ := tokenBalanceOf(ctx, token0, pair)
				pairBalace1, _ := tokenBalanceOf(ctx, token1, pair)
				Expect(tokenTotalSupply(ctx, pair)).To(Equal(lpTotalSupply))
				lpAmount, _ := ViewAmount(ctx, pair, "AdminBalance")

				var adminFee0, adminFee1 *amount.Amount
				if lp == alice {
					adminFee0 = ToAmount(MulDiv(pairBalace0.Int, lpAmount.Int, lpTotalSupply.Int))
					adminFee1 = ToAmount(MulDiv(pairBalace1.Int, lpAmount.Int, lpTotalSupply.Int))
				} else {
					adminFee0 = ToAmount(MulDiv(pairBalace0.Int, lpAmount.Int, lpTotalSupply.Add(lpAmount).Int))
					adminFee1 = ToAmount(MulDiv(pairBalace1.Int, lpAmount.Int, lpTotalSupply.Add(lpAmount).Int))
				}

				pairBalace0AfterBurn := pairBalace0.Sub(adminFee0)
				pairBalace1AfterBurn := pairBalace1.Sub(adminFee1)
				if payToken == token0 {
					swappedAmount, err := trade.UniGetAmountOut(_fee, adminFee1.Int, pairBalace1AfterBurn.Int, pairBalace0AfterBurn.Int)
					Expect(err).To(Succeed())
					adminFee0.Set(Add(adminFee0.Int, swappedAmount))
					adminFee1.Set(Zero)

					expectedWinnerFee0 := MulDivC(adminFee0.Int, big.NewInt(int64(_winnerFee)), trade.FEE_DENOMINATOR)
					expectedOwnerFee0 := Sub(adminFee0.Int, expectedWinnerFee0)

					is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
					ownerFee0 := is[1].(*amount.Amount)
					ownerFee1 := is[2].(*amount.Amount)
					winnerFee0 := is[3].(*amount.Amount)
					winnerFee1 := is[4].(*amount.Amount)

					Expect(ownerFee0).To(Equal(ToAmount(expectedOwnerFee0)))
					Expect(ownerFee1).To(Equal(ZeroAmount))
					Expect(winnerFee0).To(Equal(ToAmount(expectedWinnerFee0)))
					Expect(winnerFee1).To(Equal(ZeroAmount))
				} else {
					swappedAmount, err := trade.UniGetAmountOut(_fee, adminFee0.Int, pairBalace0AfterBurn.Int, pairBalace1AfterBurn.Int)
					Expect(err).To(Succeed())
					adminFee0.Set(Zero)
					adminFee1.Set(Add(adminFee1.Int, swappedAmount))

					expectedWinnerFee1 := MulDivCC(adminFee1.Int, int64(_winnerFee), trade.FEE_DENOMINATOR)
					expectedOwnerFee1 := Sub(adminFee1.Int, expectedWinnerFee1)

					is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
					ownerFee0 := is[1].(*amount.Amount)
					ownerFee1 := is[2].(*amount.Amount)
					winnerFee0 := is[3].(*amount.Amount)
					winnerFee1 := is[4].(*amount.Amount)

					Expect(ownerFee0).To(Equal(ZeroAmount))
					Expect(ownerFee1).To(Equal(ToAmount(expectedOwnerFee1)))
					Expect(winnerFee0).To(Equal(ZeroAmount))
					Expect(winnerFee1).To(Equal(ToAmount(expectedWinnerFee1)))
				}

				// 22.05.09 에러 대응 추가 start
				balance0, _ := tokenBalanceOf(ctx, token0, pair)
				balance1, _ := tokenBalanceOf(ctx, token1, pair)
				is, _ = Exec(ctx, alice, pair, "Reserves", []interface{}{})
				Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
				Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

				for k := 0; k < 10; k++ { // 각각 10회
					_, err := Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
					Expect(err).To(Succeed())

					_, err = Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}

				is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})

				balance0, _ = tokenBalanceOf(ctx, token0, pair)
				balance1, _ = tokenBalanceOf(ctx, token1, pair)
				is, _ = Exec(ctx, alice, pair, "Reserves", []interface{}{})
				Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
				Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

				_, err = Exec(ctx, alice, pair, "Skim", []interface{}{alice})
				Expect(err).To(Succeed())

				// 22.05.09 에러 대응 추가 end

				RemoveChain(cdx)
				afterEach()
			}
		})
	})
})
