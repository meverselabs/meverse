package test

import (
	"github.com/meverselabs/meverse/common"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TokenTransfer", func() {

	Describe("uniswap", func() {

		BeforeEach(func() {
			beforeEach()
			uniMint(genesis, bob)

			is, _ := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{token0, token1, token1, "MEV-MEFI Dex", "MEV-MEFI", alice, charlie, uint64(30000000), uint64(10000000000), uint64(0), _WhiteList, _GroupId, classMap["UniSwap"]})
			pair = is[0].(common.Address)
		})

		AfterEach(func() {
			afterEach()
		})

		It("tokentransfer", func() {

			_, err := Exec(genesis, bob, token0, "Transfer", []interface{}{pair, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(genesis, token0, pair)).To(Equal(_TestAmount))

			_, err = Exec(genesis, bob, token1, "Transfer", []interface{}{pair, _TestAmount.MulC(2)})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token1, bob)).To(Equal(_SupplyTokens[1].Sub(_TestAmount.MulC(2))))
			Expect(tokenBalanceOf(genesis, token1, pair)).To(Equal(_TestAmount.MulC(2)))

			_, err = Exec(genesis, alice, pair, "TokenTransfer", []interface{}{token0, charlie, _TestAmount.DivC(2)})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, pair)).To(Equal(_TestAmount.DivC(2)))
			Expect(tokenBalanceOf(genesis, token0, charlie)).To(Equal(_TestAmount.DivC(2)))

			_, err = Exec(genesis, alice, pair, "TokenTransfer", []interface{}{token1, charlie, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token1, pair)).To(Equal(_TestAmount))
			Expect(tokenBalanceOf(genesis, token1, charlie)).To(Equal(_TestAmount))

		})
		It("onlyOwner", func() {

			_, err := Exec(genesis, bob, pair, "TokenTransfer", []interface{}{token0, charlie, _TestAmount})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("transfer more than balance", func() {

			_, err := Exec(genesis, bob, token0, "Transfer", []interface{}{pair, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(genesis, token0, pair)).To(Equal(_TestAmount))

			_, err = Exec(genesis, alice, pair, "TokenTransfer", []interface{}{token0, charlie, _TestAmount.MulC(2)})
			Expect(err).To(HaveOccurred())

		})

		It("not base-token transfer", func() {

			otherToken := stableTokens[0]
			_, err := Exec(genesis, alice, pair, "TokenTransfer", []interface{}{otherToken, charlie, _TestAmount})
			Expect(err).To(MatchError("Exchange: NOT_EXIST_TOKEN"))

		})

	})

	Describe("stableswap", func() {

		BeforeEach(func() {
			beforeEach()
			uniMint(genesis, bob)

			is, _ := Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{token0, token1, token1, "USDT-USDC Dex", "USDT-USDC", alice, charlie, uint64(30000000), uint64(3000000000), uint64(0), _WhiteList, _GroupId, uint64(170), classMap["StableSwap"]})
			swap = is[0].(common.Address)
		})

		AfterEach(func() {
			afterEach()
		})

		It("tokentransfer", func() {

			_, err := Exec(genesis, bob, token0, "Transfer", []interface{}{swap, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(genesis, token0, swap)).To(Equal(_TestAmount))

			_, err = Exec(genesis, bob, token1, "Transfer", []interface{}{swap, _TestAmount.MulC(2)})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token1, bob)).To(Equal(_SupplyTokens[1].Sub(_TestAmount.MulC(2))))
			Expect(tokenBalanceOf(genesis, token1, swap)).To(Equal(_TestAmount.MulC(2)))

			_, err = Exec(genesis, alice, swap, "TokenTransfer", []interface{}{token0, charlie, _TestAmount.DivC(2)})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, swap)).To(Equal(_TestAmount.DivC(2)))
			Expect(tokenBalanceOf(genesis, token0, charlie)).To(Equal(_TestAmount.DivC(2)))

			_, err = Exec(genesis, alice, swap, "TokenTransfer", []interface{}{token1, charlie, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token1, swap)).To(Equal(_TestAmount))
			Expect(tokenBalanceOf(genesis, token1, charlie)).To(Equal(_TestAmount))

		})
		It("onlyOwner", func() {

			_, err := Exec(genesis, bob, swap, "TokenTransfer", []interface{}{token0, charlie, _TestAmount})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("transfer more than balance", func() {

			_, err := Exec(genesis, bob, token0, "Transfer", []interface{}{swap, _TestAmount})
			Expect(err).To(Succeed())
			Expect(tokenBalanceOf(genesis, token0, bob)).To(Equal(_SupplyTokens[0].Sub(_TestAmount)))
			Expect(tokenBalanceOf(genesis, token0, swap)).To(Equal(_TestAmount))

			_, err = Exec(genesis, alice, swap, "TokenTransfer", []interface{}{token0, charlie, _TestAmount.MulC(2)})
			Expect(err).To(HaveOccurred())

		})

		It("not base-token transfer", func() {

			otherToken := stableTokens[0]
			_, err := Exec(genesis, alice, swap, "TokenTransfer", []interface{}{otherToken, charlie, _TestAmount})
			Expect(err).To(MatchError("Exchange: NOT_EXIST_TOKEN"))

		})
	})
})
