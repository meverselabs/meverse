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

var _ = Describe("Pair", func() {
	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
		err error
	)

	Describe("PayToken Null", func() {

		BeforeEach(func() {
			beforeEachUni()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})
		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("fee, adminFee, winnerFee", func() {
			for k := 0; k < 10; k++ {
				pct := uint64(rand.Intn(1000))
				fee := uint64(float64(int64(trade.MAX_FEE*pct)) / 1000.)
				pct = uint64(rand.Intn(1000))
				adminFee := uint64(float64(int64(trade.MAX_ADMIN_FEE*pct)) / 1000.)
				pct = uint64(rand.Intn(1000))
				winnerFee := uint64(float64(int64(trade.MAX_WINNER_FEE*pct)) / 1000.)

				ctx, err = setFees(cn, ctx, pair, fee, adminFee, winnerFee, uint64(86400), aliceKey)
				Expect(err).To(Succeed())

				//fee
				is, err := Exec(ctx, admin, pair, "Fee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0]).To(Equal(fee))

				//AdminFee
				is, err = Exec(ctx, admin, pair, "AdminFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0]).To(Equal(adminFee))

				//WinnerFee
				is, err = Exec(ctx, admin, pair, "WinnerFee", []interface{}{})
				Expect(err).To(Succeed())
				Expect(is[0]).To(Equal(winnerFee))
			}
		})
		It("mint", func() {
			err = uniMint(ctx, alice)
			Expect(err).To(Succeed())

			token0Amount := amount.NewAmount(1, 0)
			token1Amount := amount.NewAmount(4, 0)

			// transfer
			_, err = Exec(ctx, alice, token0, "Transfer", []interface{}{pair, token0Amount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount))

			_, err = Exec(ctx, alice, token1, "Transfer", []interface{}{pair, token1Amount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount))

			expectedLiquidity := amount.NewAmount(2, 0)

			// mint
			is, err := Exec(ctx, alice, pair, "Mint", []interface{}{alice})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(expectedLiquidity.Sub(_ML)))

			// totalSupply
			Expect(tokenTotalSupply(ctx, pair)).To(Equal(expectedLiquidity))

			// BalanceOf pair, token0, token1
			Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(expectedLiquidity.Sub(_ML)))
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount))
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount))

			// Reserves
			is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount))
		})

		DescribeTable("getInputPrice",
			func(swapAmount, token0Amount, token1Amount, expectedOutputAmount *amount.Amount) {

				uniMint(ctx, alice)
				uniApprove(ctx, alice)
				uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

				Exec(ctx, alice, token0, "Transfer", []interface{}{pair, swapAmount})

				_, err := Exec(ctx, alice, pair, "Swap", []interface{}{ZeroAmount, expectedOutputAmount.Add(amount.NewAmount(0, 1)), alice, []byte(""), ZeroAddress})
				Expect(err).To(MatchError("Exchange: K"))

				_, err = Exec(ctx, alice, pair, "Swap", []interface{}{ZeroAmount, expectedOutputAmount, alice, []byte(""), ZeroAddress})
				Expect(err).To(Succeed())
			},

			Entry("1", amount.NewAmount(1, 0), amount.NewAmount(5, 0), amount.NewAmount(10, 0), &amount.Amount{Int: big.NewInt(1662497915624478906)}),
			Entry("2", amount.NewAmount(1, 0), amount.NewAmount(10, 0), amount.NewAmount(5, 0), &amount.Amount{Int: big.NewInt(453305446940074565)}),
			Entry("3", amount.NewAmount(2, 0), amount.NewAmount(5, 0), amount.NewAmount(10, 0), &amount.Amount{Int: big.NewInt(2851015155847869602)}),
			Entry("4", amount.NewAmount(2, 0), amount.NewAmount(10, 0), amount.NewAmount(5, 0), &amount.Amount{Int: big.NewInt(831248957812239453)}),
			Entry("5", amount.NewAmount(1, 0), amount.NewAmount(10, 0), amount.NewAmount(10, 0), &amount.Amount{Int: big.NewInt(906610893880149131)}),
			Entry("6", amount.NewAmount(1, 0), amount.NewAmount(100, 0), amount.NewAmount(100, 0), &amount.Amount{Int: big.NewInt(987158034397061298)}),
			Entry("7", amount.NewAmount(1, 0), amount.NewAmount(1000, 0), amount.NewAmount(1000, 0), &amount.Amount{Int: big.NewInt(996006981039903216)}),
		)

		DescribeTable("optimistic",
			func(outputAmount, token0Amount, token1Amount, inputAmount *amount.Amount) {

				uniMint(ctx, alice)
				uniApprove(ctx, alice)
				uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

				Exec(ctx, alice, token0, "Transfer", []interface{}{pair, inputAmount})

				_, err := Exec(ctx, alice, pair, "Swap", []interface{}{outputAmount.Add(amount.NewAmount(0, 1)), ZeroAmount, alice, []byte(""), ZeroAddress})
				Expect(err).To(MatchError("Exchange: K"))

				_, err = Exec(ctx, alice, pair, "Swap", []interface{}{outputAmount, ZeroAmount, alice, []byte(""), ZeroAddress})
				Expect(err).To(Succeed())
			},
			Entry("1", &amount.Amount{Int: big.NewInt(997000000000000000)}, amount.NewAmount(5, 0), amount.NewAmount(10, 0), amount.NewAmount(1, 0)),
			Entry("2", &amount.Amount{Int: big.NewInt(997000000000000000)}, amount.NewAmount(10, 0), amount.NewAmount(5, 0), amount.NewAmount(1, 0)),
			Entry("3", &amount.Amount{Int: big.NewInt(997000000000000000)}, amount.NewAmount(5, 0), amount.NewAmount(5, 0), amount.NewAmount(1, 0)),
			Entry("4", amount.NewAmount(1, 0), amount.NewAmount(5, 0), amount.NewAmount(5, 0), &amount.Amount{Int: big.NewInt(1003009027081243732)}),
		)

		It("swap:token0", func() {
			token0Amount := amount.NewAmount(5, 0)
			token1Amount := amount.NewAmount(10, 0)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(1662497915624478906)}

			// Transfer
			Exec(ctx, alice, token0, "Transfer", []interface{}{pair, swapAmount})

			// Swap
			_, err := Exec(ctx, alice, pair, "Swap", []interface{}{ZeroAmount, expectedOutputAmount, alice, []byte(""), ZeroAddress})
			Expect(err).To(Succeed())

			// Reserves
			is, err := Exec(ctx, alice, pair, "Reserves", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount.Add(swapAmount)))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount.Sub(expectedOutputAmount)))

			// BalanceOf
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount.Add(swapAmount)))
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount.Sub(expectedOutputAmount)))
			Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Sub(swapAmount)))
			Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Add(expectedOutputAmount)))
		})

		It("swap:token1", func() {
			token0Amount := amount.NewAmount(5, 0)
			token1Amount := amount.NewAmount(10, 0)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(453305446940074565)}

			// Transfer
			_, err := Exec(ctx, alice, token1, "Transfer", []interface{}{pair, swapAmount})
			Expect(err).To(Succeed())

			// Swap
			_, err = Exec(ctx, alice, pair, "Swap", []interface{}{expectedOutputAmount, ZeroAmount, alice, []byte(""), ZeroAddress})
			Expect(err).To(Succeed())

			// Reserves
			is, err := Exec(ctx, alice, pair, "Reserves", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount.Sub(expectedOutputAmount)))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount.Add(swapAmount)))

			// BalanceOf
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(token0Amount.Sub(expectedOutputAmount)))
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(token1Amount.Add(swapAmount)))
			Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Add(expectedOutputAmount)))
			Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Sub(swapAmount)))
		})

		It("burn", func() {
			token0Amount := amount.NewAmount(3, 0)
			token1Amount := amount.NewAmount(3, 0)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

			expectedLiquidity := amount.NewAmount(3, 0)

			// Transfer
			_, err := Exec(ctx, alice, pair, "Transfer", []interface{}{pair, expectedLiquidity.Sub(_ML)})
			Expect(err).To(Succeed())

			// Burn
			is, err := Exec(ctx, alice, pair, "Burn", []interface{}{alice})
			GPrintf("%+v", err)
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(token0Amount.Sub(_ML)))
			Expect(is[1].(*amount.Amount)).To(Equal(token1Amount.Sub(_ML)))

			// pair.BalanceOf(alice)
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(ZeroAmount))

			// pair.TotalSupply()
			Expect(tokenTotalSupply(ctx, pair)).To(Equal(_ML))

			// BalanceOf
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(_ML))
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(_ML))
			Expect(tokenBalanceOf(ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(_ML)))
			Expect(tokenBalanceOf(ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(_ML)))
		})

		It("price{0,1}CumulativeLast", func() {
			err = uniMint(ctx, alice)
			Expect(err).To(Succeed())

			token0Amount := amount.NewAmount(3, 0)
			token1Amount := amount.NewAmount(3, 0)

			//cn, cdx, ctx, err := initChain(ctx, admin)
			//Expect(err).To(Succeed())

			timestamp := uint64(time.Now().UnixNano())
			ctx, err = Sleep(cn, ctx, nil, timestamp/uint64(time.Second), aliceKey)
			Expect(err).To(Succeed())

			Exec(ctx, alice, token0, "Transfer", []interface{}{pair, token0Amount})
			Exec(ctx, alice, token1, "Transfer", []interface{}{pair, token1Amount})

			// tx : pair.Mint(alice)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Mint",
				Args:      bin.TypeWriteAll(alice),
			}

			step := uint64(1) // 1 sec
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// Reserves 1
			is, _ := Exec(ctx, alice, pair, "Reserves", []interface{}{})
			blockTimestamp := is[1].(uint64)

			// tx : pair.sync()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Sync",
				Args:      bin.TypeWriteAll(),
			}

			step = uint64(4) // 4 sec
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// Reserves 2
			initialPrice0 := token0Amount.Div(token1Amount)
			initialPrice1 := token1Amount.Div(token0Amount)

			is, _ = Exec(ctx, alice, pair, "Price0CumulativeLast", []interface{}{})
			Expect(is[0].(*amount.Amount)).To(Equal(initialPrice0))
			is, _ = Exec(ctx, alice, pair, "Price1CumulativeLast", []interface{}{})
			Expect(is[0].(*amount.Amount)).To(Equal(initialPrice1))
			is, _ = Exec(ctx, alice, pair, "Reserves", []interface{}{})
			Expect(is[1]).To(Equal(blockTimestamp + 1))

			// tx : token0.Tranfer(pair,swapAmount)
			swapAmount := amount.NewAmount(3, 0)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        token0,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(pair, swapAmount),
			}

			step = uint64(5) // 5 sec
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			//await pair.swap(0, expandTo18Decimals(1), wallet.address, '0x', overrides) // make the price nice
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Swap",
				Args:      bin.TypeWriteAll(ZeroAmount, amount.NewAmount(1, 0), alice, []byte(""), ZeroAddress),
			}

			step = uint64(10) // 10 sec
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// tx : pair.sync()
			timestamp = timestamp + 10*uint64(time.Second)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Sync",
				Args:      bin.TypeWriteAll(),
			}

			step = uint64(1) // 1 sec
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			newPrice0 := amount.NewAmount(0, amount.FractionalMax/3)
			newPrice1 := amount.NewAmount(3, 0)

			// Reserves
			is, _ = Exec(ctx, alice, pair, "Price0CumulativeLast", []interface{}{})
			Expect(is[0].(*amount.Amount)).To(Equal(initialPrice0.MulC(int64(10)).Add(newPrice0.MulC(int64(10)))))
			is, _ = Exec(ctx, alice, pair, "Price1CumulativeLast", []interface{}{})
			Expect(is[0].(*amount.Amount)).To(Equal(initialPrice1.MulC(int64(10)).Add(newPrice1.MulC(int64(10)))))
			is, _ = Exec(ctx, alice, pair, "Reserves", []interface{}{})
			Expect(is[1]).To(Equal(blockTimestamp + 20))
		})

		It("feeTo:off", func() {
			delay := uint64(86400)
			_, err = Exec(ctx, alice, pair, "CommitNewFee", []interface{}{_Fee30, uint64(0), uint64(0), delay})
			Expect(err).To(Succeed())

			ctx, err := Sleep(cn, ctx, nil, delay, aliceKey)
			Expect(err).To(Succeed())

			_, err = Exec(ctx, alice, pair, "ApplyNewFee", []interface{}{})
			Expect(err).To(Succeed())

			cc, _ := GetCC(ctx, pair, admin)
			owner, _ := Owner(cc, pair)
			Expect(owner).To(Equal(alice))

			is, err := Exec(ctx, alice, pair, "AdminFee", []interface{}{})
			Expect(is[0]).To(Equal(uint64(0)))

			token0Amount := amount.NewAmount(1000, 0)
			token1Amount := amount.NewAmount(1000, 0)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(996006981039903216)}

			// Transfer
			_, err = Exec(ctx, alice, token1, "Transfer", []interface{}{pair, swapAmount})
			Expect(err).To(Succeed())

			// Swap
			_, err = Exec(ctx, alice, pair, "Swap", []interface{}{expectedOutputAmount, ZeroAmount, alice, []byte(""), ZeroAddress})
			Expect(err).To(Succeed())

			expectedLiquidity := amount.NewAmount(1000, 0)

			// Transfer
			_, err = Exec(ctx, alice, pair, "Transfer", []interface{}{pair, expectedLiquidity.Sub(_ML)})
			Expect(err).To(Succeed())

			// Burn
			_, err = Exec(ctx, alice, pair, "Burn", []interface{}{alice})
			Expect(err).To(Succeed())

			// pair.TotalSupply()
			Expect(tokenTotalSupply(ctx, pair)).To(Equal(_ML))
		})

		It("feeTo:on", func() {
			ctx, err = setFees(cn, ctx, pair, _Fee30, _AdminFee6, _WinnerFee, uint64(86400), aliceKey)
			Expect(err).To(Succeed())

			//owner
			cc, err := GetCC(ctx, pair, alice)
			Expect(err).To(Succeed())
			owner, _ := Owner(cc, pair)
			Expect(owner).To(Equal(alice))

			token0Amount := amount.NewAmount(1000, 0)
			token1Amount := amount.NewAmount(1000, 0)

			uniMint(ctx, alice)
			uniApprove(ctx, alice)
			uniAddLiquidity(ctx, alice, token0Amount, token1Amount)

			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(996006981039903216)}

			// Transfer
			_, err = Exec(ctx, alice, token1, "Transfer", []interface{}{pair, swapAmount})
			Expect(err).To(Succeed())

			// Swap
			_, err = Exec(ctx, alice, pair, "Swap", []interface{}{expectedOutputAmount, ZeroAmount, alice, []byte(""), ZeroAddress})
			Expect(err).To(Succeed())

			expectedLiquidity := amount.NewAmount(1000, 0)

			// Transfer
			_, err = Exec(ctx, alice, pair, "Transfer", []interface{}{pair, expectedLiquidity.Sub(_ML)})
			Expect(err).To(Succeed())

			// Burn
			_, err = Exec(ctx, alice, pair, "Burn", []interface{}{alice})
			Expect(err).To(Succeed())

			// pair.TotalSupply()
			// 1/n 에서 %로 변경되면서 값이 달라짐
			// Fee protocol = 6  :                    249750499252388 = 1000 + 249750499251388
			// AdminFee = 1666666667/FEEDENOMINATOR : 249750499302338
			//            1666666666/FEEDENOMINATOR : 249750499152487
			// 					    249750499152487 < 249750499252388 < 249750499302338

			Expect(tokenTotalSupply(ctx, pair)).To(Equal(_ML.Add(&amount.Amount{Int: big.NewInt(249750499301338)}))) // 1000 고려해야 함
			Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(&amount.Amount{Int: big.NewInt(249750499301338)}))
			Expect(tokenBalanceOf(ctx, token0, pair)).To(Equal(amount.NewAmount(0, 1000).Add(&amount.Amount{Int: big.NewInt(249501683747345)})))
			Expect(tokenBalanceOf(ctx, token1, pair)).To(Equal(amount.NewAmount(0, 1000).Add(&amount.Amount{Int: big.NewInt(250000187362969)})))
		})

		It("Skim", func() {
			uniAddInitialLiquidity(ctx, alice)
			uniMint(ctx, bob)

			var transferToken common.Address
			for k := 0; k < 2; k++ {
				switch k {
				case 0:
					transferToken = token0
				case 1:
					transferToken = token1
				}

				pairBalance0, err := tokenBalanceOf(ctx, token0, pair)
				Expect(err).To(Succeed())
				pairBalance1, err := tokenBalanceOf(ctx, token1, pair)
				Expect(err).To(Succeed())

				is, err := Exec(ctx, bob, pair, "Reserves", []interface{}{})
				Expect(err).To(Succeed())
				reserve0 := is[0].([]*amount.Amount)[0]
				reserve1 := is[0].([]*amount.Amount)[1]

				Expect(pairBalance0).To(Equal(reserve0))
				Expect(pairBalance1).To(Equal(reserve1))

				_, err = Exec(ctx, bob, transferToken, "Transfer", []interface{}{pair, _TestAmount})
				Expect(err).To(Succeed())
				pairBalance0, err = tokenBalanceOf(ctx, token0, pair)
				Expect(err).To(Succeed())
				pairBalance1, err = tokenBalanceOf(ctx, token1, pair)
				Expect(err).To(Succeed())

				if k == 0 {
					Expect(pairBalance0).To(Equal(reserve0.Add(_TestAmount)))
					Expect(pairBalance1).To(Equal(reserve1))
				} else {
					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1.Add(_TestAmount)))

				}

				_, err = Exec(ctx, alice, pair, "Skim", []interface{}{charlie})
				Expect(err).To(Succeed())
				pairBalance0, err = tokenBalanceOf(ctx, token0, pair)
				Expect(err).To(Succeed())
				pairBalance1, err = tokenBalanceOf(ctx, token1, pair)
				Expect(err).To(Succeed())

				Expect(pairBalance0).To(Equal(reserve0))
				Expect(pairBalance1).To(Equal(reserve1))

				// charlie
				charlieBalance0, err := tokenBalanceOf(ctx, token0, charlie)
				Expect(err).To(Succeed())
				charlieBalance1, err := tokenBalanceOf(ctx, token1, charlie)
				Expect(err).To(Succeed())

				if k == 0 {
					Expect(charlieBalance0).To(Equal(_TestAmount))
					Expect(charlieBalance1).To(Equal(ZeroAmount))
				} else {
					// charle는 두번 받음
					Expect(charlieBalance0).To(Equal(_TestAmount))
					Expect(charlieBalance1).To(Equal(_TestAmount))
				}
			}
		})

		It("Sync", func() {
			uniAddInitialLiquidity(ctx, alice)
			uniMint(ctx, bob)

			var transferToken common.Address

			for k := 0; k < 2; k++ {
				switch k {
				case 0:
					transferToken = token0
				case 1:
					transferToken = token1
				}

				pairBalance0, err := tokenBalanceOf(ctx, token0, pair)
				Expect(err).To(Succeed())
				pairBalance1, err := tokenBalanceOf(ctx, token1, pair)
				Expect(err).To(Succeed())

				is, err := Exec(ctx, bob, pair, "Reserves", []interface{}{})
				Expect(err).To(Succeed())
				reserve0 := is[0].([]*amount.Amount)[0]
				reserve1 := is[0].([]*amount.Amount)[1]

				Expect(pairBalance0).To(Equal(reserve0))
				Expect(pairBalance1).To(Equal(reserve1))

				_, err = Exec(ctx, bob, transferToken, "Transfer", []interface{}{pair, _TestAmount})
				Expect(err).To(Succeed())
				pairBalance0, err = tokenBalanceOf(ctx, token0, pair)
				Expect(err).To(Succeed())
				pairBalance1, err = tokenBalanceOf(ctx, token1, pair)
				Expect(err).To(Succeed())

				if k == 0 {
					Expect(pairBalance0).To(Equal(reserve0.Add(_TestAmount)))
					Expect(pairBalance1).To(Equal(reserve1))
				} else {
					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1.Add(_TestAmount)))

				}

				_, err = Exec(ctx, alice, pair, "Sync", []interface{}{})
				Expect(err).To(Succeed())
				is, err = Exec(ctx, bob, pair, "Reserves", []interface{}{})
				Expect(err).To(Succeed())
				reserve0 = is[0].([]*amount.Amount)[0]
				reserve1 = is[0].([]*amount.Amount)[1]

				Expect(pairBalance0).To(Equal(reserve0))
				Expect(pairBalance1).To(Equal(reserve1))
			}
		})
	})
})
