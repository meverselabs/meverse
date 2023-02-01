package test

import (
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/exchange/trade"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("uniswap Test", func() {
	path := "_data_u"
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
	tb, ret := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	defer removeChainData(path)

	mev := ret[0].(common.Address)

	var factory, router, whiteList, tokenA, tokenB, pair *common.Address

	// default 값
	_ML := amount.NewAmount(0, trade.MINIMUM_LIQUIDITY) // minimum Liquidity
	_SupplyTokens := []*amount.Amount{amount.NewAmount(500000, 0), amount.NewAmount(1000000, 0)}

	_Fee := _Fee40
	_AdminFee := uint64(trade.MAX_ADMIN_FEE)
	_WinnerFee := _Fee5000
	_GroupId := hash.BigToHash(big.NewInt(100))

	// deployContracts create contracts
	// 1. factory, router, whitelist Deploy
	// 1. Erc20Token Contract Deploy
	// 2. Token Contract Deploy
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

		// erc20Token deploy
		erc20Token, err := erc20TokenDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		if err != nil {
			panic(err)
		}

		// // erc20Wrapper deploy
		// erc20Wrapper, _, err := erc20WrapperDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		// if err != nil {
		// 	panic(err)
		// }

		// token deploy
		token, err := tokenDeploy(tb, aliceKey, "Token", "TKN")
		if err != nil {
			panic(err)
		}

		tokenA, tokenB = erc20Token, token

		// setMinter
		for _, token := range []common.Address{*tokenA, *tokenB} {
			_, err = tb.call(aliceKey, token, "SetMinter", alice, true)
			if err != nil {
				panic(err)
			}
		}
	}

	// beforeEachDefault create contracts
	// 1. factory, router, whitelist Deploy
	// 2. Token Contract Deploy
	// 3. Pair Contract Creation
	// 4. token0, token1 sorting
	// 5. token approve to router
	beforeEach := func(fee, adminFee, winnerFee uint64) {

		deployContracts()

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
			GroupId:   _GroupId,
		}

		pair, err = pairCreate(tb, aliceKey, pC)
		Expect(err).To(Succeed())
	}

	// beforeEachDefault create contracts with default fee parameters
	beforeEachDefault := func() {
		beforeEach(_Fee, _AdminFee, _WinnerFee)
	}

	// uniMint mint tokens to to Address by token owner
	uniMint := func(to common.Address, contA, contB *common.Address) {
		_, err := tb.call(aliceKey, *contA, "Mint", to, _SupplyTokens[0])
		Expect(err).To(Succeed())

		_, err = tb.call(aliceKey, *contB, "Mint", to, _SupplyTokens[1])
		Expect(err).To(Succeed())
	}

	// uniApprove approve tokens to router
	uniApprove := func(ownerKey key.Key, contA, contB *common.Address) {
		_, err := tb.call(ownerKey, *contA, "Approve", *router, MaxUint256)
		Expect(err).To(Succeed())
		_, err = tb.call(ownerKey, *contB, "Approve", *router, MaxUint256)
		Expect(err).To(Succeed())
	}

	// uniAddLiquidity prepare initial pair'2 UniAddLiquidity
	uniAddLiquidity := func(ownerKey key.Key, contA, contB *common.Address, supplyA, supplyB *amount.Amount) {
		owner := ownerKey.PublicKey().Address()
		uniMint(owner, contA, contB)
		uniApprove(ownerKey, contA, contB)

		_, err = tb.call(ownerKey, *router, "UniAddLiquidity", *contA, *contB, supplyA, supplyB, AmountZero, AmountZero)
		Expect(err).To(Succeed())
	}

	// uniAddLiquidityDefault prepare initial default uniAddLiquidity
	uniAddLiquidityDefault := func(ownerKey key.Key) {
		uniAddLiquidity(ownerKey, tokenA, tokenB, _SupplyTokens[0], _SupplyTokens[1])

	}

	Describe("Exchange", func() {

		BeforeEach(func() {
			deployContracts()

			// pair create
			pC := &pairContractConstruction{
				TokenA:    *tokenA,
				TokenB:    *tokenB,
				PayToken:  *tokenA,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       _Fee,
				AdminFee:  _AdminFee,
				WinnerFee: _WinnerFee,
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   _GroupId,
			}

			pair, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(Succeed())
		})

		It("payToken", func() {
			is, err := tb.view(*pair, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*tokenA))
		})

		It("SetPayToken, onlyOwner", func() {

			payToken := *tokenB

			is, err := tb.call(aliceKey, *pair, "SetPayToken", payToken)
			Expect(err).To(Succeed())

			is, err = tb.view(*pair, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))

			is, err = tb.call(bobKey, *pair, "SetPayToken", payToken)
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetPayToken : ZeroAddress ", func() {

			payToken := common.Address{}

			is, err := tb.call(aliceKey, *pair, "SetPayToken", payToken)
			Expect(err).To(Succeed())

			is, err = tb.view(*pair, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))
		})

		It("SetPayToken : Another token", func() {

			payToken := common.HexToAddress("1")

			_, err := tb.call(aliceKey, *pair, "SetPayToken", payToken)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))
		})
	})

	Describe("factory", func() {

		It("Owner, allPairsLength", func() {
			factory, err = factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())

			//Owner
			is, err := tb.view(*factory, "Owner")
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(admin))

			//AllPairsLength
			is, err = tb.view(*factory, "AllPairsLength")
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(0)))
		})

		It("setOwner", func() {

			_, err := tb.call(bobKey, *factory, "SetOwner", alice)
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			_, err = tb.call(aliceKey, *factory, "SetOwner", bob)
			Expect(err).To(Succeed())

			is, err := tb.view(*factory, "Owner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			_, err = tb.call(aliceKey, *factory, "SetOwner", bob)
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("CreatePairUni : PayToken Error", func() {

			factory, err = factoryDeploy(tb, aliceKey)
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
				GroupId:   _GroupId,
			}

			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))

		})

		It("CreatePairUni", func() {

			// contract 생성
			beforeEachDefault()

			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
			Expect(err).To(Succeed())

			// PairFor
			pairFor, err := trade.PairFor(*factory, *tokenA, *tokenB)
			Expect(err).To(Succeed())
			Expect(pairFor).To(Equal(*pair))

			// AllPairs
			is, err := tb.view(*factory, "AllPairs")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(*pair))

			// AllPairsLength
			is, err = tb.view(*factory, "AllPairsLength")
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = tb.view(*factory, "GetPair", *tokenA, *tokenB)
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// GetPair of reverse tokens
			is, err = tb.view(*factory, "GetPair", *tokenB, *tokenA)
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
				GroupId:   _GroupId,
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
				GroupId:   _GroupId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = tb.view(*pair, "Factory")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*factory))

			// token0
			is, err = tb.view(*pair, "Token0")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token0))

			// token1
			is, err = tb.view(*pair, "Token1")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token1))

			// tokens
			is, err = tb.view(*pair, "Tokens")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

			// WhiteList
			is, err = tb.view(*pair, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = tb.view(*pair, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})

		It("CreatePairUni : reverse", func() {

			// contract 생성
			beforeEachDefault()

			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenB, *tokenA)
			Expect(err).To(Succeed())

			// PairFor
			pairFor, err := trade.PairFor(*factory, *tokenB, *tokenA)
			Expect(err).To(Succeed())
			Expect(pairFor).To(Equal(*pair))

			// AllPairs
			is, err := tb.view(*factory, "AllPairs")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(*pair))

			// AllPairsLength
			is, err = tb.view(*factory, "AllPairsLength")
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = tb.view(*factory, "GetPair", *tokenB, *tokenA)
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*pair))

			// GetPair of reverse tokens
			is, err = tb.view(*factory, "GetPair", *tokenA, *tokenB)
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
				GroupId:   _GroupId,
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
				GroupId:   _GroupId,
			}
			_, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = tb.view(*pair, "Factory")
			Expect(err).To(Succeed())
			Expect(is[0]).To(Equal(*factory))

			// token0
			is, err = tb.view(*pair, "Token0")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token0))

			// token1
			is, err = tb.view(*pair, "Token1")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(token1))

			// tokens
			is, err = tb.view(*pair, "Tokens")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

			// WhiteList
			is, err = tb.view(*pair, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = tb.view(*pair, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})
	})

	Describe("pair", func() {

		Describe("lp token", func() {

			var _TotalSupply *amount.Amount
			_TestAmount := amount.NewAmount(10, 0)

			BeforeEach(func() {
				beforeEachDefault()

				uniAddLiquidityDefault(aliceKey) // pair.balanceOf(alice) > 0
				_TotalSupply = tokenTotalSupply(tb.ctx, *pair)
			})

			It("Name, Symbol, TotalSupply, Decimals", func() {
				//Name string
				is, err := tb.view(*pair, "Name")
				Expect(err).To(Succeed())
				Expect(is[0].(string)).To(Equal("__UNI_NAME"))

				//Symbol string
				is, err = tb.view(*pair, "Symbol")
				Expect(err).To(Succeed())
				Expect(is[0].(string)).To(Equal("__UNI_SYMBOL"))

				//TotalSupply *amount.Amount
				is, err = tb.view(*pair, "TotalSupply")
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

				//Decimals *big.Int
				is, err = tb.view(*pair, "Decimals")
				Expect(err).To(Succeed())
				Expect(is[0].(*big.Int)).To(Equal(big.NewInt(amount.FractionalCount)))
			})

			It("Transfer", func() {
				//Transfer(To common.Address, Amount *amount.Amount)

				_, err := tb.call(aliceKey, *pair, "Transfer", bob, _TestAmount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(_TestAmount))
			})

			It("Approve", func() {
				//Approve(To common.Address, Amount *amount.Amount)
				_, err := tb.call(aliceKey, *pair, "Approve", bob, _TestAmount)
				Expect(err).To(Succeed())

				Expect(tokenAllowance(tb.ctx, *pair, alice, bob)).To(Equal(_TestAmount))
			})

			It("IncreaseAllowance", func() {
				//IncreaseAllowance(spender common.Address, addAmount *amount.Amount)
				_, err := tb.call(aliceKey, *pair, "IncreaseAllowance", bob, _TestAmount)
				Expect(err).To(Succeed())

				Expect(tokenAllowance(tb.ctx, *pair, alice, bob)).To(Equal(_TestAmount))
			})

			It("DecreaseAllowance", func() {
				//DecreaseAllowance(spender common.Address, subtractAmount *amount.Amount)

				tb.call(aliceKey, *pair, "Approve", bob, _TestAmount.MulC(3))

				_, err = tb.call(aliceKey, *pair, "DecreaseAllowance", bob, _TestAmount)
				Expect(err).To(Succeed())

				Expect(tokenAllowance(tb.ctx, *pair, alice, bob)).To(Equal(_TestAmount.MulC(2)))
			})

			It("TransferFrom", func() {
				//TransferFrom(From common.Address, To common.Address, Amount *amount.Amount)

				balance := tokenBalanceOf(tb.ctx, *pair, alice)

				tb.call(aliceKey, *pair, "Approve", bob, _TestAmount.MulC(3))

				_, err = tb.call(bobKey, *pair, "TransferFrom", alice, charlie, _TestAmount)
				Expect(err).To(Succeed())

				Expect(tokenAllowance(tb.ctx, *pair, alice, bob)).To(Equal(_TestAmount.MulC(2)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(balance.Sub(_TestAmount)))
				Expect(tokenBalanceOf(tb.ctx, *pair, charlie)).To(Equal(_TestAmount))
			})

		})

		Describe("uniswap front", func() {

			BeforeEach(func() {
				beforeEachDefault()
			})

			It("Extype, Fee, AdminFee, NTokens, Rates, PrecisionMul, Tokens, Owner, WhiteList, GroupId", func() {
				//Extype uint8
				is, err := tb.view(*pair, "ExType")
				Expect(err).To(Succeed())
				Expect(is[0].(uint8)).To(Equal(trade.UNI))

				// NTokens(cc *types.ContractContext) uint8
				is, err = tb.view(*pair, "NTokens")
				Expect(err).To(Succeed())
				Expect(is[0].(uint8)).To(Equal(uint8(2)))

				token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
				Expect(err).To(Succeed())

				// Coins []common.Address
				is, err = tb.view(*pair, "Tokens")
				Expect(err).To(Succeed())
				Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

				// Owner common.Address
				is, err = tb.view(*pair, "Owner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(alice))

				// Fee uint64
				is, err = tb.view(*pair, "Fee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(_Fee))

				// AdminFee uint64
				is, err = tb.view(*pair, "AdminFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(_AdminFee))

				// WinnerFee uint64
				is, err = tb.view(*pair, "WinnerFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(_WinnerFee))

				// WhiteList
				is, err = tb.view(*pair, "WhiteList")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(*whiteList))

				// GroupId
				is, err = tb.view(*pair, "GroupId")
				Expect(err).To(Succeed())
				Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
			})

			It("CommitNewFee, AdminActionsDeadline, FutureFee, FutureAdminFee, ApplyNewFee, RevertNewFee", func() {
				fee := uint64(rand.Intn(trade.MAX_FEE) + 1)
				admin_fee := uint64(rand.Intn(trade.MAX_ADMIN_FEE + 1))
				winner_fee := uint64(rand.Intn(trade.MAX_WINNER_FEE + 1))
				delay := uint64(3 * 86400)

				_, err := tb.call(aliceKey, *pair, "CommitNewFee", fee, admin_fee, winner_fee, delay)
				Expect(err).To(Succeed())

				// AdminActionsDeadline uint64
				is, err := tb.view(*pair, "AdminActionsDeadline")
				Expect(is[0].(uint64)).To(Equal(tb.ctx.LastTimestamp()/uint64(time.Second) - tb.step/1000 + delay))

				// FutureFee uint64
				is, err = tb.view(*pair, "FutureFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

				// FutureAdminFee uint64
				is, err = tb.view(*pair, "FutureAdminFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(admin_fee))

				// FutureWinnerFee uint64
				is, err = tb.view(*pair, "FutureWinnerFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(winner_fee))

				tb.sleep(delay*1000 - 1000)

				_, err = tb.call(aliceKey, *pair, "ApplyNewFee")
				Expect(err).To(Succeed())

				// AdminActionsDeadline uint64
				is, err = tb.view(*pair, "AdminActionsDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// Fee uint64
				is, err = tb.view(*pair, "Fee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

				// AdminFee uint64
				is, err = tb.view(*pair, "AdminFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(admin_fee))

				// WinnerFee uint64
				is, err = tb.view(*pair, "WinnerFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(winner_fee))

				_, err = tb.call(aliceKey, *pair, "CommitNewFee", uint64(1000), uint64(2000), uint64(3000), delay)
				Expect(err).To(Succeed())

				_, err = tb.call(aliceKey, *pair, "RevertNewFee")
				Expect(err).To(Succeed())

				// AdminActionsDeadline uint64
				is, err = tb.view(*pair, "AdminActionsDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// Fee uint64
				is, err = tb.view(*pair, "Fee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

				// AdminFee uint64
				is, err = tb.view(*pair, "AdminFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(admin_fee))

				// WinnerFee uint64
				is, err = tb.view(*pair, "WinnerFee")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(winner_fee))
			})

			It("CommitNewWhiteList, WhiteListDeadline, FutureWhiteList, FutureGroupId, ApplyNewWhiteList, RevertNewWhiteList", func() {

				wl, err := whiteListDeploy(tb, aliceKey)
				Expect(err).To(Succeed())

				delay := uint64(3 * 86400)
				timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)
				_, err = tb.call(aliceKey, *pair, "CommitNewWhiteList", *wl, _GroupId, delay)
				Expect(err).To(Succeed())

				// WhiteListDeadline() uint64
				//timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)
				is, err := tb.view(*pair, "WhiteListDeadline")
				Expect(is[0].(uint64)).To(Equal(timestamp + delay))

				// FutureWhiteList() common.Address
				is, err = tb.view(*pair, "FutureWhiteList")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(*wl))

				// FutureGroupId() hash.Hash256
				is, err = tb.view(*pair, "FutureGroupId")
				Expect(err).To(Succeed())
				Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))

				tb.sleep(delay*1000 - 1000)

				_, err = tb.call(aliceKey, *pair, "ApplyNewWhiteList")
				Expect(err).To(Succeed())

				// WhiteListDeadline() uint64
				is, err = tb.view(*pair, "WhiteListDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// WhiteList() common.Address
				is, err = tb.view(*pair, "WhiteList")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(*wl))

				// GroupId() hash.Hash256
				is, err = tb.view(*pair, "GroupId")
				Expect(err).To(Succeed())
				Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))

				_, err = tb.call(aliceKey, *pair, "CommitNewWhiteList", *whiteList, _GroupId, delay)
				Expect(err).To(Succeed())

				// RevertNewWhiteList()
				_, err = tb.call(aliceKey, *pair, "RevertNewWhiteList")
				Expect(err).To(Succeed())

				// AdminActionsDeadline() uint64
				is, err = tb.view(*pair, "WhiteListDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// WhiteList() common.Address
				is, err = tb.view(*pair, "WhiteList")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(*wl))

				// GroupId() hash.Hash256
				is, err = tb.view(*pair, "GroupId")
				Expect(err).To(Succeed())
				Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
			})

			It("Owner, CommitTransferOwnerWinner, TransferOwnerWinnerDeadline, ApplyTransferOwnerWinner, FutureOwner, RevertTransferOwnerWinner", func() {
				delay := uint64(3 * 86400)

				// CommitTransferOwnerWinner(_owner common.Address)
				timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)
				_, err = tb.call(aliceKey, *pair, "CommitTransferOwnerWinner", bob, charlie, delay)
				Expect(err).To(Succeed())

				// TransferOwnerWinnerDeadline uint64
				is, err := tb.view(*pair, "TransferOwnerWinnerDeadline")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(timestamp + delay))

				// FutureOwner common.Address
				is, err = tb.view(*pair, "FutureOwner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(bob))

				// FutureOwner common.Address
				is, err = tb.view(*pair, "FutureWinner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(charlie))

				tb.sleep(delay*1000 - 1000)

				// ApplyTransferOwnerWinner(cc *types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "ApplyTransferOwnerWinner")
				Expect(err).To(Succeed())

				// TransferOwnerWinnerDeadline uint64
				is, _ = tb.view(*pair, "TransferOwnerWinnerDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// Owner common.Address
				is, err = tb.view(*pair, "Owner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(bob))

				// Owner common.Address
				is, err = tb.view(*pair, "Winner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(charlie))

				_, err = tb.call(bobKey, *pair, "CommitTransferOwnerWinner", charlie, alice, delay)
				Expect(err).To(Succeed())

				// RevertTransferOwnerWinner(cc *types.ContractContext)
				_, err = tb.call(bobKey, *pair, "RevertTransferOwnerWinner")
				Expect(err).To(Succeed())

				// TransferOwnerWinnerDeadline uint64
				is, _ = tb.view(*pair, "TransferOwnerWinnerDeadline")
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

				// Owner common.Address
				is, err = tb.view(*pair, "Owner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(bob))

				// Owner common.Address
				is, err = tb.view(*pair, "Winner")
				Expect(err).To(Succeed())
				Expect(is[0].(common.Address)).To(Equal(charlie))
			})

			It("WithdrawAdminFees2, AdminBalance, Skim, Sync", func() {
				uniAddLiquidityDefault(aliceKey)
				aliceLPBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				uniMint(bob, tokenA, tokenB)
				uniApprove(bobKey, tokenA, tokenB)

				swapAmount := amount.NewAmount(1, 0)

				Expect(aliceLPBalance.Cmp(swapAmount.Int) > 0).To(BeTrue())
				// token0 -> token1 SwapExactTokensForTokens

				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenA, *tokenB})
				Expect(err).To(Succeed())

				// AdminBalance(cc *types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "AdminBalance")
				Expect(err).To(Succeed())

				// WithdrawAdminFees2(cc *types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(err).To(Succeed())

				// token1 -> token0 SwapExactTokensForTokens
				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenB, *tokenA})
				Expect(err).To(Succeed())

				// Sync
				_, err = tb.call(aliceKey, *pair, "Sync")
				Expect(err).To(Succeed())

				// transfer balance != reserve
				_, err = tb.call(bobKey, *tokenA, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				// Skim
				_, err = tb.call(aliceKey, *pair, "Skim", charlie)
				Expect(err).To(Succeed())

				// transfer balance != reserve
				_, err = tb.call(bobKey, *tokenA, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				// transfer balance != reserve
				_, err = tb.call(aliceKey, *pair, "Sync")
				Expect(err).To(Succeed())
			})

			It("KillMe, UnkillMe", func() {

				// UnkillMe(cc * types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "UnkillMe")
				Expect(err).To(Succeed())

				// KillMe(cc * types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "KillMe")
				Expect(err).To(Succeed())

				// UnkillMe(cc * types.ContractContext)
				_, err = tb.call(aliceKey, *pair, "UnkillMe")
				Expect(err).To(Succeed())
			})
		})

		Describe("TokenTransfer", func() {
			_TestAmount := amount.NewAmount(10, 0)

			BeforeEach(func() {
				beforeEachDefault()
				uniMint(bob, tokenA, tokenB)
			})

			It("tokentransfer", func() {

				_, err := tb.call(bobKey, *tokenA, "Transfer", *pair, _TestAmount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *tokenA, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
				Expect(tokenBalanceOf(tb.ctx, *tokenA, *pair)).To(Equal(_TestAmount))

				_, err = tb.call(bobKey, *tokenB, "Transfer", *pair, _TestAmount.MulC(2))
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *tokenB, bob)).To(Equal(_SupplyTokens[1].Sub(_TestAmount.MulC(2))))
				Expect(tokenBalanceOf(tb.ctx, *tokenB, *pair)).To(Equal(_TestAmount.MulC(2)))

				_, err = tb.call(aliceKey, *pair, "TokenTransfer", *tokenA, charlie, _TestAmount.DivC(2))
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *tokenA, *pair)).To(Equal(_TestAmount.DivC(2)))
				Expect(tokenBalanceOf(tb.ctx, *tokenA, charlie)).To(Equal(_TestAmount.DivC(2)))

				_, err = tb.call(aliceKey, *pair, "TokenTransfer", *tokenB, charlie, _TestAmount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *tokenB, *pair)).To(Equal(_TestAmount))
				Expect(tokenBalanceOf(tb.ctx, *tokenB, charlie)).To(Equal(_TestAmount))

			})
			It("onlyOwner", func() {

				_, err := tb.call(bobKey, *pair, "TokenTransfer", *tokenA, charlie, _TestAmount)
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("transfer more than balance", func() {

				_, err := tb.call(bobKey, *tokenA, "Transfer", *pair, _TestAmount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *tokenA, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
				Expect(tokenBalanceOf(tb.ctx, *tokenA, *pair)).To(Equal(_TestAmount))

				_, err = tb.call(aliceKey, *pair, "TokenTransfer", *tokenA, charlie, _TestAmount.MulC(2))
				Expect(err).To(HaveOccurred())

			})

			It("not base-token transfer", func() {

				otherToken := common.BigToAddress(big.NewInt(1))
				_, err := tb.call(aliceKey, *pair, "TokenTransfer", otherToken, charlie, _TestAmount)
				Expect(err).To(MatchError("Exchange: NOT_EXIST_TOKEN"))

			})
		})

		Describe("uniswap contract : PayToken Null", func() {

			var token0, token1 common.Address
			var _TestAmount = amount.NewAmount(10, 0)

			BeforeEach(func() {
				beforeEach(_Fee30, _AdminFee6, _WinnerFee)

				// token0, token1 sorting
				token0, token1, err = trade.SortTokens(*tokenA, *tokenB)
				Expect(err).To(Succeed())
			})

			It("fee, adminFee, winnerFee", func() {
				for k := 0; k < 10; k++ {
					pct := uint64(rand.Intn(1000))
					fee := uint64(float64(int64(trade.MAX_FEE*pct)) / 1000.)
					pct = uint64(rand.Intn(1000))
					adminFee := uint64(float64(int64(trade.MAX_ADMIN_FEE*pct)) / 1000.)
					pct = uint64(rand.Intn(1000))
					winnerFee := uint64(float64(int64(trade.MAX_WINNER_FEE*pct)) / 1000.)

					setFees(tb, aliceKey, *pair, fee, adminFee, winnerFee)

					//fee
					is, err := tb.view(*pair, "Fee")
					Expect(err).To(Succeed())
					Expect(is[0]).To(Equal(fee))

					//AdminFee
					is, err = tb.view(*pair, "AdminFee")
					Expect(err).To(Succeed())
					Expect(is[0]).To(Equal(adminFee))

					//WinnerFee
					is, err = tb.view(*pair, "WinnerFee")
					Expect(err).To(Succeed())
					Expect(is[0]).To(Equal(winnerFee))
				}
			})
			It("mint", func() {
				uniMint(alice, tokenA, tokenB)

				token0Amount := amount.NewAmount(1, 0)
				token1Amount := amount.NewAmount(4, 0)

				// transfer
				_, err = tb.call(aliceKey, token0, "Transfer", *pair, token0Amount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(token0Amount))

				_, err = tb.call(aliceKey, token1, "Transfer", *pair, token1Amount)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(token1Amount))

				expectedLiquidity := amount.NewAmount(2, 0)

				// mint
				is, err := tb.call(aliceKey, *pair, "Mint", alice)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLiquidity.Sub(_ML)))

				// totalSupply
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(expectedLiquidity))

				// BalanceOf *pair, token0, token1
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(expectedLiquidity.Sub(_ML)))
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(token0Amount))
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(token1Amount))

				// Reserves
				is, err = tb.view(*pair, "Reserves")
				Expect(err).To(Succeed())
				Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount))
				Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount))
			})

			DescribeTable("getInputPrice",
				func(swapAmount, token0Amount, token1Amount, expectedOutputAmount *amount.Amount) {

					uniMint(alice, tokenA, tokenB)
					uniApprove(aliceKey, tokenA, tokenB)
					_, err = tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)
					Expect(err).To(Succeed())

					tb.call(aliceKey, token0, "Transfer", *pair, swapAmount)

					_, err := tb.call(aliceKey, *pair, "Swap", AmountZero, expectedOutputAmount.Add(amount.NewAmount(0, 1)), alice, []byte(""), AddressZero)
					Expect(err).To(MatchError("Exchange: K"))

					_, err = tb.call(aliceKey, *pair, "Swap", AmountZero, expectedOutputAmount, alice, []byte(""), AddressZero)
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

					uniMint(alice, tokenA, tokenB)
					uniApprove(aliceKey, tokenA, tokenB)
					_, err = tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)
					Expect(err).To(Succeed())

					tb.call(aliceKey, token0, "Transfer", *pair, inputAmount)

					_, err := tb.call(aliceKey, *pair, "Swap", outputAmount.Add(amount.NewAmount(0, 1)), AmountZero, alice, []byte(""), AddressZero)
					Expect(err).To(MatchError("Exchange: K"))

					_, err = tb.call(aliceKey, *pair, "Swap", outputAmount, AmountZero, alice, []byte(""), AddressZero)
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

				uniMint(alice, &token0, &token1)
				uniApprove(aliceKey, &token0, &token1)
				tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)

				swapAmount := amount.NewAmount(1, 0)
				expectedOutputAmount := &amount.Amount{Int: big.NewInt(1662497915624478906)}

				// Transfer
				tb.call(aliceKey, token0, "Transfer", *pair, swapAmount)

				// Swap
				_, err = tb.call(aliceKey, *pair, "Swap", AmountZero, expectedOutputAmount, alice, []byte(""), AddressZero)
				Expect(err).To(Succeed())

				// Reserves
				is, err := tb.view(*pair, "Reserves")
				Expect(err).To(Succeed())
				Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount.Add(swapAmount)))
				Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount.Sub(expectedOutputAmount)))

				// BalanceOf
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(token0Amount.Add(swapAmount)))
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(token1Amount.Sub(expectedOutputAmount)))
				Expect(tokenBalanceOf(tb.ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Sub(swapAmount)))
				Expect(tokenBalanceOf(tb.ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Add(expectedOutputAmount)))
			})

			It("swap:token1", func() {
				token0Amount := amount.NewAmount(5, 0)
				token1Amount := amount.NewAmount(10, 0)

				uniMint(alice, &token0, &token1)
				uniApprove(aliceKey, &token0, &token1)
				tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)

				swapAmount := amount.NewAmount(1, 0)
				expectedOutputAmount := &amount.Amount{Int: big.NewInt(453305446940074565)}

				// Transfer
				_, err := tb.call(aliceKey, token1, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				// Swap
				_, err = tb.call(aliceKey, *pair, "Swap", expectedOutputAmount, AmountZero, alice, []byte(""), AddressZero)
				Expect(err).To(Succeed())

				// Reserves
				is, err := tb.view(*pair, "Reserves")
				Expect(err).To(Succeed())
				Expect(is[0].([]*amount.Amount)[0]).To(Equal(token0Amount.Sub(expectedOutputAmount)))
				Expect(is[0].([]*amount.Amount)[1]).To(Equal(token1Amount.Add(swapAmount)))

				// BalanceOf
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(token0Amount.Sub(expectedOutputAmount)))
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(token1Amount.Add(swapAmount)))
				Expect(tokenBalanceOf(tb.ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(token0Amount).Add(expectedOutputAmount)))
				Expect(tokenBalanceOf(tb.ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(token1Amount).Sub(swapAmount)))
			})

			It("burn", func() {
				token0Amount := amount.NewAmount(3, 0)
				token1Amount := amount.NewAmount(3, 0)

				uniMint(alice, &token0, &token1)
				uniApprove(aliceKey, &token0, &token1)
				tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)

				expectedLiquidity := amount.NewAmount(3, 0)

				// Transfer
				_, err := tb.call(aliceKey, *pair, "Transfer", *pair, expectedLiquidity.Sub(_ML))
				Expect(err).To(Succeed())

				// Burn
				is, err := tb.call(aliceKey, *pair, "Burn", alice)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(token0Amount.Sub(_ML)))
				Expect(is[1].(*amount.Amount)).To(Equal(token1Amount.Sub(_ML)))

				// *pair.BalanceOf(alice)
				Expect(err).To(Succeed())
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(AmountZero))

				// *pair.TotalSupply()
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(_ML))

				// BalanceOf
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(_ML))
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(_ML))
				Expect(tokenBalanceOf(tb.ctx, token0, alice)).To(Equal(_SupplyTokens[0].Sub(_ML)))
				Expect(tokenBalanceOf(tb.ctx, token1, alice)).To(Equal(_SupplyTokens[1].Sub(_ML)))
			})

			It("price{0,1}CumulativeLast", func() {
				uniMint(alice, &token0, &token1)

				token0Amount := amount.NewAmount(3, 0)
				token1Amount := amount.NewAmount(3, 0)

				tb.call(aliceKey, token0, "Transfer", *pair, token0Amount)
				tb.call(aliceKey, token1, "Transfer", *pair, token1Amount)

				// tx : *pair.Mint(alice)
				_, err := tb.call(aliceKey, *pair, "Mint", alice)
				Expect(err).To(Succeed())

				// Reserves 1
				is, _ := tb.view(*pair, "Reserves")
				blockTimestamp := is[1].(uint64)

				tb.step = 4000 // 4 sec

				// tx : *pair.sync()
				_, err = tb.call(aliceKey, *pair, "Sync")
				Expect(err).To(Succeed())

				// Reserves 2
				initialPrice0 := token0Amount.Div(token1Amount)
				initialPrice1 := token1Amount.Div(token0Amount)

				is, _ = tb.view(*pair, "Price0CumulativeLast")
				Expect(is[0].(*amount.Amount)).To(Equal(initialPrice0))
				is, _ = tb.view(*pair, "Price1CumulativeLast")
				Expect(is[0].(*amount.Amount)).To(Equal(initialPrice1))
				is, _ = tb.view(*pair, "Reserves")
				Expect(is[1]).To(Equal(blockTimestamp + 1))

				tb.step = 5000 // 5 sec

				// tx : token0.Tranfer(*pair,swapAmount)
				swapAmount := amount.NewAmount(3, 0)
				_, err = tb.call(aliceKey, token0, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				tb.step = 10000 // 10 sec

				//await *pair.swap(0, expandTo18Decimals(1), wallet.address, '0x', overrides) // make the price nice
				_, err = tb.call(aliceKey, *pair, "Swap", AmountZero, amount.NewAmount(1, 0), alice, []byte(""), AddressZero)
				Expect(err).To(Succeed())

				// tx : *pair.sync()
				_, err = tb.call(aliceKey, *pair, "Sync")
				Expect(err).To(Succeed())

				newPrice0 := amount.NewAmount(0, amount.FractionalMax/3)
				newPrice1 := amount.NewAmount(3, 0)

				// Reserves
				is, _ = tb.view(*pair, "Price0CumulativeLast")
				Expect(is[0].(*amount.Amount)).To(Equal(initialPrice0.MulC(int64(10)).Add(newPrice0.MulC(int64(10)))))
				is, _ = tb.view(*pair, "Price1CumulativeLast")
				Expect(is[0].(*amount.Amount)).To(Equal(initialPrice1.MulC(int64(10)).Add(newPrice1.MulC(int64(10)))))
				is, _ = tb.view(*pair, "Reserves")
				Expect(is[1]).To(Equal(blockTimestamp + 20))
			})

			It("feeTo:off", func() {

				delay := uint64(86400)
				tb.step = delay * 1000
				_, err = tb.call(aliceKey, *pair, "CommitNewFee", _Fee30, uint64(0), uint64(0), delay)
				Expect(err).To(Succeed())

				tb.step = _Step
				_, err = tb.call(aliceKey, *pair, "ApplyNewFee")
				Expect(err).To(Succeed())

				owner := tb.viewAddress(*pair, "Owner")
				Expect(owner).To(Equal(alice))

				is, err := tb.view(*pair, "AdminFee")
				Expect(is[0]).To(Equal(uint64(0)))

				token0Amount := amount.NewAmount(1000, 0)
				token1Amount := amount.NewAmount(1000, 0)

				uniMint(alice, &token0, &token1)
				uniApprove(aliceKey, &token0, &token1)
				tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)

				swapAmount := amount.NewAmount(1, 0)
				expectedOutputAmount := &amount.Amount{Int: big.NewInt(996006981039903216)}

				// Transfer
				_, err = tb.call(aliceKey, token1, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				// Swap
				_, err = tb.call(aliceKey, *pair, "Swap", expectedOutputAmount, AmountZero, alice, []byte(""), AddressZero)
				Expect(err).To(Succeed())

				expectedLiquidity := amount.NewAmount(1000, 0)

				// Transfer
				_, err = tb.call(aliceKey, *pair, "Transfer", *pair, expectedLiquidity.Sub(_ML))
				Expect(err).To(Succeed())

				// Burn
				_, err = tb.call(aliceKey, *pair, "Burn", alice)
				Expect(err).To(Succeed())

				// *pair.TotalSupply()
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(_ML))
			})

			It("feeTo:on", func() {
				setFees(tb, aliceKey, *pair, _Fee30, _AdminFee6, _WinnerFee)

				//owner
				owner := tb.viewAddress(*pair, "Owner")
				Expect(owner).To(Equal(alice))

				token0Amount := amount.NewAmount(1000, 0)
				token1Amount := amount.NewAmount(1000, 0)

				uniMint(alice, &token0, &token1)
				uniApprove(aliceKey, &token0, &token1)
				tb.call(aliceKey, *router, "UniAddLiquidity", token0, token1, token0Amount, token1Amount, AmountZero, AmountZero)

				swapAmount := amount.NewAmount(1, 0)
				expectedOutputAmount := &amount.Amount{Int: big.NewInt(996006981039903216)}

				// Transfer
				_, err = tb.call(aliceKey, token1, "Transfer", *pair, swapAmount)
				Expect(err).To(Succeed())

				// Swap
				_, err = tb.call(aliceKey, *pair, "Swap", expectedOutputAmount, AmountZero, alice, []byte(""), AddressZero)
				Expect(err).To(Succeed())

				expectedLiquidity := amount.NewAmount(1000, 0)

				// Transfer
				_, err = tb.call(aliceKey, *pair, "Transfer", *pair, expectedLiquidity.Sub(_ML))
				Expect(err).To(Succeed())

				// Burn
				_, err = tb.call(aliceKey, *pair, "Burn", alice)
				Expect(err).To(Succeed())

				// *pair.TotalSupply()
				// 1/n 에서 %로 변경되면서 값이 달라짐
				// Fee protocol = 6  :                    249750499252388 = 1000 + 249750499251388
				// AdminFee = 1666666667/FEEDENOMINATOR : 249750499302338
				//            1666666666/FEEDENOMINATOR : 249750499152487
				// 					    249750499152487 < 249750499252388 < 249750499302338

				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(_ML.Add(&amount.Amount{Int: big.NewInt(249750499301338)}))) // 1000 고려해야 함
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(&amount.Amount{Int: big.NewInt(249750499301338)}))
				Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(amount.NewAmount(0, 1000).Add(&amount.Amount{Int: big.NewInt(249501683747345)})))
				Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(amount.NewAmount(0, 1000).Add(&amount.Amount{Int: big.NewInt(250000187362969)})))
			})

			It("Skim", func() {
				uniAddLiquidity(aliceKey, &token0, &token1, _SupplyTokens[0], _SupplyTokens[1])
				uniMint(bob, &token0, &token1)

				var transferToken common.Address
				for k := 0; k < 2; k++ {
					switch k {
					case 0:
						transferToken = token0
					case 1:
						transferToken = token1
					}

					pairBalance0 := tokenBalanceOf(tb.ctx, token0, *pair)
					pairBalance1 := tokenBalanceOf(tb.ctx, token1, *pair)

					is, err := tb.view(*pair, "Reserves")
					Expect(err).To(Succeed())
					reserve0 := is[0].([]*amount.Amount)[0]
					reserve1 := is[0].([]*amount.Amount)[1]

					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1))

					_, err = tb.call(bobKey, transferToken, "Transfer", *pair, _TestAmount)
					Expect(err).To(Succeed())
					pairBalance0 = tokenBalanceOf(tb.ctx, token0, *pair)
					pairBalance1 = tokenBalanceOf(tb.ctx, token1, *pair)

					if k == 0 {
						Expect(pairBalance0).To(Equal(reserve0.Add(_TestAmount)))
						Expect(pairBalance1).To(Equal(reserve1))
					} else {
						Expect(pairBalance0).To(Equal(reserve0))
						Expect(pairBalance1).To(Equal(reserve1.Add(_TestAmount)))

					}

					_, err = tb.call(aliceKey, *pair, "Skim", charlie)
					Expect(err).To(Succeed())
					pairBalance0 = tokenBalanceOf(tb.ctx, token0, *pair)
					pairBalance1 = tokenBalanceOf(tb.ctx, token1, *pair)

					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1))

					// charlie
					charlieBalance0 := tokenBalanceOf(tb.ctx, token0, charlie)
					charlieBalance1 := tokenBalanceOf(tb.ctx, token1, charlie)

					if k == 0 {
						Expect(charlieBalance0).To(Equal(_TestAmount))
						Expect(charlieBalance1).To(Equal(AmountZero))
					} else {
						// charle는 두번 받음
						Expect(charlieBalance0).To(Equal(_TestAmount))
						Expect(charlieBalance1).To(Equal(_TestAmount))
					}
				}
			})

			It("Sync", func() {
				uniAddLiquidity(aliceKey, &token0, &token1, _SupplyTokens[0], _SupplyTokens[1])
				uniMint(bob, &token0, &token1)

				var transferToken common.Address

				for k := 0; k < 2; k++ {
					switch k {
					case 0:
						transferToken = token0
					case 1:
						transferToken = token1
					}

					pairBalance0 := tokenBalanceOf(tb.ctx, token0, *pair)
					pairBalance1 := tokenBalanceOf(tb.ctx, token1, *pair)

					is, err := tb.view(*pair, "Reserves")
					Expect(err).To(Succeed())
					reserve0 := is[0].([]*amount.Amount)[0]
					reserve1 := is[0].([]*amount.Amount)[1]

					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1))

					_, err = tb.call(bobKey, transferToken, "Transfer", *pair, _TestAmount)
					Expect(err).To(Succeed())
					pairBalance0 = tokenBalanceOf(tb.ctx, token0, *pair)
					pairBalance1 = tokenBalanceOf(tb.ctx, token1, *pair)

					if k == 0 {
						Expect(pairBalance0).To(Equal(reserve0.Add(_TestAmount)))
						Expect(pairBalance1).To(Equal(reserve1))
					} else {
						Expect(pairBalance0).To(Equal(reserve0))
						Expect(pairBalance1).To(Equal(reserve1.Add(_TestAmount)))

					}

					_, err = tb.call(aliceKey, *pair, "Sync")
					Expect(err).To(Succeed())
					is, err = tb.view(*pair, "Reserves")
					Expect(err).To(Succeed())
					reserve0 = is[0].([]*amount.Amount)[0]
					reserve1 = is[0].([]*amount.Amount)[1]

					Expect(pairBalance0).To(Equal(reserve0))
					Expect(pairBalance1).To(Equal(reserve1))
				}
			})
		})

	})

	Describe("router", func() {

		It("factory", func() {

			factory, err := factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())
			router, err := routerDeploy(tb, aliceKey, factory)
			Expect(err).To(Succeed())

			is, err := tb.view(*router, "Factory")
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

						beforeEach(fee, adminFee, winnerFee)
						uniMint(alice, tokenA, tokenB)

						tokenAAmount := amount.NewAmount(1, 0)
						tokenBAmount := amount.NewAmount(4, 0)

						expectedLiquidity := amount.NewAmount(2, 0)

						_, err = tb.call(aliceKey, *tokenA, "Approve", *router, MaxUint256)
						Expect(err).To(Succeed())
						_, err = tb.call(aliceKey, *tokenB, "Approve", *router, MaxUint256)
						Expect(err).To(Succeed())

						Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0]))
						Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1]))

						_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)
						Expect(err).To(Succeed())

						// BalanceOf
						Expect(tokenBalanceOf(tb.ctx, *tokenA, *pair)).To(Equal(tokenAAmount))
						Expect(tokenBalanceOf(tb.ctx, *tokenB, *pair)).To(Equal(tokenBAmount))
						Expect(tokenBalanceOf(tb.ctx, *pair, common.ZeroAddr)).To(Equal(_ML))
						Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(expectedLiquidity.Sub(_ML)))
					}
				}
			}
		})

		It("UniAddLiquidity : Price", func() {

			beforeEachDefault()
			uniMint(alice, tokenA, tokenB)

			tokenAAmount := amount.NewAmount(1, 0)
			tokenBAmount := amount.NewAmount(4, 0)

			tb.call(aliceKey, *tokenA, "Approve", *router, MaxUint256)
			tb.call(aliceKey, *tokenB, "Approve", *router, MaxUint256)

			_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)
			Expect(err).To(Succeed())

			is, err := tb.view(*pair, "Reserves")
			reserve0 := is[0].([]*amount.Amount)[0]
			reserve1 := is[0].([]*amount.Amount)[1]

			priceBefore := reserve0.Div(reserve1)

			_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenAAmount, AmountZero, AmountZero)
			Expect(err).To(Succeed())

			is, err = tb.view(*pair, "Reserves")
			reserve0 = is[0].([]*amount.Amount)[0]
			reserve1 = is[0].([]*amount.Amount)[1]

			priceAfter := reserve0.Div(reserve1)

			Expect(priceAfter).To(Equal(priceBefore))
		})

		It("AddLiquidity, UniGetLPTokenAmount : negative input", func() {
			Skip(AmountNotNegative)

			beforeEachDefault()

			_, err := tb.call(bobKey, *router, "UniGetLPTokenAmount", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), ToAmount(big.NewInt(1)))
			Expect(err).To(MatchError("Router: INSUFFICIENT_A_AMOUNT"))

			_, err = tb.call(bobKey, *router, "UniGetLPTokenAmount", *tokenA, *tokenB, ToAmount(big.NewInt(1)), ToAmount(big.NewInt(-1)))
			Expect(err).To(MatchError("Router: INSUFFICIENT_B_AMOUNT"))

			_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), ToAmount(big.NewInt(1)), AmountZero, AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_A_AMOUNT"))

			_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), ToAmount(big.NewInt(1)), AmountZero, AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_B_AMOUNT"))

		})

		It("AddLiquidity, UniGetLPTokenAmount", func() {
			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE} // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}     // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}    // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)
						uniMint(bob, tokenA, tokenB)
						uniApprove(bobKey, tokenA, tokenB)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b0, _ := ToBigInt(float64(rand.Uint64()))
							b1, _ := ToBigInt(float64(rand.Uint64()))
							token0Amount := ToAmount(b0)
							token1Amount := ToAmount(b1)

							is, err := tb.view(*router, "UniGetLPTokenAmount", *tokenA, *tokenB, token0Amount, token1Amount)
							Expect(err).To(Succeed())
							expectedLPToken := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = tb.call(bobKey, *router, "UniAddLiquidity", *tokenA, *tokenB, token0Amount, token1Amount, AmountZero, AmountZero)
							Expect(err).To(Succeed())
							amount0 := is[0].(*amount.Amount)
							amount1 := is[1].(*amount.Amount)
							liquidity := is[2].(*amount.Amount)
							Expect(liquidity).To(Equal(expectedLPToken))

							bobBalances[0] = bobBalances[0].Sub(amount0)
							bobBalances[1] = bobBalances[1].Sub(amount1)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(tb.ctx, *tokenA, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}
					}
				}
			}
		})
		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token out not match error", func() {
			beforeEachDefault()

			otherToken := common.BigToAddress(big.NewInt(1))
			_, err := tb.view(*router, "UniGetLPTokenAmountOneCoin", *tokenA, *tokenB, otherToken, ToAmount(big.NewInt(1)))
			Expect(err).To(MatchError("Router: INPUT_TOKEN_NOT_MATCH"))

			_, err = tb.call(bobKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, otherToken, ToAmount(big.NewInt(1)), AmountZero)
			Expect(err).To(MatchError("Router: INPUT_TOKEN_NOT_MATCH"))
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : negative input error", func() {
			Skip(AmountNotNegative)

			beforeEachDefault()

			_, err := tb.view(*router, "UniGetLPTokenAmountOneCoin", *tokenA, *tokenB, *tokenA, ToAmount(big.NewInt(-1)))
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = tb.view(*router, "UniGetLPTokenAmountOneCoin", *tokenA, *tokenB, *tokenB, ToAmount(big.NewInt(-1)))
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = tb.call(bobKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenA, ToAmount(big.NewInt(-1)), AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))

			_, err = tb.call(bobKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenB, ToAmount(big.NewInt(-1)), AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_AMOUNT"))
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : Initial Supply Error", func() {
			beforeEachDefault()
			uniMint(alice, tokenA, tokenB)
			uniApprove(aliceKey, tokenA, tokenB)

			_, err = tb.call(aliceKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenA, amount.NewAmount(1, 0), AmountZero)
			Expect(err).To(MatchError("Router: BOTH_RESERVE_0"))
		})

		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : tokenA", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)
						uniMint(bob, tokenA, tokenB)
						uniApprove(bobKey, tokenA, tokenB)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							tokenAmount := ToAmount(b)

							is, err := tb.view(*router, "UniGetLPTokenAmountOneCoin", *tokenA, *tokenB, *tokenA, tokenAmount)
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = tb.call(bobKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenA, tokenAmount, AmountZero)
							amountIn := is[0].(*amount.Amount)
							liquidity := is[1].(*amount.Amount)
							Expect(amountIn).To(Equal(tokenAmount))
							Expect(liquidity).To(Equal(expected))

							bobBalances[0] = bobBalances[0].Sub(amountIn)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(tb.ctx, *tokenA, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}
					}
				}
			}
		})
		It("UniAddLiquidityOneCoin, UniGetLPTokenAmountOneCoin : tokenB", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)
						uniMint(bob, tokenA, tokenB)
						uniApprove(bobKey, tokenA, tokenB)

						bobBalances := CloneAmountSlice(_SupplyTokens)
						bobLPBalance := amount.NewAmount(0, 0)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							tokenAmount := ToAmount(b)

							is, err := tb.view(*router, "UniGetLPTokenAmountOneCoin", *tokenA, *tokenB, *tokenB, tokenAmount)
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							ratio := is[1].(uint64)

							is, err = tb.call(bobKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenB, tokenAmount, AmountZero)
							amountIn := is[0].(*amount.Amount)
							liquidity := is[1].(*amount.Amount)
							Expect(amountIn).To(Equal(tokenAmount))
							Expect(liquidity).To(Equal(expected))

							bobBalances[1] = bobBalances[1].Sub(amountIn)
							bobLPBalance = bobLPBalance.Add(liquidity)

							Expect(tokenBalanceOf(tb.ctx, *tokenA, bob)).To(Equal(bobBalances[0]))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, bob)).To(Equal(bobBalances[1]))

							Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(bobLPBalance))

							lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
							Expect(MulDiv(liquidity.Int, big.NewInt(amount.FractionalMax), lpTotalSupply.Int).Uint64()).To(Equal(ratio))
						}
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

						beforeEach(fee, adminFee, winnerFee)

						tokenAAmount := amount.NewAmount(1, 0)
						tokenBAmount := amount.NewAmount(4, 0)

						uniMint(alice, tokenA, tokenB)
						uniApprove(aliceKey, tokenA, tokenB)
						tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)

						expectedLiquidity := amount.NewAmount(2, 0)

						tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)

						_, err = tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, expectedLiquidity.Sub(_ML), AmountZero, AmountZero)
						Expect(err).To(Succeed())

						Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(AmountZero))
						Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0].Sub(amount.NewAmount(0, 500))))
						Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1].Sub(amount.NewAmount(0, 2000))))
					}
				}
			}
		})

		It("RemoveLiquidity, UniGetWithdrawAmount : negative input error", func() {
			Skip("amount can't be negative")

			beforeEachDefault()

			_, err := tb.view(*router, "UniGetWithdrawAmount", *tokenA, *tokenB, ToAmount(big.NewInt(-1)))
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = tb.call(bobKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), AmountZero, AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))
		})

		It("RemoveLiquidity, UniGetWithdrawAmount", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)

						tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)

						lpBalance := tokenBalanceOf(tb.ctx, *pair, alice)
						lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := tb.view(*router, "UniGetWithdrawAmount", *tokenA, *tokenB, liquidity)
							expected0 := is[0].(*amount.Amount)
							expected1 := is[1].(*amount.Amount)

							is, err = tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity, AmountZero, AmountZero)
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							amount1Out := is[1].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected0))
							Expect(amount1Out).To(Equal(expected1))

							// balance
							balances[0] = balances[0].Add(amount0Out)
							balances[1] = balances[1].Add(amount1Out)
							Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(balances[0]))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(balances[1]))

							// lp
							lpBalance = lpBalance.Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Sub(liquidity)
							Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
						}
					}
				}
			}

		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : token out not match error", func() {
			beforeEachDefault()

			_, err := tb.view(*router, "UniGetWithdrawAmountOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(1)), alice)
			Expect(err).To(MatchError("Router: OUTPUT_TOKEN_NOT_MATCH"))

			_, err = tb.call(bobKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(1)), alice, AmountZero)
			Expect(err).To(MatchError("Router: OUTPUT_TOKEN_NOT_MATCH"))
		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : negative input error", func() {
			Skip("amount can't be negative")

			beforeEachDefault()

			_, err := tb.view(*router, "UniGetWithdrawAmountOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), *tokenA)
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = tb.view(*router, "UniGetWithdrawAmountOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), *tokenB)
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = tb.call(bobKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), *tokenA, AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))

			_, err = tb.call(bobKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, ToAmount(big.NewInt(-1)), *tokenB, AmountZero)
			Expect(err).To(MatchError("Router: INSUFFICIENT_LIQUIDITY"))
		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : tokenA", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)

						tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)

						lpBalance := tokenBalanceOf(tb.ctx, *pair, alice)
						lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := tb.view(*router, "UniGetWithdrawAmountOneCoin", *tokenA, *tokenB, liquidity, *tokenA)
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							mintFee := is[1].(*amount.Amount)

							is, err = tb.call(aliceKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, liquidity, *tokenA, AmountZero)
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected))

							// balance
							balances[0] = balances[0].Add(amount0Out)
							Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(balances[0]))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(balances[1]))

							// lp : alice(owner), totalSupply 모두에게 +
							lpBalance = lpBalance.Add(mintFee).Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Add(mintFee).Sub(liquidity)
							Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
						}
					}
				}
			}

		})

		It("UniRemoveLiquidityOneCoin, UniGetWithdrawAmountOneCoin : tokenB", func() {
			fees := []uint64{0, 30000000, 1000000000, 2500000000, trade.MAX_FEE} // 0, 30bp, 1%, 25%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}                  // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}                 // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)
						uniAddLiquidityDefault(aliceKey)

						tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)

						lpBalance := tokenBalanceOf(tb.ctx, *pair, alice)
						lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
						balances := MakeAmountSlice(2)

						for k := 0; k < _MaxIter; k++ {
							b, _ := ToBigInt(float64(rand.Uint64()))
							liquidity := ToAmount(b)

							is, err := tb.view(*router, "UniGetWithdrawAmountOneCoin", *tokenA, *tokenB, liquidity, *tokenB)
							Expect(err).To(Succeed())
							expected := is[0].(*amount.Amount)
							mintFee := is[1].(*amount.Amount)

							is, err = tb.call(aliceKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, liquidity, *tokenB, AmountZero)
							Expect(err).To(Succeed())
							amount0Out := is[0].(*amount.Amount)
							Expect(amount0Out).To(Equal(expected))

							// balance
							balances[1] = balances[1].Add(amount0Out)
							Expect(tokenBalanceOf(tb.ctx, *tokenA, alice).Cmp(balances[0].Int)).To(Equal(0))
							Expect(tokenBalanceOf(tb.ctx, *tokenB, alice).Cmp(balances[1].Int)).To(Equal(0))

							// lp : alice(owner), totalSupply 모두에게 +
							lpBalance = lpBalance.Add(mintFee).Sub(liquidity)
							lpTotalSupply = lpTotalSupply.Add(mintFee).Sub(liquidity)
							Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpBalance))
							Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
						}
					}
				}
			}
		})

		Describe("swapExactTokensForTokens, getAmountsOut", func() {

			tokenAAmount := amount.NewAmount(5, 0)
			tokenBAmount := amount.NewAmount(10, 0)
			swapAmount := amount.NewAmount(1, 0)
			expectedOutputAmount := &amount.Amount{Int: big.NewInt(1662497915624478906)}

			BeforeEach(func() {
				beforeEach(_Fee30, _AdminFee6, _WinnerFee)
				uniMint(alice, tokenA, tokenB)
				uniApprove(aliceKey, tokenA, tokenB)

				_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)
				Expect(err).To(Succeed())

				_, err = tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
				Expect(err).To(Succeed())
			})

			It("GetAmountsOut : negative input", func() {
				_, err := tb.view(*router, "GetAmountsOut", ToAmount(big.NewInt(-1)), []common.Address{*tokenA, *tokenB})
				Expect(err).To(MatchError("Router: INSUFFICIENT_IN_AMOUNT"))
			})

			It("SwapExactTokensForTokens : negative input", func() {
				_, err := tb.view(*router, "SwapExactTokensForTokens", ToAmount(big.NewInt(-1)), AmountZero, []common.Address{*tokenA, *tokenB})
				Expect(err).To(MatchError("Router: INSUFFICIENT_SWAP_AMOUNT"))
			})

			It("happy path, amounts", func() {
				is, err := tb.call(aliceKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenA, *tokenB})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(swapAmount))
				Expect(amounts[1]).To(Equal(expectedOutputAmount))

				// BalanceOf
				Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0].Sub(tokenAAmount).Sub(swapAmount)))
				Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1].Sub(tokenBAmount).Add(expectedOutputAmount)))
			})

			It("gas", func() {
				_, err := tb.call(aliceKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenA, *tokenB})

				Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0].Sub(tokenAAmount).Sub(swapAmount)))
				Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1].Sub(tokenBAmount).Add(expectedOutputAmount)))

				// gas
				balance := tokenBalanceOf(tb.ctx, mev, alice)
				Expect(err).To(Succeed())
				gas := amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance)
				log.Println("gas :", gas) // 100000000000000000 = 10^17 = 0.1 MEV
			})
		})

		Describe("swapTokensForExactTokens", func() {

			It("GetAmountsIn : negative input", func() {
				beforeEachDefault()

				_, err := tb.view(*router, "UniGetAmountsIn", ToAmount(big.NewInt(-1)), []common.Address{*tokenA, *tokenB})
				Expect(err).To(MatchError("Router: INSUFFICIENT_OUT_AMOUNT"))
			})

			It("UniSwapTokensForExactTokens : negative input", func() {
				Skip("amount can't be negative")

				beforeEachDefault()

				_, err := tb.call(aliceKey, *router, "UniSwapTokensForExactTokens", ToAmount(big.NewInt(-1)), AmountZero, []common.Address{*tokenA, *tokenB})
				Expect(err).To(MatchError("Router: INSUFFICIENT_SWAP_AMOUNT"))
			})

			It("happy path, amounts", func() {
				beforeEach(_Fee30, _AdminFee6, _WinnerFee)
				uniMint(alice, tokenA, tokenB)
				uniApprove(aliceKey, tokenA, tokenB)

				tokenAAmount := amount.NewAmount(5, 0)
				tokenBAmount := amount.NewAmount(10, 0)
				expectedSwapAmount := &amount.Amount{Int: big.NewInt(557227237267357629)}
				outputAmount := amount.NewAmount(1, 0)

				_, err = tb.call(aliceKey, *router, "UniAddLiquidity", *tokenA, *tokenB, tokenAAmount, tokenBAmount, AmountZero, AmountZero)
				Expect(err).To(Succeed())

				_, err = tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
				Expect(err).To(Succeed())

				is, err := tb.call(aliceKey, *router, "UniSwapTokensForExactTokens", outputAmount, MaxUint256, []common.Address{*tokenA, *tokenB})
				Expect(err).To(Succeed())
				amounts := is[0].([]*amount.Amount)
				Expect(amounts[0]).To(Equal(expectedSwapAmount))
				Expect(amounts[1]).To(Equal(outputAmount))

				Expect(tokenBalanceOf(tb.ctx, *tokenA, alice)).To(Equal(_SupplyTokens[0].Sub(tokenAAmount).Sub(expectedSwapAmount)))
				Expect(tokenBalanceOf(tb.ctx, *tokenB, alice)).To(Equal(_SupplyTokens[1].Sub(tokenBAmount).Add(outputAmount)))

			})
		})
	})

	Describe("WithdrawAminFee", func() {

		Describe("admin = LP", func() {

			BeforeEach(func() {
				beforeEach(30000000, 5000000000, trade.MAX_FEE)
				uniAddLiquidityDefault(aliceKey)
				uniMint(bob, tokenA, tokenB)

				_, err = tb.call(bobKey, *tokenA, "Approve", *router, MaxUint256)
				Expect(err).To(Succeed())
				_, err = tb.call(bobKey, *tokenB, "Approve", *router, MaxUint256)
				Expect(err).To(Succeed())

				swapAmount := amount.NewAmount(1, 0)
				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenA, *tokenB})
				Expect(err).To(Succeed())
			})

			It("MintedAdminBalance, AdminBalance : *pair", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // It(payToken : AddressZero) 참조

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))

				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)

				liquidity := lpOwnerBalance.Sub(expectedMintAmount) //최대
				tb.call(aliceKey, *pair, "Transfer", *pair, liquidity)
				tb.call(aliceKey, *pair, "Burn", alice)

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))

				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity).Add(expectedMintAmount)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity).Add(expectedMintAmount)))

				tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(AmountZero))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity)))
			})

			It("MintedAdminBalance, AdminBalance : Router", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // It(payToken : AddressZero) 참조

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)

				liquidity := lpOwnerBalance.Sub(expectedMintAmount) //최대

				tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
				tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity, AmountZero, AmountZero)

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity).Add(expectedMintAmount)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity).Add(expectedMintAmount)))

				tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(AmountZero))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity)))
			})

			It("admin UniRemoveLiquidity more than LPBalance : *pair", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				liquidity1 := amount.NewAmount(1, 0)
				_, err := tb.call(aliceKey, *pair, "Transfer", *pair, liquidity1)
				Expect(err).To(Succeed())

				_, err = tb.call(aliceKey, *pair, "Burn", alice)
				Expect(err).To(Succeed())

				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				liquidity2 := lpOwnerBalance.Sub(expectedMintAmount).Add(amount.NewAmount(0, 1)) // 최대 + 1

				_, err = tb.call(aliceKey, *pair, "Transfer", *pair, liquidity2)
				Expect(err).To(Succeed())
				_, err = tb.call(aliceKey, *pair, "Burn", alice)
				Expect(err).To(MatchError("Exchange: OWNER_LIQUIDITY"))
			})

			It("admin UniRemoveLiquidity more than LPBalance : Router", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				liquidity1 := amount.NewAmount(1, 0)

				tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
				_, err := tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity1, AmountZero, AmountZero)
				Expect(err).To(Succeed())

				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				liquidity2 := lpOwnerBalance.Sub(expectedMintAmount).Add(amount.NewAmount(0, 1)) // 최대 + 1
				_, err = tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity2, AmountZero, AmountZero)
				Expect(err).To(MatchError("Router: OWNER_LIQUIDITY"))
			})

			It("owner_change_before_mint_admin_fee", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)

				delay := uint64(3 * 86400)
				tb.call(aliceKey, *pair, "CommitTransferOwnerWinner", bob, charlie, delay)
				tb.step = delay * 1000
				tb.call(aliceKey, *pair, "ApplyTransferOwnerWinner")

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))

				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
			})

			It("owner_change_after_mint_admin_fee", func() {
				expectedMintAmount := amount.NewAmount(0, 1060658053645688) // 위 참조

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))
				lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, alice)
				lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)

				liquidity1 := amount.NewAmount(1, 0)
				tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
				_, err := tb.call(aliceKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity1, AmountZero, AmountZero)
				Expect(err).To(Succeed())

				delay := uint64(3 * 86400)
				_, err = tb.call(aliceKey, *pair, "CommitTransferOwnerWinner", bob, charlie, delay)
				tb.step = delay * 1000
				_, err = tb.call(aliceKey, *pair, "ApplyTransferOwnerWinner")

				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(expectedMintAmount))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(expectedMintAmount))

				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity1)))
				Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(expectedMintAmount))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity1).Add(expectedMintAmount)))

				_, err = tb.call(bobKey, *pair, "WithdrawAdminFees2")
				Expect(tb.viewAmount(*pair, "MintedAdminBalance")).To(Equal(AmountZero))
				Expect(tb.viewAmount(*pair, "AdminBalance")).To(Equal(AmountZero))
				Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply.Sub(liquidity1)))
				Expect(tokenBalanceOf(tb.ctx, *pair, alice)).To(Equal(lpOwnerBalance.Sub(liquidity1)))
				Expect(tokenBalanceOf(tb.ctx, *pair, bob)).To(Equal(AmountZero))
			})
		})

		It("payToken : AddressZero", func() {

			// token을 꼭 sorting 해서 token0, token1 으로 실행한다.

			fees := []uint64{0, 30000000, 100000000, trade.MAX_FEE} // 0, 30bp, 10%, 50%
			adminFees := []uint64{0, 5000000000, trade.MAX_FEE}     // 0, 50%, 100%
			winnerFees := []uint64{0, 5000000000, trade.MAX_FEE}    // 0, 50%, 100%

			for _, fee := range fees {
				for _, adminFee := range adminFees {
					for _, winnerFee := range winnerFees {

						beforeEach(fee, adminFee, winnerFee)

						// token0, token1 sorting
						token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
						Expect(err).To(Succeed())

						uniAddLiquidityDefault(charlieKey)
						uniMint(bob, tokenA, tokenB)
						uniApprove(bobKey, tokenA, tokenB)

						pairBalance0 := tokenBalanceOf(tb.ctx, token0, *pair)
						pairBalance1 := tokenBalanceOf(tb.ctx, token1, *pair)
						K := Sqrt(Mul(pairBalance0.Int, pairBalance1.Int)) // totalSupply

						// lpOwnerBalance = lpTotalSupply - MinimumLiqudity
						lpOwnerBalance := tokenBalanceOf(tb.ctx, *pair, charlie) // charlie = lpOwner 1명
						lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
						Expect(lpOwnerBalance).To(Equal(lpTotalSupply.Sub(_ML)))
						Expect(K).To(Equal(lpTotalSupply.Int))

						swapAmount := amount.NewAmount(1, 0)
						expectedOutputAmount, err := trade.UniGetAmountOut(fee, swapAmount.Int, pairBalance0.Int, pairBalance1.Int) // 1993996023971928199, 1991996031943904367

						is, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
						Expect(err).To(Succeed())
						amounts := is[0].([]*amount.Amount)
						Expect(amounts[0]).To(Equal(swapAmount))
						Expect(amounts[1].Int).To(Equal(expectedOutputAmount))

						// Fee =  10^18*(30bp,40bp) = (300000000000000,400000000000000)
						// *pairBalance0 += swapAmount
						// *pairBalance1 -= expectedOutputAmount
						pairBalance0AfterSwap := pairBalance0.Add(swapAmount)
						pairBalance1AfterSwap := pairBalance1.Sub(amounts[1])
						Expect(tokenBalanceOf(tb.ctx, token0, *pair)).To(Equal(pairBalance0AfterSwap))
						Expect(tokenBalanceOf(tb.ctx, token1, *pair)).To(Equal(pairBalance1AfterSwap))
						Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))

						// totalSupply change 를 K값 변화로 계산 - 아직 totalSupply에 반영되지 않음
						cK := Sqrt(Mul(pairBalance0AfterSwap.Int, pairBalance1AfterSwap.Int)) // changedK : cK - K = 2121316110473344, 2828421484873749

						is, err = tb.call(aliceKey, *pair, "AdminBalance")
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

						is, err = tb.call(aliceKey, *pair, "WithdrawAdminFees2")
						Expect(err).To(Succeed())
						Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
						Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
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
					}
				}
			}
		})

		It("payToken : token0", func() {
			_fee := uint64(30000000)
			_adminFee := uint64(10000000000)
			_winnerFee := uint64(0)

			lp := eve
			deployContracts()

			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
			Expect(err).To(Succeed())

			payToken := token0

			// pair create
			pC := &pairContractConstruction{
				TokenA:    *tokenA,
				TokenB:    *tokenB,
				PayToken:  payToken,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       _fee,
				AdminFee:  _adminFee,
				WinnerFee: _winnerFee,
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   _GroupId,
			}

			pair, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(Succeed())

			uniAddLiquidityDefault(eveKey)
			uniMint(bob, tokenA, tokenB)
			uniApprove(bobKey, tokenA, tokenB)

			lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
			swapAmount := amount.NewAmount(1, 0)
			for k := 0; k < 10; k++ { // 각각 10회
				_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
				Expect(err).To(Succeed())

				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
				Expect(err).To(Succeed())
			}

			pairBalace0 := tokenBalanceOf(tb.ctx, token0, *pair)
			pairBalace1 := tokenBalanceOf(tb.ctx, token1, *pair)
			Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
			lpAmount := tb.viewAmount(*pair, "AdminBalance")

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

				is, err := tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(ToAmount(expectedOwnerFee0)))
				Expect(ownerFee1).To(Equal(AmountZero))
				Expect(winnerFee0).To(Equal(ToAmount(expectedWinnerFee0)))
				Expect(winnerFee1).To(Equal(AmountZero))
			} else {
				swappedAmount, err := trade.UniGetAmountOut(_fee, adminFee0.Int, pairBalace0AfterBurn.Int, pairBalace1AfterBurn.Int)
				Expect(err).To(Succeed())
				adminFee0.Set(Zero)
				adminFee1.Set(Add(adminFee1.Int, swappedAmount))

				expectedWinnerFee1 := MulDivCC(adminFee1.Int, int64(_winnerFee), trade.FEE_DENOMINATOR)
				expectedOwnerFee1 := Sub(adminFee1.Int, expectedWinnerFee1)

				is, err := tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(AmountZero))
				Expect(ownerFee1).To(Equal(ToAmount(expectedOwnerFee1)))
				Expect(winnerFee0).To(Equal(AmountZero))
				Expect(winnerFee1).To(Equal(ToAmount(expectedWinnerFee1)))
			}

			// 22.05.09 에러 대응 추가 start
			balance0 := tokenBalanceOf(tb.ctx, token0, *pair)
			balance1 := tokenBalanceOf(tb.ctx, token1, *pair)
			is, _ := tb.view(*pair, "Reserves")
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

			for k := 0; k < 10; k++ { // 각각 10회
				_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
				Expect(err).To(Succeed())

				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
				Expect(err).To(Succeed())
			}

			is, err = tb.call(aliceKey, *pair, "WithdrawAdminFees2")

			balance0 = tokenBalanceOf(tb.ctx, token0, *pair)
			balance1 = tokenBalanceOf(tb.ctx, token1, *pair)
			is, _ = tb.view(*pair, "Reserves")
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

			_, err = tb.call(aliceKey, *pair, "Skim", alice)
			Expect(err).To(Succeed())

			// 22.05.09 에러 대응 추가 end
			//}
		})

		It("payToken : token1", func() {
			_fee := uint64(30000000)
			_adminFee := uint64(10000000000)
			_winnerFee := uint64(0)

			lp := eve
			deployContracts()

			// token0, token1 sorting
			token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
			Expect(err).To(Succeed())

			payToken := token1

			// pair create
			pC := &pairContractConstruction{
				TokenA:    *tokenA,
				TokenB:    *tokenB,
				PayToken:  payToken,
				Name:      "__UNI_NAME",
				Symbol:    "__UNI_SYMBOL",
				Owner:     alice,
				Winner:    charlie,
				Fee:       _fee,
				AdminFee:  _adminFee,
				WinnerFee: _winnerFee,
				Factory:   *factory,
				WhiteList: *whiteList,
				GroupId:   _GroupId,
			}

			pair, err = pairCreate(tb, aliceKey, pC)
			Expect(err).To(Succeed())

			uniAddLiquidityDefault(eveKey)
			uniMint(bob, tokenA, tokenB)
			uniApprove(bobKey, tokenA, tokenB)

			lpTotalSupply := tokenTotalSupply(tb.ctx, *pair)
			swapAmount := amount.NewAmount(1, 0)
			for k := 0; k < 10; k++ { // 각각 10회
				_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
				Expect(err).To(Succeed())

				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
				Expect(err).To(Succeed())
			}

			pairBalace0 := tokenBalanceOf(tb.ctx, token0, *pair)
			pairBalace1 := tokenBalanceOf(tb.ctx, token1, *pair)
			Expect(tokenTotalSupply(tb.ctx, *pair)).To(Equal(lpTotalSupply))
			lpAmount := tb.viewAmount(*pair, "AdminBalance")

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

				is, err := tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(ToAmount(expectedOwnerFee0)))
				Expect(ownerFee1).To(Equal(AmountZero))
				Expect(winnerFee0).To(Equal(ToAmount(expectedWinnerFee0)))
				Expect(winnerFee1).To(Equal(AmountZero))
			} else {
				swappedAmount, err := trade.UniGetAmountOut(_fee, adminFee0.Int, pairBalace0AfterBurn.Int, pairBalace1AfterBurn.Int)
				Expect(err).To(Succeed())
				adminFee0.Set(Zero)
				adminFee1.Set(Add(adminFee1.Int, swappedAmount))

				expectedWinnerFee1 := MulDivCC(adminFee1.Int, int64(_winnerFee), trade.FEE_DENOMINATOR)
				expectedOwnerFee1 := Sub(adminFee1.Int, expectedWinnerFee1)

				is, err := tb.call(aliceKey, *pair, "WithdrawAdminFees2")
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(lpAmount))
				ownerFee0 := is[1].(*amount.Amount)
				ownerFee1 := is[2].(*amount.Amount)
				winnerFee0 := is[3].(*amount.Amount)
				winnerFee1 := is[4].(*amount.Amount)

				Expect(ownerFee0).To(Equal(AmountZero))
				Expect(ownerFee1).To(Equal(ToAmount(expectedOwnerFee1)))
				Expect(winnerFee0).To(Equal(AmountZero))
				Expect(winnerFee1).To(Equal(ToAmount(expectedWinnerFee1)))
			}

			// 22.05.09 에러 대응 추가 start
			balance0 := tokenBalanceOf(tb.ctx, token0, *pair)
			balance1 := tokenBalanceOf(tb.ctx, token1, *pair)
			is, _ := tb.view(*pair, "Reserves")
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

			for k := 0; k < 10; k++ { // 각각 10회
				_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
				Expect(err).To(Succeed())

				_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
				Expect(err).To(Succeed())
			}

			is, err = tb.call(aliceKey, *pair, "WithdrawAdminFees2")

			balance0 = tokenBalanceOf(tb.ctx, token0, *pair)
			balance1 = tokenBalanceOf(tb.ctx, token1, *pair)
			is, _ = tb.view(*pair, "Reserves")
			Expect(is[0].([]*amount.Amount)[0]).To(Equal(balance0))
			Expect(is[0].([]*amount.Amount)[1]).To(Equal(balance1))

			_, err = tb.call(aliceKey, *pair, "Skim", alice)
			Expect(err).To(Succeed())

			// 22.05.09 에러 대응 추가 end
		})
	})

	Describe("Uni Integration", func() {

		It("random", func() {
			// eve : whitelist

			beforeEachDefault()

			uniAddLiquidityDefault(aliceKey)

			uniMint(alice, tokenA, tokenB)
			uniApprove(aliceKey, tokenA, tokenB)
			uniMint(bob, tokenA, tokenB)
			uniApprove(bobKey, tokenA, tokenB)
			uniMint(charlie, tokenA, tokenB)
			uniApprove(charlieKey, tokenA, tokenB)
			uniMint(eve, tokenA, tokenB)
			uniApprove(eveKey, tokenA, tokenB)

			tb.call(aliceKey, *pair, "Approve", *router, MaxUint256)
			tb.call(bobKey, *pair, "Approve", *router, MaxUint256)
			tb.call(charlieKey, *pair, "Approve", *router, MaxUint256)
			tb.call(eveKey, *pair, "Approve", *router, MaxUint256)

			for k := 0; k < 1000; k++ {
				senderKey := userKeys[rand.Intn(4)]
				who := senderKey.PublicKey().Address()

				switch rand.Intn(7) {
				case 0:
					token0Amount, _ := ToBigInt(float64(rand.Uint64()))
					token1Amount, _ := ToBigInt(float64(rand.Uint64()))
					balance0 := tokenBalanceOf(tb.ctx, *tokenA, who)
					balance1 := tokenBalanceOf(tb.ctx, *tokenB, who)
					if balance0.Cmp(token0Amount) > 0 && balance1.Cmp(token1Amount) > 0 {
						_, err = tb.call(senderKey, *router, "UniAddLiquidity", *tokenA, *tokenB, token0Amount, token1Amount, AmountZero, AmountZero)
						Expect(err).To(Succeed())
					}
				case 1:
					token0Amount, _ := ToBigInt(float64(rand.Uint64()))
					token1Amount, _ := ToBigInt(float64(rand.Uint64()))
					balance0 := tokenBalanceOf(tb.ctx, *tokenA, who)
					balance1 := tokenBalanceOf(tb.ctx, *tokenB, who)
					switch rand.Intn(2) {
					case 0:
						if balance0.Cmp(token0Amount) > 0 {
							_, err = tb.call(senderKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenA, token0Amount, AmountZero)
							Expect(err).To(Succeed())
						}
					case 1:
						if balance1.Cmp(token1Amount) > 0 {
							_, err = tb.call(senderKey, *router, "UniAddLiquidityOneCoin", *tokenA, *tokenB, *tokenB, token1Amount, AmountZero)
							Expect(err).To(Succeed())
						}
					}
				case 2:
					liquidity, _ := ToBigInt(float64(rand.Uint64()))
					minted := amount.NewAmount(0, 0)
					if who == alice { // owner
						minted = tb.viewAmount(*pair, "MintedAdminBalance")
					}
					balance := tokenBalanceOf(tb.ctx, *pair, who)
					if balance.Sub(minted).Cmp(liquidity) > 0 {
						_, err = tb.call(senderKey, *router, "UniRemoveLiquidity", *tokenA, *tokenB, liquidity, AmountZero, AmountZero)
						Expect(err).To(Succeed())
					}
				case 3:
					liquidity, _ := ToBigInt(float64(rand.Uint64()))
					minted := amount.NewAmount(0, 0)
					if who == alice { // owner
						minted = tb.viewAmount(*pair, "MintedAdminBalance")
					}
					balance := tokenBalanceOf(tb.ctx, *pair, who)
					if balance.Sub(minted).Cmp(liquidity) > 0 {
						switch rand.Intn(2) {
						case 0:
							_, err = tb.call(senderKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, liquidity, *tokenA, AmountZero)
							Expect(err).To(Succeed())
						case 1:
							_, err = tb.call(senderKey, *router, "UniRemoveLiquidityOneCoin", *tokenA, *tokenB, liquidity, *tokenB, AmountZero)
							Expect(err).To(Succeed())
						}
					}
				case 4:
					swapAmount, _ := ToBigInt(float64(rand.Uint64()))
					switch rand.Intn(2) {
					case 0:
						_, err := tb.call(senderKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenA, *tokenB})
						Expect(err).To(Succeed())
					case 1:
						_, err := tb.call(senderKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{*tokenB, *tokenA})
						Expect(err).To(Succeed())
					}
				case 5:
					outputAmount, _ := ToBigInt(float64(rand.Uint64()))
					switch rand.Intn(2) {
					case 0:
						_, err := tb.call(senderKey, *router, "UniSwapTokensForExactTokens", outputAmount, MaxUint256, []common.Address{*tokenA, *tokenB})
						Expect(err).To(Succeed())
					case 1:
						_, err := tb.call(senderKey, *router, "UniSwapTokensForExactTokens", outputAmount, MaxUint256, []common.Address{*tokenB, *tokenA})
						Expect(err).To(Succeed())
					}
				case 6:
					// owner = alice
					_, err = tb.call(aliceKey, *pair, "WithdrawAdminFees2")
					Expect(err).To(Succeed())
				default:
				}
			}
		})
	})

	Describe("WhiteList", func() {

		Describe("FeeWhiteList, FeeAddress, Swap", func() {
			BeforeEach(func() {
				beforeEach(_Fee30, _AdminFee6, _WinnerFee)
			})

			It("FeeWhiteList, FeeAddress : not WhiteList", func() {

				is, err := tb.view(*pair, "FeeWhiteList", alice)
				Expect(err).To(Succeed())
				Expect(is[0].([]byte)).To(Equal([]byte{}))

				is, err = tb.view(*pair, "FeeAddress", alice)
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(_Fee30))

			})

			It("FeeWhiteList, FeeAddress : WhiteList", func() {

				fee := uint64(0) // 0%

				is, err := tb.view(*pair, "FeeWhiteList", eve)
				Expect(err).To(Succeed())
				Expect(is[0].([]byte)).To(Equal(bin.Uint64Bytes(fee)))

				is, err = tb.view(*pair, "FeeAddress", eve)
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(fee))

			})

			Describe("Swap", func() {

				var token0, token1 common.Address

				swapAmount := amount.NewAmount(1, 0)
				token0Amount := amount.NewAmount(1000, 0)
				token1Amount := amount.NewAmount(1000, 0)
				expectedOutput := amount.NewAmount(0, 996006981039903216)   // 0.3%
				wlExpectedOutput := amount.NewAmount(0, 999000999000999000) // 0.0%

				BeforeEach(func() {
					// token0, token1 sorting
					token0, token1, _ = trade.SortTokens(*tokenA, *tokenB)

					uniAddLiquidity(aliceKey, &token0, &token1, token0Amount, token1Amount)
				})

				It("not WhiteList fee = pair.fee(cc)", func() {

					tb.call(aliceKey, token0, "Transfer", *pair, swapAmount)

					_, err := tb.call(aliceKey, *pair, "Swap", AmountZero, expectedOutput.Add(amount.NewAmount(0, 1)), alice, []byte(""), AddressZero)
					Expect(err).To(MatchError("Exchange: K"))

					_, err = tb.call(aliceKey, *pair, "Swap", AmountZero, expectedOutput, alice, []byte(""), AddressZero)
					Expect(err).To(Succeed())
				})

				It("WhiteList from = AddressZero fee = feeWhiteList ", func() {

					uniMint(eve, &token0, &token1)

					tb.call(eveKey, token0, "Transfer", *pair, swapAmount)

					_, err := tb.call(eveKey, *pair, "Swap", AmountZero, wlExpectedOutput.Add(amount.NewAmount(0, 1)), eve, []byte(""), AddressZero)
					Expect(err).To(MatchError("Exchange: K"))

					_, err = tb.call(eveKey, *pair, "Swap", AmountZero, wlExpectedOutput, eve, []byte(""), AddressZero)
					Expect(err).To(Succeed())
				})

				It("WhiteList from = whitelist fee = feeWhiteList", func() {
					Skip("eve isn't a whitelist any more")

					uniMint(eve, &token0, &token1)

					tb.call(eveKey, token0, "Transfer", *pair, swapAmount)

					_, err := tb.call(eveKey, *pair, "Swap", AmountZero, wlExpectedOutput.Add(amount.NewAmount(0, 1)), eve, []byte(""), eve)
					Expect(err).To(MatchError("Exchange: K"))

					_, err = tb.call(eveKey, *pair, "Swap", AmountZero, wlExpectedOutput, eve, []byte(""), eve)
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

				deployContracts()

				// token0, token1 sorting
				token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
				Expect(err).To(Succeed())

				// pair create
				pC := &pairContractConstruction{
					TokenA:    token0,
					TokenB:    token1,
					PayToken:  token0,
					Name:      "__UNI_NAME",
					Symbol:    "__UNI_SYMBOL",
					Owner:     alice,
					Winner:    charlie,
					Fee:       _Fee30,
					AdminFee:  _AdminFee6,
					WinnerFee: _WinnerFee,
					Factory:   *factory,
					WhiteList: *whiteList,
					GroupId:   _GroupId,
				}

				pair, err = pairCreate(tb, aliceKey, pC)
				Expect(err).To(Succeed())

				uniAddLiquidity(charlieKey, &token0, &token1, _SupplyTokens[0], _SupplyTokens[1])

				uniMint(bob, &token0, &token1)
				uniApprove(bobKey, &token0, &token1)

				swapAmount := amount.NewAmount(1, 0)
				for k := 0; k < 10; k++ {
					_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
					Expect(err).To(Succeed())
					_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
					Expect(err).To(Succeed())
				}

				lpAmount := tb.viewAmount(*pair, "AdminBalance")
				is, err := tb.call(aliceKey, *pair, "WithdrawAdminFees2")
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

			})

			It("WhiteList", func() {

				deployContracts()

				// token0, token1 sorting
				token0, token1, err := trade.SortTokens(*tokenA, *tokenB)
				Expect(err).To(Succeed())

				// pair create
				pC := &pairContractConstruction{
					TokenA:    token0,
					TokenB:    token1,
					PayToken:  token0,
					Name:      "__UNI_NAME",
					Symbol:    "__UNI_SYMBOL",
					Owner:     eve,
					Winner:    charlie,
					Fee:       _Fee30,
					AdminFee:  _AdminFee6,
					WinnerFee: _WinnerFee,
					Factory:   *factory,
					WhiteList: *whiteList,
					GroupId:   _GroupId,
				}

				pair, err = pairCreate(tb, aliceKey, pC)
				Expect(err).To(Succeed())

				uniAddLiquidity(charlieKey, &token0, &token1, _SupplyTokens[0], _SupplyTokens[1])

				uniMint(bob, &token0, &token1)
				uniApprove(bobKey, &token0, &token1)

				swapAmount := amount.NewAmount(1, 0)
				for k := 0; k < 10; k++ {
					_, err := tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token0, token1})
					Expect(err).To(Succeed())
					_, err = tb.call(bobKey, *router, "SwapExactTokensForTokens", swapAmount, AmountZero, []common.Address{token1, token0})
					Expect(err).To(Succeed())
				}

				lpAmount := tb.viewAmount(*pair, "AdminBalance")
				// whitelist
				is, err := tb.call(eveKey, *pair, "WithdrawAdminFees2")
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

			})
		})

	})
})
