package test

import (
	"math/big"
	"math/rand"
	"strconv"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SwapExactTokensForTokens", func() {

	path := "_data_s"
	chainID := big.NewInt(65535)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		panic(err)
	}
	aliceKey, bobKey, charlieKey, eveKey := userKeys[0], userKeys[1], userKeys[2], userKeys[3]
	alice, bob, charlie, eve := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address(), eveKey.PublicKey().Address()
	admin := alice

	args := []interface{}{alice, bob, charlie, eve} // alice(admin), bob, charlie, eve

	// 체인생성 및  mev 생성
	tb, _ := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	//tb, ret := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	defer removeChainData(path)

	//mev := ret[0].(common.Address)

	var factory, router, whiteList *common.Address
	var tokens []common.Address

	// default 값
	_SupplyTokens := []*amount.Amount{amount.NewAmount(500000, 0), amount.NewAmount(1000000, 0)}

	_WinnerFee := _Fee5000
	_GroupId := hash.BigToHash(big.NewInt(100))

	_PairName := "__UNI_NAME"
	_PairSymbol := "__UNI_SYMBOL"
	_SwapName := "__STABLE_NAME"
	_SwapSymbol := "__STABLE_SYMBOL"
	_Amp := int64(360 * 2)

	deployContracts := func() {

		whiteList, err = whiteListDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

		factory, err = factoryDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

		router, err = routerDeploy(tb, aliceKey, factory)
		if err != nil {
			panic(err)
		}

	}

	deployTokens := func(senderKey key.Key, size int) []common.Address {

		_coins := make([]common.Address, size, size)

		// erc20Token deploy
		erc20Token, err := erc20TokenDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		if err != nil {
			panic(err)
		}

		idx := rand.Intn(int(size))

		for k := 0; k < size; k++ {
			if k != idx {
				token, err := tokenDeploy(tb, senderKey, "Token"+strconv.Itoa(int(k)), "TOKEN"+strconv.Itoa(int(k)))
				if err != nil {
					panic(err)
				}
				_coins[k] = *token
			} else {
				_coins[k] = *erc20Token
			}

			_, err = tb.call(aliceKey, _coins[k], "SetMinter", alice, true)
			if err != nil {
				panic(err)
			}
		}

		return _coins
	}

	It("0->1->2", func() {
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
			Expect(err).To(Succeed())
		}

		is, err := tb.call(aliceKey, *factory, "CreatePairUni", tokens[0], tokens[1], AddressZero, _PairName, _PairSymbol, alice, charlie, _Fee30, _AdminFee6, _WinnerFee, *whiteList, _GroupId, ClassMap["UniSwap"])
		Expect(err).To(Succeed())
		//pair = &is[0].(common.Address)

		is, err = tb.call(aliceKey, *factory, "CreatePairStable", tokens[1], tokens[2], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee30, _AdminFee6, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
		Expect(err).To(Succeed())
		swap := is[0].(common.Address)

		tb.call(aliceKey, tokens[0], "Mint", alice, _SupplyTokens[0])
		tb.call(aliceKey, tokens[1], "Mint", alice, _SupplyTokens[1])
		tb.call(aliceKey, tokens[0], "Approve", *router, MaxUint256)
		tb.call(aliceKey, tokens[1], "Approve", *router, MaxUint256)

		_, err = tb.call(aliceKey, *router, "UniAddLiquidity", tokens[0], tokens[1], _SupplyTokens[0], _SupplyTokens[1], AmountZero, AmountZero)
		Expect(err).To(Succeed())

		tb.call(aliceKey, tokens[1], "Mint", bob, _SupplyTokens[0])
		tb.call(aliceKey, tokens[2], "Mint", bob, _SupplyTokens[1])
		tb.call(bobKey, tokens[1], "Approve", swap, MaxUint256)
		tb.call(bobKey, tokens[2], "Approve", swap, MaxUint256)

		_, err = tb.call(bobKey, swap, "AddLiquidity", _SupplyTokens, amount.NewAmount(0, 0))
		Expect(err).To(Succeed())

		swapAmount := amount.NewAmount(1, 0)
		expectedOutputAmount1 := amount.NewAmount(0, 1993996023971928199)
		expectedOutputAmount2 := amount.NewAmount(0, 1990340394780245684)

		tb.call(aliceKey, tokens[0], "Mint", charlie, swapAmount)
		tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)

		is, err = tb.call(charlieKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{tokens[0], tokens[1], tokens[2]})
		Expect(err).To(Succeed())
		amounts := is[0].([]*amount.Amount)
		Expect(amounts[0]).To(Equal(swapAmount))
		Expect(amounts[1]).To(Equal(expectedOutputAmount1))
		Expect(amounts[2]).To(Equal(expectedOutputAmount2))
		Expect(tokenBalanceOf(tb.ctx, tokens[0], charlie).Cmp(AmountZero.Int)).To(Equal(0))
		Expect(tokenBalanceOf(tb.ctx, tokens[1], charlie).Cmp(AmountZero.Int)).To(Equal(0))
		Expect(tokenBalanceOf(tb.ctx, tokens[2], charlie)).To(Equal(expectedOutputAmount2))
	})

	It("Random", func() {

		for k := 0; k < 10; k++ {
			for length := 6; length < 7; length++ {
				pairs := make([]common.Address, length-1)

				deployContracts()
				tokens = deployTokens(aliceKey, length)

				swapAmount := amount.NewAmount(1, 0)
				tb.call(aliceKey, tokens[0], "Mint", charlie, swapAmount)

				for i := 0; i < length-1; i++ {
					switch rand.Intn(2) {
					//switch i % 2 {
					//switch 1 {
					case 0:
						is, err := tb.call(aliceKey, *factory, "CreatePairUni", tokens[i], tokens[i+1], AddressZero, _PairName, _PairSymbol, admin, charlie, _Fee30, _AdminFee6, _WinnerFee, *whiteList, _GroupId, ClassMap["UniSwap"])
						Expect(err).To(Succeed())
						pairs[i] = is[0].(common.Address)

						tb.call(aliceKey, tokens[i], "Mint", alice, _SupplyTokens[0])
						tb.call(aliceKey, tokens[i+1], "Mint", alice, _SupplyTokens[1])
						tb.call(aliceKey, tokens[i], "Approve", *router, MaxUint256)
						tb.call(aliceKey, tokens[i+1], "Approve", *router, MaxUint256)

						_, err = tb.call(aliceKey, *router, "UniAddLiquidity", tokens[i], tokens[i+1], _SupplyTokens[0], _SupplyTokens[1], AmountZero, AmountZero)
						Expect(err).To(Succeed())
						//GPrintln("U")

					case 1:
						is, err := tb.call(aliceKey, *factory, "CreatePairStable", tokens[i], tokens[i+1], AddressZero, _SwapName, _SwapSymbol, admin, charlie, _Fee30, _AdminFee6, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
						Expect(err).To(Succeed())
						pairs[i] = is[0].(common.Address)

						tb.call(aliceKey, tokens[i], "Mint", bob, _SupplyTokens[0])
						tb.call(aliceKey, tokens[i+1], "Mint", bob, _SupplyTokens[1])
						tb.call(bobKey, tokens[i], "Approve", pairs[i], MaxUint256)
						tb.call(bobKey, tokens[i+1], "Approve", pairs[i], MaxUint256)

						_, err = tb.call(bobKey, pairs[i], "AddLiquidity", _SupplyTokens, amount.NewAmount(0, 0))
						Expect(err).To(Succeed())
						//GPrintln("S")
					}
				}

				tb.call(charlieKey, tokens[0], "Approve", *router, MaxUint256)

				is, err := tb.call(charlieKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, tokens)
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				for j := 0; j < length-1; j++ {
					Expect(tokenBalanceOf(tb.ctx, tokens[j], charlie).Cmp(AmountZero.Int)).To(Equal(0))
				}
				Expect(tokenBalanceOf(tb.ctx, tokens[length-1], charlie)).To(Equal(amounts[length-1]))
			}
		}
	})
})
