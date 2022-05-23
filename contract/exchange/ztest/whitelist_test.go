package test

import (
	"math/big"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WhiteList", func() {
	var (
		cdx int
		cn  *chain.Chain
		ctx *types.Context
	)

	Describe("Uniswap", func() {
		Describe("FeeWhiteList, FeeAddress, Swap", func() {
			BeforeEach(func() {
				beforeEachUni()
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
			})
			AfterEach(func() {
				RemoveChain(cdx)
				afterEach()
			})

			It("FeeWhiteList, FeeAddress : not WhiteList", func() {

				is, err := Exec(ctx, alice, pair, "FeeWhiteList", []interface{}{alice})
				Expect(err).To(Succeed())
				Expect(is[0].([]byte)).To(Equal([]byte{}))

				is, err = Exec(ctx, alice, pair, "FeeAddress", []interface{}{alice})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(_Fee30))

			})

			It("FeeWhiteList, FeeAddress : WhiteList", func() {
				fee := uint64(0) // 0%

				is, err := Exec(ctx, alice, pair, "FeeWhiteList", []interface{}{eve})
				Expect(err).To(Succeed())
				Expect(is[0].([]byte)).To(Equal(bin.Uint64Bytes(fee)))

				is, err = Exec(ctx, alice, pair, "FeeAddress", []interface{}{eve})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

			})

			Describe("Swap", func() {
				swapAmount := amount.NewAmount(1, 0)
				token0Amount := amount.NewAmount(1000, 0)
				token1Amount := amount.NewAmount(1000, 0)
				expectedOutput := amount.NewAmount(0, 996006981039903216)   // 0.3%
				wlExpectedOutput := amount.NewAmount(0, 999000999000999000) // 0.0%

				BeforeEach(func() {
					uniMint(ctx, alice)
					uniApprove(ctx, alice)
					uniAddLiquidity(ctx, alice, token0Amount, token1Amount)
				})

				It("not WhiteList fee = pair.fee(cc)", func() {

					Exec(ctx, alice, token0, "Transfer", []interface{}{pair, swapAmount})

					_, err := Exec(ctx, alice, pair, "Swap", []interface{}{ZeroAmount, expectedOutput.Add(amount.NewAmount(0, 1)), alice, []byte(""), ZeroAddress})
					Expect(err).To(MatchError("Exchange: K"))

					_, err = Exec(ctx, alice, pair, "Swap", []interface{}{ZeroAmount, expectedOutput, alice, []byte(""), ZeroAddress})
					Expect(err).To(Succeed())
				})

				It("WhiteList from = ZeroAddress fee = feeWhiteList ", func() {
					uniMint(ctx, eve)

					Exec(ctx, eve, token0, "Transfer", []interface{}{pair, swapAmount})

					_, err := Exec(ctx, eve, pair, "Swap", []interface{}{ZeroAmount, wlExpectedOutput.Add(amount.NewAmount(0, 1)), eve, []byte(""), ZeroAddress})
					Expect(err).To(MatchError("Exchange: K"))

					_, err = Exec(ctx, eve, pair, "Swap", []interface{}{ZeroAmount, wlExpectedOutput, eve, []byte(""), ZeroAddress})
					Expect(err).To(Succeed())
				})

				It("WhiteList from = whitelist fee = feeWhiteList", func() {
					uniMint(ctx, eve)

					Exec(ctx, eve, token0, "Transfer", []interface{}{pair, swapAmount})

					_, err := Exec(ctx, eve, pair, "Swap", []interface{}{ZeroAmount, wlExpectedOutput.Add(amount.NewAmount(0, 1)), eve, []byte(""), eve})
					Expect(err).To(MatchError("Exchange: K"))

					_, err = Exec(ctx, eve, pair, "Swap", []interface{}{ZeroAmount, wlExpectedOutput, eve, []byte(""), eve})
					Expect(err).To(Succeed())
				})
			})
		})

		Describe("WithdrawAdminFees2", func() {
			expectedOwnerFee := []*amount.Amount{amount.NewAmount(0, 3744403031754912), amount.NewAmount(0, 0)}
			expectedWinnerFee := []*amount.Amount{amount.NewAmount(0, 3744403031754911), amount.NewAmount(0, 0)}

			wlExpectedOwnerFee := []*amount.Amount{amount.NewAmount(0, 3750028073802461), amount.NewAmount(0, 0)}
			wlExpectedWinnerFee := []*amount.Amount{amount.NewAmount(0, 3750028073802460), amount.NewAmount(0, 0)}

			It("not WhiteList", func() {
				beforeEach()
				is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{token0, token1, token0, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
				Expect(err).To(Succeed())
				pair = is[0].(common.Address)
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)

				uniAddInitialLiquidity(ctx, charlie)

				uniMint(ctx, bob)
				tokenApprove(ctx, token0, bob, routerAddr)
				tokenApprove(ctx, token1, bob, routerAddr)
				swapAmount := amount.NewAmount(1, 0)
				for k := 0; k < 10; k++ {
					_, err := Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
					Expect(err).To(Succeed())
					_, err = Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}

				lpAmount, _ := ViewAmount(ctx, pair, "AdminBalance")
				is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(expectedOwnerFee[0]))
				Expect(ownerFee1).To(Equal(expectedOwnerFee[1]))
				Expect(winnerFee0).To(Equal(expectedWinnerFee[0]))
				Expect(winnerFee1).To(Equal(expectedWinnerFee[1]))

				RemoveChain(cdx)
				afterEach()
			})

			It("WhiteList", func() {
				beforeEach()
				is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{token0, token1, token0, _PairName, _PairSymbol, eve, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
				Expect(err).To(Succeed())
				pair = is[0].(common.Address)
				cn, cdx, ctx, _ = initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)

				uniAddInitialLiquidity(ctx, charlie)

				uniMint(ctx, bob)
				tokenApprove(ctx, token0, bob, routerAddr)
				tokenApprove(ctx, token1, bob, routerAddr)
				swapAmount := amount.NewAmount(1, 0)
				for k := 0; k < 10; k++ {
					_, err := Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token0, token1}})
					Expect(err).To(Succeed())
					_, err = Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{token1, token0}})
					Expect(err).To(Succeed())
				}

				lpAmount, _ := ViewAmount(ctx, pair, "AdminBalance")
				is, err = Exec(ctx, eve, pair, "WithdrawAdminFees2", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(wlExpectedOwnerFee[0]))
				Expect(ownerFee1).To(Equal(wlExpectedOwnerFee[1]))
				Expect(winnerFee0).To(Equal(wlExpectedWinnerFee[0]))
				Expect(winnerFee1).To(Equal(wlExpectedWinnerFee[1]))

				RemoveChain(cdx)
				afterEach()
			})
		})
	})

	Describe("StableSwap", func() {

		BeforeEach(func() {
			beforeEachStable()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})
		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("FeeWhiteList, FeeAddress : not WhiteList", func() {

			is, err := Exec(ctx, alice, swap, "FeeWhiteList", []interface{}{alice})
			Expect(err).To(Succeed())
			Expect(is[0].([]byte)).To(Equal([]byte{}))

			is, err = Exec(ctx, alice, swap, "FeeAddress", []interface{}{alice})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(_Fee))

		})

		It("FeeWhiteList, FeeAddress : WhiteList", func() {
			fee := uint64(0) // 0%

			is, err := Exec(ctx, alice, swap, "FeeWhiteList", []interface{}{eve})
			Expect(err).To(Succeed())
			Expect(is[0].([]byte)).To(Equal(bin.Uint64Bytes(fee)))

			is, err = Exec(ctx, alice, swap, "FeeAddress", []interface{}{eve})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

		})

		Describe("AddLiquidity", func() {
			initialAmounts := MakeAmountSlice(N)
			initialAmounts[0].Set(Mul(Exp(big.NewInt(10), big.NewInt(int64(decimals[0]))), big.NewInt(1)))

			expectedLPAmount := amount.NewAmount(0, 998001999537911375)   // fee = 0.4%
			wlExpectedLPAmount := amount.NewAmount(0, 999999999537679406) // fee = 0.0%

			BeforeEach(func() {
				stableAddInitialLiquidity(ctx, alice)

			})

			It("CalcLPTokenAmount", func() {
				is, err := Exec(ctx, bob, swap, "CalcLPTokenAmount", []interface{}{initialAmounts, true})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

				is, err = Exec(ctx, eve, swap, "CalcLPTokenAmount", []interface{}{initialAmounts, true})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})

			It("AddLiquidity : not WhiteList", func() {

				stableMint(ctx, bob)
				stableApprove(ctx, bob)

				is, err := Exec(ctx, bob, swap, "AddLiquidity", []interface{}{initialAmounts, ZeroAmount})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))
			})

			It("AddLiquidity : WhiteList", func() {

				stableMint(ctx, eve)
				stableApprove(ctx, eve)

				is, err := Exec(ctx, eve, swap, "AddLiquidity", []interface{}{initialAmounts, ZeroAmount})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})
		})

		Describe("Exchange", func() {

			send := uint8(0)
			recv := uint8(1)
			amt := ToAmount(Pow10(decimals[send]))
			expectedOutput := amount.NewAmount(0, 995999)   // fee = 0.4%
			wlExpectedOutput := amount.NewAmount(0, 999999) // fee = 0.0%

			BeforeEach(func() {
				stableAddInitialLiquidity(ctx, alice)
			})

			It("Output", func() {
				Expect(expectedOutput.Cmp(wlExpectedOutput.Int) < 0).To(BeTrue())
			})

			It("Exchange, GetDy : not WhiteList fee = pair.fee(cc)", func() {

				stableApprove(ctx, bob)
				tokenMint(ctx, stableTokens[send], bob, amt)

				is, err := Exec(ctx, bob, swap, "GetDy", []interface{}{send, recv, amt, ZeroAddress})
				Expect(err).To(Succeed())
				received := is[0].(*amount.Amount)
				Expect(received).To(Equal(expectedOutput))
				//GPrintln("received", received)

				_, err = Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, amt, ZeroAmount, ZeroAddress})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(ctx, stableTokens[send], bob)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(ctx, stableTokens[recv], bob)).To(Equal(expectedOutput))

			})

			It("Exchange, GetDy : WhiteList from = ZeroAddress fee = feeWhiteList ", func() {
				stableApprove(ctx, eve)
				tokenMint(ctx, stableTokens[send], eve, amt)

				is, err := Exec(ctx, eve, swap, "GetDy", []interface{}{send, recv, amt, ZeroAddress})
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutput))

				_, err = Exec(ctx, eve, swap, "Exchange", []interface{}{send, recv, amt, ZeroAmount, ZeroAddress})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(ctx, stableTokens[send], eve)).To(Equal(ZeroAmount))
				received, _ := tokenBalanceOf(ctx, stableTokens[recv], eve)
				Expect(received).To(Equal(wlExpectedOutput))
			})

			It("Exchange, GetDy : WhiteList from = whitelist fee = feeWhiteList", func() {
				stableApprove(ctx, eve)
				tokenMint(ctx, stableTokens[send], eve, amt)

				is, err := Exec(ctx, eve, swap, "GetDy", []interface{}{send, recv, amt, eve})
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutput))

				_, err = Exec(ctx, eve, swap, "Exchange", []interface{}{send, recv, amt, ZeroAmount, eve})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(ctx, stableTokens[send], eve)).To(Equal(ZeroAmount))
				received, _ := tokenBalanceOf(ctx, stableTokens[recv], eve)
				Expect(received).To(Equal(wlExpectedOutput))
			})

		})

		Describe("RemoveLiquidityImbalance", func() {
			amts := MakeAmountSlice(N)
			amts[0].Set(Pow10(decimals[0]))

			max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)

			expectedLPAmount := amount.NewAmount(0, 1002000000462552385)   // fee = 0.4%
			wlExpectedLPAmount := amount.NewAmount(0, 1000000000462321109) // fee = 0%

			It("CalcLPTokenAmount : not Whitelist", func() {
				stableAddInitialLiquidity(ctx, alice)

				is, err := Exec(ctx, alice, swap, "CalcLPTokenAmount", []interface{}{amts, false})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

			})

			It("CalcLPTokenAmount : Whitelist", func() {
				stableAddInitialLiquidity(ctx, eve)

				is, err := Exec(ctx, eve, swap, "CalcLPTokenAmount", []interface{}{amts, false})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})

			It("RemoveLiquidityImbalance : not WhiteList", func() {
				stableAddInitialLiquidity(ctx, alice)

				is, err := Exec(ctx, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, max_burn})
				Expect(err).To(Succeed())

				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))
			})

			It("RemoveLiquidityImbalance : WhiteList", func() {
				stableAddInitialLiquidity(ctx, eve)

				is, err := Exec(ctx, eve, swap, "RemoveLiquidityImbalance", []interface{}{amts, max_burn})
				Expect(err).To(Succeed())

				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})
		})

		Describe("RemoveLiquidityOneCoin, CalcWithdrawOneCoin", func() {
			amt := ToAmount(Pow10(decimals[0]))
			idx := uint8(0)
			expectedOutputAmount := amount.NewAmount(0, 997999999539758297)   // fee = 0.4%
			wlExpectedOutputAmount := amount.NewAmount(0, 999999999537678891) // fee = 0%

			It("CalcWithdrawOneCoin : not WhiteList", func() {

				stableAddInitialLiquidity(ctx, alice)
				is, err := Exec(ctx, alice, swap, "CalcWithdrawOneCoin", []interface{}{amt, idx})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
			})

			It("RemoveLiquidityOneCoin : not WhiteList", func() {

				stableAddInitialLiquidity(ctx, alice)

				is, err := Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
			})

			It("CalcWithdrawOneCoin : WhiteList", func() {

				stableAddInitialLiquidity(ctx, eve)
				is, err := Exec(ctx, eve, swap, "CalcWithdrawOneCoin", []interface{}{amt, idx})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
			})

			It("RemoveLiquidityOneCoin : WhiteList", func() {
				stableAddInitialLiquidity(ctx, eve)

				is, err := Exec(ctx, eve, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
			})
		})

		Describe("WithdrawAdminFee, AdminBalances", func() {

			pi := uint8(0)

			expectedAdminBalances := []*amount.Amount{amount.NewAmount(11967, 900766209766638784), amount.NewAmount(0, 0), amount.NewAmount(0, 0)}   // fee = 0.4%
			wlExpectedAdminBalances := []*amount.Amount{amount.NewAmount(11999, 234276186563205263), amount.NewAmount(0, 0), amount.NewAmount(0, 0)} // fee = 0%

			It("not WhiteList", func() {
				sbc := &trade.StableSwapConstruction{
					Name:         _SwapName,
					Symbol:       _SwapSymbol,
					Factory:      ZeroAddress,
					NTokens:      uint8(N),
					Tokens:       stableTokens,
					PayToken:     stableTokens[pi],
					Owner:        alice,
					Winner:       charlie,
					Fee:          _Fee,
					AdminFee:     _AdminFee,
					WinnerFee:    _WinnerFee,
					WhiteList:    _WhiteList,
					GroupId:      _GroupId,
					Amp:          big.NewInt(_Amp),
					PrecisionMul: _PrecisionMul,
					Rates:        _Rates,
				}
				swap, _ = stablebase(ctx, sbc)

				stableAddInitialLiquidity(ctx, alice)
				stableMint(ctx, bob)
				stableApprove(ctx, bob)

				// 0,1 -> 1,2 -> 2,0
				for i := uint8(0); i < N; i++ {
					send := i
					recv := i + 1
					if recv == N {
						recv = 0
					}
					_, err := Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
					Expect(err).To(Succeed())
				}

				admin_balances := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
					Expect(err).To(Succeed())
					admin_balances[i].Set(is[0].(*amount.Amount).Int)
				}

				Expect(admin_balances).To(Equal(expectedAdminBalances))

				_, err := Exec(ctx, alice, swap, "WithdrawAdminFees", []interface{}{})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					if i == pi {
						wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(_WinnerFee)), trade.FEE_DENOMINATOR))
						Expect(tokenBalanceOf(ctx, stableTokens[i], charlie)).To(Equal(wF))
						Expect(tokenBalanceOf(ctx, stableTokens[i], alice)).To(Equal(admin_balances[i].Sub(wF)))
					} else {
						Expect(tokenBalanceOf(ctx, stableTokens[i], charlie)).To(Equal(ZeroAmount))
						Expect(tokenBalanceOf(ctx, stableTokens[i], alice)).To(Equal(ZeroAmount))
					}
				}
			})

			// owner whitelist(eve) 로 변경
			It("WhiteList", func() {
				sbc := &trade.StableSwapConstruction{
					Name:         _SwapName,
					Symbol:       _SwapSymbol,
					Factory:      ZeroAddress,
					NTokens:      uint8(N),
					Tokens:       stableTokens,
					PayToken:     stableTokens[pi],
					Owner:        eve,
					Winner:       charlie,
					Fee:          _Fee,
					AdminFee:     _AdminFee,
					WinnerFee:    _WinnerFee,
					WhiteList:    _WhiteList,
					GroupId:      _GroupId,
					Amp:          big.NewInt(_Amp),
					PrecisionMul: _PrecisionMul,
					Rates:        _Rates,
				}
				swap, _ = stablebase(ctx, sbc)

				stableAddInitialLiquidity(ctx, alice)
				stableMint(ctx, bob)
				stableApprove(ctx, bob)

				// 0,1 -> 1,2 -> 2,0
				for i := uint8(0); i < N; i++ {
					send := i
					recv := i + 1
					if recv == N {
						recv = 0
					}
					_, err := Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
					Expect(err).To(Succeed())
				}

				admin_balances := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
					Expect(err).To(Succeed())
					admin_balances[i].Set(is[0].(*amount.Amount).Int)
				}
				Expect(admin_balances).To(Equal(wlExpectedAdminBalances))

				_, err := Exec(ctx, eve, swap, "WithdrawAdminFees", []interface{}{})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					if i == pi {
						wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(_WinnerFee)), trade.FEE_DENOMINATOR))
						Expect(tokenBalanceOf(ctx, stableTokens[i], charlie)).To(Equal(wF))
						Expect(tokenBalanceOf(ctx, stableTokens[i], eve)).To(Equal(admin_balances[i].Sub(wF)))
					} else {
						Expect(tokenBalanceOf(ctx, stableTokens[i], charlie)).To(Equal(ZeroAmount))
						Expect(tokenBalanceOf(ctx, stableTokens[i], eve)).To(Equal(ZeroAmount))
					}
				}
			})
		})

	})

	Describe("Router", func() {

		var tokens []common.Address

		token0Amount := amount.NewAmount(5, 0)
		token1Amount := amount.NewAmount(10, 0)

		BeforeEach(func() {
			beforeEachWithoutTokens()

			tokens = DeployTokens(genesis, classMap["Token"], 3, admin)

			is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{tokens[0], tokens[1], tokens[0], _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
			Expect(err).To(Succeed())
			pair = is[0].(common.Address)

			is, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{tokens[1], tokens[2], tokens[1], _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
			Expect(err).To(Succeed())
			swap = is[0].(common.Address)

			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
			ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
			ctx, _ = setFees(cn, ctx, swap, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)

			//initial add liquidity
			tokenMint(ctx, tokens[0], alice, token0Amount)
			tokenMint(ctx, tokens[1], alice, token1Amount)
			tokenApprove(ctx, tokens[0], alice, routerAddr)
			tokenApprove(ctx, tokens[1], alice, routerAddr)

			_, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{tokens[0], tokens[1], token0Amount, token1Amount, ZeroAmount, ZeroAmount})
			Expect(err).To(Succeed())

		})
		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		Describe("UniSwap", func() {

			Describe("uniAddLiquidityOneCoin, uniGetLPTokenAmountOneCoin", func() {

				supplyAmount := amount.NewAmount(1, 0)

				expectedLPAmount := amount.NewAmount(0, 673811701330274830)   // fee = 0.4%
				wlExpectedLPAmount := amount.NewAmount(0, 674898880549347191) // fee = 0%

				It("not WhiteList", func() {

					tokenMint(ctx, tokens[0], charlie, supplyAmount)
					tokenApprove(ctx, tokens[0], charlie, routerAddr)

					is, err := Exec(ctx, charlie, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{tokens[0], tokens[1], tokens[0], supplyAmount})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

					is, err = Exec(ctx, charlie, routerAddr, "UniAddLiquidityOneCoin", []interface{}{tokens[0], tokens[1], tokens[0], supplyAmount, ZeroAmount})
					Expect(err).To(Succeed())
					Expect(is[1].(*amount.Amount)).To(Equal(expectedLPAmount))

				})

				It("WhiteList", func() {

					tokenMint(ctx, tokens[0], eve, supplyAmount)
					tokenApprove(ctx, tokens[0], eve, routerAddr)

					is, err := Exec(ctx, eve, routerAddr, "UniGetLPTokenAmountOneCoin", []interface{}{tokens[0], tokens[1], tokens[0], supplyAmount})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))

					is, err = Exec(ctx, eve, routerAddr, "UniAddLiquidityOneCoin", []interface{}{tokens[0], tokens[1], tokens[0], supplyAmount, ZeroAmount})
					Expect(err).To(Succeed())
					Expect(is[1].(*amount.Amount)).To(Equal(wlExpectedLPAmount))

				})
			})

			Describe("GetAmountsOut, SwapExactTokensForTokens", func() {
				// uni_router_test.go 참조
				swapAmount := amount.NewAmount(1, 0)
				expectedOutputAmount := amount.NewAmount(0, 1662497915624478906)   // fee = 0.4%
				wlExpectedOutputAmount := amount.NewAmount(0, 1666666666666666666) // fee = 0%

				It("GetAmountsOut, SwapExactTokensForTokens : not WhiteList", func() {

					tokenMint(ctx, tokens[0], charlie, swapAmount)
					tokenApprove(ctx, tokens[0], charlie, routerAddr)

					is, err := Exec(ctx, charlie, routerAddr, "GetAmountsOut", []interface{}{swapAmount, []common.Address{tokens[0], tokens[1]}})
					GPrintf("%+v", err)
					Expect(err).To(Succeed())
					amounts := is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(swapAmount))
					Expect(amounts[1]).To(Equal(expectedOutputAmount))

					is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts = is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(swapAmount))
					Expect(amounts[1]).To(Equal(expectedOutputAmount))
					Expect(tokenBalanceOf(ctx, tokens[0], charlie)).To(Equal(ZeroAmount))
					Expect(tokenBalanceOf(ctx, tokens[1], charlie)).To(Equal(expectedOutputAmount))
				})

				It("GetAmountsOut, SwapExactTokensForTokens : WhiteList", func() {

					tokenMint(ctx, tokens[0], eve, swapAmount)
					tokenApprove(ctx, tokens[0], eve, routerAddr)

					is, err := Exec(ctx, eve, routerAddr, "GetAmountsOut", []interface{}{swapAmount, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts := is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(swapAmount))
					Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))

					is, err = Exec(ctx, eve, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts = is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(swapAmount))
					Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))
					Expect(tokenBalanceOf(ctx, tokens[0], eve)).To(Equal(ZeroAmount))
					Expect(tokenBalanceOf(ctx, tokens[1], eve)).To(Equal(wlExpectedOutputAmount))
				})
			})

			Describe("UniGetAmountsIn, UniSwapTokensForExactTokens", func() {

				outputAmount := amount.NewAmount(1, 0)
				expectedSwapAmount := amount.NewAmount(0, 557227237267357629)   // fee = 0.4%
				wlExpectedSwapAmount := amount.NewAmount(0, 555555555555555556) // fee = 0%

				It("not WhiteList", func() {

					tokenMint(ctx, tokens[0], charlie, expectedSwapAmount)
					tokenApprove(ctx, tokens[0], charlie, routerAddr)

					is, err := Exec(ctx, charlie, routerAddr, "UniGetAmountsIn", []interface{}{outputAmount, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts := is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(expectedSwapAmount))
					Expect(amounts[1]).To(Equal(outputAmount))

					is, err = Exec(ctx, charlie, routerAddr, "UniSwapTokensForExactTokens", []interface{}{outputAmount, MaxUint256, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts = is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(expectedSwapAmount))
					Expect(amounts[1]).To(Equal(outputAmount))
					Expect(tokenBalanceOf(ctx, tokens[0], charlie)).To(Equal(ZeroAmount))
					Expect(tokenBalanceOf(ctx, tokens[1], charlie)).To(Equal(outputAmount))
				})

				It("WhiteList", func() {

					tokenMint(ctx, tokens[0], eve, wlExpectedSwapAmount)
					tokenApprove(ctx, tokens[0], eve, routerAddr)

					is, err := Exec(ctx, eve, routerAddr, "UniGetAmountsIn", []interface{}{outputAmount, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts := is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(wlExpectedSwapAmount))
					Expect(amounts[1]).To(Equal(outputAmount))

					is, err = Exec(ctx, eve, routerAddr, "UniSwapTokensForExactTokens", []interface{}{outputAmount, MaxUint256, []common.Address{tokens[0], tokens[1]}})
					Expect(err).To(Succeed())
					amounts = is[0].([]*amount.Amount)
					Expect(amounts[0]).To(Equal(wlExpectedSwapAmount))
					Expect(amounts[1]).To(Equal(outputAmount))
					Expect(tokenBalanceOf(ctx, tokens[0], eve)).To(Equal(ZeroAmount))
					Expect(tokenBalanceOf(ctx, tokens[1], eve)).To(Equal(outputAmount))
				})
			})

			Describe("uniRemoveLiquidityOneCoin, uniGetWithdrawAmountOneCoin", func() {

				withdrawLPAmount := amount.NewAmount(1, 0)

				expectedOutputAmount := amount.NewAmount(0, 1498586552649812788)   // fee = 0.4%
				wlExpectedOutputAmount := amount.NewAmount(0, 1500602061002483409) // fee = 0%
				expctedMintFee := amount.NewAmount(0, 321441396770565)

				It("not whiteList : output and mintfee", func() {

					tokenMint(ctx, tokens[0], charlie, token0Amount)
					tokenMint(ctx, tokens[1], charlie, token1Amount)
					tokenApprove(ctx, tokens[0], charlie, routerAddr)
					tokenApprove(ctx, tokens[1], charlie, routerAddr)
					Exec(ctx, charlie, routerAddr, "UniAddLiquidity", []interface{}{tokens[0], tokens[1], token0Amount, token1Amount, ZeroAmount, ZeroAmount})

					// bob swap : mintFee > 0
					swapAmount := amount.NewAmount(1, 0)
					tokenMint(ctx, tokens[0], bob, swapAmount)
					tokenApprove(ctx, tokens[0], bob, routerAddr)
					Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1]}})

					tokenApprove(ctx, pair, charlie, routerAddr)

					is, err := Exec(ctx, charlie, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{tokens[0], tokens[1], withdrawLPAmount, tokens[0]})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
					Expect(is[1].(*amount.Amount)).To(Equal(expctedMintFee))

					is, err = Exec(ctx, charlie, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{tokens[0], tokens[1], withdrawLPAmount, tokens[0], ZeroAmount})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
				})

				It("whiteList", func() {

					tokenMint(ctx, tokens[0], eve, token0Amount)
					tokenMint(ctx, tokens[1], eve, token1Amount)
					tokenApprove(ctx, tokens[0], eve, routerAddr)
					tokenApprove(ctx, tokens[1], eve, routerAddr)
					Exec(ctx, eve, routerAddr, "UniAddLiquidity", []interface{}{tokens[0], tokens[1], token0Amount, token1Amount, ZeroAmount, ZeroAmount})

					// bob swap : mintFee > 0
					swapAmount := amount.NewAmount(1, 0)
					tokenMint(ctx, tokens[0], bob, swapAmount)
					tokenApprove(ctx, tokens[0], bob, routerAddr)
					Exec(ctx, bob, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1]}})

					tokenApprove(ctx, pair, eve, routerAddr)

					is, err := Exec(ctx, eve, routerAddr, "UniGetWithdrawAmountOneCoin", []interface{}{tokens[0], tokens[1], withdrawLPAmount, tokens[0]})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
					Expect(is[1].(*amount.Amount)).To(Equal(expctedMintFee))

					is, err = Exec(ctx, eve, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{tokens[0], tokens[1], withdrawLPAmount, tokens[0], ZeroAmount})
					Expect(err).To(Succeed())
					Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
				})
			})
		})

		Describe("StableSwap", func() {
			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := amount.NewAmount(0, 997864184682683252)    // fee = 0.4%
			wlExpectedOutputAmount := amount.NewAmount(0, 1000866785037796641) // fee = 0%

			BeforeEach(func() {
				token0Amount := amount.NewAmount(5, 0)
				token1Amount := amount.NewAmount(10, 0)

				tokenMint(ctx, tokens[1], bob, token0Amount)
				tokenMint(ctx, tokens[2], bob, token1Amount)
				tokenApprove(ctx, tokens[1], bob, swap)
				tokenApprove(ctx, tokens[2], bob, swap)

				_, err := Exec(ctx, bob, swap, "AddLiquidity", []interface{}{[]*amount.Amount{token0Amount, token1Amount}, amount.NewAmount(0, 0)})
				Expect(err).To(Succeed())
			})

			It("Stableswap : GetAmountsOut, SwapExactTokensForTokens : not WhiteList", func() {

				tokenMint(ctx, tokens[1], charlie, swapAmount)
				tokenApprove(ctx, tokens[1], charlie, routerAddr)

				is, err := Exec(ctx, charlie, routerAddr, "GetAmountsOut", []interface{}{swapAmount, []common.Address{tokens[1], tokens[2]}})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))

				is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[1], tokens[2]}})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))
				Expect(tokenBalanceOf(ctx, tokens[1], charlie)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(ctx, tokens[2], charlie)).To(Equal(expectedOutputAmount))
			})

			It("Uniswap : GetAmountsOut, SwapExactTokensForTokens : WhiteList", func() {

				tokenMint(ctx, tokens[1], eve, swapAmount)
				tokenApprove(ctx, tokens[1], eve, routerAddr)

				is, err := Exec(ctx, eve, routerAddr, "GetAmountsOut", []interface{}{swapAmount, []common.Address{tokens[1], tokens[2]}})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))

				is, err = Exec(ctx, eve, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[1], tokens[2]}})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))
				Expect(tokenBalanceOf(ctx, tokens[1], eve)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(ctx, tokens[2], eve)).To(Equal(wlExpectedOutputAmount))
			})

		})
	})
})
