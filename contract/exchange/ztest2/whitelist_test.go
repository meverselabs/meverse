package test

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("whitelist : Router", func() {

	path := "_data_w"
	chainID := big.NewInt(65535)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		panic(err)
	}
	aliceKey, bobKey, charlieKey, eveKey := userKeys[0], userKeys[1], userKeys[2], userKeys[3]
	alice, bob, charlie, eve := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address(), eveKey.PublicKey().Address()

	args := []interface{}{alice, bob, charlie, eve} // alice(admin), bob, charlie, eve

	// 체인생성 및  mev 생성
	tb, _ := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	//tb, ret := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	defer removeChainData(path)

	//mev := ret[0].(common.Address)

	var factory, router, whiteList *common.Address
	var pair, swap common.Address
	var tokens []common.Address

	// default 값

	token0Amount := amount.NewAmount(5, 0)
	token1Amount := amount.NewAmount(10, 0)

	_GroupId := hash.BigToHash(big.NewInt(100))

	_PairName := "__UNI_NAME"
	_PairSymbol := "__UNI_SYMBOL"
	_SwapName := "__STABLE_NAME"
	_SwapSymbol := "__STABLE_SYMBOL"
	_Amp := int64(360 * 2)

	deployContracts := func() {

		factory, err = factoryDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

		router, err = routerDeploy(tb, aliceKey, factory)
		if err != nil {
			panic(err)
		}

		whiteList, err = whiteListDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

	}

	BeforeEach(func() {
		deployContracts()

		// erc20Token deploy
		erc20Token, err := erc20TokenDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		Expect(err).To(Succeed())

		// 2 tokens deploy
		token0, err := tokenDeploy(tb, aliceKey, "Token1", "TKN1")
		Expect(err).To(Succeed())

		token2, err := tokenDeploy(tb, aliceKey, "Token2", "TKN2")
		Expect(err).To(Succeed())

		tokens = []common.Address{*token0, *erc20Token, *token2}

		// setMinter
		for _, token := range tokens {
			_, err = tb.call(aliceKey, token, "SetMinter", alice, true)
			if err != nil {
				panic(err)
			}
		}

		is, err := tb.call(aliceKey, *factory, "CreatePairUni", tokens[0], tokens[1], tokens[0], _PairName, _PairSymbol, alice, charlie, _Fee30, _AdminFee6, _Fee5000, *whiteList, _GroupId, ClassMap["UniSwap"])
		Expect(err).To(Succeed())
		pair = is[0].(common.Address)

		is, err = tb.call(aliceKey, *factory, "CreatePairStable", tokens[1], tokens[2], tokens[1], _SwapName, _SwapSymbol, bob, charlie, _Fee30, _AdminFee6, _Fee5000, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
		Expect(err).To(Succeed())
		swap = is[0].(common.Address)

		//initial add liquidity
		tb.call(aliceKey, tokens[0], "Mint", alice, token0Amount)
		tb.call(aliceKey, tokens[1], "Mint", alice, token1Amount)
		tb.call(aliceKey, tokens[0], "Approve", *router, MaxUint256)
		tb.call(aliceKey, tokens[1], "Approve", *router, MaxUint256)

		_, err = tb.call(aliceKey, *router, "UniAddLiquidity", tokens[0], tokens[1], token0Amount, token1Amount, AmountZero, AmountZero)
		Expect(err).To(Succeed())

	})

	Describe("UniSwap", func() {

		Describe("uniAddLiquidityOneCoin, uniGetLPTokenAmountOneCoin", func() {

			supplyAmount := amount.NewAmount(1, 0)

			expectedLPAmount := amount.NewAmount(0, 673811701330274830)   // fee = 0.4%
			wlExpectedLPAmount := amount.NewAmount(0, 674898880549347191) // fee = 0%

			It("not WhiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", charlie, supplyAmount)
				tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.view(*router, "UniGetLPTokenAmountOneCoin", tokens[0], tokens[1], tokens[0], supplyAmount)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

				is, err = tb.call(charlieKey, *router, "UniAddLiquidityOneCoin", tokens[0], tokens[1], tokens[0], supplyAmount, AmountZero)
				Expect(err).To(Succeed())
				Expect(is[1].(*amount.Amount)).To(Equal(expectedLPAmount))

			})

			It("WhiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", eve, supplyAmount)
				tb.call(eveKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.viewFrom(eve, *router, "UniGetLPTokenAmountOneCoin", tokens[0], tokens[1], tokens[0], supplyAmount)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))

				is, err = tb.call(eveKey, *router, "UniAddLiquidityOneCoin", tokens[0], tokens[1], tokens[0], supplyAmount, AmountZero)
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

				tb.call(aliceKey, tokens[0], "Mint", charlie, swapAmount)
				tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.view(*router, "GetAmountsOut", swapAmount, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))

				is, err = tb.call(charlieKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))
				Expect(tokenBalanceOf(tb.ctx, tokens[0], charlie)).To(Equal(AmountZero))
				Expect(tokenBalanceOf(tb.ctx, tokens[1], charlie)).To(Equal(expectedOutputAmount))
			})

			It("GetAmountsOut, SwapExactTokensForTokens : WhiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", eve, swapAmount)
				tb.call(eveKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.viewFrom(eve, *router, "GetAmountsOut", swapAmount, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))

				is, err = tb.call(eveKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))
				Expect(tokenBalanceOf(tb.ctx, tokens[0], eve)).To(Equal(AmountZero))
				Expect(tokenBalanceOf(tb.ctx, tokens[1], eve)).To(Equal(wlExpectedOutputAmount))
			})
		})

		Describe("UniGetAmountsIn, UniSwapTokensForExactTokens", func() {

			outputAmount := amount.NewAmount(1, 0)
			expectedSwapAmount := amount.NewAmount(0, 557227237267357629)   // fee = 0.4%
			wlExpectedSwapAmount := amount.NewAmount(0, 555555555555555556) // fee = 0%

			It("not WhiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", charlie, expectedSwapAmount)
				tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.view(*router, "UniGetAmountsIn", outputAmount, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(expectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))

				is, err = tb.call(charlieKey, *router, "UniSwapTokensForExactTokens", outputAmount, MaxUint256, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(expectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))
				Expect(tokenBalanceOf(tb.ctx, tokens[0], charlie)).To(Equal(AmountZero))
				Expect(tokenBalanceOf(tb.ctx, tokens[1], charlie)).To(Equal(outputAmount))
			})

			It("WhiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", eve, wlExpectedSwapAmount)
				tb.call(eveKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.viewFrom(eve, *router, "UniGetAmountsIn", outputAmount, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(wlExpectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))

				is, err = tb.call(eveKey, *router, "UniSwapTokensForExactTokens", outputAmount, MaxUint256, []common.Address{tokens[0], tokens[1]})
				Expect(err).To(Succeed())
				amounts = is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(wlExpectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))
				Expect(tokenBalanceOf(tb.ctx, tokens[0], eve)).To(Equal(AmountZero))
				Expect(tokenBalanceOf(tb.ctx, tokens[1], eve)).To(Equal(outputAmount))
			})
		})

		Describe("uniRemoveLiquidityOneCoin, uniGetWithdrawAmountOneCoin", func() {

			withdrawLPAmount := amount.NewAmount(1, 0)

			expectedOutputAmount := amount.NewAmount(0, 1498586552649812788)   // fee = 0.4%
			wlExpectedOutputAmount := amount.NewAmount(0, 1500602061002483409) // fee = 0%
			expctedMintFee := amount.NewAmount(0, 321441396770565)

			It("not whiteList : output and mintfee", func() {

				tb.call(aliceKey, tokens[0], "Mint", charlie, token0Amount)
				tb.call(aliceKey, tokens[1], "Mint", charlie, token1Amount)
				tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)
				tb.call(charlieKey, tokens[1], "Approve", *router, MaxUint256)

				tb.call(charlieKey, *router, "UniAddLiquidity", tokens[0], tokens[1], token0Amount, token1Amount, AmountZero, AmountZero)

				// bob swap : mintFee > 0
				swapAmount := amount.NewAmount(1, 0)
				tb.call(aliceKey, tokens[0], "Mint", bob, swapAmount)
				tb.call(bobKey, tokens[0], "Approve", *router, MaxUint256)

				tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[0], tokens[1]})

				tb.call(charlieKey, pair, "Approve", *router, MaxUint256)

				is, err := tb.view(*router, "UniGetWithdrawAmountOneCoin", tokens[0], tokens[1], withdrawLPAmount, tokens[0])
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
				Expect(is[1].(*amount.Amount)).To(Equal(expctedMintFee))

				is, err = tb.call(charlieKey, *router, "UniRemoveLiquidityOneCoin", tokens[0], tokens[1], withdrawLPAmount, tokens[0], AmountZero)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
			})

			It("whiteList", func() {

				tb.call(aliceKey, tokens[0], "Mint", eve, token0Amount)
				tb.call(aliceKey, tokens[1], "Mint", eve, token1Amount)
				tb.call(eveKey, tokens[0], "Approve", *router, MaxUint256)
				tb.call(eveKey, tokens[1], "Approve", *router, MaxUint256)

				tb.call(eveKey, *router, "UniAddLiquidity", tokens[0], tokens[1], token0Amount, token1Amount, AmountZero, AmountZero)

				// bob swap : mintFee > 0
				swapAmount := amount.NewAmount(1, 0)
				tb.call(aliceKey, tokens[0], "Mint", bob, swapAmount)
				tb.call(bobKey, tokens[0], "Approve", *router, MaxUint256)

				tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[0], tokens[1]})

				tb.call(eveKey, pair, "Approve", *router, MaxUint256)

				is, err := tb.viewFrom(eve, *router, "UniGetWithdrawAmountOneCoin", tokens[0], tokens[1], withdrawLPAmount, tokens[0])
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
				Expect(is[1].(*amount.Amount)).To(Equal(expctedMintFee))

				is, err = tb.call(eveKey, *router, "UniRemoveLiquidityOneCoin", tokens[0], tokens[1], withdrawLPAmount, tokens[0], AmountZero)
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

			tb.call(aliceKey, tokens[1], "Mint", bob, token0Amount)
			tb.call(aliceKey, tokens[2], "Mint", bob, token1Amount)
			tb.call(bobKey, tokens[1], "Approve", swap, MaxUint256)
			tb.call(bobKey, tokens[2], "Approve", swap, MaxUint256)

			_, err := tb.call(bobKey, swap, "AddLiquidity", []*amount.Amount{token0Amount, token1Amount}, amount.NewAmount(0, 0))
			Expect(err).To(Succeed())
		})

		It("Stableswap : GetAmountsOut, SwapExactTokensForTokens : not WhiteList", func() {

			tb.call(aliceKey, tokens[1], "Mint", charlie, swapAmount)
			tb.call(charlieKey, tokens[1], "Approve", *router, MaxUint256)

			is, err := tb.view(*router, "GetAmountsOut", swapAmount, []common.Address{tokens[1], tokens[2]})
			Expect(err).To(Succeed())
			amounts := is[0].([]*amount.Amount)
			Expect(amounts[0]).To(Equal(swapAmount))
			Expect(amounts[1]).To(Equal(expectedOutputAmount))

			is, err = tb.call(charlieKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[1], tokens[2]})
			Expect(err).To(Succeed())
			amounts = is[0].([]*amount.Amount)
			Expect(amounts[0]).To(Equal(swapAmount))
			Expect(amounts[1]).To(Equal(expectedOutputAmount))
			Expect(tokenBalanceOf(tb.ctx, tokens[1], charlie).Cmp(AmountZero.Int)).To(Equal(0))
			Expect(tokenBalanceOf(tb.ctx, tokens[2], charlie)).To(Equal(expectedOutputAmount))
		})

		It("Uniswap : GetAmountsOut, SwapExactTokensForTokens : WhiteList", func() {

			tb.call(aliceKey, tokens[1], "Mint", eve, swapAmount)
			tb.call(eveKey, tokens[1], "Approve", *router, MaxUint256)

			is, err := tb.viewFrom(eve, *router, "GetAmountsOut", swapAmount, []common.Address{tokens[1], tokens[2]})
			Expect(err).To(Succeed())
			amounts := is[0].([]*amount.Amount)
			Expect(amounts[0]).To(Equal(swapAmount))
			Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))

			is, err = tb.call(eveKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[1], tokens[2]})
			Expect(err).To(Succeed())
			amounts = is[0].([]*amount.Amount)
			Expect(amounts[0]).To(Equal(swapAmount))
			Expect(amounts[1]).To(Equal(wlExpectedOutputAmount))
			Expect(tokenBalanceOf(tb.ctx, tokens[1], eve).Cmp(AmountZero.Int)).To(Equal(0))
			Expect(tokenBalanceOf(tb.ctx, tokens[2], eve)).To(Equal(wlExpectedOutputAmount))
		})

	})
})
