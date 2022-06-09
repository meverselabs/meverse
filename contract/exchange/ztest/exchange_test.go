package test

import (
	"math/big"

	"github.com/meverselabs/meverse/common"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	contract : curve-contract
	directory : tests/token
	files :
		test_approve.py
		test_mint_burn.py
		test_transfer.py
		test_transferFrom.py
*/

var _ = Describe("Exchange", func() {

	Describe("Uniswap", func() {

		BeforeEach(func() {
			beforeEach()
			is, _ := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], uniTokens[0], _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
			pair = is[0].(common.Address)
		})

		AfterEach(func() {
			afterEach()
		})

		It("payToken", func() {
			is, err := Exec(genesis, alice, pair, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(uniTokens[0]))
		})

		It("SetPayToken, onlyOwner", func() {

			payToken := uniTokens[1]

			is, err := Exec(genesis, alice, pair, "SetPayToken", []interface{}{payToken})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, pair, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))

			is, err = Exec(genesis, bob, pair, "SetPayToken", []interface{}{payToken})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetPayToken : ZeroAddress ", func() {

			payToken := ZeroAddress

			is, err := Exec(genesis, alice, pair, "SetPayToken", []interface{}{payToken})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, pair, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))
		})

		It("SetPayToken : Another token", func() {

			payToken := common.HexToAddress("1")

			_, err := Exec(genesis, alice, pair, "SetPayToken", []interface{}{payToken})
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))
		})

	})

	Describe("StableSwap", func() {

		BeforeEach(func() {
			beforeEach()
			sbc := &trade.StableSwapConstruction{
				Name:         _SwapName,
				Symbol:       _SwapSymbol,
				Factory:      ZeroAddress,
				NTokens:      uint8(N),
				Tokens:       stableTokens,
				PayToken:     stableTokens[0],
				Owner:        alice,
				Winner:       charlie,
				Fee:          _Fee,
				AdminFee:     _AdminFee,
				WinnerFee:    _WinnerFee,
				Amp:          big.NewInt(_Amp),
				PrecisionMul: _PrecisionMul,
				Rates:        _Rates,
			}
			swap, _ = stablebase(genesis, sbc)
		})

		AfterEach(func() {
			afterEach()
		})

		It("payToken", func() {
			is, err := Exec(genesis, alice, swap, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(stableTokens[0]))
		})

		It("SetPayToken, onlyOwner", func() {

			payToken := stableTokens[1]

			is, err := Exec(genesis, alice, swap, "SetPayToken", []interface{}{payToken})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, swap, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))

			is, err = Exec(genesis, bob, swap, "SetPayToken", []interface{}{payToken})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetPayToken : ZeroAddress ", func() {

			payToken := ZeroAddress

			is, err := Exec(genesis, alice, swap, "SetPayToken", []interface{}{payToken})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, swap, "PayToken", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(payToken))
		})

		It("SetPayToken : Another token", func() {

			payToken := common.HexToAddress("1")

			_, err := Exec(genesis, alice, swap, "SetPayToken", []interface{}{payToken})
			Expect(err).To(MatchError("Exchange: NOT_EXIST_PAYTOKEN"))
		})

	})
})
