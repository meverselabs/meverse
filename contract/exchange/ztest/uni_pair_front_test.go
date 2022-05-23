package test

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/whitelist"

	. "github.com/meverselabs/meverse/contract/exchange/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	transaction 테스트
*/

var _ = Describe("Uni Front Tx", func() {
	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
		err error

		step = uint64(3600)
	)

	Describe("Token", func() {
		BeforeEach(func() {
			beforeEachUni()
			lpTokenMint(genesis, pair, alice, _TotalSupply) // initial Token Supply
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("Name, Symbol, TotalSupply, Decimals", func() {
			//Name string
			is, err := Exec(ctx, alice, pair, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal("__UNI_NAME"))

			//Symbol string
			is, err = Exec(ctx, alice, pair, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal("__UNI_SYMBOL"))

			//TotalSupply *amount.Amount
			is, err = Exec(ctx, alice, pair, "TotalSupply", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			//Decimals *big.Int
			is, err = Exec(ctx, alice, pair, "Decimals", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(big.NewInt(amount.FractionalCount)))
		})

		It("Transfer", func() {
			//Transfer(To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(ctx, pair, bob)).To(Equal(_TestAmount))
		})

		It("Approve", func() {
			//Approve(To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, pair, alice, bob)).To(Equal(_TestAmount))
		})

		It("IncreaseAllowance", func() {
			//IncreaseAllowance(spender common.Address, addAmount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "IncreaseAllowance",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, pair, alice, bob)).To(Equal(_TestAmount))
		})

		It("DecreaseAllowance", func() {
			//DecreaseAllowance(spender common.Address, subtractAmount *amount.Amount)

			tSeconds := ctx.LastTimestamp() / uint64(time.Second)

			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount.MulC(3)),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			tSeconds += step

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "DecreaseAllowance",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, pair, alice, bob)).To(Equal(_TestAmount.MulC(2)))
		})

		It("TransferFrom", func() {
			//TransferFrom(From common.Address, To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount.MulC(3)),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "TransferFrom",
				Args:      bin.TypeWriteAll(alice, charlie, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, pair, alice, bob)).To(Equal(_TestAmount.MulC(2)))
			Expect(tokenBalanceOf(ctx, pair, alice)).To(Equal(_TotalSupply.Sub(_TestAmount)))
			Expect(tokenBalanceOf(ctx, pair, charlie)).To(Equal(_TestAmount))
		})

	})

	Describe("Uni", func() {

		BeforeEach(func() {
			beforeEachUni()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, err = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("Extype, Fee, AdminFee, NTokens, Rates, PrecisionMul, Tokens, Owner, WhiteList, GroupId", func() {
			//Extype uint8
			is, err := Exec(ctx, alice, pair, "ExType", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(trade.UNI))

			ctx, _ = setFees(cn, ctx, pair, uint64(1234), uint64(3456), uint64(67890), uint64(86400), aliceKey)

			// NTokens(cc *types.ContractContext) uint8
			is, err = Exec(ctx, alice, pair, "NTokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(uint8(2)))

			sortToken0, sortToken1, err := trade.SortTokens(uniTokens[0], uniTokens[1])
			Expect(err).To(Succeed())

			// Coins []common.Address
			is, err = Exec(ctx, alice, pair, "Tokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal([]common.Address{sortToken0, sortToken1}))

			// Owner common.Address
			is, err = Exec(ctx, alice, pair, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(alice))

			// Fee uint64
			is, err = Exec(ctx, alice, pair, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(1234)))

			// AdminFee uint64
			is, err = Exec(ctx, alice, pair, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(3456)))

			// WinnerFee uint64
			is, err = Exec(ctx, alice, pair, "WinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(67890)))

			// WhiteList
			is, err = Exec(ctx, alice, pair, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(_WhiteList))

			// GroupId
			is, err = Exec(ctx, alice, pair, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})

		It("CommitNewFee, AdminActionsDeadline, FutureFee, FutureAdminFee, ApplyNewFee, RevertNewFee", func() {
			fee := uint64(rand.Intn(trade.MAX_FEE) + 1)
			admin_fee := uint64(rand.Intn(trade.MAX_ADMIN_FEE + 1))
			winner_fee := uint64(rand.Intn(trade.MAX_WINNER_FEE + 1))
			delay := uint64(3 * 86400)

			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitNewFee",
				Args:      bin.TypeWriteAll(fee, admin_fee, winner_fee, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline uint64
			is, err := Exec(ctx, alice, pair, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureFee uint64
			is, err = Exec(ctx, alice, pair, "FutureFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// FutureAdminFee uint64
			is, err = Exec(ctx, alice, pair, "FutureAdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// FutureWinnerFee uint64
			is, err = Exec(ctx, alice, pair, "FutureWinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "ApplyNewFee",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline uint64
			is, err = Exec(ctx, alice, pair, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee uint64
			is, err = Exec(ctx, alice, pair, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee uint64
			is, err = Exec(ctx, alice, pair, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee uint64
			is, err = Exec(ctx, alice, pair, "WinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitNewFee",
				Args:      bin.TypeWriteAll(uint64(1000), uint64(2000), uint64(3000), delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// RevertNewParameters(cc *types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "RevertNewFee",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline uint64
			is, err = Exec(ctx, alice, pair, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee uint64
			is, err = Exec(ctx, alice, pair, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee uint64
			is, err = Exec(ctx, alice, pair, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee uint64
			is, err = Exec(ctx, alice, pair, "WinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))
		})

		It("CommitNewWhiteList, WhiteListDeadline, FutureWhiteList, FutureGroupId, ApplyNewWhiteList, RevertNewWhiteList", func() {

			// whitelist contract
			whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
			bs, _, err := bin.WriterToBytes(whitelistConstrunction)
			Expect(err).To(Succeed())
			v, err := genesis.DeployContract(admin, classMap["WhiteList"], bs)
			Expect(err).To(Succeed())
			wl := v.(*whitelist.WhiteListContract).Address()
			gId := hash.BigToHash(big.NewInt(100))

			delay := uint64(3 * 86400)

			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitNewWhiteList",
				Args:      bin.TypeWriteAll(wl, gId, delay),
			}

			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err := Exec(ctx, alice, pair, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureWhiteList() common.Address
			is, err = Exec(ctx, alice, pair, "FutureWhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// FutureGroupId() hash.Hash256
			is, err = Exec(ctx, alice, pair, "FutureGroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "ApplyNewWhiteList",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err = Exec(ctx, alice, pair, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = Exec(ctx, alice, pair, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// GroupId() hash.Hash256
			is, err = Exec(ctx, alice, pair, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitNewWhiteList",
				Args:      bin.TypeWriteAll(_WhiteList, _GroupId, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// RevertNewWhiteList()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "RevertNewWhiteList",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = Exec(ctx, alice, pair, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = Exec(ctx, alice, pair, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// GroupId() hash.Hash256
			is, err = Exec(ctx, alice, pair, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))
		})

		It("Owner, CommitTransferOwnerWinner, TransferOwnerWinnerDeadline, ApplyTransferOwnerWinner, FutureOwner, RevertTransferOwnerWinner", func() {
			delay := uint64(3 * 86400)

			// CommitTransferOwnerWinner(_owner common.Address)
			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitTransferOwnerWinner",
				Args:      bin.TypeWriteAll(bob, charlie, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline uint64
			is, err := Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureOwner common.Address
			is, err = Exec(ctx, alice, pair, "FutureOwner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// FutureOwner common.Address
			is, err = Exec(ctx, alice, pair, "FutureWinner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			// ApplyTransferOwnerWinner(cc *types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "ApplyTransferOwnerWinner",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline uint64
			is, _ = Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner common.Address
			is, err = Exec(ctx, alice, pair, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner common.Address
			is, err = Exec(ctx, alice, pair, "Winner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "CommitTransferOwnerWinner",
				Args:      bin.TypeWriteAll(charlie, alice, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// RevertTransferOwnerWinner(cc *types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "RevertTransferOwnerWinner",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline uint64
			is, _ = Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner common.Address
			is, err = Exec(ctx, alice, pair, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner common.Address
			is, err = Exec(ctx, alice, pair, "Winner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))
		})

		It("WithdrawAdminFees2, AdminBalance, Skim, Sync", func() {
			uniAddInitialLiquidity(ctx, alice)
			aliceLPBalance, _ := tokenBalanceOf(ctx, pair, alice)
			uniMint(ctx, bob)
			uniApprove(ctx, bob)
			ctx, _ = setFees(cn, ctx, pair, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(86400), aliceKey)

			swapAmount := amount.NewAmount(1, 0)

			Expect(aliceLPBalance.Cmp(swapAmount.Int) > 0).To(BeTrue())
			// token0 -> token1 SwapExactTokensForTokens
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        routerAddr,
				Method:    "SwapExactTokensForTokens",
				Args:      bin.TypeWriteAll(swapAmount, ZeroAmount, []common.Address{token0, token1}),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			GPrintf("%+v", err)
			Expect(err).To(Succeed())

			// AdminBalance(cc *types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "AdminBalance",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// WithdrawAdminFees2(cc *types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "WithdrawAdminFees2",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// token1 -> token0 SwapExactTokensForTokens
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        routerAddr,
				Method:    "SwapExactTokensForTokens",
				Args:      bin.TypeWriteAll(swapAmount, ZeroAmount, []common.Address{token1, token0}),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// Sync
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Sync",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// transfer balance != reserve
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        token0,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(pair, swapAmount),
			}

			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// Skim
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Skim",
				Args:      bin.TypeWriteAll(charlie),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// transfer balance != reserve
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        token0,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(pair, swapAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "Sync",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())
		})

		It("KillMe, UnkillMe", func() {
			// UnkillMe(cc * types.ContractContext)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "UnkillMe",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// KillMe(cc * types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "KillMe",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// UnkillMe(cc * types.ContractContext)
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        pair,
				Method:    "UnkillMe",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())
		})
	})
})
