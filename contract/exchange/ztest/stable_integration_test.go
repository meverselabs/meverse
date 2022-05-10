package test

import (
	"math"
	"math/big"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common/amount"

	"github.com/meverselabs/meverse/contract/exchange/trade"
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

var _ = Describe("Integration", func() {

	Describe("test_curve.py", func() {

		It("test_curve_in_contract", func() {

			st_pct := []float64{}
			for i := 0; i < 50; i++ {
				pct := float64(rand.Intn(1000))
				st_pct = append(st_pct, pct/1000)
			}
			st_seed_amount := []*big.Int{}
			for i := 0; i < 5; i++ {
				st_seed_amount = append(st_seed_amount, big.NewInt(int64(math.Pow(float64(10), float64((rand.Intn(120-50)+50)/10)))))
			}

			for k := 0; k < len(st_seed_amount); k++ {
				beforeEachStable()
				cn, cdx, ctx, _ := initChain(genesis, admin)
				ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
				ctx, _ = setFees(cn, ctx, swap, 0, 0, 0, uint64(86400), aliceKey)

				// add initial pool liquidity
				// we add at an imbalance of +10% for each subsequent coin
				initial_liquidity := []*amount.Amount{}

				for i := uint8(0); i < N; i++ {
					amount := ToAmount(Add(Mul(st_seed_amount[k], big.NewInt(int64(math.Pow10(decimals[i])))), big.NewInt(1)))

					_, err := Exec(ctx, admin, stableTokens[i], "Mint", []interface{}{alice, amount})
					Expect(err).To(Succeed())

					balance, err := tokenBalanceOf(ctx, stableTokens[i], alice)
					Expect(err).To(Succeed())

					if balance.Int.Cmp(amount.Int) < 0 {
						Fail("test_curve_in_contract : Balance")
					}

					initial_liquidity = append(initial_liquidity, amount.DivC(10))
					_, err = Exec(ctx, alice, stableTokens[i], "Approve", []interface{}{swap, amount.DivC(10)})
					Expect(err).To(Succeed())

				}

				_, err := Exec(ctx, alice, swap, "AddLiquidity", []interface{}{initial_liquidity, ZeroAmount})
				Expect(err).To(Succeed())

				is, err := Exec(ctx, alice, swap, "Reserves", []interface{}{})
				Expect(err).To(Succeed())
				balances := is[0].([]*amount.Amount)

				is, err = Exec(ctx, alice, swap, "Rates", []interface{}{})
				Expect(err).To(Succeed())
				rates := is[0].([]*big.Int)

				curve_model := &Curve{}
				curve_model.init(_Amp, ToBigInts(balances), N, rates, big.NewInt(0))

				// execute a series of swaps and compare the go model(simulation.go) to the contract results
				for i := uint8(0); i < N; i++ {
					rates[i] = big.NewInt(trade.PRECISION)
				}
				for i := 0; i < len(st_pct); i++ {
					send := 0
					recv := 1

					dx := Mul(Mul(big.NewInt(2), st_seed_amount[k]), big.NewInt(int64(math.Pow10(decimals[send])*st_pct[i])))
					if dx.Cmp(Zero) <= 0 {
						continue
					}

					is, err = Exec(ctx, alice, swap, "GetDy", []interface{}{uint8(send), uint8(recv), ToAmount(dx), ZeroAddress})
					Expect(err).To(Succeed())

					dy_1_c := is[0].(*amount.Amount).Int

					dy_1 := MulDivC(dy_1_c, rates[recv], trade.PRECISION)

					s := big.NewInt(int64(math.Pow10(18 - decimals[send])))
					r := big.NewInt(int64(math.Pow10(18 - decimals[recv])))
					dy_2 := curve_model.dy(uint8(send), uint8(recv), Mul(dx, s))
					dy_2 = Div(dy_2, r)

					if !(Approx(ToFloat64(dy_1), ToFloat64(dy_2), 1e-8) || Abs(Sub(dy_1, dy_2)).Cmp(big.NewInt(2)) <= 0) {
						Fail("Approx")
					}
				}

				RemoveChain(cdx)
				afterEach()
			}
		})
	})

	Describe("test_heavily_imbalanced.py", func() {

		It("test_imbalanced_swaps", func() {
			for idx := uint8(0); idx < N; idx++ {

				beforeEachStable()
				stableAddInitialLiquidity(genesis, alice)
				stableMint(genesis, bob)
				stableApprove(genesis, bob)

				amounts := MakeAmountSlice(N)
				amounts[idx].Set(_InitialAmounts[idx].MulC(1000).Int)

				tokenMint(genesis, stableTokens[idx], alice, amounts[idx])
				Exec(genesis, alice, swap, "AddLiquidity", []interface{}{amounts, ZeroAmount})

				swap_indexes := []int{}
				for i := uint8(0); i < N; i++ {
					if i == idx {
						continue
					}
					swap_indexes = append(swap_indexes, int(i))
				}

				//# make a few swaps where the input asset is the one we already have too much of
				//# bob is getting rekt here, but the contract shouldn't bork
				for k := 0; k < len(swap_indexes); k++ {
					m := uint8(swap_indexes[k])
					_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{idx, m, _InitialAmounts[idx].DivC(int64(N)), ZeroAmount, ZeroAddress})
					Expect(err).To(Succeed())
				}

				//# now we go the other direction, swaps where the output asset is the imbalanced one
				//# lucky bob is about to get rich!
				for k := 0; k < len(swap_indexes); k++ {
					m := uint8(swap_indexes[k])
					_, err := Exec(genesis, bob, swap, "Exchange", []interface{}{m, idx, _InitialAmounts[m].DivC(int64(N)), ZeroAmount, ZeroAddress})
					Expect(err).To(Succeed())
				}

				afterEach()
			}
		})
	})

	Describe("test_virtual_price_increases.py", func() {

		It("test_number_always_go_up", func() {

			max_examples := 1
			for max_examples > 0 {

				st_pct := []float64{}
				for i := 0; i < 10; i++ {
					pct := float64(rand.Intn(101-50) + 50)
					st_pct = append(st_pct, pct/50)
				}
				st_seed_amount := []float64{}
				for i := 0; i < 3; i++ {
					st_seed_amount = append(st_seed_amount, float64(rand.Intn(1005-1001)+1001)/1000.)
				}

				for k := 0; k < len(st_seed_amount); k++ {

					for m := 0; m < len(st_pct); m++ {

						beforeEachStable()
						stableAddInitialLiquidity(genesis, alice)
						is, _ := Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
						virtual_price := is[0].(*amount.Amount)

						cn, cdx, ctx, _ := initChain(genesis, admin)
						ctx, _ = setFees(cn, ctx, swap, uint64(10000000), uint64(0), uint64(0), uint64(86400), aliceKey)
						for i := uint8(0); i < N; i++ {
							amt := amount.NewAmount(uint64(_BaseAmount), 0)
							tokenMint(ctx, stableTokens[i], alice, amt)
						}

						r := rand.Intn(5)

						if r == 0 {
							// rule_ramp_A
							is, _ := Exec(ctx, alice, swap, "A", []interface{}{})
							new_A := is[0].(*big.Int)
							timestamp := ctx.LastTimestamp()/uint64(time.Second) + 86410
							_, err := Exec(ctx, alice, swap, "RampA", []interface{}{new_A, timestamp})
							Expect(err).To(Succeed())
						} else if r == 1 {
							// rule_increase_rates
							//     not existent

							// rule_exchange

							send, recv, _, _ := _min_max(ctx, swap, alice)
							amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[send])) * st_pct[m])))
							_, err := Exec(ctx, alice, swap, "Exchange", []interface{}{uint8(send), recv, amt, ZeroAmount, ZeroAddress})
							Expect(err).To(Succeed())
						} else if r == 2 {
							// rule_exchange_underlying
							//     not existent

							// rule_remove_one_coin

							_, idx, _, _ := _min_max(ctx, swap, alice)
							amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[idx])) * st_pct[m])))

							_, err := Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amt, idx, ZeroAmount})
							Expect(err).To(Succeed())
						} else if r == 3 {
							// rule_remove_imbalance
							_, idx, _, _ := _min_max(ctx, swap, alice)
							amts := MakeAmountSlice(N)
							amts[idx] = ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[idx])) * st_pct[m])))
							_, err := Exec(ctx, alice, swap, "RemoveLiquidityImbalance", []interface{}{amts, MaxUint256})
							Expect(err).To(Succeed())
						} else if r == 4 {
							// rule_remove
							amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(18)) * st_pct[m])))
							_, err := Exec(ctx, alice, swap, "RemoveLiquidity", []interface{}{amt, MakeAmountSlice(N)})
							Expect(err).To(Succeed())
						}

						// invariant_check_virtual_price
						is, _ = Exec(ctx, alice, swap, "GetVirtualPrice", []interface{}{})
						virtual_price2 := is[0].(*amount.Amount)
						if virtual_price2.Cmp(virtual_price.Int) < 0 {
							Fail("test_number_always_go_up")
						}
						virtual_price.Set(virtual_price.Int)

						// invariant_advance_time
						ctx, _ = Sleep(cn, ctx, nil, 3600, aliceKey)

						RemoveChain(cdx)
						afterEach()
					}
				}
				max_examples--
			}
		})
	})
})
