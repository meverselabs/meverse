package test

import (
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Senario", func() {
	var (
		cdx int
		cn  *chain.Chain
		ctx *types.Context
	)

	It("MEV-MEFI", func() {
		beforeEach()
		MEV := uniTokens[0]
		MEFI := uniTokens[1]
		uniMint(genesis, alice)

		// pair 생성 : fee 0.3%, adminfee 100%, winnerfee 0%
		is, _ := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{MEV, MEFI, MEV, "MEV-MEFI Dex", "MEV-MEFI", alice, charlie, uint64(30000000), uint64(10000000000), uint64(0), _WhiteList, _GroupId, classMap["UniSwap"]})
		pair = is[0].(common.Address)
		GPrintln("MEV-MEFI", pair)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

		tokenApprove(ctx, MEV, alice, routerAddr)
		tokenApprove(ctx, MEFI, alice, routerAddr)
		tokenApprove(ctx, pair, alice, routerAddr)

		// 1. UniAddLiqudity (MEV, MEFI, 1000, 1000, 0, 0)
		is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{MEV, MEFI, amount.NewAmount(1000, 0), amount.NewAmount(1000, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 1", is[0], is[1], is[2])

		// 2. UniAddLiqudity (MEFI, MEV, 500, 300, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{MEFI, MEV, amount.NewAmount(500, 0), amount.NewAmount(300, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 2", is[0], is[1], is[2])

		// 3. SwapExactTokensForTokens (1, 0, [MEV,MEFI])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, MEFI}})
		Expect(err).To(Succeed())
		GPrintln(" 3", is[0])

		// 4. SwapExactTokensForTokens (1, 0, [MEFI,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEFI, MEV}})
		Expect(err).To(Succeed())
		GPrintln(" 4", is[0])

		// 5. UniAddLiquidityOneCoin (MEV, MEFI, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, MEFI, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 5", is[0], is[1])

		// 6. UniAddLiquidityOneCoin (MEV, MEFI, MEFI, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, MEFI, MEFI, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 6", is[0], is[1])

		// 7. UniAddLiquidityOneCoin (MEFI, MEV, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEFI, MEV, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 7", is[0], is[1])

		// 8. UniAddLiquidityOneCoin (MEFI, MEV, MEFI, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEFI, MEV, MEFI, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 8", is[0], is[1])

		// 9. UniRemoveLiquidity (MEFI, MEV, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{MEFI, MEV, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 9", is[0], is[1])

		// 10. UniRemoveLiquidity (MEV, MEFI, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{MEV, MEFI, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("10", is[0], is[1])

		// 11. SwapExactTokensForTokens (1, 0, [MEV,MEFI])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, MEFI}})
		Expect(err).To(Succeed())
		GPrintln("11", is[0])

		// 12. SwapExactTokensForTokens (1, 0, [MEFI,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEFI, MEV}})
		Expect(err).To(Succeed())
		GPrintln("12", is[0])

		// 13. UniRemoveLiquidityOneCoin (MEFI, MEV, 1, MEFI, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEFI, MEV, amount.NewAmount(1, 0), MEFI, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("13", is[0])

		// 14. UniRemoveLiquidityOneCoin (MEFI, MEV, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEFI, MEV, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("14", is[0])

		// 15. UniRemoveLiquidityOneCoin (MEV, MEFI, 1, MEFI, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, MEFI, amount.NewAmount(1, 0), MEFI, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("15", is[0])

		// 16. UniRemoveLiquidityOneCoin (MEV, MEFI, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, MEFI, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("16", is[0])

		// 17. Reserves
		is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("17", is[0], is[1])

		// 18. WithdrawAdminFees2
		is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("18", is[0], is[1], is[2], is[3], is[4])

		// 19. UniAddLiquidityOneCoin (MEFI, MEV, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEFI, MEV, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("19", is[0], is[1])

		RemoveChain(cdx)
		afterEach()
	})

	It("MEV-USDT", func() {
		beforeEach()
		MEV := uniTokens[0]
		USDT := uniTokens[1]
		uniMint(genesis, alice)

		// pair 생성 : fee 0.3%, adminfee 50%, winnerfee 0%
		is, _ := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{MEV, USDT, MEV, "MEV-USDT Dex", "MEV-USDT", alice, charlie, uint64(30000000), uint64(5000000000), uint64(0), _WhiteList, _GroupId, classMap["UniSwap"]})
		pair = is[0].(common.Address)
		GPrintln("MEV-USDT", pair)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

		tokenApprove(ctx, MEV, alice, routerAddr)
		tokenApprove(ctx, USDT, alice, routerAddr)
		tokenApprove(ctx, pair, alice, routerAddr)

		// 1. UniAddLiqudity (MEV, USDT, 1000, 100, 0, 0)
		is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{MEV, USDT, amount.NewAmount(1000, 0), amount.NewAmount(100, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 1", is[0], is[1], is[2])

		// 2. UniAddLiqudity (USDT, MEV, 500, 300, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{USDT, MEV, amount.NewAmount(500, 0), amount.NewAmount(300, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 2", is[0], is[1], is[2])

		// 3. SwapExactTokensForTokens (1, 0, [MEV,USDT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, USDT}})
		Expect(err).To(Succeed())
		GPrintln(" 3", is[0])

		// 4. SwapExactTokensForTokens (1, 0, [USDT,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDT, MEV}})
		Expect(err).To(Succeed())
		GPrintln(" 4", is[0])

		// 5. UniAddLiquidityOneCoin (MEV, USDT, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, USDT, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 5", is[0], is[1])

		// 6. UniAddLiquidityOneCoin (MEV, USDT, USDT, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, USDT, USDT, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 6", is[0], is[1])

		// 7. UniAddLiquidityOneCoin (USDT, MEV, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{USDT, MEV, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 7", is[0], is[1])

		// 8. UniAddLiquidityOneCoin (USDT, MEV, USDT, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{USDT, MEV, USDT, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 8", is[0], is[1])

		// 9. UniRemoveLiquidity (USDT, MEV, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{USDT, MEV, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 9", is[0], is[1])

		// 10. UniRemoveLiquidity (MEV, USDT, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{MEV, USDT, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("10", is[0], is[1])

		// 11. SwapExactTokensForTokens (1, 0, [MEV,USDT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, USDT}})
		Expect(err).To(Succeed())
		GPrintln("11", is[0])

		// 12. SwapExactTokensForTokens (1, 0, [USDT,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDT, MEV}})
		Expect(err).To(Succeed())
		GPrintln("12", is[0])

		// 13. UniRemoveLiquidityOneCoin (USDT, MEV, 1, USDT, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{USDT, MEV, amount.NewAmount(1, 0), USDT, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("13", is[0])

		// 14. UniRemoveLiquidityOneCoin (USDT, MEV, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{USDT, MEV, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("14", is[0])

		// 15. UniRemoveLiquidityOneCoin (MEV, USDT, 1, USDT, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, USDT, amount.NewAmount(1, 0), USDT, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("15", is[0])

		// 16. UniRemoveLiquidityOneCoin (MEV, USDT, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, USDT, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("16", is[0])

		// 17. Reserves
		is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("17", is[0], is[1])

		// 18. WithdrawAdminFees2
		is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("18", is[0], is[1], is[2], is[3], is[4])

		// 19. UniAddLiquidityOneCoin (MEV, USDT, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, USDT, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("19", is[0], is[1])

		RemoveChain(cdx)
		afterEach()
	})

	It("USDT-USDC", func() {
		beforeEach()
		USDT := uniTokens[0]
		USDC := uniTokens[1]
		uniMint(genesis, alice)

		// swap 생성 : fee 0.3%, adminfee 30%, winnerfee 0%, Amp 170
		is, _ := Exec(genesis, admin, factoryAddr, "CreatePairStable", []interface{}{USDT, USDC, USDC, "USDT-USDC Dex", "USDT-USDC", alice, charlie, uint64(30000000), uint64(3000000000), uint64(0), _WhiteList, _GroupId, uint64(170), classMap["StableSwap"]})
		swap = is[0].(common.Address)
		GPrintln("USDT-USDC", swap)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

		tokenApprove(ctx, USDT, alice, swap)
		tokenApprove(ctx, USDC, alice, swap)
		tokenApprove(ctx, USDT, alice, routerAddr)
		tokenApprove(ctx, USDC, alice, routerAddr)
		tokenApprove(ctx, swap, alice, routerAddr)

		// 0. CalcLPTokenAmount ([1000, 1000])
		is, err := Exec(ctx, alice, swap, "CalcLPTokenAmount", []interface{}{[]*amount.Amount{amount.NewAmount(1000, 0), amount.NewAmount(1000, 0)}, true})
		Expect(err).To(Succeed())
		GPrintln(" 0", is[0])

		// 1. UniAddLiqudity ([1000, 1000], 0)
		is, err = Exec(ctx, alice, swap, "AddLiquidity", []interface{}{[]*amount.Amount{amount.NewAmount(1000, 0), amount.NewAmount(1000, 0)}, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 1", is[0])

		// 2. UniAddLiqudity ([500, 300], 0)
		is, err = Exec(ctx, alice, swap, "AddLiquidity", []interface{}{[]*amount.Amount{amount.NewAmount(500, 0), amount.NewAmount(300, 0)}, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 2", is[0])

		// 3. SwapExactTokensForTokens (1, 0, [USDT,USDC])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDT, USDC}})
		Expect(err).To(Succeed())
		GPrintln(" 3", is[0])

		// 4. SwapExactTokensForTokens (1, 0, [USDC,USDT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDC, USDT}})
		Expect(err).To(Succeed())
		GPrintln(" 4", is[0])

		// 5. AddLiquidity ([10,0], 0)
		is, err = Exec(ctx, alice, swap, "AddLiquidity", []interface{}{[]*amount.Amount{amount.NewAmount(10, 0), ZeroAmount}, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 5", is[0])

		// 6. AddLiquidity ([0,10], 0)
		is, err = Exec(ctx, alice, swap, "AddLiquidity", []interface{}{[]*amount.Amount{ZeroAmount, amount.NewAmount(10, 0)}, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 6", is[0])

		// 7. RemoveLiquidity (1, [0,0])
		is, err = Exec(ctx, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(1, 0), []*amount.Amount{ZeroAmount, ZeroAmount}})
		Expect(err).To(Succeed())
		GPrintln(" 7", is[0])

		// 8. RemoveLiquidity (1, [0,0])
		is, err = Exec(ctx, alice, swap, "RemoveLiquidity", []interface{}{amount.NewAmount(1, 0), []*amount.Amount{ZeroAmount, ZeroAmount}})
		Expect(err).To(Succeed())
		GPrintln(" 8", is[0])

		// 9. SwapExactTokensForTokens (1, 0, [USDT,USDC])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDT, USDC}})
		Expect(err).To(Succeed())
		GPrintln(" 9", is[0])

		// 10. SwapExactTokensForTokens (1, 0, [USDC,USDT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{USDC, USDT}})
		Expect(err).To(Succeed())
		GPrintln("10", is[0])

		// 11. RemoveLiquidityOneCoin (1, 0, 0)
		is, err = Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amount.NewAmount(1, 0), uint8(0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("11", is[0])

		// 12. RemoveLiquidityOneCoin (1, 1, 0)
		is, err = Exec(ctx, alice, swap, "RemoveLiquidityOneCoin", []interface{}{amount.NewAmount(1, 0), uint8(1), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("12", is[0])

		// 13. RemoveLiquidityImbalance ([1,2], MaxUint256)
		is, err = Exec(ctx, alice, swap, "RemoveLiquidityImbalance", []interface{}{[]*amount.Amount{amount.NewAmount(1, 0), amount.NewAmount(2, 0)}, MaxUint256})
		Expect(err).To(Succeed())
		GPrintln("13", is[0])

		// 14. RemoveLiquidityImbalance ([2,1], MaxUint256)
		is, err = Exec(ctx, alice, swap, "RemoveLiquidityImbalance", []interface{}{[]*amount.Amount{amount.NewAmount(2, 0), amount.NewAmount(1, 0)}, MaxUint256})
		Expect(err).To(Succeed())
		GPrintln("14", is[0])

		// 15. Reserves
		is, err = Exec(ctx, alice, swap, "Reserves", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("15", is[0], is[1])

		// 16. WithdrawAdminFees
		is, err = Exec(ctx, alice, pair, "WithdrawAdminFees", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("16", is[0], is[1])

		// 17. AddLiquidity ([10,0], 0)
		is, err = Exec(ctx, alice, swap, "AddLiquidity", []interface{}{[]*amount.Amount{amount.NewAmount(10, 0), ZeroAmount}, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("17", is[0])

		RemoveChain(cdx)
		afterEach()
	})

	It("PocketToken-MEV", func() {
		beforeEach()
		PT := uniTokens[0]
		MEV := uniTokens[1]
		uniMint(genesis, alice)

		// pair 생성 : fee 0.3%, adminfee 50%, winnerfee 50%
		is, _ := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{PT, MEV, PT, "PocketToken-MEV Dex", "PocketToken-MEV", alice, charlie, uint64(30000000), uint64(5000000000), uint64(5000000000), _WhiteList, _GroupId, classMap["UniSwap"]})
		pair = is[0].(common.Address)
		GPrintln("PocketToken-MEV", pair)

		cn, cdx, ctx, _ = initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

		tokenApprove(ctx, PT, alice, routerAddr)
		tokenApprove(ctx, MEV, alice, routerAddr)
		tokenApprove(ctx, pair, alice, routerAddr)

		// 1. UniAddLiqudity (PT, MEV, 100, 1000, 0, 0)
		is, err := Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{PT, MEV, amount.NewAmount(100, 0), amount.NewAmount(1000, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 1", is[0], is[1], is[2])

		// 2. UniAddLiqudity (MEV, PT, 500, 300, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidity", []interface{}{MEV, PT, amount.NewAmount(500, 0), amount.NewAmount(300, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 2", is[0], is[1], is[2])

		// 3. SwapExactTokensForTokens (1, 0, [PT,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{PT, MEV}})
		Expect(err).To(Succeed())
		GPrintln(" 3", is[0])

		// 4. SwapExactTokensForTokens (1, 0, [MEV,PT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, PT}})
		Expect(err).To(Succeed())
		GPrintln(" 4", is[0])

		// 5. UniAddLiquidityOneCoin (PT, MEV, PT, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{PT, MEV, PT, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 5", is[0], is[1])

		// 6. UniAddLiquidityOneCoin (PT, MEV, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{PT, MEV, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 6", is[0], is[1])

		// 7. UniAddLiquidityOneCoin (MEV, PT, PT, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, PT, PT, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 7", is[0], is[1])

		// 8. UniAddLiquidityOneCoin (MEV, PT, MEV, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{MEV, PT, MEV, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 8", is[0], is[1])

		// 9. UniRemoveLiquidity (MEV, PT, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{MEV, PT, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln(" 9", is[0], is[1])

		// 10. UniRemoveLiquidity (PT, MEV, 1, 0, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidity", []interface{}{PT, MEV, amount.NewAmount(1, 0), ZeroAmount, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("10", is[0], is[1])

		// 11. SwapExactTokensForTokens (1, 0, [PT,MEV])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{PT, MEV}})
		Expect(err).To(Succeed())
		GPrintln("11", is[0])

		// 12. SwapExactTokensForTokens (1, 0, [MEV,PT])
		is, err = Exec(ctx, alice, routerAddr, "SwapExactTokensForTokens", []interface{}{amount.NewAmount(1, 0), ZeroAmount, []common.Address{MEV, PT}})
		Expect(err).To(Succeed())
		GPrintln("12", is[0])

		// 13. UniRemoveLiquidityOneCoin (MEV, PT, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, PT, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("13", is[0])

		// 14. UniRemoveLiquidityOneCoin (MEV, PT, 1, PT, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{MEV, PT, amount.NewAmount(1, 0), PT, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("14", is[0])

		// 15. UniRemoveLiquidityOneCoin (PT, MEV, 1, MEV, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{PT, MEV, amount.NewAmount(1, 0), MEV, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("15", is[0])

		// 16. UniRemoveLiquidityOneCoin (PT, MEV, 1, PT, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniRemoveLiquidityOneCoin", []interface{}{PT, MEV, amount.NewAmount(1, 0), PT, ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("16", is[0])

		// 17. Reserves
		is, err = Exec(ctx, alice, pair, "Reserves", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("17", is[0], is[1])

		// 18. WithdrawAdminFees2
		is, err = Exec(ctx, alice, pair, "WithdrawAdminFees2", []interface{}{})
		Expect(err).To(Succeed())
		GPrintln("18", is[0], is[1], is[2], is[3], is[4])

		// 19. UniAddLiquidityOneCoin (PT, MEV, PT, 1, 0)
		is, err = Exec(ctx, alice, routerAddr, "UniAddLiquidityOneCoin", []interface{}{PT, MEV, PT, amount.NewAmount(1, 0), ZeroAmount})
		Expect(err).To(Succeed())
		GPrintln("19", is[0], is[1])

		RemoveChain(cdx)
		afterEach()
	})
})
