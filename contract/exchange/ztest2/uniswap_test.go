package test2

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// test : ginkgo
//        ginkgo -v  (verbose mode)
// skip : It("...", func() {
//           if(condition)  {
//			 	Skip("생략이유")
//           }
//         })
// focus : It -> FIt,  Describe -> FDescribe
var _ = Describe("Uniswap Router Test", func() {
	path := "_data"
	chainID := big.NewInt(65535)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		panic(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()
	admin := alice

	args := []interface{}{alice, bob, charlie} // alice(admin), bob, charlie

	// 체인생성 및  mev 생성
	tb, _, err := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	if err != nil {
		panic(err)
	}
	defer removeChainData(path)

	//mev := ret[0].(common.Address)

	// 1. factory, router, whitelist Deploy
	// 1. Erc20TokenWrapper Contract Deploy
	// 2. Token Contract Deploy
	// 3. Pair Contract Creation
	// 4. token0, token1 sorting
	// 5. token approve to router
	beforeEach := func(fee, adminFee, winnerFee uint64) (*common.Address, *common.Address, *common.Address, *common.Address, *common.Address, *common.Address) {
		factory, err := factoryDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

		router, err := routerDeploy(tb, aliceKey, factory)
		if err != nil {
			panic(err)
		}

		whiteList, err := whiteListDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}

		// erc20Wrapper deploy
		erc20Wrapper, _, err := erc20WrapperDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		if err != nil {
			panic(err)
		}

		// token deploy
		token, err := tokenDeploy(tb, aliceKey, "Token", "TKN")
		if err != nil {
			panic(err)
		}

		// setMinter
		for _, token := range []common.Address{*erc20Wrapper, *token} {
			err = TokenSetMinter(tb, aliceKey, token, alice, true)
			if err != nil {
				panic(err)
			}
		}

		tokenA, tokenB := erc20Wrapper, token

		// pair create
		pC := &pairContractConstruction{
			TokenA:    *tokenA,
			TokenB:    *tokenB,
			PayToken:  common.Address{},
			Name:      "__UNI_NAME",
			Symbol:    "__UNI_SYMBOL",
			Owner:     alice,
			Winner:    charlie,
			Fee:       fee,       // uint64(40000000)
			AdminFee:  adminFee,  // uint64(trade.MAX_ADMIN_FEE),
			WinnerFee: winnerFee, // uint64(5000000000),
			Factory:   *factory,
			WhiteList: *whiteList,
			GroupId:   hash.BigToHash(big.NewInt(1)),
		}

		pair, err := pairCreate(tb, aliceKey, pC)
		if err != nil {
			panic(err)
		}

		// // token0, token1 sorting
		// token0, token1, err = trade.SortTokens(*erc20Wrapper, *token)
		// if err != nil {
		// 	panic(err)
		// }

		// approve
		// for _, token := range []common.Address{token0, token1} {
		// 	for _, senderKey := range []key.Key{aliceKey, bobKey, charlieKey} {
		// 		ctx, err = TokenApprove(tb, ctx, senderKey, token, router, MaxUint256)
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 	}
		// }

		return factory, router, whiteList, tokenA, tokenB, pair
	}

	beforeEachDefault := func() (*common.Address, *common.Address, *common.Address, *common.Address, *common.Address, *common.Address) {
		fee, adminFee, winnerFee := uint64(40000000), uint64(trade.MAX_ADMIN_FEE), uint64(5000000000)
		return beforeEach(fee, adminFee, winnerFee)
	}

	initMint := func(to common.Address, args ...any) {
		for i, arg := range args {
			err = TokenMint(tb, aliceKey, *(arg.(*common.Address)), to, _SupplyTokens[i])
			if err != nil {
				panic(err)
			}
		}
	}

	// AfterEach(func() {
	// 	removeChainData(path)
	// })

	Describe("factory", func() {

		It("Owner, allPairsLength", func() {
			factory, err := factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())

			//Owner
			is, err := Exec(tb.ctx, AddressZero, *factory, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(admin))

			//AllPairsLength
			is, err = Exec(tb.ctx, admin, *factory, "AllPairsLength", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(0)))
		})

		It("CreatePairUni : PayToken Error", func() {

			factory, err := factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())

			// pair create
			pC := &pairContractConstruction{
				TokenA:    common.BigToAddress(big.NewInt(1)),
				TokenB:    common.BigToAddress(big.NewInt(2)),
				PayToken:  common.BigToAddress(big.NewInt(3)), // not exist paytoken
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       uint64(40000000),
				AdminFee:  uint64(trade.MAX_ADMIN_FEE),
				WinnerFee: uint64(5000000000),
				Factory:   *factory,
				WhiteList: common.Address{},
				GroupId:   hash.BigToHash(big.NewInt(1)),
			}

			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))

			// _, err := Exec(genesis, admin, factory, "CreatePairUni", []interface{}{*tokenA, *tokenB, alice, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})

		})

		It("CreatePairUni", func() {

			// contract 생성
			factory, _, whiteList, tokenA, tokenB, pair := beforeEachDefault()

			gId := hash.BigToHash(big.NewInt(1))

			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
			Expect(err).To(Succeed())

			// PairFor
			pairFor, err := trade.PairFor(*factory, *tokenA, *tokenB)
			Expect(err).To(Succeed())
			Expect(pairFor).To(Equal(*pair))

			// AllPairs
			is, err := Exec(tb.ctx, admin, *factory, "AllPairs", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(*pair))

			// AllPairsLength
			is, err = Exec(tb.ctx, admin, *factory, "AllPairsLength", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = Exec(tb.ctx, admin, *factory, "GetPair", []interface{}{*tokenA, *tokenB})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// GetPair of reverse tokens
			is, err = Exec(tb.ctx, admin, *factory, "GetPair", []interface{}{*tokenB, *tokenA})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// CreatePairUni of same tokens
			pC := &pairContractConstruction{
				TokenA:    *tokenA,
				TokenB:    *tokenB,
				PayToken:  AddressZero,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       uint64(40000000),
				AdminFee:  uint64(trade.MAX_ADMIN_FEE),
				WinnerFee: uint64(5000000000),
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   gId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// CreatePairUni of same reverse tokens
			pC = &pairContractConstruction{
				TokenA:    *tokenB,
				TokenB:    *tokenA,
				PayToken:  AddressZero,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       uint64(40000000),
				AdminFee:  uint64(trade.MAX_ADMIN_FEE),
				WinnerFee: uint64(5000000000),
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   gId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = Exec(tb.ctx, alice, *pair, "Factory", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*factory))

			// token0
			is, err = Exec(tb.ctx, alice, *pair, "Token0", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token0))

			// token1
			is, err = Exec(tb.ctx, alice, *pair, "Token1", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token1))

			// tokens
			is, err = Exec(tb.ctx, alice, *pair, "Tokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

			// WhiteList
			is, err = Exec(tb.ctx, alice, *pair, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = Exec(tb.ctx, alice, *pair, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))
		})

		It("CreatePairUni : reverse", func() {

			// contract 생성
			factory, _, whiteList, tokenA, tokenB, pair := beforeEachDefault()

			gId := hash.BigToHash(big.NewInt(1))
			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenB, *tokenA)
			Expect(err).To(Succeed())

			// PairFor
			pairFor, err := trade.PairFor(*factory, *tokenB, *tokenA)
			Expect(err).To(Succeed())
			Expect(pairFor).To(Equal(*pair))

			// AllPairs
			is, err := Exec(tb.ctx, admin, *factory, "AllPairs", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(*pair))

			// AllPairsLength
			is, err = Exec(tb.ctx, admin, *factory, "AllPairsLength", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = Exec(tb.ctx, admin, *factory, "GetPair", []interface{}{*tokenB, *tokenA})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// GetPair of reverse tokens
			is, err = Exec(tb.ctx, admin, *factory, "GetPair", []interface{}{*tokenA, *tokenB})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// CreatePairUni of same tokens
			pC := &pairContractConstruction{
				TokenA:    *tokenB,
				TokenB:    *tokenA,
				PayToken:  AddressZero,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       uint64(40000000),
				AdminFee:  uint64(trade.MAX_ADMIN_FEE),
				WinnerFee: uint64(5000000000),
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   gId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// CreatePairUni of same reverse tokens
			pC = &pairContractConstruction{
				TokenA:    *tokenA,
				TokenB:    *tokenB,
				PayToken:  AddressZero,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       uint64(40000000),
				AdminFee:  uint64(trade.MAX_ADMIN_FEE),
				WinnerFee: uint64(5000000000),
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   gId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = Exec(tb.ctx, alice, *pair, "Factory", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(*factory))

			// token0
			is, err = Exec(tb.ctx, alice, *pair, "Token0", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token0))

			// token1
			is, err = Exec(tb.ctx, alice, *pair, "Token1", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token1))

			// tokens
			is, err = Exec(tb.ctx, alice, *pair, "Tokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

			// WhiteList
			is, err = Exec(tb.ctx, alice, *pair, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = Exec(tb.ctx, alice, *pair, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))
		})
	})

	Describe("router", func() {

		It("factory", func() {

			factory, err := factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())
			router, err := routerDeploy(tb, aliceKey, factory)
			Expect(err).To(Succeed())

			is, err := Exec(tb.ctx, alice, *router, "Factory", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(*factory))
		})

		It("UniAddLiquidity : Initial Supply", func() {
			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE}     // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_ADMIN_FEE}   // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_WINNER_FEE} // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						_, router, _, tokenA, tokenB, pair := beforeEach(fee, adminFee, winnerFee)
						initMint(alice, tokenA, tokenB)

						tokenAAmount := amount.NewAmount(1, 0)
						tokenBAmount := amount.NewAmount(4, 0)

						expectedLiquidity := amount.NewAmount(2, 0)

						err = TokenApprove(tb, aliceKey, *tokenA, *router, MaxUint256)
						Expect(err).To(Succeed())
						err = TokenApprove(tb, aliceKey, *tokenB, *router, MaxUint256)
						Expect(err).To(Succeed())

						Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0]))
						Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1]))

						err = UniAddLiquidity(tb, aliceKey, *router, *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)
						Expect(err).To(Succeed())

						// BalanceOf
						Expect(tokenBalanceOf(tb.ctx, *tokenA, *pair)).To(Equal(tokenAAmount))
						Expect(tokenBalanceOf(tb.ctx, *tokenB, *pair)).To(Equal(tokenBAmount))
						Expect(tokenBalanceOf(tb.ctx, *pair, common.ZeroAddr)).To(Equal(_MininumLiquidity))
						Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(expectedLiquidity.Sub(_MininumLiquidity)))
					}
				}
			}
		})
	})
})
