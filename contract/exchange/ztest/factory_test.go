package test

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/exchange/trade"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("factory", func() {

	BeforeEach(func() {
		beforeEach()
	})

	AfterEach(func() {
		afterEach()
	})

	It("Owner, allPairsLength", func() {
		//Owner
		is, err := Exec(genesis, admin, factoryAddr, "Owner", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0]).To(Equal(admin))

		//AllPairsLength
		is, err = Exec(genesis, admin, factoryAddr, "AllPairsLength", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(uint16)).To(Equal(uint16(0)))
	})

	It("CreatePairUni : PayToken Error", func() {
		// CreatePairUni
		_, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], alice, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))
	})

	It("CreatePairUni", func() {

		// CreatePairUni
		is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(Succeed())
		pair, err := trade.PairFor(factoryAddr, uniTokens[0], uniTokens[1])
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// AllPairs
		is, err = Exec(genesis, admin, factoryAddr, "AllPairs", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)[0]).To(Equal(pair))

		// AllPairsLength
		is, err = Exec(genesis, admin, factoryAddr, "AllPairsLength", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(uint16)).To(Equal(uint16(1)))

		// GetPair of tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[0], uniTokens[1]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// GetPair of reverse tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[1], uniTokens[0]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// CreatePairUni of same tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// CreatePairUni of same reverse tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// factory
		is, err = Exec(genesis, alice, pair, "Factory", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(factoryAddr))

		// token0
		is, err = Exec(genesis, alice, pair, "Token0", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(token0))

		// token1
		is, err = Exec(genesis, alice, pair, "Token1", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(token1))

		// tokens
		is, err = Exec(genesis, alice, pair, "Tokens", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

		// WhiteList
		is, err = Exec(genesis, alice, pair, "WhiteList", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(_WhiteList))

		// GroupId
		is, err = Exec(genesis, alice, pair, "GroupId", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
	})

	It("CreatePairUni : reverse", func() {

		// CreatePairUni
		is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(Succeed())
		pair, err := trade.PairFor(factoryAddr, uniTokens[1], uniTokens[0])
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// AllPairs
		is, err = Exec(genesis, admin, factoryAddr, "AllPairs", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)[0]).To(Equal(pair))

		// AllPairsLength
		is, err = Exec(genesis, admin, factoryAddr, "AllPairsLength", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(uint16)).To(Equal(uint16(1)))

		// GetPair of tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[1], uniTokens[0]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// GetPair of reverse tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[0], uniTokens[1]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// CreatePairUni of same tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// CreatePairUni of same reverse tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// factory
		is, err = Exec(genesis, alice, pair, "Factory", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0]).To(Equal(factoryAddr))

		// token0
		is, err = Exec(genesis, alice, pair, "Token0", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(token0))

		// token1
		is, err = Exec(genesis, alice, pair, "Token1", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(token1))

		// tokens
		is, err = Exec(genesis, alice, pair, "Tokens", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)).To(Equal([]common.Address{token0, token1}))

		// WhiteList
		is, err = Exec(genesis, alice, pair, "WhiteList", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(_WhiteList))

		// GroupId
		is, err = Exec(genesis, alice, pair, "GroupId", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
	})

	It("CreatePairStable", func() {

		// CreatePairUni
		is, err := Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})

		Expect(err).To(Succeed())
		pair, err := trade.PairFor(factoryAddr, uniTokens[0], uniTokens[1])
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// AllPairs
		is, err = Exec(genesis, admin, factoryAddr, "AllPairs", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)[0]).To(Equal(pair))

		// AllPairsLength
		is, err = Exec(genesis, admin, factoryAddr, "AllPairsLength", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(uint16)).To(Equal(uint16(1)))

		// GetPair of tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[0], uniTokens[1]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// GetPair of reverse tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[1], uniTokens[0]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// CreatePairUni of same tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// CreatePairUni of same reverse tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// factory
		is, err = Exec(genesis, alice, pair, "Factory", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(factoryAddr))

		// tokens
		is, err = Exec(genesis, alice, pair, "Tokens", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)).To(Equal([]common.Address{uniTokens[0], uniTokens[1]}))

		// WhiteList
		is, err = Exec(genesis, alice, pair, "WhiteList", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(_WhiteList))

		// GroupId
		is, err = Exec(genesis, alice, pair, "GroupId", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
	})

	It("CreatePairStable : reverse", func() {

		// CreatePairUni
		is, err := Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _SwapName, _SwapSymbol, bob, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})

		Expect(err).To(Succeed())
		pair, err := trade.PairFor(factoryAddr, uniTokens[0], uniTokens[1])
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// AllPairs
		is, err = Exec(genesis, admin, factoryAddr, "AllPairs", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)[0]).To(Equal(pair))

		// AllPairsLength
		is, err = Exec(genesis, admin, factoryAddr, "AllPairsLength", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(uint16)).To(Equal(uint16(1)))

		// GetPair of tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[0], uniTokens[1]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// GetPair of reverse tokens
		is, err = Exec(genesis, admin, factoryAddr, "GetPair", []interface{}{uniTokens[1], uniTokens[0]})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(pair))

		// CreatePairUni of same tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[1], uniTokens[0], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// CreatePairUni of same reverse tokens
		_, err = Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, uint64(_Amp), classMap["StableSwap"]})
		Expect(err).To(MatchError("Exchange: PAIR_EXISTS"))

		// factory
		is, err = Exec(genesis, alice, pair, "Factory", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(factoryAddr))

		// tokens
		is, err = Exec(genesis, alice, pair, "Tokens", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].([]common.Address)).To(Equal([]common.Address{uniTokens[1], uniTokens[0]}))

		// WhiteList
		is, err = Exec(genesis, alice, pair, "WhiteList", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(_WhiteList))

		// GroupId
		is, err = Exec(genesis, alice, pair, "GroupId", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))

	})

	It("setOwner", func() {

		_, err := Exec(genesis, alice, factoryAddr, "SetOwner", []interface{}{alice})
		Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		_, err = Exec(genesis, admin, factoryAddr, "SetOwner", []interface{}{bob})
		Expect(err).To(Succeed())

		is, err := Exec(genesis, admin, factoryAddr, "Owner", []interface{}{})
		Expect(err).To(Succeed())
		Expect(is[0].(common.Address)).To(Equal(bob))

		_, err = Exec(genesis, admin, factoryAddr, "SetOwner", []interface{}{bob})
		Expect(err).To(MatchError("Exchange: FORBIDDEN"))

	})
})
