package test

import (
	"strconv"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	. "github.com/meverselabs/meverse/contract/exchange/util"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	contract : curve-contract
	files :
		tests/forked/test_gas.py
		tests/forked/test_insufficient_balances.py
*/

var _ = Describe("Forked", func() {
	var err error
	Describe("test_gas.py", func() {
		BeforeEach(func() {
			beforeEachStable()
		})

		AfterEach(func() {
			afterEach()
		})

		It("test_swap_gas", func() {
			// 아래값과 같아야 함
			// alice balances 1 [500000000000000000000000, 500000000000, 500000000000]
			// alice balances 2 [300000000000000000000000, 300000000000, 300000000000]
			// bob balances 2 [1000000000000000000000000, 1000000000000, 1000000000000]
			// send, recv, amt in exchange 0 1 1000000000000000000
			// send, recv, amt in exchange 0 2 1000000000000000000
			// send, recv, amt in exchange 1 0 1000000
			// send, recv, amt in exchange 1 2 1000000
			// send, recv, amt in exchange 2 0 1000000
			// send, recv, amt in exchange 2 1 1000000
			// alice balances 3 [1300000999600003379351116, 1300000999601, 1300000999599]
			// bob balances 3 [999600002736027652, 999600, 999601]
			// alice balances 4 [1300000999600003379351116, 1300000999601, 1300000999599]
			// bob balances 4 [999600002736027652, 999600, 999601]
			// alice balances 5 [1300001332938765204076108, 1300001332939, 1300001332938]
			// alice balances 6 [1300002332938765204076108, 1300002332939, 1300002332938]
			// alice balances 7 [1300004332938765204076108, 1300004332939, 1300004332938]
			// alice balances 8 [1300005332755480461826063, 1300004332939, 1300004332938]

			stableMint(genesis, alice)
			stableApprove(genesis, alice)
			stableMint(genesis, bob)
			stableApprove(genesis, bob)

			step := uint64(3600)

			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, uint64(4000000), uint64(5000000000), uint64(5000000000), uint64(86400), aliceKey)

			amts := CloneAmountSlice(_InitialAmounts)
			for i := uint8(0); i < N; i++ {
				amts[i].Set(DivC(amts[i].Int, 2))
			}
			Exec(ctx, alice, swap, "AddLiquidity", []interface{}{amts, ZeroAmount})

			ctx, _ = Sleep(cn, ctx, nil, step, aliceKey)

			balances, _ := stableTokenBalances(ctx, alice)
			GPrintln("alice balances 1", balances)

			//# add liquidity imbalanced
			for idx := uint8(0); idx < N; idx++ {
				amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(DivC(amts[i].Int, 10))
				}
				amts[idx] = amount.NewAmount(0, 0)

				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "AddLiquidity",
					Args:      bin.TypeWriteAll(amts, ZeroAmount),
				}
				ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
				Expect(err).To(Succeed())
			}

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 2", balances)

			balances, _ = stableTokenBalances(ctx, bob)
			GPrintln("bob balances 2", balances)

			//# perform swaps between each coin
			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}

					amt := Pow10(decimals[send])
					tokenMint(ctx, stableTokens[send], bob, ToAmount(AddC(amt, 1)))

					recv_balance, _ := tokenBalanceOf(ctx, stableTokens[recv], bob)
					if recv_balance.Cmp(Zero) > 0 {
						Expect(safeTransfer(ctx, bob, stableTokens[recv], alice, recv_balance)).To(Succeed())
					}
					GPrintln("send, recv, amt in exchange", send, recv, amt)

					tx := &types.Transaction{
						ChainID:   ctx.ChainID(),
						Timestamp: ctx.LastTimestamp(),
						To:        swap,
						Method:    "Exchange",
						Args:      bin.TypeWriteAll(send, recv, ToAmount(amt), ZeroAmount, ZeroAddress),
					}
					ctx, err = Sleep(cn, ctx, tx, step, bobKey)
					Expect(err).To(Succeed())
				}
			}

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 3", balances)

			balances, _ = stableTokenBalances(ctx, bob)
			GPrintln("bob balances 3", balances)

			//# remove liquidity balanced
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RemoveLiquidity",
				Args:      bin.TypeWriteAll(amount.NewAmount(1, 0), MakeAmountSlice(N)),
			}
			ctx, _ = Sleep(cn, ctx, tx, step, aliceKey)

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 5", balances)

			amts = MakeAmountSlice(N)
			for i := uint8(0); i < N; i++ {
				amts[i].Set(Pow10(decimals[i]))
			}

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RemoveLiquidityImbalance",
				Args:      bin.TypeWriteAll(amts, MaxUint256),
			}
			ctx, _ = Sleep(cn, ctx, tx, step, aliceKey)

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 6", balances)

			//# remove liquidity balanced
			for idx := uint8(0); idx < N; idx++ {
				amts = MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(Pow10(decimals[i]))
				}
				amts[idx].Set(Zero)

				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "RemoveLiquidityImbalance",
					Args:      bin.TypeWriteAll(amts, MaxUint256),
				}
				ctx, _ = Sleep(cn, ctx, tx, step, aliceKey)
			}

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 7", balances)

			//# remove_liquidity_one_coin
			for idx := uint8(0); idx < N; idx++ {
				amt := ToAmount(Pow10(decimals[idx]))
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "RemoveLiquidityOneCoin",
					Args:      bin.TypeWriteAll(amt, idx, ZeroAmount),
				}
				ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
				if err != nil {
					Fail(strconv.Itoa(int(idx)) + " : " + err.Error())
				}
			}

			balances, _ = stableTokenBalances(ctx, alice)
			GPrintln("alice balances 8", balances)

			//	mainToken 가스계산
			mToken := *ctx.MainToken()
			balance, _ := tokenBalanceOf(ctx, mToken, alice)
			GPrintln("gas-used of alice", amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance))

			balance, _ = tokenBalanceOf(ctx, mToken, bob)
			GPrintln("gas-used of bob", amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance))

			RemoveChain(cdx)
		})

		It("test_zap_gas", func() {
			// no zap
		})
	})

	Describe("test_insufficient_balances.py", func() {

		It("test_swap_gas", func() {
			beforeEachStable()
			stableMint(genesis, alice)
			stableApprove(genesis, alice)
			stableMint(genesis, bob)
			stableApprove(genesis, bob)
			step := uint64(3600)
			cn, cdx, ctx, _ := initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, swap, uint64(4000000), uint64(5000000000), uint64(5000000000), uint64(86400), aliceKey)

			//# attempt to deposit more funds than user has
			for idx := uint8(0); idx < N; idx++ {
				amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(DivC(amts[i].Int, 2))
				}

				balance, _ := tokenBalanceOf(ctx, stableTokens[idx], alice)
				amts[idx] = balance.Add(amount.NewAmount(0, 1))

				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "AddLiquidity",
					Args:      bin.TypeWriteAll(amts, ZeroAmount),
				}
				_, err := Sleep(cn, ctx, tx, step, aliceKey)
				Expect(err).To(MatchError("the token holding quantity is insufficient"))
			}
			//# add liquidity balanced
			amts := CloneAmountSlice(_InitialAmounts)
			for i := uint8(0); i < N; i++ {
				amts[i].Set(DivC(amts[i].Int, 2))
			}
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "AddLiquidity",
				Args:      bin.TypeWriteAll(amts, ZeroAmount),
			}
			var err error
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			//# attempt to perform swaps between stableTokens with insufficient funds
			for send := uint8(0); send < N; send++ {
				for recv := uint8(0); recv < N; recv++ {
					if send == recv {
						continue
					}

					amt := Clone(DivC(_InitialAmounts[send].Int, 4))

					tx := &types.Transaction{
						ChainID:   ctx.ChainID(),
						Timestamp: ctx.LastTimestamp(),
						To:        swap,
						Method:    "Exchange",
						Args:      bin.TypeWriteAll(send, recv, ToAmount(amt), ZeroAmount, ZeroAddress),
					}
					_, err = Sleep(cn, ctx, tx, step, charlieKey)
					Expect(err).To(MatchError("the token holding quantity is insufficient"))
				}
			}

			//# remove liquidity balanced
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RemoveLiquidity",
				Args:      bin.TypeWriteAll(amount.NewAmount(1, 0), MakeAmountSlice(N)),
			}

			_, err = Sleep(cn, ctx, tx, step, charlieKey)
			Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))

			//# remove liquidity imbalanced
			for idx := uint8(0); idx < N; idx++ {
				amts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(Pow10(decimals[i]))
				}

				balance, _ := tokenBalanceOf(ctx, stableTokens[idx], swap)

				amts[idx].Set(AddC(balance.Int, 1))
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "RemoveLiquidityImbalance",
					Args:      bin.TypeWriteAll(amts, MaxUint256),
				}
				_, err = Sleep(cn, ctx, tx, step, charlieKey)
				Expect(err).To(HaveOccurred())
			}

			//# remove_liquidity_one_coin
			for idx := uint8(0); idx < N; idx++ {
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: ctx.LastTimestamp(),
					To:        swap,
					Method:    "RemoveLiquidityOneCoin",
					Args:      bin.TypeWriteAll(ToAmount(Pow10(decimals[idx])), idx, ZeroAmount),
				}
				_, err = Sleep(cn, ctx, tx, step, charlieKey)
				Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))
			}

			RemoveChain(cdx)
			afterEach()

		})
	})

})
