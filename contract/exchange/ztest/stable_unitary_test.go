package test

import (
	"math"
	"math/big"
	"time"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	contract : curve-contract
	files :
		tests/pools/common/unitary/test_add_liquidity.py
		tests/pools/common/unitary/test_add_liquidity_initial.py
		tests/pools/common/unitary/test_claim_fees.py
		tests/pools/common/unitary/test_exchange.py
		tests/pools/common/unitary/test_exchange_reverts.py
		tests/pools/common/unitary/test_get_virtual_price.py
		tests/pools/common/unitary/test_kill.py
		tests/pools/common/unitary/test_modify_fees.py
		tests/pools/common/unitary/test_nonpayable.py
		tests/pools/common/unitary/test_ramp_A_precise.py
		tests/pools/common/unitary/test_remove_liquidity.py
		tests/pools/common/unitary/test_remove_liquidity_imbalance.py
		tests/pools/common/unitary/test_remove_liquidity_one_coin.py
		tests/pools/common/unitary/test_transfer_ownership.py              -> ownerwinner_test.go
		tests/pools/common/unitary/test_xfer_to_contract.py
*/

var _ = Describe("Unitary", func() {

	It("CalcLPTokenAmount", func() {
		beforeEachStable()

		for i := uint8(0); i < N; i++ {
			amounts := MakeAmountSlice(N)
			amounts[i].Int.Set(big.NewInt(-1))
			_, err := Exec(genesis, alice, swap, "CalcLPTokenAmount", []interface{}{amounts, true})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))
		}

		afterEach()
	})

	Describe("test_add_liquidity_initial.py", func() {
		BeforeEach(func() {
			beforeEachStable()
			stableMint(genesis, alice)
			stableApprove(genesis, alice)
		})

		AfterEach(func() {
			cleanUp()
		})

		It("negative input amount", func() {
			stableAddLiquidity(genesis, alice, swap, _InitialAmounts)
			for i := uint8(0); i < N; i++ {
				amounts := MakeAmountSlice(N)

				amounts[i].Int.Set(big.NewInt(-1))
				_, err := Exec(genesis, alice, swap, "AddLiquidity", []interface{}{amounts, amount.NewAmount(0, 0)})
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))
			}

			afterEach()
		})

		DescribeTable("test_initial",
			func(min_amount *amount.Amount) {
				amounts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amounts[i].Set(Pow10(decimals[i]))
				}

				_, err := Exec(genesis, alice, swap, "AddLiquidity", []interface{}{amounts, min_amount})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					Expect(tokenBalanceOf(genesis, stableTokens[i], alice)).To(Equal(_InitialAmounts[i].Sub(amounts[i])))
					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(amounts[i]))
				}

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(amount.NewAmount(uint64(N), 0)))
				Expect(tokenTotalSupply(genesis, swap)).To(Equal(amount.NewAmount(uint64(N), 0)))
			},

			Entry("1", amount.NewAmount(0, 0)),
			Entry("2", amount.NewAmount(2, 0)),
		)

		It("test_initial_liquidity_missing_coin", func() {
			amounts := MakeAmountSlice(N)
			for i := uint8(0); i < N; i++ {
				amounts[i].Set(Pow10(decimals[i]))
			}

			for i := uint8(0); i < N; i++ {
				amounts[i].Int.Set(Zero)
				_, err := Exec(genesis, alice, swap, "AddLiquidity", []interface{}{amounts, amount.NewAmount(0, 0)})
				Expect(err).To(MatchError("Exchange: INITILAL_DEPOSIT"))
			}
		})
	})

	Describe("test_add_liquidity.py", func() {

		BeforeEach(func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)
			stableMint(genesis, bob)
			stableApprove(genesis, bob)
		})

		AfterEach(func() {
			afterEach()
		})

		It("test_add_liquidity", func() {
			_, err := Exec(genesis, bob, swap, "AddLiquidity", []interface{}{_InitialAmounts, ZeroAmount})
			Expect(err).To(Succeed())

			for i := uint8(0); i < N; i++ {
				Expect(tokenBalanceOf(genesis, stableTokens[i], bob)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(_InitialAmounts[i].MulC(2)))
			}
			Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(amount.NewAmount(uint64(N), 0).MulC(_BaseAmount)))
			Expect(tokenTotalSupply(genesis, swap)).To(Equal(amount.NewAmount(uint64(N), 0).MulC(_BaseAmount * 2)))

		})

		It("test_add_with_slippage", func() {
			amounts := MakeSlice(N)
			for i := uint8(0); i < N; i++ {
				amounts[i].Set(Pow10(decimals[i]))
			}

			amt0, err := MulF(amounts[0], 0.99)
			Expect(err).To(Succeed())
			amounts[0].Set(amt0)

			amt1, err := MulF(amounts[1], 1.01)
			Expect(err).To(Succeed())
			amounts[1].Set(amt1)

			_, err = Exec(genesis, bob, swap, "AddLiquidity", []interface{}{ToAmounts(amounts), ZeroAmount})
			Expect(err).To(Succeed())

			balance, _ := tokenBalanceOf(genesis, swap, bob)
			if balance.Cmp(big.NewInt(int64(float64(amount.NewAmount(uint64(N), 0).Int64())*0.999))) <= 0 {
				Fail("Lower")
			}
			if balance.Cmp(amount.NewAmount(uint64(N), 0).Int) >= 0 {
				Fail("Upper")
			}

		})

		It("test_add_one_coin", func() {
			for idx := uint8(0); idx < N; idx++ {
				// idx != 0 초기화 다시한번
				if idx != 0 {
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)
					stableMint(genesis, bob)
					stableApprove(genesis, bob)
				}

				amounts := MakeAmountSlice(N)
				amounts[idx].Set(_InitialAmounts[idx].Int)

				_, err := Exec(genesis, bob, swap, "AddLiquidity", []interface{}{amounts, ZeroAmount})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					balance, _ := tokenBalanceOf(genesis, stableTokens[i], bob)
					if balance.Cmp(_InitialAmounts[i].Sub(amounts[i]).Int) != 0 {
						Fail("Not Equal")
					}
					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(_InitialAmounts[i].Add(amounts[i])))
				}

				balance, _ := tokenBalanceOf(genesis, swap, bob)
				if balance.Cmp(big.NewInt(int64(float64(amount.NewAmount(uint64(_BaseAmount), 0).Int64())*0.999))) <= 0 {
					Fail("Lower")
				}
				if balance.Cmp(amount.NewAmount(uint64(_BaseAmount), 0).Int) >= 0 {
					Fail("Upper")
				}

				afterEach()
			}
		})

		It("test_insufficient_balance", func() {
			amounts := MakeAmountSlice(N)
			for i := uint8(0); i < N; i++ {
				amounts[i].Set(Pow10(decimals[i]))
			}

			_, err := Exec(genesis, charlie, swap, "AddLiquidity", []interface{}{amounts, ZeroAmount})
			Expect(err).To(MatchError("the token holding quantity is insufficient"))
		})

		It("test_min_amount_too_high", func() {
			amounts := MakeAmountSlice(N)
			for i := uint8(0); i < N; i++ {
				amounts[i].Set(Pow10(decimals[i]))
			}

			min_amount := amount.NewAmount(uint64(N), 1)
			_, err := Exec(genesis, bob, swap, "AddLiquidity", []interface{}{amounts, min_amount})
			Expect(err).To(MatchError("Exchange: SLIPPAGE"))
		})

		It("test_min_amount_with_slippage", func() {
			amounts := MakeAmountSlice(N)
			for i := uint8(0); i < N; i++ {
				amounts[i].Set(Pow10(decimals[i]))
			}
			amounts[0].Set(big.NewInt(int64(float64(amounts[0].Int64()) * 0.99)))
			amounts[1].Set(big.NewInt(int64(float64(amounts[1].Int64()) * 1.01)))

			min_amount := amount.NewAmount(uint64(N), 0)
			_, err := Exec(genesis, bob, swap, "AddLiquidity", []interface{}{amounts, min_amount})
			Expect(err).To(MatchError("Exchange: SLIPPAGE"))
		})

		It("test_event", func() {
			// no event
		})

		It("test_wrong_eth_amount", func() {
			// no Ether send
		})
	})

	Describe("test_claim_fees.py", func() {
		It("test_withdraw_only_owner", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)
			stableMint(genesis, bob)
			stableApprove(genesis, bob)
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(3*86400), aliceKey)

			_, err := Exec(ctx, bob, swap, "WithdrawAdminFees", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			RemoveChain(cdx)
			afterEach()
		})

		It("AdminBalances index error", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)
			stableMint(genesis, bob)
			stableApprove(genesis, bob)
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(3*86400), aliceKey)

			_, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(4)})
			Expect(err).To(MatchError("Exchange: IDX"))

			RemoveChain(cdx)
			afterEach()
		})

		It("test_admin_balances", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {
						for sending := uint8(0); sending < N; sending++ {
							for receiving := uint8(0); receiving < N; receiving++ {
								if sending == receiving {
									continue
								}
								beforeEachStable()
								stableAddInitialLiquidity(genesis, alice)
								stableMint(genesis, bob)
								stableApprove(genesis, bob)
								cn, cdx, ctx, _ := initChain(genesis, admin)
								ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
								ctx, _ = setFees(cn, ctx, swap, fee, adminFee, winnerFee, uint64(86400), aliceKey)

								send, recv := sending, receiving
								is, err := Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
								Expect(err).To(Succeed())
								dy_receiving := is[0].(*amount.Amount)
								dy_receiving = ToAmount(MulDivCC(dy_receiving.Int, trade.FEE_DENOMINATOR, int64(trade.FEE_DENOMINATOR-fee))) // y = y' * DENOM / (DENOM-FEE)
								fee_receiving := getFee(dy_receiving.Int, fee)
								adminfee_receiving := getFee(fee_receiving, adminFee)

								send, recv = receiving, sending
								is, err = Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
								Expect(err).To(Succeed())
								dy_sending := is[0].(*amount.Amount)
								dy_sending = ToAmount(MulDivCC(dy_sending.Int, trade.FEE_DENOMINATOR, int64(trade.FEE_DENOMINATOR-fee))) // y = y' * DENOM / (DENOM-FEE)
								fee_sending := getFee(dy_sending.Int, fee)
								adminfee_sending := getFee(fee_sending, adminFee)

								is, err = Exec(ctx, bob, swap, "Reserves", []interface{}{})
								Expect(err).To(Succeed())

								reserves := is[0].([]*amount.Amount)

								for k := uint8(0); k < N; k++ {
									balance_coin, _ := tokenBalanceOf(ctx, stableTokens[k], swap)

									admin_fee := balance_coin.Sub(reserves[k]).Int

									switch k {
									case sending:
										if admin_fee.Cmp(Zero) == 0 {
											Expect(adminfee_sending).To(Equal(Zero))
										} else {
											Expect(admin_fee).To(Equal(adminfee_sending))
										}
									case receiving:
										if admin_fee.Cmp(Zero) == 0 {
											Expect(adminfee_receiving).To(Equal(Zero))
										} else {
											Expect(admin_fee).To(Equal(adminfee_receiving))
										}
									default:
										Expect(admin_fee.Cmp(Zero) == 0).To(BeTrue())
									}
								}

								RemoveChain(cdx)
								afterEach()
							}
						}
					}
				}
			}
		})

		It("test_withdraw_one_coin", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						for send := uint8(0); send < N; send++ {
							for recv := uint8(0); recv < N; recv++ {
								if send == recv {
									continue
								}
								if !(send == 0 && recv == 1) {
									continue
								}
								beforeEachStable()
								stableAddInitialLiquidity(genesis, alice)
								stableMint(genesis, bob)
								stableApprove(genesis, bob)
								cn, cdx, ctx, _ := initChain(genesis, admin)
								ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
								ctx, _ = setFees(cn, ctx, swap, fee, adminFee, winnerFee, uint64(86400), aliceKey)

								_, err := Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
								Expect(err).To(Succeed())

								admin_balances, err := stableGetAdminBalances(ctx) // fixture
								Expect(err).To(Succeed())
								for i := uint8(0); i < N; i++ {
									is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
									Expect(err).To(Succeed())
									if admin_balances[i].Cmp(Zero) == 0 {
										Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
									} else {
										Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i]))
									}
								}

								if admin_balances[recv].Cmp(Zero) < 0 {
									Fail("Admin Balance")
								}
								Expect(Sum(ToBigInts(admin_balances))).To(Equal(admin_balances[recv].Int))

								_, err = Exec(ctx, alice, swap, "WithdrawAdminFees", []interface{}{})
								Expect(err).To(Succeed())

								winner, _ := ViewAddress(ctx, swap, "Winner")

								wF := ToAmount(MulDivCC(admin_balances[recv].Int, int64(winnerFee), trade.FEE_DENOMINATOR))
								Expect(tokenBalanceOf(ctx, stableTokens[recv], winner)).To(Equal(wF))
								Expect(tokenBalanceOf(ctx, stableTokens[recv], alice)).To(Equal(admin_balances[recv].Sub(wF)))

								is, err := Exec(ctx, bob, swap, "Reserves", []interface{}{})
								Expect(err).To(Succeed())

								swap_reserves := is[0].([]*amount.Amount)
								Expect(tokenBalanceOf(ctx, stableTokens[recv], swap)).To(Equal(swap_reserves[recv]))

								RemoveChain(cdx)
								afterEach()
							}
						}
					}
				}
			}
		})

		It("test_withdraw_all_coins", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEachStable()
						stableAddInitialLiquidity(genesis, alice)
						stableMint(genesis, bob)
						stableApprove(genesis, bob)
						cn, cdx, ctx, _ := initChain(genesis, admin)
						ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
						ctx, _ = setFees(cn, ctx, swap, fee, adminFee, winnerFee, uint64(86400), aliceKey)

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

						admin_balances, err := stableGetAdminBalances(ctx) // fixture
						Expect(err).To(Succeed())
						for i := uint8(0); i < N; i++ {
							is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
							Expect(err).To(Succeed())
							if admin_balances[i].Cmp(Zero) == 0 {
								Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
							} else {
								Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i]))
							}
						}

						_, err = Exec(ctx, alice, swap, "WithdrawAdminFees", []interface{}{})
						Expect(err).To(Succeed())

						winner, _ := ViewAddress(ctx, swap, "Winner")

						for i := uint8(0); i < N; i++ {
							wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(winnerFee)), trade.FEE_DENOMINATOR))
							Expect(tokenBalanceOf(ctx, stableTokens[i], winner)).To(Equal(wF))
							Expect(tokenBalanceOf(ctx, stableTokens[i], alice)).To(Equal(admin_balances[i].Sub(wF)))
						}

						RemoveChain(cdx)
						afterEach()
					}
				}
			}
		})

		It("test_withdraw_all_coins : payToken Error", func() {

			err := beforeEach()
			Expect(err).To(Succeed())
			sbc := &trade.StableSwapConstruction{
				Name:         _SwapName,
				Symbol:       _SwapSymbol,
				Factory:      ZeroAddress,
				NTokens:      uint8(N),
				Tokens:       stableTokens,
				PayToken:     alice, // payToken Error 유발
				Owner:        alice,
				Winner:       charlie,
				Fee:          _Fee,
				AdminFee:     _AdminFee,
				WinnerFee:    _WinnerFee,
				Amp:          big.NewInt(_Amp),
				PrecisionMul: _PrecisionMul,
				Rates:        _Rates,
			}
			swap, err = stablebase(genesis, sbc)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))

		})

		It("test_withdraw_all_coins : payToken", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {
						for pi := uint8(0); pi < N; pi++ { //	payTokenIndex
							err := beforeEach()
							Expect(err).To(Succeed())
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
							swap, err = stablebase(genesis, sbc)
							Expect(err).To(Succeed())

							stableAddInitialLiquidity(genesis, alice)
							stableMint(genesis, bob)
							stableApprove(genesis, bob)
							cn, cdx, ctx, _ := initChain(genesis, admin)
							ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
							ctx, _ = setFees(cn, ctx, swap, fee, adminFee, winnerFee, uint64(86400), aliceKey)

							// 0,1 -> 1,2 -> 2,0
							for i := uint8(0); i < N; i++ {
								send := i
								recv := i + 1
								if recv == N {
									recv = 0
								}
								_, err := Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, _InitialAmounts[send], ZeroAmount, ZeroAddress})
								GPrintf("%+v", err)
								Expect(err).To(Succeed())
							}

							admin_balances := MakeAmountSlice(N)
							for i := uint8(0); i < N; i++ {
								is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
								Expect(err).To(Succeed())
								admin_balances[i].Set(is[0].(*amount.Amount).Int)
							}

							_, err = Exec(ctx, alice, swap, "WithdrawAdminFees", []interface{}{})
							Expect(err).To(Succeed())

							winner, _ := ViewAddress(ctx, swap, "Winner")

							for i := uint8(0); i < N; i++ {
								if i == pi {
									wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(winnerFee)), trade.FEE_DENOMINATOR))
									Expect(tokenBalanceOf(ctx, stableTokens[i], winner)).To(Equal(wF))
									Expect(tokenBalanceOf(ctx, stableTokens[i], alice)).To(Equal(admin_balances[i].Sub(wF)))
								} else {
									Expect(tokenBalanceOf(ctx, stableTokens[i], winner)).To(Equal(ZeroAmount))
									Expect(tokenBalanceOf(ctx, stableTokens[i], alice)).To(Equal(ZeroAmount))
								}
							}

							RemoveChain(cdx)
							afterEach()
						}
					}
				}
			}
		})
	})

	Describe("test_exchange_reverts.py", func() {

		It("GetDy non-positive input amount", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			_, err := Exec(genesis, alice, swap, "GetDy", []interface{}{uint8(0), uint8(1), ToAmount(big.NewInt(-1)), ZeroAddress})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			afterEach()
		})

		It("Exchange non-positive input amount", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			_, err := Exec(genesis, alice, swap, "Exchange", []interface{}{uint8(0), uint8(1), ToAmount(big.NewInt(-1)), ZeroAmount, ZeroAddress})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			afterEach()
		})

		It("test_insufficient_balance", func() {
			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)
					stableApprove(genesis, bob)

					amt := ToAmount(Pow10(decimals[send]))
					tokenMint(genesis, stableTokens[send], bob, amt)

					_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{send, recv, ToAmount(Add(amt.Int, big.NewInt(1))), ZeroAmount, ZeroAddress})
					Expect(err).To(MatchError("the token holding quantity is insufficient"))

					afterEach()
				}
			}
		})

		It("test_min_dy_too_high", func() {
			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)
					stableApprove(genesis, bob)

					amt := ToAmount(Pow10(decimals[send]))
					tokenMint(genesis, stableTokens[send], bob, amt)

					is, err := Exec(genesis, bob, swap, "GetDy", []interface{}{send, recv, amt, ZeroAddress})
					Expect(err).To(Succeed())

					min_dy := is[0].(*amount.Amount)

					_, err = Exec(genesis, bob, swap, "Exchange", []interface{}{send, recv, amt, ToAmount(AddC(min_dy.Int, 2)), ZeroAddress})
					Expect(err).To(HaveOccurred())

					afterEach()
				}
			}
		})

		It("test_same_coin", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableApprove(genesis, bob)

				_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{idx, idx, ZeroAmount, ZeroAmount, ZeroAddress})
				Expect(err).To(MatchError("Exchange: OUT"))

				afterEach()
			}
		})

		It("test_i_below_zero", func() {
			// uint8 can't be below zero
		})

		It("test_i_above_n_coins", func() {
			idxes := []uint8{9, 140, 255}
			for i := 0; i < len(idxes); i++ {
				idx := idxes[i]
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableApprove(genesis, bob)

				_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{idx, uint8(0), ZeroAmount, ZeroAmount, ZeroAddress})
				Expect(err).To(MatchError("Exchange: IN"))

				afterEach()
			}
		})

		It("test_j_below_zero", func() {
			//uint8 can'b below zero
		})

		It("test_j_above_n_coins", func() {
			idxes := []uint8{9, 100}
			for i := 0; i < len(idxes); i++ {
				idx := idxes[i]
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableApprove(genesis, bob)

				_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{uint8(0), idx, ZeroAmount, ZeroAmount, ZeroAddress})
				Expect(err).To(MatchError("Exchange: OUT"))

				afterEach()
			}
		})

		It("test_nonpayable", func() {
			// no ether
		})

	})

	Describe("test_exchange.py", func() {

		It("test_exchange", func() {
			var err error

			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}

					fees := []float64{0., 0.04, 0.1337, 0.5}
					admin_fees := []float64{0., 0.04, 0.1337, 0.5}
					//winner_fees := []float64{0., 0.04, 0.1337, 0.5}
					winner_fees := []float64{0.}

					for k := 0; k < len(fees); k++ {
						fee := fees[k]
						for m := 0; m < len(admin_fees); m++ {
							admin_fee := admin_fees[m]
							for n := 0; n < len(winner_fees); n++ {
								winner_fee := winner_fees[n]

								beforeEachStable()
								stableAddInitialLiquidity(genesis, alice)
								stableApprove(genesis, bob)
								cn, cdx, ctx, _ := initChain(genesis, admin)

								ctx, err = setFees(cn, ctx, swap, uint64(fee*float64(trade.FEE_DENOMINATOR)), uint64(admin_fee*float64(trade.FEE_DENOMINATOR)), uint64(winner_fee*float64(trade.FEE_DENOMINATOR)), uint64(86400), aliceKey)
								Expect(err).To(Succeed())

								amt := ToAmount(Pow10(decimals[send]))
								tokenMint(ctx, stableTokens[send], bob, amt)

								_, err = Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, amt, ZeroAmount, ZeroAddress})
								Expect(err).To(Succeed())

								Expect(tokenBalanceOf(ctx, stableTokens[send], bob)).To(Equal(ZeroAmount))
								received, _ := tokenBalanceOf(ctx, stableTokens[recv], bob)

								received_float, _ := new(big.Float).SetInt(received.Int).Float64()
								r := received_float / float64(Pow10(decimals[recv]).Uint64())
								if 1-math.Max(math.Pow10(-4), 1/received_float)-fee >= r || r >= 1-fee {
									Fail("Fee")
								}

								expected_admin_fee := ToFloat64(Pow10(decimals[recv])) * fee * admin_fee
								a_fees, err := stableGetAdminBalances(ctx)
								Expect(err).To(Succeed())
								for i := uint8(0); i < N; i++ {
									is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
									Expect(err).To(Succeed())
									if a_fees[i].Cmp(Zero) == 0 {
										Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
									} else {
										Expect(is[0].(*amount.Amount)).To(Equal(a_fees[i]))
									}
								}

								if expected_admin_fee >= 1 {
									if !Approx(expected_admin_fee/ToFloat64(a_fees[recv].Int), 1, math.Max(1e-3, 1/(expected_admin_fee-1.1))) {
										Fail("Apporx")
									}
								} else {
									if a_fees[recv].Cmp(big.NewInt(1)) > 0 {
										Fail("Admin Fee")
									}
								}
								RemoveChain(cdx)
								afterEach()
							}
						}
					}
				}
			}
		})

		It("test_min_dy", func() {
			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)
					stableApprove(genesis, bob)

					amt := ToAmount(Pow10(decimals[send]))
					tokenMint(genesis, stableTokens[send], bob, amt)

					is, err := Exec(genesis, bob, swap, "GetDy", []interface{}{send, recv, amt, ZeroAddress})
					Expect(err).To(Succeed())

					min_dy := is[0].(*amount.Amount)

					_, err = Exec(genesis, bob, swap, "Exchange", []interface{}{send, recv, amt, ToAmount(Sub(min_dy.Int, big.NewInt(1))), ZeroAddress})
					Expect(err).To(Succeed())

					received, _ := tokenBalanceOf(genesis, stableTokens[recv], bob)

					if Abs(Sub(received.Int, min_dy.Int)).Cmp(big.NewInt(1)) > 0 {
						Fail("Recived")
					}

					afterEach()
				}
			}
		})
	})

	Describe("test_get_virtual_price.py", func() {
		It("test_number_go_up", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)
			stableApprove(genesis, bob)
			stableMint(genesis, bob)
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(86400), aliceKey)
			ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

			is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
			Expect(err).To(Succeed())

			virtual_price := is[0].(*amount.Amount)

			for i := uint8(0); i < N; i++ {
				amts := MakeSlice(N)
				amts[i] = Clone(_InitialAmounts[i].Int)

				_, err := Exec(ctx, bob, swap, "AddLiquidity", []interface{}{ToAmounts(amts), ZeroAmount})
				Expect(err).To(Succeed())

				is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
				Expect(err).To(Succeed())

				// admin_fee = 100% 인경우 계산오차가 차이날 수 있음  new_virtual_price >= virtual_price - 1
				new_virtual_price := is[0].(*amount.Amount)
				if AddC(new_virtual_price.Int, 2).Cmp(virtual_price.Int) <= 0 {
					GPrintln("new_virtual_price, virtual_price", new_virtual_price, virtual_price)
					Fail("Vitrual Price")
				}
				virtual_price.Int.Set(new_virtual_price.Int)
			}
			RemoveChain(cdx)
			afterEach()
		})

		It("test_remove_one_coin", func() {

			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableApprove(genesis, bob)
				stableMint(genesis, bob)
				cn, cdx, ctx, _ := initChain(genesis, admin)
				ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(86400), aliceKey)
				ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

				amt, err := tokenBalanceOf(genesis, swap, alice)
				Expect(err).To(Succeed())
				amt = amt.DivC(10)

				is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
				Expect(err).To(Succeed())

				virtual_price := is[0].(*amount.Amount)

				_, err = Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
				Expect(err).To(Succeed())

				is, err = Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
				Expect(err).To(Succeed())
				new_virtual_price := is[0].(*amount.Amount)

				// admin_fee = 100% 인경우 계산오차가 차이날 수 있음  new_virtual_price >= virtual_price - 1
				if AddC(new_virtual_price.Int, 2).Cmp(virtual_price.Int) <= 0 {
					GPrintln("new_virtual_price, virtual_price", new_virtual_price, virtual_price)
					Fail("Virtual Price")
				}

				RemoveChain(cdx)
				afterEach()
			}

		})

		It("test_remove_imbalance", func() {

			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableApprove(genesis, bob)
				stableMint(genesis, bob)
				cn, cdx, ctx, _ := initChain(genesis, admin)
				ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(3*86400), aliceKey)
				ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

				amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amts[i] = amts[i].DivC(2)
				}
				amts[idx] = amount.NewAmount(0, 0)

				is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
				Expect(err).To(Succeed())

				virtual_price := is[0].(*amount.Amount)

				_, err = Exec(ctx, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)})
				Expect(err).To(Succeed())

				is, err = Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
				Expect(err).To(Succeed())

				new_virtual_price := is[0].(*amount.Amount)

				if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
					Fail("Virtual Price")
				}

				RemoveChain(cdx)
				afterEach()
			}

		})

		It("test_remove", func() {

			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)
			stableApprove(genesis, bob)
			stableMint(genesis, bob)
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(3*86400), aliceKey)
			ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

			withdraw_amount := DivC(Sum(ToBigInts(_InitialAmounts)), 2)

			is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
			Expect(err).To(Succeed())

			virtual_price := is[0].(*amount.Amount)

			_, err = Exec(ctx, alice, swap, "RemoveLiquidity", []interface{}{withdraw_amount, MakeAmountSlice(N)})
			Expect(err).To(Succeed())

			is, err = Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
			Expect(err).To(Succeed())

			new_virtual_price := is[0].(*amount.Amount)

			if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
				Fail("Virtual Price")
			}

			RemoveChain(cdx)
			afterEach()
		})

		It("test_exchange", func() {

			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)
					stableApprove(genesis, bob)
					stableMint(genesis, bob)
					cn, cdx, ctx, _ := initChain(genesis, admin)
					ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(3*86400), aliceKey)
					ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

					is, err := Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
					Expect(err).To(Succeed())
					virtual_price := is[0].(*amount.Amount)

					amt := ToAmount(Pow10(decimals[send]))

					_, err = Exec(ctx, bob, swap, "Exchange", []interface{}{send, recv, amt, ZeroAmount, ZeroAddress})
					Expect(err).To(Succeed())

					is, err = Exec(ctx, bob, swap, "GetVirtualPrice", []interface{}{})
					Expect(err).To(Succeed())

					new_virtual_price := is[0].(*amount.Amount)

					if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
						Fail("Virtual Price")
					}

					RemoveChain(cdx)
					afterEach()
				}
			}
		})

		It("test_exchange_underlying", func() {
			// no lending
		})

	})

	Describe("test_kill.py", func() {
		BeforeEach(func() {
			beforeEachStable()
		})

		AfterEach(func() {
			afterEach()
		})

		It("kill_me", func() {
			_, err := Exec(genesis, alice, swap, "KillMe", []interface{}{})
			Expect(err).To(Succeed())
		})

		It("test_kill_approaching_deadline", func() {
			// no dead line
		})

		It("test_kill_only_owner", func() {
			_, err := Exec(genesis, bob, swap, "KillMe", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_kill_beyond_deadline", func() {
			// no deadline
		})

		It("test_kill_and_unkill", func() {
			Exec(genesis, alice, swap, "KillMe", []interface{}{})

			_, err := Exec(genesis, alice, swap, "UnkillMe", []interface{}{})
			Expect(err).To(Succeed())
		})

		It("test_unkill_without_kill", func() {
			_, err := Exec(genesis, alice, swap, "UnkillMe", []interface{}{})
			Expect(err).To(Succeed())

		})

		It("test_unkill_only_owner", func() {
			_, err := Exec(genesis, bob, swap, "UnkillMe", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("test_remove_liquidity", func() {
			stableAddInitialLiquidity(genesis, alice)

			Exec(genesis, alice, swap, "KillMe", []interface{}{})
			_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), MakeAmountSlice(N)})
			Expect(err).To(Succeed())

		})

		It("test_remove_liquidity_imbalance", func() {
			Exec(genesis, alice, swap, "KillMe", []interface{}{})
			_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{MakeAmountSlice(N), ZeroAmount})
			Expect(err).To(MatchError("Exchange: KILLED"))

		})

		It("test_remove_liquidity_one_coin", func() {
			Exec(genesis, alice, swap, "KillMe", []interface{}{})
			_, err := Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{ZeroAmount, uint8(0), ZeroAmount})
			Expect(err).To(MatchError("Exchange: KILLED"))

		})

		It("test_exchange", func() {

			Exec(genesis, alice, swap, "KillMe", []interface{}{})
			_, err := Exec(genesis, alice, swap, "Exchange", []interface{}{uint8(0), uint8(0), ZeroAmount, ZeroAmount, ZeroAddress})
			Expect(err).To(MatchError("Exchange: KILLED"))

		})
	})

	Describe("test_modify_fees.py", func() {
		It("test_commit", func() {

			fees := [4][3]uint64{{0, 0, 0}, {23, 42, 18}, {trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}, {1, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}}

			for k := 0; k < len(fees); k++ {
				beforeEachStable()

				fee := fees[k][0]
				admin_fee := fees[k][1]
				winner_fee := fees[k][2]
				delay := uint64(3 * 86400)
				_, err := Exec(genesis, alice, swap, "CommitNewFee", []interface{}{fee, admin_fee, winner_fee, delay})
				Expect(err).To(Succeed())

				is, err := Exec(genesis, alice, swap, "AdminActionsDeadline", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(delay))

				is, err = Exec(genesis, alice, swap, "FutureFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

				is, err = Exec(genesis, alice, swap, "FutureAdminFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(admin_fee))

				is, err = Exec(genesis, alice, swap, "FutureWinnerFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(winner_fee))

				afterEach()
			}

		})

		It("test_commit_only_owner", func() {
			beforeEachStable()

			_, err := Exec(genesis, bob, swap, "CommitNewFee", []interface{}{uint64(23), uint64(42), uint64(18), uint64(3 * 86400)})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			afterEach()
		})

		It("test_commit_already_active", func() {
			beforeEachStable()

			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

			_, err := Exec(ctx, alice, swap, "CommitNewFee", []interface{}{uint64(23), uint64(42), uint64(18), uint64(0)})

			_, err = Exec(ctx, alice, swap, "CommitNewFee", []interface{}{uint64(23), uint64(42), uint64(18), uint64(0)})
			Expect(err).To(MatchError("Exchange: ADMIN_ACTIONS_DEADLINE"))

			RemoveChain(cdx)
			afterEach()
		})

		It("test_commit_admin_fee_too_high", func() {
			fees := [2]uint64{2 * trade.MAX_ADMIN_FEE, 3 * trade.MAX_ADMIN_FEE}

			for k := 0; k < len(fees); k++ {
				beforeEachStable()

				admin_fee := fees[k]

				_, err := Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), admin_fee, uint64(0), uint64(0)})
				Expect(err).To(MatchError("Exchange: FUTURE_ADMIN_FEE_EXCEED_MAXADMINFEE"))

				afterEach()
			}
		})

		It("test_commit_fee_too_high", func() {
			fees := [2]uint64{2 * trade.MAX_ADMIN_FEE, 3 * trade.MAX_ADMIN_FEE}

			for k := 0; k < len(fees); k++ {
				beforeEachStable()

				fee := fees[k]

				_, err := Exec(genesis, alice, swap, "CommitNewFee", []interface{}{fee, uint64(0), uint64(0), uint64(86400)})
				Expect(err).To(MatchError("Exchange: FUTURE_FEE_EXCEED_MAXFEE"))

				afterEach()
			}
		})

		It("test_commit_winner_fee_too_high", func() {
			fees := [2]uint64{2 * trade.MAX_WINNER_FEE, 3 * trade.MAX_WINNER_FEE}

			for k := 0; k < len(fees); k++ {
				beforeEachStable()

				winner_fee := fees[k]

				_, err := Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), uint64(0), winner_fee, uint64(86400)})
				Expect(err).To(MatchError("Exchange: FUTURE_WINNER_FEE_EXCEED_MAXADMINFEE"))

				afterEach()
			}
		})

		It("test_apply", func() {
			fees := [4][3]uint64{{0, 0, 0}, {23, 42, 18}, {trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}, {1, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}}

			for k := 0; k < len(fees); k++ {
				beforeEachStable()

				fee := fees[k][0]
				admin_fee := fees[k][1]
				winner_fee := fees[k][2]
				delay := uint64(86400)

				Exec(genesis, alice, swap, "CommitNewFee", []interface{}{fee, admin_fee, winner_fee, delay})
				cn, cdx, ctx, _ := initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

				_, err := Exec(ctx, alice, swap, "ApplyNewFee", []interface{}{})
				Expect(err).To(Succeed())

				is, err := Exec(ctx, alice, swap, "AdminActionsDeadline", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				is, err = Exec(genesis, alice, swap, "FutureFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

				is, err = Exec(genesis, alice, swap, "FutureAdminFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(admin_fee))

				is, err = Exec(genesis, alice, swap, "FutureWinnerFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(winner_fee))

				RemoveChain(cdx)
				afterEach()
			}
		})

		It("test_apply_only_owner", func() {
			beforeEachStable()

			delay := uint64(0)
			Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), uint64(0), uint64(0), delay})
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

			_, err := Exec(ctx, bob, swap, "ApplyNewFee", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			RemoveChain(cdx)
			afterEach()
		})

		It("test_apply_insufficient_time", func() {
			_DELAY := uint64(3 * 86400)
			delays := [2]uint64{0, _DELAY - 2}

			for k := 0; k < len(delays); k++ {
				beforeEachStable()

				delay := delays[k]

				Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), uint64(0), uint64(0), _DELAY})
				cn, cdx, ctx, _ := initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

				_, err := Exec(ctx, alice, swap, "ApplyNewFee", []interface{}{})
				Expect(err).To(MatchError("Exchange: ADMIN_ACTIONS_DEADLINE"))

				RemoveChain(cdx)
				afterEach()
			}
		})

		It("test_apply_no_action", func() {
			beforeEachStable()

			_, err := Exec(genesis, alice, swap, "ApplyNewFee", []interface{}{})
			Expect(err).To(MatchError("Exchange: NO_ACTIVE_ACTION"))

			afterEach()
		})

		It("test_revert", func() {
			beforeEachStable()

			Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), uint64(0), uint64(0), uint64(3 * 86400)})

			_, err := Exec(genesis, alice, swap, "RevertNewFee", []interface{}{})
			Expect(err).To(Succeed())

			afterEach()
		})

		It("test_revert_only_owner", func() {
			beforeEachStable()

			Exec(genesis, alice, swap, "CommitNewFee", []interface{}{uint64(0), uint64(0), uint64(0), uint64(3 * 86400)})

			_, err := Exec(genesis, bob, swap, "RevertNewFee", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			afterEach()
		})

		It("test_revert_without_commit", func() {
			beforeEachStable()

			_, err := Exec(genesis, alice, swap, "RevertNewFee", []interface{}{})
			Expect(err).To(Succeed())

			is, err := Exec(genesis, alice, swap, "AdminActionsDeadline", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			afterEach()
		})

		It("test_withdraw_only_owner", func() {
			beforeEachStable()

			_, err := Exec(genesis, bob, swap, "WithdrawAdminFees", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			afterEach()
		})
	})

	Describe("test_nonpayable.py", func() {
		It("test_fallback_reverts", func() {
			// no fallback in Meverse
		})
	})

	Describe("test_ramp_A_precise.py", func() {
		var (
			cdx              int
			err              error
			cn               *chain.Chain
			ctx              *types.Context
			timestamp_second uint64
		)

		BeforeEach(func() {
			genesis = types.NewEmptyContext()
			stableTokens = DeployTokens(genesis, classMap["Token"], N, admin)
			cn, cdx, ctx, _ = initChain(genesis, admin)
			timestamp_second = uint64(time.Now().UnixNano()) / uint64(time.Second)
			ctx, _ = Sleep(cn, ctx, nil, timestamp_second, aliceKey)
			swap, err = stablebase(ctx, stableBaseContruction())
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("test_ramp_A", func() {
			is, err := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			Expect(err).To(Succeed())
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
			future_time := ctx.LastTimestamp()/uint64(time.Second) + trade.MIN_RAMP_TIME + 5

			_, err = Exec(ctx, alice, swap, "RampA", []interface{}{MulC(initial_A, 2), future_time})
			Expect(err).To(Succeed())

			is, err = Exec(ctx, alice, swap, "InitialA", []interface{}{})
			Expect(DivC(is[0].(*big.Int), trade.A_PRECISION)).To(Equal(initial_A))

			is, err = Exec(ctx, alice, swap, "FutureA", []interface{}{})
			Expect(DivC(is[0].(*big.Int), trade.A_PRECISION)).To(Equal(MulC(initial_A, 2)))

			is, err = Exec(ctx, alice, swap, "InitialATime", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(timestamp_second))

			is, err = Exec(ctx, alice, swap, "FutureATime", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(future_time))

		})

		It("test_ramp_A_final", func() {
			is, _ := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

			future_time := ctx.LastTimestamp()/uint64(time.Second) + 1000000

			Exec(ctx, alice, swap, "RampA", []interface{}{MulC(initial_A, 2), future_time})

			ctx, _ = Sleep(cn, ctx, nil, 1000000, aliceKey)

			is, _ = Exec(ctx, alice, swap, "A", []interface{}{})
			Expect(is[0].(*big.Int)).To(Equal(MulC(initial_A, 2)))

		})

		It("test_ramp_A_value_up", func() {
			is, _ := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

			future_time := ctx.LastTimestamp()/uint64(time.Second) + 1000000

			Exec(ctx, alice, swap, "RampA", []interface{}{MulC(initial_A, 2), future_time})

			initial_time := ctx.LastTimestamp() / uint64(time.Second)
			duration := future_time - initial_time

			for ctx.LastTimestamp()/uint64(time.Second) < future_time {
				ctx, _ = Sleep(cn, ctx, nil, 100000, aliceKey)
				elapsed := float64(ctx.LastTimestamp()/uint64(time.Second)-uint64(initial_time)) / float64(duration)
				expected := AddC(initial_A, int64(ToFloat64(initial_A)*elapsed))
				is, err := Exec(ctx, alice, swap, "A", []interface{}{})
				Expect(err).To(Succeed())
				A := is[0].(*big.Int)
				if 0.999 >= ToFloat64(expected)/ToFloat64(A) {
					Fail("Lower")
				}
				if ToFloat64(expected)/ToFloat64(A) > 1 {
					Fail("Upper")
				}
			}
		})

		It("test_ramp_A_value_down", func() {
			is, _ := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

			future_time := ctx.LastTimestamp()/uint64(time.Second) + 1000000

			Exec(ctx, alice, swap, "RampA", []interface{}{DivC(initial_A, 10), future_time})

			initial_time := ctx.LastTimestamp() / uint64(time.Second)
			duration := future_time - initial_time

			for ctx.LastTimestamp()/uint64(time.Second) < future_time {
				ctx, _ = Sleep(cn, ctx, nil, 100000, aliceKey)
				elapsed := float64(ctx.LastTimestamp()/uint64(time.Second)-uint64(initial_time)) / float64(duration)
				expected := SubC(initial_A, int64(elapsed*ToFloat64(initial_A)/10.*9.))
				is, _ := Exec(ctx, alice, swap, "A", []interface{}{})
				A := is[0].(*big.Int)

				if expected.Cmp(Zero) == 0 {
					Expect(A).To(Equal(DivC(initial_A, 10)))
				} else {
					if Abs(Sub(A, expected)).Cmp(big.NewInt(1)) > 0 {
						Fail("A")
					}
				}
			}
		})

		It("test_stop_ramp_A", func() {
			is, _ := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
			future_time := ctx.LastTimestamp()/uint64(time.Second) + 1000000
			Exec(ctx, alice, swap, "RampA", []interface{}{DivC(initial_A, 10), future_time})

			ctx, _ = Sleep(cn, ctx, nil, 31337, aliceKey)

			is, _ = Exec(ctx, alice, swap, "A", []interface{}{})
			current_A := is[0].(*big.Int)

			_, err = Exec(ctx, alice, swap, "StopRampA", []interface{}{})

			is, err = Exec(ctx, alice, swap, "InitialA", []interface{}{})
			Expect(DivC(is[0].(*big.Int), 100)).To(Equal(current_A))

			is, err = Exec(ctx, alice, swap, "FutureA", []interface{}{})
			Expect(DivC(is[0].(*big.Int), 100)).To(Equal(current_A))

			is, err = Exec(ctx, alice, swap, "InitialATime", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(ctx.LastTimestamp() / uint64(time.Second)))

			is, err = Exec(ctx, alice, swap, "FutureATime", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(ctx.LastTimestamp() / uint64(time.Second)))
		})

		It("test_ramp_A_only_owner", func() {
			future_time := ctx.LastTimestamp()/uint64(time.Second) + 1000000
			_, err = Exec(ctx, bob, swap, "RampA", []interface{}{Zero, future_time})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("test_ramp_A_insufficient_time", func() {
			future_time := ctx.LastTimestamp()/uint64(time.Second) + trade.MIN_RAMP_TIME - 1
			_, err = Exec(ctx, alice, swap, "RampA", []interface{}{Zero, future_time})
			Expect(err).To(MatchError("Exchange: Ramp_A_BIG"))

		})

		It("test_stop_ramp_A_only_owner", func() {
			_, err = Exec(ctx, bob, swap, "StopRampA", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})
	})

	Describe("test_remove_liquidity.py", func() {

		It("non-positive input amount", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			_, err := Exec(genesis, alice, swap, "CalcWithdrawCoins", []interface{}{ToAmount(big.NewInt(-1))})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			afterEach()
		})

		It("non-positive input amount", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{ToAmount(big.NewInt(-1)), MakeAmountSlice(N)})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			afterEach()
		})

		It("test_remove_liquidity", func() {
			min_amt := []int{0, 1}

			for k := 0; k < len(min_amt); k++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				min_amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					min_amts[i].Set(Mul(big.NewInt(int64(min_amt[k])), min_amts[i].Int))
				}
				_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), min_amts})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					Expect(tokenBalanceOf(genesis, stableTokens[i], alice)).To(Equal(_InitialAmounts[i]))
					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(ZeroAmount))
				}

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(ZeroAmount))
				Expect(tokenTotalSupply(genesis, swap)).To(Equal(ZeroAmount))

				afterEach()
			}
		})

		It("test_remove_partial", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			withdraw_amount := Div(Sum(ToBigInts(_InitialAmounts)), big.NewInt(2))

			min_amts := MakeAmountSlice(N)
			_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{withdraw_amount, min_amts})
			Expect(err).To(Succeed())

			for i := uint8(0); i < N; i++ {
				pool_balance, _ := tokenBalanceOf(genesis, stableTokens[i], swap)
				alice_balance, _ := tokenBalanceOf(genesis, stableTokens[i], alice)

				Expect(Add(pool_balance.Int, alice_balance.Int)).To(Equal(_InitialAmounts[i].Int))
			}

			Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(ToAmount(withdraw_amount))))
			Expect(tokenTotalSupply(genesis, swap)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(ToAmount(withdraw_amount))))

			afterEach()
		})

		It("test_below_min_amount", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				min_amts := CloneAmountSlice(_InitialAmounts)
				min_amts[idx] = ToAmount(Add(min_amts[idx].Int, big.NewInt(1)))

				_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), min_amts})
				Expect(err).To(MatchError("Exchange: WITHDRAWAL_RESULTED_IN_FEWER_COINS_THAN_EXPECTED"))

				afterEach()
			}
		})

		It("test_amount_exceeds_balance", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			min_amts := MakeAmountSlice(N)

			_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(uint64(N)*uint64(_BaseAmount), 1), min_amts})
			Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))

			afterEach()
		})

		It("test_event", func() {
			// no event
		})
	})

	Describe("test_remove_liquidity_imbalance.py", func() {

		It("non-positive input amount", func() {
			for i := uint8(0); i < N; i++ {
				amts := MakeAmountSlice(N)
				amts[i].Set(big.NewInt(-1))

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, MaxUint256})
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				afterEach()
			}
		})

		It("test_remove_balanced", func() {
			divisor := []int{2, 5, 10}

			for k := uint8(0); k < N; k++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				amts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(Div(_InitialAmounts[i].Int, big.NewInt(int64(divisor[k]))))
				}
				max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).DivC(int64(divisor[k]))
				_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, ToAmount(Add(max_burn.Int, big.NewInt(1)))})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					Expect(tokenBalanceOf(genesis, stableTokens[i], alice)).To(Equal(amts[i]))
					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
				}

				alice_pool_token_balance, _ := tokenBalanceOf(genesis, swap, alice)
				pool_total_supply, _ := tokenTotalSupply(genesis, swap)

				if !(Abs(alice_pool_token_balance.Sub(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(max_burn)).Int).Cmp(big.NewInt(1)) <= 0) {
					Fail("Lower")
				}

				if !(Abs(pool_total_supply.Sub(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(max_burn)).Int).Cmp(big.NewInt(1)) <= 0) {
					Fail("Upper")
				}

				afterEach()
			}
		})

		It("test_remove_some", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				amts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(_InitialAmounts[i].DivC(2).Int)
				}
				amts[idx].Set(Zero)

				max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)
				_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, max_burn})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					balance, _ := tokenBalanceOf(genesis, stableTokens[i], alice)
					if balance.Cmp(amts[i].Int) != 0 {
						Fail("Balance")
					}

					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
				}

				actual_balance, _ := tokenBalanceOf(genesis, swap, alice)
				actual_total_supply, _ := tokenTotalSupply(genesis, swap)

				ideal_balance := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amount.NewAmount(uint64(_BaseAmount)/2*uint64(N-1), 0))

				Expect(actual_balance).To(Equal(actual_total_supply))

				if ToFloat64(ideal_balance.Int)*0.99 >= ToFloat64(actual_balance.Int) {
					Fail("Lower")
				}
				if actual_balance.Int.Cmp(ideal_balance.Int) >= 0 {
					Fail("Upper")
				}

				afterEach()
			}
		})

		It("test_remove_one", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				amts := MakeAmountSlice(N)
				amts[idx].Set(_InitialAmounts[idx].DivC(2).Int)

				max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)
				_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, max_burn})
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					Expect(tokenBalanceOf(genesis, stableTokens[i], alice)).To(Equal(amts[i]))
					Expect(tokenBalanceOf(genesis, stableTokens[i], swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
				}

				actual_balance, _ := tokenBalanceOf(genesis, swap, alice)
				actual_total_supply, _ := tokenTotalSupply(genesis, swap)

				ideal_balance := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amount.NewAmount(uint64(_BaseAmount)/2, 0))

				Expect(actual_balance).To(Equal(actual_total_supply))
				if ToFloat64(ideal_balance.Int)*0.99 >= ToFloat64(actual_balance.Int) {
					Fail("Lower")
				}
				if actual_balance.Int.Cmp(ideal_balance.Int) >= 0 {
					Fail("Upper")
				}

				afterEach()
			}
		})

		It("test_exceed_max_burn", func() {
			divisor := []int{1, 2, 10}

			for k := uint8(0); k < N; k++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				amts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(Div(_InitialAmounts[i].Int, big.NewInt(int64(divisor[k]))))
				}
				max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).DivC(int64(divisor[k]))

				_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, ToAmount(SubC(max_burn.Int, 1))})
				Expect(err).To(MatchError("Exchange: SLIPPAGE"))

				afterEach()
			}
		})

		It("test_cannot_remove_zero", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			amts := MakeAmountSlice(N)

			_, err := Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, ZeroAmount})
			Expect(err).To(MatchError("Exchange: ZERO_TOKEN_BURN"))

			afterEach()
		})

		It("test_no_totalsupply", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			amts := MakeAmountSlice(N)
			total_supply, _ := tokenTotalSupply(genesis, swap)

			_, err := Exec(genesis, alice, swap, "RemoveLiquidity", []interface{}{total_supply, amts})
			Expect(err).To(Succeed())

			_, err = Exec(genesis, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, ZeroAmount})
			Expect(err).To(MatchError("Exchange: D0_IS_ZERO"))

			afterEach()
		})

		It("test_event", func() {
			// no event
		})
	})

	Describe("test_remove_liquidity_one_coin.py", func() {

		It("CalcWithdrawOneCoin non-positive input amount", func() {
			for i := uint8(0); i < N; i++ {
				amts := MakeAmountSlice(N)
				amts[i].Set(big.NewInt(-1))

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, alice, swap, "CalcWithdrawOneCoin", []interface{}{ToAmount(big.NewInt(-1)), uint8(1)})
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				afterEach()
			}
		})

		It("RemoveLiquidityOneCoin non-positive input amount", func() {
			for i := uint8(0); i < N; i++ {
				amts := MakeAmountSlice(N)
				amts[i].Set(big.NewInt(-1))

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{ToAmount(big.NewInt(-1)), uint8(1), ZeroAmount})
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				afterEach()
			}
		})

		It("CalcWithdrawOneCoin index error", func() {
			for i := uint8(0); i < N; i++ {
				amts := MakeAmountSlice(N)
				amts[i].Set(big.NewInt(-1))

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, alice, swap, "CalcWithdrawOneCoin", []interface{}{big.NewInt(1), uint8(N)})
				Expect(err).To(MatchError("Exchange: IDX"))

				afterEach()
			}
		})

		It("RemoveLiquidityOneCoin index error", func() {
			for i := uint8(0); i < N; i++ {
				amts := MakeAmountSlice(N)
				amts[i].Set(big.NewInt(-1))

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{big.NewInt(1), uint8(N), ZeroAmount})
				Expect(err).To(MatchError("Exchange: IDX"))

				afterEach()
			}
		})

		It("test_amount_received", func() {

			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				cn, cdx, ctx, err := initChain(genesis, admin)
				Expect(err).To(Succeed())
				ctx, err = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				Expect(err).To(Succeed())
				ctx, err = setFees(cn, ctx, swap, 0, 0, 0, uint64(86400), aliceKey)
				Expect(err).To(Succeed())

				rate_mod := 1.00001

				_, err = Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amount.NewAmount(1, 0), idx, ZeroAmount})
				Expect(err).To(Succeed())

				balance, _ := tokenBalanceOf(ctx, stableTokens[idx], alice)

				if big.NewInt(int64(ToFloat64(Pow10(decimals[idx]))/rate_mod)).Cmp(balance.Int) > 0 {
					Fail("Lower")
				}

				if balance.Int.Cmp(Pow10(decimals[idx])) > 0 {
					Fail("Upper")
				}

				RemoveChain(cdx)
				afterEach()
			}
		})

		It("test_lp_token_balance", func() {
			divisors := []int{1, 5, 42}
			for k := 0; k < len(divisors); k++ {
				divisor := divisors[k]

				for idx := uint8(0); idx < N; idx++ {
					beforeEachStable()
					stableAddInitialLiquidity(genesis, alice)

					balance, _ := tokenBalanceOf(genesis, swap, alice)
					amt := balance.DivC(int64(divisor))

					_, err := Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
					Expect(err).To(Succeed())

					balance, _ = tokenBalanceOf(genesis, swap, alice)
					if balance.Cmp(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amt).Int) != 0 {
						Fail("Balance")
					}

					afterEach()
				}
			}
		})

		It("test_expected_vs_actual", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				balance, _ := tokenBalanceOf(genesis, swap, alice)
				amt := balance.DivC(int64(10))

				is, err := Exec(genesis, alice, swap, "CalcWithdrawOneCoin", []interface{}{amt, idx})
				Expect(err).To(Succeed())
				expected := is[0].(*amount.Amount)

				_, err = Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, stableTokens[idx], alice)).To(Equal(expected))
				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amt)))

				afterEach()
			}
		})

		It("test_below_min_amount", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				amt, _ := tokenBalanceOf(genesis, swap, alice)

				is, err := Exec(genesis, alice, swap, "CalcWithdrawOneCoin", []interface{}{amt, idx})
				Expect(err).To(Succeed())

				expected := is[0].(*amount.Amount)

				_, err = Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, AddC(expected.Int, 1)})
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_OUTPUT_AMOUNT"))

				afterEach()
			}
		})

		It("test_amount_exceeds_balance", func() {
			for idx := uint8(0); idx < N; idx++ {
				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)

				_, err := Exec(genesis, bob, swap, "RemoveLiquidityOneCoin", []interface{}{amount.NewAmount(0, 1), idx, ZeroAmount})
				Expect(err).To(HaveOccurred())

				afterEach()
			}
		})

		It("test_below_zero", func() {
			//uint8 can't be below zero
		})

		It("test_above_n_coins", func() {
			beforeEachStable()
			stableAddInitialLiquidity(genesis, alice)

			_, err := Exec(genesis, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amount.NewAmount(0, 1), N, ZeroAmount})
			Expect(err).To(MatchError("Exchange: IDX"))

			afterEach()
		})

		It("test_event", func() {
			// no event
		})
	})

	Describe("test_xfer_to_contract.py", func() {

		It("test_unexpected_eth", func() {
			//동일한 주소를 갖는 swap contract를 다시 만들었을 경우 그 값이 보존되는 지
			genesis = types.NewEmptyContext()

			stableTokens = DeployTokens(genesis, classMap["Token"], N, admin)

			var err error
			sbc := stableBaseContruction()
			swap, err = stablebase(genesis, sbc)

			bs, _, err := bin.WriterToBytes(sbc)
			Expect(err).To(Succeed())

			stableAddInitialLiquidity(genesis, alice)

			is, _ := Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			virtual_price := is[0].(*amount.Amount)

			cn, cdx, ctx, _ := initChain(genesis, admin)
			timestamp_second := uint64(time.Now().UnixNano()) / uint64(time.Second)
			ctx, _ = Sleep(cn, ctx, nil, timestamp_second, aliceKey)

			// Deploy with Same Address
			v, err := ctx.DeployContractWithAddress(bob, classMap["StableSwap"], swap, bs)
			swap2 := v.(*trade.StableSwap).Address() // swap2 == swap

			Expect(swap).To(Equal(swap2))

			is, _ = Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			virtual_price2 := is[0].(*amount.Amount)

			Expect(virtual_price2).To(Equal(virtual_price))

			admin_balances, _ := stableGetAdminBalances(ctx)
			for i := uint8(0); i < N; i++ {
				is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
				Expect(err).To(Succeed())
				if admin_balances[i].Cmp(Zero) == 0 {
					Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
				} else {
					Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i].Int))
				}
			}
			Expect(Sum(ToBigInts(admin_balances))).To(Equal(Zero))

			RemoveChain(cdx)
			afterEach()
		})

		It("test_unexpected_coin", func() {
			beforeEachStable()

			stableAddInitialLiquidity(genesis, alice)

			is, _ := Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			virtual_price := is[0].(*amount.Amount)

			tokenMint(genesis, stableTokens[N-1], swap, amount.NewAmount(0, 123456))

			is, _ = Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			virtual_price2 := is[0].(*amount.Amount)

			Expect(virtual_price).To(Equal(virtual_price2))

			admin_balances, _ := stableGetAdminBalances(genesis)
			for i := uint8(0); i < N; i++ {
				is, err := Exec(genesis, bob, swap, "AdminBalances", []interface{}{uint8(i)})
				Expect(err).To(Succeed())
				if admin_balances[i].Cmp(Zero) == 0 {
					Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
				} else {
					Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i]))
				}
			}
			Expect(Sum(ToBigInts(admin_balances))).To(Equal(big.NewInt(123456)))

			afterEach()
		})
	})
})
