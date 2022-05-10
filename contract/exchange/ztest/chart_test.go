package test

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chart", func() {

	It("dy : stable vs uni", func() {
		beforeEachWithoutTokens()

		ctx := genesis
		sN := 100 // supply number
		tN := 6   // token number

		tokens := DeployTokens(ctx, classMap["Token"], uint8(tN), admin)
		//GPrintln("tokens", tokens)
		fee, adminfee, winnerfee := uint64(0), uint64(0), uint64(0)
		supply := amount.NewAmount(uint64(sN), 0)

		is, err := Exec(ctx, admin, factoryAddr, "CreatePairUni", []interface{}{tokens[0], tokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, fee, adminfee, winnerfee, _WhiteList, _GroupId, classMap["UniSwap"]})
		Expect(err).To(Succeed())
		pair = is[0].(common.Address)

		is, err = Exec(ctx, admin, factoryAddr, "CreatePairStable", []interface{}{tokens[2], tokens[3], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, fee, adminfee, winnerfee, _WhiteList, _GroupId, uint64(50), classMap["StableSwap"]})
		Expect(err).To(Succeed())
		swap1 := is[0].(common.Address)

		is, err = Exec(ctx, admin, factoryAddr, "CreatePairStable", []interface{}{tokens[4], tokens[5], ZeroAddress, _SwapName, _SwapSymbol, alice, charlie, fee, adminfee, winnerfee, _WhiteList, _GroupId, uint64(100), classMap["StableSwap"]})
		Expect(err).To(Succeed())
		swap2 := is[0].(common.Address)

		for i := 0; i < tN; i++ {
			tokenMint(ctx, tokens[i], alice, supply)
			tokenApprove(ctx, tokens[i], alice, routerAddr)
			tokenApprove(ctx, tokens[i], alice, swap1)
			tokenApprove(ctx, tokens[i], alice, swap2)
		}

		_, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{tokens[0], tokens[1], supply, supply, ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())

		_, err = Exec(ctx, alice, swap1, "AddLiquidity", []interface{}{[]*amount.Amount{supply, supply}, ZeroAmount})
		//GPrintf("%+v", err)
		Expect(err).To(Succeed())

		_, err = Exec(ctx, alice, swap2, "AddLiquidity", []interface{}{[]*amount.Amount{supply, supply}, ZeroAmount})
		Expect(err).To(Succeed())

		for i := 0; i < tN; i += 2 {
			tokenMint(ctx, tokens[i], charlie, supply)
			tokenApprove(ctx, tokens[i], charlie, routerAddr)
		}
		swapAmount := amount.NewAmount(1, 0)
		for k := 0; k < sN; k++ {
			is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[0], tokens[1]}})
			Expect(err).To(Succeed())
			outputUni := is[0].([]*amount.Amount)[1]

			is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[2], tokens[3]}})
			Expect(err).To(Succeed())
			outputStable1 := is[0].([]*amount.Amount)[1]

			is, err = Exec(ctx, charlie, routerAddr, "SwapExactTokensForTokens", []interface{}{swapAmount, ZeroAmount, []common.Address{tokens[4], tokens[5]}})
			Expect(err).To(Succeed())
			outputStable2 := is[0].([]*amount.Amount)[1]

			GPrintln(k, outputUni, outputStable1, outputStable2)
		}

		afterEach()
	})
})
