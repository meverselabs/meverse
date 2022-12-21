package test

import (
	"log"
	"math"
	"math/big"
	"math/rand"
	"strconv"
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

var _ = Describe("stableswap Test", func() {
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

	var factory, whiteList, swap *common.Address
	var stableTokens []common.Address
	decimals := []int{18, 6, 6}
	N := uint8(len(decimals))

	// default 값
	_Fee := _Fee40
	_AdminFee := uint64(trade.MAX_ADMIN_FEE)
	_WinnerFee := _Fee5000
	_GroupId := hash.BigToHash(big.NewInt(100))

	_SwapName := "__STABLE_NAME"
	_SwapSymbol := "__STABLE_SYMBOL"
	_Amp := int64(360 * 2)

	//1e6 of each coin - used to make an even initial deposit in many test setups
	_InitialAmounts := MakeAmountSlice(N)
	for k := uint8(0); k < N; k++ {
		_InitialAmounts[k].Set(Mul(Exp(big.NewInt(10), big.NewInt(int64(decimals[k]))), big.NewInt(_BaseAmount)))
	}

	// _load_pool_data in curve-contract/brownie_hooks.py
	_PrecisionMul := make([]uint64, N, N)
	_Rates := make([]*big.Int, N, N)
	for k := uint8(0); k < N; k++ {
		_PrecisionMul[k] = uint64(trade.PRECISION / int64(math.Pow10(decimals[k])))
		_Rates[k] = MulC(big.NewInt(int64(_PrecisionMul[k])), trade.PRECISION)
	}

	deployContracts := func() {
		// erc20Token deploy
		erc20Token, err := erc20TokenDeploy(tb, aliceKey, amount.NewAmount(0, 0))
		if err != nil {
			panic(err)
		}

		// 2 tokens deploy
		token1, err := tokenDeploy(tb, aliceKey, "Token1", "TKN1")
		if err != nil {
			panic(err)
		}
		token2, err := tokenDeploy(tb, aliceKey, "Token2", "TKN2")
		if err != nil {
			panic(err)
		}

		stableTokens = []common.Address{*erc20Token, *token1, *token2}

		// setMinter
		for _, token := range stableTokens {
			_, err = tb.call(aliceKey, token, "SetMinter", alice, true)
			if err != nil {
				panic(err)
			}
		}

		whiteList, err = whiteListDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}
	}

	// deployPairContracts deploy factory and router
	deployFactoryAndRouter := func() {

		factory, err = factoryDeploy(tb, aliceKey)
		if err != nil {
			panic(err)
		}
	}

	// beforeEachDefault create contracts
	// 1. factory, router, whitelist Deploy
	// 2. Token Contract Deploy
	// 3. swap Contract Creation
	beforeEach := func(fee, adminFee, winnerFee uint64, plus bool) {

		if plus {
			deployFactoryAndRouter()
		}

		deployContracts()

		// swap Deploy
		sD := &trade.StableSwapConstruction{
			Name:         "__STABLE_NAME",
			Symbol:       "__STABLE_SYMBOL",
			NTokens:      uint8(N),
			Tokens:       stableTokens,
			PayToken:     common.Address{},
			Owner:        alice,
			Winner:       charlie,
			Fee:          fee,
			AdminFee:     adminFee,
			WinnerFee:    winnerFee,
			WhiteList:    *whiteList,
			GroupId:      _GroupId,
			Amp:          big.NewInt(_Amp),
			PrecisionMul: _PrecisionMul,
			Rates:        _Rates,
		}

		if plus {
			sD.Factory = *factory
		} else {
			sD.Factory = common.Address{}
		}

		swap, err = swapDeploy(tb, aliceKey, sD)
		Expect(err).To(Succeed())
	}

	// beforeEachDefault create contracts with default fee parameters
	beforeEachDefault := func() {
		beforeEach(_Fee, _AdminFee, _WinnerFee, false)
	}

	// stableMint mint tokens to to Address by token owner
	stableMint := func(to common.Address) {
		for i := uint8(0); i < N; i++ {
			_, err := tb.call(aliceKey, stableTokens[i], "Mint", to, _InitialAmounts[i])
			Expect(err).To(Succeed())
		}
	}

	// stableApprove approve tokens to router
	stableApprove := func(ownerKey key.Key) {
		for i := uint8(0); i < N; i++ {
			_, err := tb.call(ownerKey, stableTokens[i], "Approve", *swap, MaxUint256)
			Expect(err).To(Succeed())
		}
	}

	// stableAddLiquidity prepare initial swap's AddLiquidity
	stableAddLiquidity := func(ownerKey key.Key, amts []*amount.Amount) {
		owner := ownerKey.PublicKey().Address()
		stableMint(owner)
		stableApprove(ownerKey)

		_, err = tb.call(ownerKey, *swap, "AddLiquidity", amts, amount.NewAmount(0, 0))
		Expect(err).To(Succeed())
	}

	// addInitialLiquidityDefault prepare initial default uniAddLiquidity
	stableAddLiquidityDefault := func(ownerKey key.Key) {
		stableAddLiquidity(ownerKey, _InitialAmounts)
	}

	getFee := func(b *big.Int, fee uint64) *big.Int {
		return MulDivCC(b, int64(fee), trade.FEE_DENOMINATOR)
	}

	stableGetAdminBalances := func() ([]*amount.Amount, error) {
		is, err := tb.view(*swap, "Reserves")
		if err != nil {
			return nil, err
		}
		swap_reserves := is[0].([]*amount.Amount)

		admin_balances := MakeSlice(N)

		for k := uint8(0); k < N; k++ {
			tBalance := tokenBalanceOf(tb.ctx, stableTokens[k], *swap)
			if err != nil {
				return nil, err
			}
			admin_balances[k].Set(Sub(tBalance.Int, swap_reserves[k].Int))
		}
		return ToAmounts(admin_balances), nil
	}

	stableTokenBalances := func(owner common.Address) []*amount.Amount {

		balances := MakeAmountSlice(N)
		for k, token := range stableTokens {
			bal := tokenBalanceOf(tb.ctx, token, owner)
			balances[k].Set(bal.Int)
		}
		return balances
	}

	_min_max := func() (uint8, uint8, *big.Int, *big.Int) {
		is, _ := tb.view(*swap, "Reserves")
		reserves := ToBigInts(is[0].([]*amount.Amount))
		min_idx, min_amount := MinInArray(reserves)
		max_idx, max_amount := MaxInArray(reserves)
		if min_idx == max_idx {
			min_idx = int(Abs(SubC(big.NewInt(int64(min_idx)), 1)).Int64()) // abs(min_idx -1)
		}
		return uint8(min_idx), uint8(max_idx), min_amount, max_amount
	}

	Describe("Exchange", func() {

		BeforeEach(func() {

			deployContracts()

			// swap Deploy
			sD := &trade.StableSwapConstruction{
				Name:         "__STABLE_NAME",
				Symbol:       "__STABLE_SYMBOL",
				Factory:      common.Address{},
				NTokens:      uint8(N),
				Tokens:       stableTokens,
				PayToken:     stableTokens[0],
				Owner:        alice,
				Winner:       charlie,
				Fee:          _Fee,
				AdminFee:     _AdminFee,
				WinnerFee:    _WinnerFee,
				WhiteList:    *whiteList,
				GroupId:      _GroupId,
				Amp:          big.NewInt(_Amp),
				PrecisionMul: _PrecisionMul,
				Rates:        _Rates,
			}

			swap, err = swapDeploy(tb, aliceKey, sD)
			Expect(err).To(Succeed())
		})

		It("payToken", func() {
			is, err := tb.view(*swap, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(stableTokens[0]))
		})

		It("SetPayToken, onlyOwner", func() {

			payToken := stableTokens[1]

			is, err := tb.call(aliceKey, *swap, "SetPayToken", payToken)
			Expect(err).To(Succeed())

			is, err = tb.view(*swap, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))

			is, err = tb.call(bobKey, *swap, "SetPayToken", payToken)
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetPayToken : AddressZero ", func() {

			payToken := common.Address{}

			is, err := tb.call(aliceKey, *swap, "SetPayToken", payToken)
			Expect(err).To(Succeed())

			is, err = tb.view(*swap, "PayToken")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))
		})

		It("SetPayToken : Another token", func() {

			payToken := common.HexToAddress("1")

			_, err := tb.call(aliceKey, *swap, "SetPayToken", payToken)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))
		})

	})

	Describe("factory", func() {

		BeforeEach(func() {
			deployContracts()

			factory, err = factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())
		})

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

		It("deploy stableswap : PayToken Error", func() {

			payToken := common.BigToAddress(big.NewInt(1))

			// swap Deploy
			sD := &trade.StableSwapConstruction{
				Name:         "__STABLE_NAME",
				Symbol:       "__STABLE_SYMBOL",
				Factory:      common.Address{},
				NTokens:      uint8(N),
				Tokens:       stableTokens,
				PayToken:     payToken,
				Owner:        alice,
				Winner:       charlie,
				Fee:          _Fee,
				AdminFee:     _AdminFee,
				WinnerFee:    _WinnerFee,
				WhiteList:    *whiteList,
				GroupId:      _GroupId,
				Amp:          big.NewInt(_Amp),
				PrecisionMul: _PrecisionMul,
				Rates:        _Rates,
			}

			_, err = swapDeploy(tb, aliceKey, sD)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))

		})

		It("CreatePairStable", func() {

			// CreatePairUni
			is, err := tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[0], stableTokens[1], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])

			Expect(err).To(Succeed())
			pair, err := trade.PairFor(*factory, stableTokens[0], stableTokens[1])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// AllPairs
			is, err = tb.view(*factory, "AllPairs")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(pair))

			// AllPairsLength
			is, err = tb.view(*factory, "AllPairsLength")
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = tb.view(*factory, "GetPair", stableTokens[0], stableTokens[1])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// GetPair of reverse tokens
			is, err = tb.view(*factory, "GetPair", stableTokens[1], stableTokens[0])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// CreatePairUni of same tokens
			is, err = tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[0], stableTokens[1], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// CreatePairUni of same reverse tokens
			is, err = tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[1], stableTokens[0], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = tb.view(pair, "Factory")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*factory))

			// tokens
			is, err = tb.view(pair, "Tokens")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{stableTokens[0], stableTokens[1]}))

			// WhiteList
			is, err = tb.view(pair, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = tb.view(pair, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})

		It("CreatePairStable : reverse", func() {

			// CreatePairStable
			is, err := tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[1], stableTokens[0], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
			Expect(err).To(Succeed())

			pair, err := trade.PairFor(*factory, stableTokens[0], stableTokens[1])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// AllPairs
			is, err = tb.view(*factory, "AllPairs")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)[0]).To(Equal(pair))

			// AllPairsLength
			is, err = tb.view(*factory, "AllPairsLength")
			Expect(err).To(Succeed())
			Expect(is[0].(uint16)).To(Equal(uint16(1)))

			// GetPair of tokens
			is, err = tb.view(*factory, "GetPair", stableTokens[0], stableTokens[1])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// GetPair of reverse tokens
			is, err = tb.view(*factory, "GetPair", stableTokens[1], stableTokens[0])
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(pair))

			// CreatePairUni of same tokens
			is, err = tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[1], stableTokens[0], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// CreatePairUni of same reverse tokens
			is, err = tb.call(aliceKey, *factory, "CreatePairStable", stableTokens[0], stableTokens[1], AddressZero, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, *whiteList, _GroupId, uint64(_Amp), ClassMap["StableSwap"])
			Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

			// factory
			is, err = tb.view(pair, "Factory")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*factory))

			// tokens
			is, err = tb.view(pair, "Tokens")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{stableTokens[1], stableTokens[0]}))

			// WhiteList
			is, err = tb.view(pair, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = tb.view(pair, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))

		})
	})

	Describe("Token", func() {

		var _TotalSupply *amount.Amount
		_TestAmount := amount.NewAmount(10, 0)

		BeforeEach(func() {
			beforeEachDefault()
			stableAddLiquidityDefault(aliceKey)
		})

		It("Name, Symbol, TotalSupply, Decimals", func() {
			// Name() string
			is, err := tb.view(*swap, "Name")
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapName))

			// Symbol() string
			is, err = tb.view(*swap, "Symbol")
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapSymbol))

			// TotalSupply() *amount.Amount
			is, err = tb.view(*swap, "TotalSupply")
			Expect(err).To(Succeed())
			_TotalSupply = is[0].(*amount.Amount)

			// Decimals() *big.Int
			is, err = tb.view(*swap, "Decimals")
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(big.NewInt(amount.FractionalCount)))
		})

		It("Transfer", func() {
			//Transfer(To common.Address, Amount *amount.Amount)
			_, err := tb.call(aliceKey, *swap, "Transfer", bob, _TestAmount)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, *swap, bob)).To(Equal(_TestAmount))
		})

		It("Approve", func() {
			//Approve(To common.Address, Amount *amount.Amount)
			_, err := tb.call(aliceKey, *swap, "Approve", bob, _TestAmount)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, *swap, alice, bob)).To(Equal(_TestAmount))
		})

		It("IncreaseAllowance", func() {
			//IncreaseAllowance(spender common.Address, addAmount *amount.Amount)
			_, err := tb.call(aliceKey, *swap, "IncreaseAllowance", bob, _TestAmount)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, *swap, alice, bob)).To(Equal(_TestAmount))
		})

		It("DecreaseAllowance", func() {
			//DecreaseAllowance(spender common.Address, subtractAmount *amount.Amount)
			tb.call(aliceKey, *swap, "Approve", bob, _TestAmount.MulC(3))

			_, err := tb.call(aliceKey, *swap, "DecreaseAllowance", bob, _TestAmount)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, *swap, alice, bob)).To(Equal(_TestAmount.MulC(2)))
		})

		It("TransferFrom", func() {
			//TransferFrom(From common.Address, To common.Address, Amount *amount.Amount)
			tb.call(aliceKey, *swap, "Approve", bob, _TestAmount.MulC(3))

			_, err := tb.call(bobKey, *swap, "TransferFrom", alice, charlie, _TestAmount)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, *swap, alice, bob)).To(Equal(_TestAmount.MulC(2)))

			Expect(tokenBalanceOf(tb.ctx, *swap, alice)).To(Equal(_TotalSupply.Sub(_TestAmount)))
			Expect(tokenBalanceOf(tb.ctx, *swap, charlie)).To(Equal(_TestAmount))

		})

	})

	Describe("stableswap front", func() {

		BeforeEach(func() {
			beforeEach(uint64(1234), uint64(3456), uint64(67890), false)
		})

		It("Extype, Fee, AdminFee, NTokens, Rates, PrecisionMul, Tokens, Owner, WhiteList, GroupId", func() {
			//Extype() uint8``
			is, err := tb.view(*swap, "ExType")
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(trade.STABLE))

			// NTokens() uint8
			is, err = tb.view(*swap, "NTokens")
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(uint8(N)))

			// RATES() []*big.Int
			is, err = tb.view(*swap, "Rates")
			Expect(err).To(Succeed())
			Expect(is[0].([]*big.Int)).To(Equal(_Rates))

			// PRECISION_MUL() []uint64
			is, err = tb.view(*swap, "PrecisionMul")
			Expect(err).To(Succeed())
			Expect(is[0].([]uint64)).To(Equal(_PrecisionMul))

			// Coins() []common.Address
			is, err = tb.view(*swap, "Tokens")
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal(stableTokens))

			// Owner() common.Address
			is, err = tb.view(*swap, "Owner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(alice))

			// Fee() uint64
			is, err = tb.view(*swap, "Fee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(1234)))

			// AdminFee() uint64
			is, err = tb.view(*swap, "AdminFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(3456)))

			// WinnerFee() uint64
			is, err = tb.view(*swap, "WinnerFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(67890)))

			// WhiteList
			is, err = tb.view(*swap, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*whiteList))

			// GroupId
			is, err = tb.view(*swap, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})

		It("InitialA, FutureA, RampA, InitialA, FutureA, InitialATime, FutureATime, StopRampA", func() {

			is, err := tb.view(*swap, "InitialA")
			Expect(err).To(Succeed())
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
			future_time := tb.ctx.LastTimestamp()/uint64(time.Second) + trade.MIN_RAMP_TIME + 5

			timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)
			_, err = tb.call(aliceKey, *swap, "RampA", MulC(initial_A, 2), future_time)
			Expect(err).To(Succeed())

			// InitialA() *big.Int
			is, err = tb.view(*swap, "InitialA")
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(MulC(initial_A, trade.A_PRECISION)))

			// FutureA() *big.Int
			is, err = tb.view(*swap, "FutureA")
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(MulC(MulC(initial_A, 2), trade.A_PRECISION)))

			// InitialATime() uint64
			is, err = tb.view(*swap, "InitialATime")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(timestamp))

			// FutureATime() uint64
			is, err = tb.view(*swap, "FutureATime")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(future_time))

			//StopRampA()
			_, err = tb.call(aliceKey, *swap, "StopRampA")
			Expect(err).To(Succeed())

		})

		It("CommitNewFee, AdminActionsDeadline, FutureFee, FutureAdminFee, ApplyNewFee, RevertNewParameters", func() {

			fee := uint64(rand.Intn(trade.MAX_FEE) + 1)
			admin_fee := uint64(rand.Intn(trade.MAX_ADMIN_FEE + 1))
			winner_fee := uint64(rand.Intn(trade.MAX_WINNER_FEE + 1))
			delay := uint64(3 * 86400)

			timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)

			_, err = tb.call(aliceKey, *swap, "CommitNewFee", fee, admin_fee, winner_fee, delay)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err := tb.view(*swap, "AdminActionsDeadline")
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureFee() uint64
			is, err = tb.view(*swap, "FutureFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// FutureAdminFee() uint64
			is, err = tb.view(*swap, "FutureAdminFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// FutureWinnerFee() uint64
			is, err = tb.view(*swap, "FutureWinnerFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			tb.sleep(delay*1000 - 1000)

			_, err = tb.call(aliceKey, *swap, "ApplyNewFee")
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = tb.view(*swap, "AdminActionsDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee() uint64
			is, err = tb.view(*swap, "Fee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee() uint64
			is, err = tb.view(*swap, "AdminFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee() uint64
			is, err = tb.view(*swap, "WinnerFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			_, err = tb.call(aliceKey, *swap, "CommitNewFee", uint64(1000), uint64(2000), uint64(3000), delay)
			Expect(err).To(Succeed())

			// RevertNewParameters()
			_, err = tb.call(aliceKey, *swap, "RevertNewFee")
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = tb.view(*swap, "AdminActionsDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee() uint64
			is, err = tb.view(*swap, "Fee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee() uint64
			is, err = tb.view(*swap, "AdminFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee() uint64
			is, err = tb.view(*swap, "WinnerFee")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

		})

		It("CommitNewWhiteList, WhiteListDeadline, FutureWhiteList, FutureGroupId, ApplyNewWhiteList, RevertNewWhiteList", func() {

			wl, err := whiteListDeploy(tb, aliceKey)
			Expect(err).To(Succeed())
			gId := hash.BigToHash(big.NewInt(100))

			delay := uint64(3 * 86400)
			timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)

			_, err = tb.call(aliceKey, *swap, "CommitNewWhiteList", *wl, gId, delay)
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err := tb.view(*swap, "WhiteListDeadline")
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureWhiteList() common.Address
			is, err = tb.view(*swap, "FutureWhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*wl))

			// FutureGroupId() hash.Hash256
			is, err = tb.view(*swap, "FutureGroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			tb.sleep(delay*1000 - 1000)

			_, err = tb.call(aliceKey, *swap, "ApplyNewWhiteList")
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err = tb.view(*swap, "WhiteListDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = tb.view(*swap, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*wl))

			// GroupId() hash.Hash256
			is, err = tb.view(*swap, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			_, err = tb.call(aliceKey, *swap, "CommitNewWhiteList", *whiteList, _GroupId, delay)
			Expect(err).To(Succeed())

			// RevertNewWhiteList()
			_, err = tb.call(aliceKey, *swap, "RevertNewWhiteList")
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = tb.view(*swap, "WhiteListDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = tb.view(*swap, "WhiteList")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(*wl))

			// GroupId() hash.Hash256
			is, err = tb.view(*swap, "GroupId")
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))
		})

		It("Owner, CommitTransferOwnerWinner, TransferOwnerWinnerDeadline, ApplyTransferOwnerWinner, FutureOwner, RevertTransferOwnerWinner", func() {

			delay := uint64(3 * 86400)
			timestamp := tb.ctx.LastTimestamp() / uint64(time.Second)

			// CommitTransferOwnerWinner( _owner common.Address)
			_, err = tb.call(aliceKey, *swap, "CommitTransferOwnerWinner", bob, charlie, delay)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, err := tb.view(*swap, "TransferOwnerWinnerDeadline")
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureOwner() common.Address
			is, err = tb.view(*swap, "FutureOwner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// FutureOwner() common.Address
			is, err = tb.view(*swap, "FutureWinner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			tb.sleep(delay*1000 - 1000)

			// ApplyTransferOwnerWinner()
			_, err = tb.call(aliceKey, *swap, "ApplyTransferOwnerWinner")
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, _ = tb.view(*swap, "TransferOwnerWinnerDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner() common.Address
			is, err = tb.view(*swap, "Owner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner() common.Address
			is, err = tb.view(*swap, "Winner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			_, err = tb.call(bobKey, *swap, "CommitTransferOwnerWinner", charlie, alice, delay)
			Expect(err).To(Succeed())

			// RevertTransferOwnerWinner()
			_, err = tb.call(bobKey, *swap, "RevertTransferOwnerWinner")
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, _ = tb.view(*swap, "TransferOwnerWinnerDeadline")
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner() common.Address
			is, err = tb.view(*swap, "Owner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner() common.Address
			is, err = tb.view(*swap, "Winner")
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

		})

		It("Withdraw_admin_fees, Donate_admin_fees", func() {
			stableAddLiquidityDefault(aliceKey)
			stableMint(bob)
			stableApprove(bobKey)
			setFees(tb, aliceKey, *swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE)

			// 0 -> 1 Exchange
			_, err = tb.call(bobKey, *swap, "Exchange", uint8(0), uint8(1), _InitialAmounts[0], AmountZero, AddressZero)
			Expect(err).To(Succeed())

			// Withdraw_admin_fees()
			_, err = tb.call(aliceKey, *swap, "WithdrawAdminFees")
			Expect(err).To(Succeed())

			// 1 -> 0 Exchange
			_, err = tb.call(bobKey, *swap, "Exchange", uint8(1), uint8(0), _InitialAmounts[1], AmountZero, AddressZero)
			Expect(err).To(Succeed())

			// Donate_admin_fees()
			_, err = tb.call(aliceKey, *swap, "DonateAdminFees")
			Expect(err).To(Succeed())
		})

		It("KillMe, UnkillMe", func() {
			// KillMe()
			_, err = tb.call(aliceKey, *swap, "KillMe")
			Expect(err).To(Succeed())

			// UnkillMe()
			_, err = tb.call(aliceKey, *swap, "UnkillMe")
			Expect(err).To(Succeed())
		})

	})

	Describe("tokenTransfer", func() {

		var token0, token1 *common.Address
		var pair common.Address

		_SupplyTokens := []*amount.Amount{amount.NewAmount(500000, 0), amount.NewAmount(1000000, 0)}
		_TestAmount := amount.NewAmount(10, 0)

		BeforeEach(func() {
			// erc20Token deploy
			token0, err = erc20TokenDeploy(tb, aliceKey, amount.NewAmount(0, 0))
			Expect(err).To(Succeed())

			//  token deploy
			token1, err = tokenDeploy(tb, aliceKey, "Token1", "TKN1")
			Expect(err).To(Succeed())

			// setMinter
			for _, token := range []common.Address{*token0, *token1} {
				_, err = tb.call(aliceKey, token, "SetMinter", alice, true)
				Expect(err).To(Succeed())
			}

			whiteList, err = whiteListDeploy(tb, aliceKey)
			Expect(err).To(Succeed())

			factory, err = factoryDeploy(tb, aliceKey)
			Expect(err).To(Succeed())

			tb.call(aliceKey, *token0, "Mint", bob, _SupplyTokens[0])
			tb.call(aliceKey, *token1, "Mint", bob, _SupplyTokens[1])

			is, err := tb.call(aliceKey, *factory, "CreatePairStable", *token0, *token1, *token1, "USDT-USDC Dex", "USDT-USDC", alice, charlie, uint64(30000000), uint64(3000000000), uint64(0), *whiteList, _GroupId, uint64(170), ClassMap["StableSwap"])
			Expect(err).To(Succeed())
			pair = is[0].(common.Address)
		})

		It("tokentransfer", func() {

			_, err := tb.call(bobKey, *token0, "Transfer", pair, _TestAmount)
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(tb.ctx, *token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(tb.ctx, *token0, pair)).To(Equal(_TestAmount))

			_, err = tb.call(bobKey, *token1, "Transfer", pair, _TestAmount.MulC(2))
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(tb.ctx, *token1, bob)).To(Equal(_SupplyTokens[1].Sub(_TestAmount.MulC(2))))
			Expect(tokenBalanceOf(tb.ctx, *token1, pair)).To(Equal(_TestAmount.MulC(2)))

			_, err = tb.call(aliceKey, pair, "TokenTransfer", *token0, charlie, _TestAmount.DivC(2))
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(tb.ctx, *token0, pair)).To(Equal(_TestAmount.DivC(2)))
			Expect(tokenBalanceOf(tb.ctx, *token0, charlie)).To(Equal(_TestAmount.DivC(2)))

			_, err = tb.call(aliceKey, pair, "TokenTransfer", *token1, charlie, _TestAmount)
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(tb.ctx, *token1, pair)).To(Equal(_TestAmount))
			Expect(tokenBalanceOf(tb.ctx, *token1, charlie)).To(Equal(_TestAmount))

		})
		It("onlyOwner", func() {

			_, err := tb.call(bobKey, pair, "TokenTransfer", *token0, charlie, _TestAmount)
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("transfer more than balance", func() {

			_, err := tb.call(bobKey, *token0, "Transfer", pair, _TestAmount)
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(tb.ctx, *token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(tb.ctx, *token0, pair)).To(Equal(_TestAmount))

			_, err = tb.call(aliceKey, pair, "TokenTransfer", *token0, charlie, _TestAmount.MulC(2))
			Expect(err).To(HaveOccurred())

		})

		It("not base-token transfer", func() {

			otherToken := common.BigToAddress(big.NewInt(1))
			_, err := tb.call(aliceKey, pair, "TokenTransfer", otherToken, charlie, _TestAmount)
			Expect(err).To(MatchError("Exchange: NOT_EXIST_TOKEN"))

		})
	})

	Describe("unitary", func() {

		It("CalcLPTokenAmount", func() {
			Skip(AmountNotNegative)

			beforeEachDefault()

			for i := uint8(0); i < N; i++ {
				amounts := MakeAmountSlice(N)
				amounts[i].Int.Set(big.NewInt(-1))
				_, err := tb.call(aliceKey, *swap, "CalcLPTokenAmount", amounts, true)
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))
			}
		})

		Describe("test_add_liquidity_initial.py", func() {
			BeforeEach(func() {
				beforeEachDefault()
				stableMint(alice)
				stableApprove(aliceKey)
			})

			It("negative input amount", func() {
				Skip(AmountNotNegative)

				stableAddLiquidity(aliceKey, _InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amounts := MakeAmountSlice(N)

					amounts[i].Int.Set(big.NewInt(-1))
					_, err := tb.call(aliceKey, *swap, "AddLiquidity", amounts, amount.NewAmount(0, 0))
					Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))
				}
			})

			DescribeTable("test_initial",
				func(min_amount *amount.Amount) {
					amounts := MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amounts[i].Set(Pow10(decimals[i]))
					}

					_, err := tb.call(aliceKey, *swap, "AddLiquidity", amounts, min_amount)
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice)).To(Equal(_InitialAmounts[i].Sub(amounts[i])))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(amounts[i]))
					}

					Expect(tokenBalanceOf(tb.ctx, *swap, alice)).To(Equal(amount.NewAmount(uint64(N), 0)))
					Expect(tokenTotalSupply(tb.ctx, *swap)).To(Equal(amount.NewAmount(uint64(N), 0)))
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
					_, err := tb.call(aliceKey, *swap, "AddLiquidity", amounts, amount.NewAmount(0, 0))
					Expect(err).To(MatchError("Exchange: INITILAL_DEPOSIT"))
				}
			})
		})

		Describe("test_add_liquidity.py", func() {

			BeforeEach(func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)
			})

			AfterEach(func() {

			})

			It("test_add_liquidity", func() {
				_, err := tb.call(bobKey, *swap, "AddLiquidity", _InitialAmounts, AmountZero)
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					Expect(tokenBalanceOf(tb.ctx, stableTokens[i], bob).Cmp(AmountZero.Int)).To(Equal(0))
					Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(_InitialAmounts[i].MulC(2)))
				}
				Expect(tokenBalanceOf(tb.ctx, *swap, bob)).To(Equal(amount.NewAmount(uint64(N), 0).MulC(_BaseAmount)))
				Expect(tokenTotalSupply(tb.ctx, *swap)).To(Equal(amount.NewAmount(uint64(N), 0).MulC(_BaseAmount * 2)))

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

				_, err = tb.call(bobKey, *swap, "AddLiquidity", ToAmounts(amounts), AmountZero)
				Expect(err).To(Succeed())

				balance := tokenBalanceOf(tb.ctx, *swap, bob)
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
						beforeEachDefault()
						stableAddLiquidityDefault(aliceKey)
						stableMint(bob)
						stableApprove(bobKey)
					}

					amounts := MakeAmountSlice(N)
					amounts[idx].Set(_InitialAmounts[idx].Int)

					_, err := tb.call(bobKey, *swap, "AddLiquidity", amounts, AmountZero)
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						balance := tokenBalanceOf(tb.ctx, stableTokens[i], bob)
						if balance.Cmp(_InitialAmounts[i].Sub(amounts[i]).Int) != 0 {
							Fail("Not Equal")
						}
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(_InitialAmounts[i].Add(amounts[i])))
					}

					balance := tokenBalanceOf(tb.ctx, *swap, bob)
					if balance.Cmp(big.NewInt(int64(float64(amount.NewAmount(uint64(_BaseAmount), 0).Int64())*0.999))) <= 0 {
						Fail("Lower")
					}
					if balance.Cmp(amount.NewAmount(uint64(_BaseAmount), 0).Int) >= 0 {
						Fail("Upper")
					}

				}
			})

			It("test_insufficient_balance", func() {
				amounts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amounts[i].Set(Pow10(decimals[i]))
				}

				_, err := tb.call(charlieKey, *swap, "AddLiquidity", amounts, AmountZero)

				insufficientBalanceErrorCheck(err)
			})

			It("test_min_amount_too_high", func() {
				amounts := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amounts[i].Set(Pow10(decimals[i]))
				}

				min_amount := amount.NewAmount(uint64(N), 1)
				_, err := tb.call(bobKey, *swap, "AddLiquidity", amounts, min_amount)
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
				_, err := tb.call(bobKey, *swap, "AddLiquidity", amounts, min_amount)
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
				beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
				stableAddLiquidityDefault(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				_, err := tb.call(bobKey, *swap, "WithdrawAdminFees")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("AdminBalances index error", func() {
				beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
				stableAddLiquidityDefault(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				_, err := tb.call(bobKey, *swap, "AdminBalances", uint8(4))
				Expect(err).To(MatchError("Exchange: IDX"))

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
									beforeEach(fee, adminFee, winnerFee, false)
									stableAddLiquidityDefault(aliceKey)
									stableMint(bob)
									stableApprove(bobKey)

									send, recv := sending, receiving
									is, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
									Expect(err).To(Succeed())
									dy_receiving := is[0].(*amount.Amount)
									dy_receiving = ToAmount(MulDivCC(dy_receiving.Int, trade.FEE_DENOMINATOR, int64(trade.FEE_DENOMINATOR-fee))) // y = y' * DENOM / (DENOM-FEE)
									fee_receiving := getFee(dy_receiving.Int, fee)
									adminfee_receiving := getFee(fee_receiving, adminFee)

									send, recv = receiving, sending
									is, err = tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
									Expect(err).To(Succeed())
									dy_sending := is[0].(*amount.Amount)
									dy_sending = ToAmount(MulDivCC(dy_sending.Int, trade.FEE_DENOMINATOR, int64(trade.FEE_DENOMINATOR-fee))) // y = y' * DENOM / (DENOM-FEE)
									fee_sending := getFee(dy_sending.Int, fee)
									adminfee_sending := getFee(fee_sending, adminFee)

									is, err = tb.view(*swap, "Reserves")
									Expect(err).To(Succeed())

									reserves := is[0].([]*amount.Amount)

									for k := uint8(0); k < N; k++ {
										balance_coin := tokenBalanceOf(tb.ctx, stableTokens[k], *swap)

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
									beforeEach(fee, adminFee, winnerFee, false)
									stableAddLiquidityDefault(aliceKey)
									stableMint(bob)
									stableApprove(bobKey)

									_, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
									Expect(err).To(Succeed())

									admin_balances, err := stableGetAdminBalances() // fixture
									Expect(err).To(Succeed())
									for i := uint8(0); i < N; i++ {
										is, err := tb.view(*swap, "AdminBalances", uint8(i))
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

									_, err = tb.call(aliceKey, *swap, "WithdrawAdminFees")
									Expect(err).To(Succeed())

									winner := tb.viewAddress(*swap, "Winner")

									wF := ToAmount(MulDivCC(admin_balances[recv].Int, int64(winnerFee), trade.FEE_DENOMINATOR))
									Expect(tokenBalanceOf(tb.ctx, stableTokens[recv], winner)).To(Equal(wF))
									Expect(tokenBalanceOf(tb.ctx, stableTokens[recv], alice)).To(Equal(admin_balances[recv].Sub(wF)))

									is, err := tb.view(*swap, "Reserves")
									Expect(err).To(Succeed())

									swap_reserves := is[0].([]*amount.Amount)
									Expect(tokenBalanceOf(tb.ctx, stableTokens[recv], *swap)).To(Equal(swap_reserves[recv]))

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

							beforeEach(fee, adminFee, winnerFee, false)
							stableAddLiquidityDefault(aliceKey)
							stableMint(bob)
							stableApprove(bobKey)

							// 0,1 -> 1,2 -> 2,0
							for i := uint8(0); i < N; i++ {
								send := i
								recv := i + 1
								if recv == N {
									recv = 0
								}

								_, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
								Expect(err).To(Succeed())
							}

							admin_balances, err := stableGetAdminBalances() // fixture
							Expect(err).To(Succeed())
							for i := uint8(0); i < N; i++ {
								is, err := tb.view(*swap, "AdminBalances", uint8(i))
								Expect(err).To(Succeed())
								if admin_balances[i].Cmp(Zero) == 0 {
									Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
								} else {
									Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i]))
								}
							}

							_, err = tb.call(aliceKey, *swap, "WithdrawAdminFees")
							Expect(err).To(Succeed())

							winner := tb.viewAddress(*swap, "Winner")

							for i := uint8(0); i < N; i++ {
								wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(winnerFee)), trade.FEE_DENOMINATOR))
								Expect(tokenBalanceOf(tb.ctx, stableTokens[i], winner).Cmp(wF.Int)).To(Equal(0))
								Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice).Cmp(admin_balances[i].Sub(wF).Int)).To(Equal(0))
							}

						}
					}
				}
			})

			It("test_withdraw_all_coins : payToken Error", func() {

				payToken := common.BigToAddress(big.NewInt(1))

				// swap Deploy
				sD := &trade.StableSwapConstruction{
					Name:         "__STABLE_NAME",
					Symbol:       "__STABLE_SYMBOL",
					Factory:      common.Address{},
					NTokens:      uint8(N),
					Tokens:       stableTokens,
					PayToken:     payToken,
					Owner:        alice,
					Winner:       charlie,
					Fee:          _Fee,
					AdminFee:     _AdminFee,
					WinnerFee:    _WinnerFee,
					WhiteList:    *whiteList,
					GroupId:      _GroupId,
					Amp:          big.NewInt(_Amp),
					PrecisionMul: _PrecisionMul,
					Rates:        _Rates,
				}

				swap, err = swapDeploy(tb, aliceKey, sD)
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

								deployContracts()

								// swap Deploy
								sD := &trade.StableSwapConstruction{
									Name:         "__STABLE_NAME",
									Symbol:       "__STABLE_SYMBOL",
									NTokens:      uint8(N),
									Tokens:       stableTokens,
									PayToken:     stableTokens[pi],
									Owner:        alice,
									Winner:       charlie,
									Fee:          fee,
									AdminFee:     adminFee,
									WinnerFee:    winnerFee,
									WhiteList:    *whiteList,
									GroupId:      _GroupId,
									Amp:          big.NewInt(_Amp),
									PrecisionMul: _PrecisionMul,
									Rates:        _Rates,
								}

								swap, err = swapDeploy(tb, aliceKey, sD)

								stableAddLiquidityDefault(aliceKey)
								stableMint(bob)
								stableApprove(bobKey)

								// 0,1 -> 1,2 -> 2,0
								for i := uint8(0); i < N; i++ {
									send := i
									recv := i + 1
									if recv == N {
										recv = 0
									}
									_, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
									Expect(err).To(Succeed())
								}

								admin_balances := MakeAmountSlice(N)
								for i := uint8(0); i < N; i++ {
									is, err := tb.view(*swap, "AdminBalances", uint8(i))
									Expect(err).To(Succeed())
									admin_balances[i].Set(is[0].(*amount.Amount).Int)
								}

								_, err = tb.call(aliceKey, *swap, "WithdrawAdminFees")
								Expect(err).To(Succeed())

								winner := tb.viewAddress(*swap, "Winner")

								for i := uint8(0); i < N; i++ {
									if i == pi {
										wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(winnerFee)), trade.FEE_DENOMINATOR))
										Expect(tokenBalanceOf(tb.ctx, stableTokens[i], winner).Cmp(wF.Int)).To(Equal(0))
										Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice).Cmp(admin_balances[i].Sub(wF).Int)).To(Equal(0))
									} else {
										Expect(tokenBalanceOf(tb.ctx, stableTokens[i], winner).Cmp(AmountZero.Int)).To(Equal(0))
										Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice).Cmp(AmountZero.Int)).To(Equal(0))
									}
								}

							}
						}
					}
				}
			})
		})

		Describe("test_exchange_reverts.py", func() {

			It("GetDy non-positive input amount", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				_, err := tb.view(*swap, "GetDy", uint8(0), uint8(1), ToAmount(big.NewInt(-1)), AddressZero)
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			})

			It("Exchange non-positive input amount", func() {
				Skip(AmountNotNegative)

				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				_, err := tb.call(aliceKey, *swap, "Exchange", uint8(0), uint8(1), ToAmount(big.NewInt(-1)), AmountZero, AddressZero)
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			})

			It("test_insufficient_balance", func() {
				for send := uint8(0); send < N; send++ {
					for recv := uint8(0); recv < N; recv++ {
						if send == recv {
							continue
						}
						beforeEachDefault()
						stableAddLiquidityDefault(aliceKey)
						stableApprove(bobKey)

						amt := ToAmount(Pow10(decimals[send]))
						tb.call(aliceKey, stableTokens[send], "Mint", bob, amt)

						_, err := tb.call(bobKey, *swap, "Exchange", send, recv, ToAmount(Add(amt.Int, big.NewInt(1))), AmountZero, AddressZero)

						insufficientBalanceErrorCheck(err)
					}
				}
			})

			It("test_min_dy_too_high", func() {
				for send := uint8(0); send < N; send++ {
					for recv := uint8(0); recv < N; recv++ {
						if send == recv {
							continue
						}
						beforeEachDefault()
						stableAddLiquidityDefault(aliceKey)
						stableApprove(bobKey)

						amt := ToAmount(Pow10(decimals[send]))
						tb.call(aliceKey, stableTokens[send], "Mint", bob, amt)

						is, err := tb.view(*swap, "GetDy", send, recv, amt, AddressZero)
						Expect(err).To(Succeed())

						min_dy := is[0].(*amount.Amount)

						_, err = tb.call(bobKey, *swap, "Exchange", send, recv, amt, ToAmount(AddC(min_dy.Int, 2)), AddressZero)
						Expect(err).To(HaveOccurred())

					}
				}
			})

			It("test_same_coin", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)
					stableApprove(bobKey)

					_, err := tb.call(bobKey, *swap, "Exchange", idx, idx, AmountZero, AmountZero, AddressZero)
					Expect(err).To(MatchError("Exchange: OUT"))

				}
			})

			It("test_i_below_zero", func() {
				// uint8 can't be below zero
			})

			It("test_i_above_n_coins", func() {
				idxes := []uint8{9, 140, 255}
				for i := 0; i < len(idxes); i++ {
					idx := idxes[i]
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)
					stableApprove(bobKey)

					_, err := tb.call(bobKey, *swap, "Exchange", idx, uint8(0), AmountZero, AmountZero, AddressZero)
					Expect(err).To(MatchError("Exchange: IN"))

				}
			})

			It("test_j_below_zero", func() {
				//uint8 can'b below zero
			})

			It("test_j_above_n_coins", func() {
				idxes := []uint8{9, 100}
				for i := 0; i < len(idxes); i++ {
					idx := idxes[i]
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)
					stableApprove(bobKey)

					_, err := tb.call(bobKey, *swap, "Exchange", uint8(0), idx, AmountZero, AmountZero, AddressZero)
					Expect(err).To(MatchError("Exchange: OUT"))

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

									beforeEach(uint64(fee*float64(trade.FEE_DENOMINATOR)), uint64(admin_fee*float64(trade.FEE_DENOMINATOR)), uint64(winner_fee*float64(trade.FEE_DENOMINATOR)), false)
									stableAddLiquidityDefault(aliceKey)
									stableApprove(bobKey)

									amt := ToAmount(Pow10(decimals[send]))
									tb.call(aliceKey, stableTokens[send], "Mint", bob, amt)

									_, err = tb.call(bobKey, *swap, "Exchange", send, recv, amt, AmountZero, AddressZero)
									Expect(err).To(Succeed())

									Expect(tokenBalanceOf(tb.ctx, stableTokens[send], bob).Cmp(AmountZero.Int)).To(Equal(0))
									received := tokenBalanceOf(tb.ctx, stableTokens[recv], bob)

									received_float, _ := new(big.Float).SetInt(received.Int).Float64()
									r := received_float / float64(Pow10(decimals[recv]).Uint64())
									if 1-math.Max(math.Pow10(-4), 1/received_float)-fee >= r || r >= 1-fee {
										Fail("Fee")
									}

									expected_admin_fee := ToFloat64(Pow10(decimals[recv])) * fee * admin_fee
									a_fees, err := stableGetAdminBalances()
									Expect(err).To(Succeed())
									for i := uint8(0); i < N; i++ {
										is, err := tb.view(*swap, "AdminBalances", uint8(i))
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
						beforeEachDefault()
						stableAddLiquidityDefault(aliceKey)
						stableApprove(bobKey)

						amt := ToAmount(Pow10(decimals[send]))
						tb.call(aliceKey, stableTokens[send], "Mint", bob, amt)

						is, err := tb.view(*swap, "GetDy", send, recv, amt, AddressZero)
						Expect(err).To(Succeed())

						min_dy := is[0].(*amount.Amount)

						_, err = tb.call(bobKey, *swap, "Exchange", send, recv, amt, ToAmount(Sub(min_dy.Int, big.NewInt(1))), AddressZero)
						Expect(err).To(Succeed())

						received := tokenBalanceOf(tb.ctx, stableTokens[recv], bob)

						if Abs(Sub(received.Int, min_dy.Int)).Cmp(big.NewInt(1)) > 0 {
							Fail("Recived")
						}

					}
				}
			})
		})

		Describe("test_get_virtual_price.py", func() {
			It("test_number_go_up", func() {
				beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
				stableAddLiquidityDefault(aliceKey)
				stableApprove(bobKey)
				stableMint(bob)

				is, err := tb.view(*swap, "GetVirtualPrice")
				Expect(err).To(Succeed())

				virtual_price := is[0].(*amount.Amount)

				for i := uint8(0); i < N; i++ {
					amts := MakeSlice(N)
					amts[i] = Clone(_InitialAmounts[i].Int)

					_, err := tb.call(bobKey, *swap, "AddLiquidity", ToAmounts(amts), AmountZero)
					Expect(err).To(Succeed())

					is, err := tb.view(*swap, "GetVirtualPrice")
					Expect(err).To(Succeed())

					// admin_fee = 100% 인경우 계산오차가 차이날 수 있음  new_virtual_price >= virtual_price - 1
					new_virtual_price := is[0].(*amount.Amount)
					if AddC(new_virtual_price.Int, 2).Cmp(virtual_price.Int) <= 0 {
						log.Println("new_virtual_price, virtual_price", new_virtual_price, virtual_price)
						Fail("Vitrual Price")
					}
					virtual_price.Int.Set(new_virtual_price.Int)
				}

			})

			It("test_remove_one_coin", func() {

				for idx := uint8(0); idx < N; idx++ {
					beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
					stableAddLiquidityDefault(aliceKey)
					stableApprove(bobKey)
					stableMint(bob)

					amt := tokenBalanceOf(tb.ctx, *swap, alice)
					amt = amt.DivC(10)

					is, err := tb.view(*swap, "GetVirtualPrice")
					Expect(err).To(Succeed())

					virtual_price := is[0].(*amount.Amount)

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
					Expect(err).To(Succeed())

					is, err = tb.view(*swap, "GetVirtualPrice")
					Expect(err).To(Succeed())
					new_virtual_price := is[0].(*amount.Amount)

					// admin_fee = 100% 인경우 계산오차가 차이날 수 있음  new_virtual_price >= virtual_price - 1
					if AddC(new_virtual_price.Int, 2).Cmp(virtual_price.Int) <= 0 {
						log.Println("new_virtual_price, virtual_price", new_virtual_price, virtual_price)
						Fail("Virtual Price")
					}

				}

			})

			It("test_remove_imbalance", func() {

				for idx := uint8(0); idx < N; idx++ {
					beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
					stableAddLiquidityDefault(aliceKey)
					stableApprove(bobKey)
					stableMint(bob)

					amts := CloneAmountSlice(_InitialAmounts)
					for i := uint8(0); i < N; i++ {
						amts[i] = amts[i].DivC(2)
					}
					amts[idx] = amount.NewAmount(0, 0)

					is, err := tb.view(*swap, "GetVirtualPrice")
					Expect(err).To(Succeed())

					virtual_price := is[0].(*amount.Amount)

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0))
					Expect(err).To(Succeed())

					is, err = tb.view(*swap, "GetVirtualPrice")
					Expect(err).To(Succeed())

					new_virtual_price := is[0].(*amount.Amount)

					if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
						Fail("Virtual Price")
					}

				}

			})

			It("test_remove", func() {

				beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
				stableAddLiquidityDefault(aliceKey)
				stableApprove(bobKey)
				stableMint(bob)

				withdraw_amount := DivC(Sum(ToBigInts(_InitialAmounts)), 2)

				is, err := tb.view(*swap, "GetVirtualPrice")
				Expect(err).To(Succeed())

				virtual_price := is[0].(*amount.Amount)

				_, err = tb.call(aliceKey, *swap, "RemoveLiquidity", withdraw_amount, MakeAmountSlice(N))
				Expect(err).To(Succeed())

				is, err = tb.view(*swap, "GetVirtualPrice")
				Expect(err).To(Succeed())

				new_virtual_price := is[0].(*amount.Amount)

				if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
					Fail("Virtual Price")
				}

			})

			It("test_exchange", func() {

				for send := uint8(0); send < N; send++ {
					for recv := uint8(0); recv < N; recv++ {
						if send == recv {
							continue
						}
						beforeEach(trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, false)
						stableAddLiquidityDefault(aliceKey)
						stableApprove(bobKey)
						stableMint(bob)

						is, err := tb.view(*swap, "GetVirtualPrice")
						Expect(err).To(Succeed())
						virtual_price := is[0].(*amount.Amount)

						amt := ToAmount(Pow10(decimals[send]))

						_, err = tb.call(bobKey, *swap, "Exchange", send, recv, amt, AmountZero, AddressZero)
						Expect(err).To(Succeed())

						is, err = tb.view(*swap, "GetVirtualPrice")
						Expect(err).To(Succeed())

						new_virtual_price := is[0].(*amount.Amount)

						if new_virtual_price.Int.Cmp(virtual_price.Int) < 0 {
							Fail("Virtual Price")
						}

					}
				}
			})

			It("test_exchange_underlying", func() {
				// no lending
			})

		})

		Describe("test_kill.py", func() {
			BeforeEach(func() {
				beforeEachDefault()
			})

			AfterEach(func() {

			})

			It("kill_me", func() {
				_, err := tb.call(aliceKey, *swap, "KillMe")
				Expect(err).To(Succeed())
			})

			It("test_kill_approaching_deadline", func() {
				// no dead line
			})

			It("test_kill_only_owner", func() {
				_, err := tb.call(bobKey, *swap, "KillMe")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))
			})

			It("test_kill_beyond_deadline", func() {
				// no deadline
			})

			It("test_kill_and_unkill", func() {
				tb.call(aliceKey, *swap, "KillMe")

				_, err := tb.call(aliceKey, *swap, "UnkillMe")
				Expect(err).To(Succeed())
			})

			It("test_unkill_without_kill", func() {
				_, err := tb.call(aliceKey, *swap, "UnkillMe")
				Expect(err).To(Succeed())

			})

			It("test_unkill_only_owner", func() {
				_, err := tb.call(bobKey, *swap, "UnkillMe")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("test_remove_liquidity", func() {
				stableAddLiquidityDefault(aliceKey)

				tb.call(aliceKey, *swap, "KillMe")
				_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), MakeAmountSlice(N))
				Expect(err).To(Succeed())

			})

			It("test_remove_liquidity_imbalance", func() {
				tb.call(aliceKey, *swap, "KillMe")
				_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", MakeAmountSlice(N), AmountZero)
				Expect(err).To(MatchError("Exchange: KILLED"))

			})

			It("test_remove_liquidity_one_coin", func() {
				tb.call(aliceKey, *swap, "KillMe")
				_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", AmountZero, uint8(0), AmountZero)
				Expect(err).To(MatchError("Exchange: KILLED"))

			})

			It("test_exchange", func() {

				tb.call(aliceKey, *swap, "KillMe")
				_, err := tb.call(aliceKey, *swap, "Exchange", uint8(0), uint8(0), AmountZero, AmountZero, AddressZero)
				Expect(err).To(MatchError("Exchange: KILLED"))

			})
		})

		Describe("test_modify_fees.py", func() {
			It("test_commit", func() {

				fees := [4][3]uint64{{0, 0, 0}, {23, 42, 18}, {trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}, {1, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}}

				for k := 0; k < len(fees); k++ {
					beforeEachDefault()

					fee, admin_fee, winner_fee := fees[k][0], fees[k][1], fees[k][2]
					delay := uint64(3 * 86400)
					_, err := tb.call(aliceKey, *swap, "CommitNewFee", fee, admin_fee, winner_fee, delay)
					Expect(err).To(Succeed())

					is, err := tb.view(*swap, "AdminActionsDeadline")
					Expect(err).To(Succeed())
					//Expect(is[0].(uint64)).To(Equal(delay))
					Expect(is[0].(uint64)).To(Equal(tb.ctx.LastTimestamp()/uint64(time.Second) - tb.step/1000 + delay))

					is, err = tb.view(*swap, "FutureFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(fee))

					is, err = tb.view(*swap, "FutureAdminFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(admin_fee))

					is, err = tb.view(*swap, "FutureWinnerFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(winner_fee))

				}

			})

			It("test_commit_only_owner", func() {
				beforeEachDefault()

				_, err := tb.call(bobKey, *swap, "CommitNewFee", uint64(23), uint64(42), uint64(18), uint64(3*86400))
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("test_commit_already_active", func() {
				beforeEachDefault()

				_, err := tb.call(aliceKey, *swap, "CommitNewFee", uint64(23), uint64(42), uint64(18), uint64(0))
				tb.sleep(3600 * 1000)
				_, err = tb.call(aliceKey, *swap, "CommitNewFee", uint64(23), uint64(42), uint64(18), uint64(0))
				Expect(err).To(MatchError("Exchange: ADMIN_ACTIONS_DEADLINE"))

			})

			It("test_commit_admin_fee_too_high", func() {
				fees := [2]uint64{2 * trade.MAX_ADMIN_FEE, 3 * trade.MAX_ADMIN_FEE}

				for k := 0; k < len(fees); k++ {
					beforeEachDefault()

					admin_fee := fees[k]

					_, err := tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), admin_fee, uint64(0), uint64(0))
					Expect(err).To(MatchError("Exchange: FUTURE_ADMIN_FEE_EXCEED_MAXADMINFEE"))

				}
			})

			It("test_commit_fee_too_high", func() {
				fees := [2]uint64{2 * trade.MAX_ADMIN_FEE, 3 * trade.MAX_ADMIN_FEE}

				for k := 0; k < len(fees); k++ {
					beforeEachDefault()

					fee := fees[k]

					_, err := tb.call(aliceKey, *swap, "CommitNewFee", fee, uint64(0), uint64(0), uint64(86400))
					Expect(err).To(MatchError("Exchange: FUTURE_FEE_EXCEED_MAXFEE"))

				}
			})

			It("test_commit_winner_fee_too_high", func() {
				fees := [2]uint64{2 * trade.MAX_WINNER_FEE, 3 * trade.MAX_WINNER_FEE}

				for k := 0; k < len(fees); k++ {
					beforeEachDefault()

					winner_fee := fees[k]

					_, err := tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), uint64(0), winner_fee, uint64(86400))
					Expect(err).To(MatchError("Exchange: FUTURE_WINNER_FEE_EXCEED_MAXADMINFEE"))

				}
			})

			It("test_apply", func() {
				fees := [4][3]uint64{{0, 0, 0}, {23, 42, 18}, {trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}, {1, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE}}

				for k := 0; k < len(fees); k++ {
					beforeEachDefault()

					fee := fees[k][0]
					admin_fee := fees[k][1]
					winner_fee := fees[k][2]
					delay := uint64(86400)

					tb.step = delay * 1000
					tb.call(aliceKey, *swap, "CommitNewFee", fee, admin_fee, winner_fee, delay)

					tb.step = _Step
					_, err := tb.call(aliceKey, *swap, "ApplyNewFee")
					Expect(err).To(Succeed())

					is, err := tb.view(*swap, "AdminActionsDeadline")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(uint64(0)))

					is, err = tb.view(*swap, "FutureFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(fee))

					is, err = tb.view(*swap, "FutureAdminFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(admin_fee))

					is, err = tb.view(*swap, "FutureWinnerFee")
					Expect(err).To(Succeed())
					Expect(is[0].(uint64)).To(Equal(winner_fee))

				}
			})

			It("test_apply_only_owner", func() {
				beforeEachDefault()

				delay := uint64(0)
				tb.step = delay * 1000
				tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), uint64(0), uint64(0), delay)

				tb.step = _Step
				_, err := tb.call(bobKey, *swap, "ApplyNewFee")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("test_apply_insufficient_time", func() {
				_DELAY := uint64(3 * 86400)
				delays := [2]uint64{1, _DELAY - 2} // delay can't be zero

				for k := 0; k < len(delays); k++ {
					beforeEachDefault()

					delay := delays[k]

					tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), uint64(0), uint64(0), _DELAY)
					tb.sleep(delay * 1000)

					_, err := tb.call(aliceKey, *swap, "ApplyNewFee")
					Expect(err).To(MatchError("Exchange: ADMIN_ACTIONS_DEADLINE"))

				}
			})

			It("test_apply_no_action", func() {
				beforeEachDefault()

				_, err := tb.call(aliceKey, *swap, "ApplyNewFee")
				Expect(err).To(MatchError("Exchange: NO_ACTIVE_ACTION"))

			})

			It("test_revert", func() {
				beforeEachDefault()

				tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), uint64(0), uint64(0), uint64(3*86400))

				_, err := tb.call(aliceKey, *swap, "RevertNewFee")
				Expect(err).To(Succeed())

			})

			It("test_revert_only_owner", func() {
				beforeEachDefault()

				tb.call(aliceKey, *swap, "CommitNewFee", uint64(0), uint64(0), uint64(0), uint64(3*86400))

				_, err := tb.call(bobKey, *swap, "RevertNewFee")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("test_revert_without_commit", func() {
				beforeEachDefault()

				_, err := tb.call(aliceKey, *swap, "RevertNewFee")
				Expect(err).To(Succeed())

				is, err := tb.call(aliceKey, *swap, "AdminActionsDeadline")
				Expect(err).To(Succeed())
				Expect(is[0].(uint64)).To(Equal(uint64(0)))

			})

			It("test_withdraw_only_owner", func() {
				beforeEachDefault()

				_, err := tb.call(bobKey, *swap, "WithdrawAdminFees")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})
		})

		Describe("test_nonpayable.py", func() {
			It("test_fallback_reverts", func() {
				// no fallback in Meverse
			})
		})

		Describe("test_ramp_A_precise.py", func() {

			BeforeEach(func() {

				beforeEachDefault()
			})

			It("test_ramp_A", func() {
				is, err := tb.view(*swap, "InitialA")
				Expect(err).To(Succeed())
				initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
				lastTimestamp := tb.ctx.LastTimestamp()
				future_time := lastTimestamp + trade.MIN_RAMP_TIME + 5

				_, err = tb.call(aliceKey, *swap, "RampA", MulC(initial_A, 2), future_time)
				Expect(err).To(Succeed())

				is, err = tb.view(*swap, "InitialA")
				Expect(DivC(is[0].(*big.Int), trade.A_PRECISION)).To(Equal(initial_A))

				is, err = tb.view(*swap, "FutureA")
				Expect(DivC(is[0].(*big.Int), trade.A_PRECISION)).To(Equal(MulC(initial_A, 2)))

				is, err = tb.view(*swap, "InitialATime")
				Expect(is[0].(uint64)).To(Equal(lastTimestamp / uint64(time.Second)))

				is, err = tb.view(*swap, "FutureATime")
				Expect(is[0].(uint64)).To(Equal(future_time))

			})

			It("test_ramp_A_final", func() {
				lastTimestamp := tb.ctx.LastTimestamp()

				is, _ := tb.view(*swap, "InitialA")
				initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

				tb.step = 1000000 * 1000
				future_time := lastTimestamp/uint64(time.Second) + 1000000

				tb.call(aliceKey, *swap, "RampA", MulC(initial_A, 2), future_time)

				is, _ = tb.view(*swap, "A")
				Expect(is[0].(*big.Int)).To(Equal(MulC(initial_A, 2)))
			})

			It("test_ramp_A_value_up", func() {
				is, _ := tb.view(*swap, "InitialA")
				initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

				future_time := tb.ctx.LastTimestamp()/uint64(time.Second) + 1000000

				tb.call(aliceKey, *swap, "RampA", MulC(initial_A, 2), future_time)

				initial_time := tb.ctx.LastTimestamp() / uint64(time.Second)
				duration := future_time - initial_time

				for tb.ctx.LastTimestamp()/uint64(time.Second) < future_time {
					tb.sleep(100000)
					elapsed := float64(tb.ctx.LastTimestamp()/uint64(time.Second)-uint64(initial_time)) / float64(duration)
					expected := AddC(initial_A, int64(ToFloat64(initial_A)*elapsed))
					is, err := tb.view(*swap, "A")
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
				is, _ := tb.view(*swap, "InitialA")
				initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)

				future_time := tb.ctx.LastTimestamp()/uint64(time.Second) + 1000000

				tb.call(aliceKey, *swap, "RampA", DivC(initial_A, 10), future_time)

				initial_time := tb.ctx.LastTimestamp() / uint64(time.Second)
				duration := future_time - initial_time

				for tb.ctx.LastTimestamp()/uint64(time.Second) < future_time {
					tb.sleep(100000)
					elapsed := float64(tb.ctx.LastTimestamp()/uint64(time.Second)-uint64(initial_time)) / float64(duration)
					expected := SubC(initial_A, int64(elapsed*ToFloat64(initial_A)/10.*9.))
					is, _ := tb.view(*swap, "A")
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
				lastTimestamp := tb.ctx.LastTimestamp()

				is, _ := tb.view(*swap, "InitialA")
				initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
				future_time := lastTimestamp/uint64(time.Second) + 1000000
				tb.call(aliceKey, *swap, "RampA", DivC(initial_A, 10), future_time)

				tb.sleep(31337 * 1000)

				is, _ = tb.view(*swap, "A")
				current_A := is[0].(*big.Int)

				lastTimestamp = tb.ctx.LastTimestamp()

				_, err = tb.call(aliceKey, *swap, "StopRampA")

				is, err = tb.view(*swap, "InitialA")
				Expect(DivC(is[0].(*big.Int), 100)).To(Equal(current_A))

				is, err = tb.view(*swap, "FutureA")
				Expect(DivC(is[0].(*big.Int), 100)).To(Equal(current_A))

				is, err = tb.view(*swap, "InitialATime")
				Expect(is[0].(uint64)).To(Equal(lastTimestamp / uint64(time.Second)))

				is, err = tb.view(*swap, "FutureATime")
				Expect(is[0].(uint64)).To(Equal(lastTimestamp / uint64(time.Second)))
			})

			It("test_ramp_A_only_owner", func() {
				future_time := tb.ctx.LastTimestamp()/uint64(time.Second) + 1000000
				_, err = tb.call(bobKey, *swap, "RampA", Zero, future_time)
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})

			It("test_ramp_A_insufficient_time", func() {
				future_time := tb.ctx.LastTimestamp()/uint64(time.Second) + trade.MIN_RAMP_TIME - 1
				_, err = tb.call(aliceKey, *swap, "RampA", Zero, future_time)
				Expect(err).To(MatchError("Exchange: Ramp_A_BIG"))

			})

			It("test_stop_ramp_A_only_owner", func() {
				_, err = tb.call(bobKey, *swap, "StopRampA")
				Expect(err).To(MatchError("Exchange: FORBIDDEN"))

			})
		})

		Describe("test_remove_liquidity.py", func() {

			It("non-positive input amount", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				_, err := tb.view(*swap, "CalcWithdrawCoins", ToAmount(big.NewInt(-1)))
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			})

			It("non-positive input amount", func() {
				Skip(AmountNotNegative)

				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", ToAmount(big.NewInt(-1)), MakeAmountSlice(N))
				Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

			})

			It("test_remove_liquidity", func() {
				min_amt := []int{0, 1}

				for k := 0; k < len(min_amt); k++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					min_amts := CloneAmountSlice(_InitialAmounts)
					for i := uint8(0); i < N; i++ {
						min_amts[i].Set(Mul(big.NewInt(int64(min_amt[k])), min_amts[i].Int))
					}
					_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), min_amts)
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice)).To(Equal(_InitialAmounts[i]))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap).Cmp(AmountZero.Int)).To(Equal(0))
					}

					Expect(tokenBalanceOf(tb.ctx, *swap, alice)).To(Equal(AmountZero))
					Expect(tokenTotalSupply(tb.ctx, *swap)).To(Equal(AmountZero))

				}
			})

			It("test_remove_partial", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				withdraw_amount := Div(Sum(ToBigInts(_InitialAmounts)), big.NewInt(2))

				min_amts := MakeAmountSlice(N)
				_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", withdraw_amount, min_amts)
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					pool_balance := tokenBalanceOf(tb.ctx, stableTokens[i], *swap)
					alice_balance := tokenBalanceOf(tb.ctx, stableTokens[i], alice)

					Expect(Add(pool_balance.Int, alice_balance.Int)).To(Equal(_InitialAmounts[i].Int))
				}

				Expect(tokenBalanceOf(tb.ctx, *swap, alice)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(ToAmount(withdraw_amount))))
				Expect(tokenTotalSupply(tb.ctx, *swap)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(ToAmount(withdraw_amount))))

			})

			It("test_below_min_amount", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					min_amts := CloneAmountSlice(_InitialAmounts)
					min_amts[idx] = ToAmount(Add(min_amts[idx].Int, big.NewInt(1)))

					_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0), min_amts)
					Expect(err).To(MatchError("Exchange: WITHDRAWAL_RESULTED_IN_FEWER_COINS_THAN_EXPECTED"))

				}
			})

			It("test_amount_exceeds_balance", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				min_amts := MakeAmountSlice(N)

				_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", amount.NewAmount(uint64(N)*uint64(_BaseAmount), 1), min_amts)
				Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))

			})

			It("test_event", func() {
				// no event
			})
		})

		Describe("test_remove_liquidity_imbalance.py", func() {

			It("non-positive input amount", func() {
				Skip(AmountNotNegative)

				for i := uint8(0); i < N; i++ {
					amts := MakeAmountSlice(N)
					amts[i].Set(big.NewInt(-1))

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, MaxUint256)
					Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				}
			})

			It("test_remove_balanced", func() {
				divisor := []int{2, 5, 10}

				for k := uint8(0); k < N; k++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					amts := MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(Div(_InitialAmounts[i].Int, big.NewInt(int64(divisor[k]))))
					}
					max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).DivC(int64(divisor[k]))
					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, ToAmount(Add(max_burn.Int, big.NewInt(1))))
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice)).To(Equal(amts[i]))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
					}

					alice_pool_token_balance := tokenBalanceOf(tb.ctx, *swap, alice)
					pool_total_supply := tokenTotalSupply(tb.ctx, *swap)

					if !(Abs(alice_pool_token_balance.Sub(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(max_burn)).Int).Cmp(big.NewInt(1)) <= 0) {
						Fail("Lower")
					}

					if !(Abs(pool_total_supply.Sub(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(max_burn)).Int).Cmp(big.NewInt(1)) <= 0) {
						Fail("Upper")
					}

				}
			})

			It("test_remove_some", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					amts := MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(_InitialAmounts[i].DivC(2).Int)
					}
					amts[idx].Set(Zero)

					max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)
					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, max_burn)
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						balance := tokenBalanceOf(tb.ctx, stableTokens[i], alice)
						if balance.Cmp(amts[i].Int) != 0 {
							Fail("Balance")
						}

						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
					}

					actual_balance := tokenBalanceOf(tb.ctx, *swap, alice)
					actual_total_supply := tokenTotalSupply(tb.ctx, *swap)

					ideal_balance := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amount.NewAmount(uint64(_BaseAmount)/2*uint64(N-1), 0))

					Expect(actual_balance).To(Equal(actual_total_supply))

					if ToFloat64(ideal_balance.Int)*0.99 >= ToFloat64(actual_balance.Int) {
						Fail("Lower")
					}
					if actual_balance.Int.Cmp(ideal_balance.Int) >= 0 {
						Fail("Upper")
					}

				}
			})

			It("test_remove_one", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					amts := MakeAmountSlice(N)
					amts[idx].Set(_InitialAmounts[idx].DivC(2).Int)

					max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0)
					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, max_burn)
					Expect(err).To(Succeed())

					for i := uint8(0); i < N; i++ {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice).Cmp(amts[i].Int)).To(Equal(0))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], *swap)).To(Equal(_InitialAmounts[i].Sub(amts[i])))
					}

					actual_balance := tokenBalanceOf(tb.ctx, *swap, alice)
					actual_total_supply := tokenTotalSupply(tb.ctx, *swap)

					ideal_balance := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amount.NewAmount(uint64(_BaseAmount)/2, 0))

					Expect(actual_balance).To(Equal(actual_total_supply))
					if ToFloat64(ideal_balance.Int)*0.99 >= ToFloat64(actual_balance.Int) {
						Fail("Lower")
					}
					if actual_balance.Int.Cmp(ideal_balance.Int) >= 0 {
						Fail("Upper")
					}

				}
			})

			It("test_exceed_max_burn", func() {
				divisor := []int{1, 2, 10}

				for k := uint8(0); k < N; k++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					amts := MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(Div(_InitialAmounts[i].Int, big.NewInt(int64(divisor[k]))))
					}
					max_burn := amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).DivC(int64(divisor[k]))

					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, ToAmount(SubC(max_burn.Int, 1)))
					Expect(err).To(MatchError("Exchange: SLIPPAGE"))

				}
			})

			It("test_cannot_remove_zero", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				amts := MakeAmountSlice(N)

				_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, AmountZero)
				Expect(err).To(MatchError("Exchange: ZERO_TOKEN_BURN"))

			})

			It("test_no_totalsupply", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				amts := MakeAmountSlice(N)
				total_supply := tokenTotalSupply(tb.ctx, *swap)

				_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", total_supply, amts)
				Expect(err).To(Succeed())

				_, err = tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, AmountZero)
				Expect(err).To(MatchError("Exchange: D0_IS_ZERO"))

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

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.view(*swap, "CalcWithdrawOneCoin", ToAmount(big.NewInt(-1)), uint8(1))
					Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				}
			})

			It("RemoveLiquidityOneCoin non-positive input amount", func() {
				Skip(AmountNotNegative)

				for i := uint8(0); i < N; i++ {
					amts := MakeAmountSlice(N)
					amts[i].Set(big.NewInt(-1))

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", ToAmount(big.NewInt(-1)), uint8(1), AmountZero)
					Expect(err).To(MatchError("Exchange: INSUFFICIENT_INPUT"))

				}
			})

			It("CalcWithdrawOneCoin index error", func() {
				for i := uint8(0); i < N; i++ {
					amts := MakeAmountSlice(N)
					amts[i].Set(big.NewInt(-1))

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.view(*swap, "CalcWithdrawOneCoin", big.NewInt(1), uint8(N))
					Expect(err).To(MatchError("Exchange: IDX"))

				}
			})

			It("RemoveLiquidityOneCoin index error", func() {
				for i := uint8(0); i < N; i++ {
					amts := MakeAmountSlice(N)
					amts[i].Set(big.NewInt(-1))

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", big.NewInt(1), uint8(N), AmountZero)
					Expect(err).To(MatchError("Exchange: IDX"))

				}
			})

			It("test_amount_received", func() {

				for idx := uint8(0); idx < N; idx++ {
					beforeEach(0, 0, 0, false)
					stableAddLiquidityDefault(aliceKey)

					rate_mod := 1.00001

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amount.NewAmount(1, 0), idx, AmountZero)
					Expect(err).To(Succeed())

					balance := tokenBalanceOf(tb.ctx, stableTokens[idx], alice)

					if big.NewInt(int64(ToFloat64(Pow10(decimals[idx]))/rate_mod)).Cmp(balance.Int) > 0 {
						Fail("Lower")
					}

					if balance.Int.Cmp(Pow10(decimals[idx])) > 0 {
						Fail("Upper")
					}

				}
			})

			It("test_lp_token_balance", func() {
				divisors := []int{1, 5, 42}
				for k := 0; k < len(divisors); k++ {
					divisor := divisors[k]

					for idx := uint8(0); idx < N; idx++ {
						beforeEachDefault()
						stableAddLiquidityDefault(aliceKey)

						balance := tokenBalanceOf(tb.ctx, *swap, alice)
						amt := balance.DivC(int64(divisor))

						_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
						Expect(err).To(Succeed())

						balance = tokenBalanceOf(tb.ctx, *swap, alice)
						if balance.Cmp(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amt).Int) != 0 {
							Fail("Balance")
						}

					}
				}
			})

			It("test_expected_vs_actual", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					balance := tokenBalanceOf(tb.ctx, *swap, alice)
					amt := balance.DivC(int64(10))

					is, err := tb.view(*swap, "CalcWithdrawOneCoin", amt, idx)
					Expect(err).To(Succeed())
					expected := is[0].(*amount.Amount)

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
					Expect(err).To(Succeed())

					Expect(tokenBalanceOf(tb.ctx, stableTokens[idx], alice)).To(Equal(expected))
					Expect(tokenBalanceOf(tb.ctx, *swap, alice)).To(Equal(amount.NewAmount(uint64(N)*uint64(_BaseAmount), 0).Sub(amt)))

				}
			})

			It("test_below_min_amount", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					amt := tokenBalanceOf(tb.ctx, *swap, alice)

					is, err := tb.view(*swap, "CalcWithdrawOneCoin", amt, idx)
					Expect(err).To(Succeed())

					expected := is[0].(*amount.Amount)

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AddC(expected.Int, 1))
					Expect(err).To(MatchError("Exchange: INSUFFICIENT_OUTPUT_AMOUNT"))

				}
			})

			It("test_amount_exceeds_balance", func() {
				for idx := uint8(0); idx < N; idx++ {
					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)

					_, err := tb.call(bobKey, *swap, "RemoveLiquidityOneCoin", amount.NewAmount(0, 1), idx, AmountZero)
					Expect(err).To(HaveOccurred())

				}
			})

			It("test_below_zero", func() {
				//uint8 can't be below zero
			})

			It("test_above_n_coins", func() {
				beforeEachDefault()
				stableAddLiquidityDefault(aliceKey)

				_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amount.NewAmount(0, 1), N, AmountZero)
				Expect(err).To(MatchError("Exchange: IDX"))

			})

			It("test_event", func() {
				// no event
			})
		})

		Describe("test_xfer_to_contract.py", func() {

			// tx를 보낼때는 동일 주소 생성 불가
			// It("test_unexpected_eth", func() {
			// 	//동일한 주소를 갖는 *swap contract를 다시 만들었을 경우 그 값이 보존되는 지

			// 	genesis = types.NewEmptyContext()

			// 	stableTokens = DeployTokens(genesis, classMap["Token"], N, admin)

			// 	var err error
			// 	sbc := stableBaseContruction()
			// 	swap, err = stablebase(genesis, sbc)

			// 	bs, _, err := bin.WriterToBytes(sbc)
			// 	Expect(err).To(Succeed())

			// 	stableAddLiquidityDefault(genesis, alice)

			// 	is, _ := Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			// 	virtual_price := is[0].(*amount.Amount)

			// 	cn, cdx, ctx, _ := initChain(genesis, admin)
			// 	timestamp_second := uint64(time.Now().UnixNano()) / uint64(time.Second)
			// 	ctx, _ = Sleep(cn, ctx, nil, timestamp_second, aliceKey)

			// 	// Deploy with Same Address
			// 	v, err := ctx.DeployContractWithAddress(bob, classMap["StableSwap"], swap, bs)
			// 	swap2 := v.(*trade.StableSwap).Address() // swap2 == swap

			// 	Expect(swap).To(Equal(swap2))

			// 	is, _ = Exec(genesis, alice, swap, "GetVirtualPrice", []interface{}{})
			// 	virtual_price2 := is[0].(*amount.Amount)

			// 	Expect(virtual_price2).To(Equal(virtual_price))

			// 	admin_balances, _ := stableGetAdminBalances(ctx)
			// 	for i := uint8(0); i < N; i++ {
			// 		is, err := Exec(ctx, bob, swap, "AdminBalances", []interface{}{uint8(i)})
			// 		Expect(err).To(Succeed())
			// 		if admin_balances[i].Cmp(Zero) == 0 {
			// 			Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
			// 		} else {
			// 			Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i].Int))
			// 		}
			// 	}
			// 	Expect(Sum(ToBigInts(admin_balances))).To(Equal(Zero))

			// 	RemoveChain(cdx)
			// 	afterEach()
			// })

			It("test_unexpected_coin", func() {
				beforeEachDefault()

				stableAddLiquidityDefault(aliceKey)

				is, _ := tb.view(*swap, "GetVirtualPrice")
				virtual_price := is[0].(*amount.Amount)

				tb.call(aliceKey, stableTokens[N-1], "Mint", *swap, amount.NewAmount(0, 123456))

				is, _ = tb.view(*swap, "GetVirtualPrice")
				virtual_price2 := is[0].(*amount.Amount)

				Expect(virtual_price).To(Equal(virtual_price2))

				admin_balances, _ := stableGetAdminBalances()
				for i := uint8(0); i < N; i++ {
					is, err := tb.view(*swap, "AdminBalances", uint8(i))
					Expect(err).To(Succeed())
					if admin_balances[i].Cmp(Zero) == 0 {
						Expect(is[0].(*amount.Amount).Cmp(Zero) == 0).To(BeTrue())
					} else {
						Expect(is[0].(*amount.Amount)).To(Equal(admin_balances[i]))
					}
				}
				Expect(Sum(ToBigInts(admin_balances))).To(Equal(big.NewInt(123456)))

			})
		})
	})
	Describe("Forked", func() {

		Describe("test_gas.py", func() {
			BeforeEach(func() {
				beforeEachDefault()
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

				stableMint(alice)
				stableApprove(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				setFees(tb, aliceKey, *swap, uint64(4000000), uint64(5000000000), uint64(5000000000))

				amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(DivC(amts[i].Int, 2))
				}
				_, err := tb.call(aliceKey, *swap, "AddLiquidity", amts, AmountZero)
				Expect(err).To(Succeed())

				balances := stableTokenBalances(alice)
				log.Println("alice balances 1", balances)

				//# add liquidity imbalanced
				for idx := uint8(0); idx < N; idx++ {
					amts := CloneAmountSlice(_InitialAmounts)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(DivC(amts[i].Int, 10))
					}
					amts[idx] = amount.NewAmount(0, 0)

					_, err := tb.call(aliceKey, *swap, "AddLiquidity", amts, AmountZero)
					Expect(err).To(Succeed())
				}

				balances = stableTokenBalances(alice)
				log.Println("alice balances 2", balances)

				balances = stableTokenBalances(bob)
				log.Println("bob balances 2", balances)

				//# perform swaps between each coin
				for send := uint8(0); send < N; send++ {
					for recv := uint8(0); recv < N; recv++ {
						if send == recv {
							continue
						}

						amt := Pow10(decimals[send])
						_, err := tb.call(aliceKey, stableTokens[send], "Mint", bob, ToAmount(AddC(amt, 1)))
						Expect(err).To(Succeed())

						recv_balance := tokenBalanceOf(tb.ctx, stableTokens[recv], bob)
						if recv_balance.Cmp(Zero) > 0 {
							_, err = tb.call(bobKey, stableTokens[recv], "Transfer", alice, recv_balance)
							Expect(err).To(Succeed())
						}
						log.Println("send, recv, amt in exchange", send, recv, amt)

						_, err = tb.call(bobKey, *swap, "Exchange", send, recv, ToAmount(amt), AmountZero, AddressZero)
						Expect(err).To(Succeed())
					}
				}

				balances = stableTokenBalances(alice)
				log.Println("alice balances 3", balances)

				balances = stableTokenBalances(bob)
				log.Println("bob balances 3", balances)

				//# remove liquidity balanced
				_, err = tb.call(aliceKey, *swap, "RemoveLiquidity", amount.NewAmount(1, 0), MakeAmountSlice(N))
				Expect(err).To(Succeed())

				balances = stableTokenBalances(alice)
				log.Println("alice balances 5", balances)

				amts = MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(Pow10(decimals[i]))
				}

				_, err = tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, MaxUint256)
				Expect(err).To(Succeed())

				balances = stableTokenBalances(alice)
				log.Println("alice balances 6", balances)

				//# remove liquidity balanced
				for idx := uint8(0); idx < N; idx++ {
					amts = MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(Pow10(decimals[i]))
					}
					amts[idx].Set(Zero)

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, MaxUint256)
					Expect(err).To(Succeed())
				}

				balances = stableTokenBalances(alice)
				log.Println("alice balances 7", balances)

				//# remove_liquidity_one_coin
				for idx := uint8(0); idx < N; idx++ {
					amt := ToAmount(Pow10(decimals[idx]))

					_, err = tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
					if err != nil {
						Fail(strconv.Itoa(int(idx)) + " : " + err.Error())
					}
				}

				balances = stableTokenBalances(alice)
				log.Println("alice balances 8", balances)

				//	mainToken 가스계산
				mToken := *tb.ctx.MainToken()
				balance := tokenBalanceOf(tb.ctx, mToken, alice)
				log.Println("gas-used of alice", amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance))

				balance = tokenBalanceOf(tb.ctx, mToken, bob)
				log.Println("gas-used of bob", amount.NewAmount(uint64(_BaseAmount), 0).Sub(balance))
			})

			It("test_zap_gas", func() {
				// no zap
			})
		})

		Describe("test_insufficient_balances.py", func() {

			It("test_swap_gas", func() {
				beforeEach(uint64(4000000), uint64(5000000000), uint64(5000000000), false)
				stableMint(alice)
				stableApprove(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				//# attempt to deposit more funds than user has
				for idx := uint8(0); idx < N; idx++ {
					amts := CloneAmountSlice(_InitialAmounts)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(DivC(amts[i].Int, 2))
					}

					balance := tokenBalanceOf(tb.ctx, stableTokens[idx], alice)
					amts[idx] = balance.Add(amount.NewAmount(0, 1))

					_, err = tb.call(aliceKey, *swap, "AddLiquidity", amts, AmountZero)
					insufficientBalanceErrorCheck(err)
				}
				//# add liquidity balanced
				amts := CloneAmountSlice(_InitialAmounts)
				for i := uint8(0); i < N; i++ {
					amts[i].Set(DivC(amts[i].Int, 2))
				}
				_, err = tb.call(aliceKey, *swap, "AddLiquidity", amts, AmountZero)
				Expect(err).To(Succeed())

				//# attempt to perform swaps between stableTokens with insufficient funds
				for send := uint8(0); send < N; send++ {
					for recv := uint8(0); recv < N; recv++ {
						if send == recv {
							continue
						}

						amt := Clone(DivC(_InitialAmounts[send].Int, 4))

						_, err = tb.call(charlieKey, *swap, "Exchange", send, recv, ToAmount(amt), AmountZero, AddressZero)
						insufficientBalanceErrorCheck(err)
					}
				}

				//# remove liquidity balanced
				_, err = tb.call(charlieKey, *swap, "RemoveLiquidity", amount.NewAmount(1, 0), MakeAmountSlice(N))
				Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))

				//# remove liquidity imbalanced
				for idx := uint8(0); idx < N; idx++ {
					amts := MakeAmountSlice(N)
					for i := uint8(0); i < N; i++ {
						amts[i].Set(Pow10(decimals[i]))
					}

					balance := tokenBalanceOf(tb.ctx, stableTokens[idx], *swap)
					amts[idx].Set(AddC(balance.Int, 1))
					_, err = tb.call(charlieKey, *swap, "RemoveLiquidityImbalance", amts, MaxUint256)
					Expect(err).To(HaveOccurred())
				}

				//# remove_liquidity_one_coin
				for idx := uint8(0); idx < N; idx++ {
					_, err = tb.call(charlieKey, *swap, "RemoveLiquidityOneCoin", ToAmount(Pow10(decimals[idx])), idx, AmountZero)
					Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))
				}
			})
		})
	})

	Describe("Integration", func() {

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
					beforeEach(0, 0, 0, false)

					// add initial pool liquidity
					// we add at an imbalance of +10% for each subsequent coin
					initial_liquidity := []*amount.Amount{}

					for i := uint8(0); i < N; i++ {
						amount := ToAmount(Add(Mul(st_seed_amount[k], big.NewInt(int64(math.Pow10(decimals[i])))), big.NewInt(1)))

						_, err := tb.call(aliceKey, stableTokens[i], "Mint", alice, amount)
						Expect(err).To(Succeed())

						balance := tokenBalanceOf(tb.ctx, stableTokens[i], alice)

						if balance.Int.Cmp(amount.Int) < 0 {
							Fail("test_curve_in_contract : Balance")
						}

						initial_liquidity = append(initial_liquidity, amount.DivC(10))
						_, err = tb.call(aliceKey, stableTokens[i], "Approve", *swap, amount.DivC(10))
						Expect(err).To(Succeed())

					}

					_, err := tb.call(aliceKey, *swap, "AddLiquidity", initial_liquidity, AmountZero)
					Expect(err).To(Succeed())

					is, err := tb.view(*swap, "Reserves")
					Expect(err).To(Succeed())
					balances := is[0].([]*amount.Amount)

					is, err = tb.view(*swap, "Rates")
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

						is, err = tb.view(*swap, "GetDy", uint8(send), uint8(recv), ToAmount(dx), AddressZero)
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

				}
			})
		})

		Describe("test_heavily_imbalanced.py", func() {

			It("test_imbalanced_swaps", func() {
				for idx := uint8(0); idx < N; idx++ {

					beforeEachDefault()
					stableAddLiquidityDefault(aliceKey)
					stableMint(bob)
					stableApprove(bobKey)

					amounts := MakeAmountSlice(N)
					amounts[idx].Set(_InitialAmounts[idx].MulC(1000).Int)

					tb.call(aliceKey, stableTokens[idx], "Mint", alice, amounts[idx])
					tb.call(aliceKey, *swap, "AddLiquidity", amounts, AmountZero)

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
						_, err := tb.call(bobKey, *swap, "Exchange", idx, m, _InitialAmounts[idx].DivC(int64(N)), AmountZero, AddressZero)
						Expect(err).To(Succeed())
					}

					//# now we go the other direction, swaps where the output asset is the imbalanced one
					//# lucky bob is about to get rich!
					for k := 0; k < len(swap_indexes); k++ {
						m := uint8(swap_indexes[k])
						_, err := tb.call(bobKey, *swap, "Exchange", m, idx, _InitialAmounts[m].DivC(int64(N)), AmountZero, AddressZero)
						Expect(err).To(Succeed())
					}
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

							beforeEach(uint64(10000000), uint64(0), uint64(0), false)
							stableAddLiquidityDefault(aliceKey)
							is, _ := tb.view(*swap, "GetVirtualPrice")
							virtual_price := is[0].(*amount.Amount)

							for i := uint8(0); i < N; i++ {
								amt := amount.NewAmount(uint64(_BaseAmount), 0)
								tb.call(aliceKey, stableTokens[i], "Mint", alice, amt)
							}

							r := rand.Intn(5)

							if r == 0 {
								// rule_ramp_A
								is, _ := tb.view(*swap, "A")
								new_A := is[0].(*big.Int)
								timestamp := tb.ctx.LastTimestamp()/uint64(time.Second) + 86410
								_, err := tb.call(aliceKey, *swap, "RampA", new_A, timestamp)
								Expect(err).To(Succeed())
							} else if r == 1 {
								// rule_increase_rates
								//     not existent

								// rule_exchange

								send, recv, _, _ := _min_max()
								amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[send])) * st_pct[m])))
								_, err := tb.call(aliceKey, *swap, "Exchange", uint8(send), recv, amt, AmountZero, AddressZero)
								Expect(err).To(Succeed())
							} else if r == 2 {
								// rule_exchange_underlying
								//     not existent

								// rule_remove_one_coin

								_, idx, _, _ := _min_max()
								amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[idx])) * st_pct[m])))

								_, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
								Expect(err).To(Succeed())
							} else if r == 3 {
								// rule_remove_imbalance
								_, idx, _, _ := _min_max()
								amts := MakeAmountSlice(N)
								amts[idx] = ToAmount(big.NewInt(int64(ToFloat64(Pow10(decimals[idx])) * st_pct[m])))
								_, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, MaxUint256)
								Expect(err).To(Succeed())
							} else if r == 4 {
								// rule_remove
								amt := ToAmount(big.NewInt(int64(ToFloat64(Pow10(18)) * st_pct[m])))
								_, err := tb.call(aliceKey, *swap, "RemoveLiquidity", amt, MakeAmountSlice(N))
								Expect(err).To(Succeed())
							}

							// invariant_check_virtual_price
							is, _ = tb.call(aliceKey, *swap, "GetVirtualPrice")
							virtual_price2 := is[0].(*amount.Amount)
							if virtual_price2.Cmp(virtual_price.Int) < 0 {
								Fail("test_number_always_go_up")
							}
							virtual_price.Set(virtual_price.Int)

							// invariant_advance_time
							tb.sleep(3600 * 1000)
						}
					}
					max_examples--
				}
			})
		})
	})
	Describe("WhiteList", func() {

		// whitelist = eve

		BeforeEach(func() {
			beforeEachDefault()
		})

		It("FeeWhiteList, FeeAddress : not WhiteList", func() {

			is, err := tb.view(*swap, "FeeWhiteList", alice)
			Expect(err).To(Succeed())
			Expect(is[0].([]byte)).To(Equal([]byte{}))

			is, err = tb.view(*swap, "FeeAddress", alice)
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(_Fee))

		})

		It("FeeWhiteList, FeeAddress : WhiteList", func() {
			fee := uint64(0) // 0%

			is, err := tb.view(*swap, "FeeWhiteList", eve)
			Expect(err).To(Succeed())
			Expect(is[0].([]byte)).To(Equal(bin.Uint64Bytes(fee)))

			is, err = tb.view(*swap, "FeeAddress", eve)
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

		})

		Describe("AddLiquidity", func() {
			initialAmounts := MakeAmountSlice(N)
			initialAmounts[0].Set(Mul(Exp(big.NewInt(10), big.NewInt(int64(decimals[0]))), big.NewInt(1)))

			expectedLPAmount := amount.NewAmount(0, 998001999537911375)   // fee = 0.4%
			wlExpectedLPAmount := amount.NewAmount(0, 999999999537679406) // fee = 0.0%

			BeforeEach(func() {
				stableAddLiquidityDefault(aliceKey)

			})

			It("CalcLPTokenAmount", func() {
				is, err := tb.view(*swap, "CalcLPTokenAmount", initialAmounts, true)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

				is, err = tb.viewFrom(eve, *swap, "CalcLPTokenAmount", initialAmounts, true)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})

			It("AddLiquidity : not WhiteList", func() {

				stableMint(bob)
				stableApprove(bobKey)

				is, err := tb.call(bobKey, *swap, "AddLiquidity", initialAmounts, AmountZero)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))
			})

			It("AddLiquidity : WhiteList", func() {

				stableMint(eve)
				stableApprove(eveKey)

				is, err := tb.call(eveKey, *swap, "AddLiquidity", initialAmounts, AmountZero)
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
				stableAddLiquidityDefault(aliceKey)
			})

			It("Output", func() {
				Expect(expectedOutput.Cmp(wlExpectedOutput.Int) < 0).To(BeTrue())
			})

			It("Exchange, GetDy : not WhiteList fee = pair.fee(cc)", func() {

				stableApprove(bobKey)
				tb.call(aliceKey, stableTokens[send], "Mint", bob, amt)

				is, err := tb.view(*swap, "GetDy", send, recv, amt, AddressZero)
				Expect(err).To(Succeed())
				received := is[0].(*amount.Amount)
				Expect(received).To(Equal(expectedOutput))
				//GPrintln("received", received)

				_, err = tb.call(bobKey, *swap, "Exchange", send, recv, amt, AmountZero, AddressZero)
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(tb.ctx, stableTokens[send], bob).Cmp(AmountZero.Int)).To(Equal(0))
				Expect(tokenBalanceOf(tb.ctx, stableTokens[recv], bob)).To(Equal(expectedOutput))

			})

			It("Exchange, GetDy : WhiteList from = AddressZero fee = feeWhiteList ", func() {
				stableApprove(eveKey)
				tb.call(aliceKey, stableTokens[send], "Mint", eve, amt)

				is, err := tb.viewFrom(eve, *swap, "GetDy", send, recv, amt, AddressZero)
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutput))

				_, err = tb.call(eveKey, *swap, "Exchange", send, recv, amt, AmountZero, AddressZero)
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(tb.ctx, stableTokens[send], eve).Cmp(AmountZero.Int)).To(Equal(0))
				received := tokenBalanceOf(tb.ctx, stableTokens[recv], eve)
				Expect(received).To(Equal(wlExpectedOutput))
			})

			It("Exchange, GetDy : WhiteList from = whitelist fee = feeWhiteList", func() {
				stableApprove(eveKey)
				tb.call(aliceKey, stableTokens[send], "Mint", eve, amt)

				is, err := tb.view(*swap, "GetDy", send, recv, amt, eve)
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutput))

				_, err = tb.call(eveKey, *swap, "Exchange", send, recv, amt, AmountZero, eve)
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(tb.ctx, stableTokens[send], eve).Cmp(AmountZero.Int)).To(Equal(0))
				received := tokenBalanceOf(tb.ctx, stableTokens[recv], eve)
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
				stableAddLiquidityDefault(aliceKey)

				is, err := tb.view(*swap, "CalcLPTokenAmount", amts, false)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))

			})

			It("CalcLPTokenAmount : Whitelist", func() {
				stableAddLiquidityDefault(eveKey)

				is, err := tb.viewFrom(eve, *swap, "CalcLPTokenAmount", amts, false)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedLPAmount))
			})

			It("RemoveLiquidityImbalance : not WhiteList", func() {
				stableAddLiquidityDefault(aliceKey)

				is, err := tb.call(aliceKey, *swap, "RemoveLiquidityImbalance", amts, max_burn)
				Expect(err).To(Succeed())

				Expect(is[0].(*amount.Amount)).To(Equal(expectedLPAmount))
			})

			It("RemoveLiquidityImbalance : WhiteList", func() {
				stableAddLiquidityDefault(eveKey)

				is, err := tb.call(eveKey, *swap, "RemoveLiquidityImbalance", amts, max_burn)
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

				stableAddLiquidityDefault(aliceKey)
				is, err := tb.view(*swap, "CalcWithdrawOneCoin", amt, idx)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
			})

			It("RemoveLiquidityOneCoin : not WhiteList", func() {

				stableAddLiquidityDefault(aliceKey)

				is, err := tb.call(aliceKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(expectedOutputAmount))
			})

			It("CalcWithdrawOneCoin : WhiteList", func() {

				stableAddLiquidityDefault(eveKey)
				is, err := tb.viewFrom(eve, *swap, "CalcWithdrawOneCoin", amt, idx)
				Expect(err).To(Succeed())
				Expect(is[0].(*amount.Amount)).To(Equal(wlExpectedOutputAmount))
			})

			It("RemoveLiquidityOneCoin : WhiteList", func() {
				stableAddLiquidityDefault(eveKey)

				is, err := tb.call(eveKey, *swap, "RemoveLiquidityOneCoin", amt, idx, AmountZero)
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
					Factory:      AddressZero,
					NTokens:      uint8(N),
					Tokens:       stableTokens,
					PayToken:     stableTokens[pi],
					Owner:        alice,
					Winner:       charlie,
					Fee:          _Fee,
					AdminFee:     _AdminFee,
					WinnerFee:    _WinnerFee,
					WhiteList:    *whiteList,
					GroupId:      _GroupId,
					Amp:          big.NewInt(_Amp),
					PrecisionMul: _PrecisionMul,
					Rates:        _Rates,
				}
				swap, err = swapDeploy(tb, aliceKey, sbc)
				Expect(err).To(Succeed())

				stableAddLiquidityDefault(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				// 0,1 -> 1,2 -> 2,0
				for i := uint8(0); i < N; i++ {
					send := i
					recv := i + 1
					if recv == N {
						recv = 0
					}
					_, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
					Expect(err).To(Succeed())
				}

				admin_balances := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					is, err := tb.call(bobKey, *swap, "AdminBalances", uint8(i))
					Expect(err).To(Succeed())
					admin_balances[i].Set(is[0].(*amount.Amount).Int)
				}

				Expect(admin_balances).To(Equal(expectedAdminBalances))

				_, err := tb.call(aliceKey, *swap, "WithdrawAdminFees")
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					if i == pi {
						wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(_WinnerFee)), trade.FEE_DENOMINATOR))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], charlie)).To(Equal(wF))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice)).To(Equal(admin_balances[i].Sub(wF)))
					} else {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], charlie)).To(Equal(AmountZero))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], alice)).To(Equal(AmountZero))
					}
				}
			})

			// owner whitelist(eve) 로 변경
			It("WhiteList", func() {
				sbc := &trade.StableSwapConstruction{
					Name:         _SwapName,
					Symbol:       _SwapSymbol,
					Factory:      AddressZero,
					NTokens:      uint8(N),
					Tokens:       stableTokens,
					PayToken:     stableTokens[pi],
					Owner:        eve,
					Winner:       charlie,
					Fee:          _Fee,
					AdminFee:     _AdminFee,
					WinnerFee:    _WinnerFee,
					WhiteList:    *whiteList,
					GroupId:      _GroupId,
					Amp:          big.NewInt(_Amp),
					PrecisionMul: _PrecisionMul,
					Rates:        _Rates,
				}

				swap, err = swapDeploy(tb, aliceKey, sbc)
				Expect(err).To(Succeed())

				stableAddLiquidityDefault(aliceKey)
				stableMint(bob)
				stableApprove(bobKey)

				// 0,1 -> 1,2 -> 2,0
				for i := uint8(0); i < N; i++ {
					send := i
					recv := i + 1
					if recv == N {
						recv = 0
					}
					_, err := tb.call(bobKey, *swap, "Exchange", send, recv, _InitialAmounts[send], AmountZero, AddressZero)
					Expect(err).To(Succeed())
				}

				admin_balances := MakeAmountSlice(N)
				for i := uint8(0); i < N; i++ {
					is, err := tb.call(bobKey, *swap, "AdminBalances", uint8(i))
					Expect(err).To(Succeed())
					admin_balances[i].Set(is[0].(*amount.Amount).Int)
				}
				Expect(admin_balances).To(Equal(wlExpectedAdminBalances))

				_, err := tb.call(eveKey, *swap, "WithdrawAdminFees")
				Expect(err).To(Succeed())

				for i := uint8(0); i < N; i++ {
					if i == pi {
						wF := ToAmount(MulDivC(admin_balances[i].Int, big.NewInt(int64(_WinnerFee)), trade.FEE_DENOMINATOR))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], charlie)).To(Equal(wF))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], eve)).To(Equal(admin_balances[i].Sub(wF)))
					} else {
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], charlie)).To(Equal(AmountZero))
						Expect(tokenBalanceOf(tb.ctx, stableTokens[i], eve)).To(Equal(AmountZero))
					}
				}
			})
		})

	})

})
