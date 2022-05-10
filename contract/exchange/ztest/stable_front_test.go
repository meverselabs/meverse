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

var _ = Describe("Stable Front Tx", func() {
	var (
		cn  *chain.Chain
		cdx int
		ctx *types.Context
		err error

		step = uint64(3600)
	)

	Describe("Token", func() {
		BeforeEach(func() {
			beforeEachStable()
			lpTokenMint(genesis, swap, alice, _TotalSupply) // initial Token Supply
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})
		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("Name, Symbol, TotalSupply, Decimals", func() {
			// Name() string
			is, err := Exec(ctx, alice, swap, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapName))

			// Symbol() string
			is, err = Exec(ctx, alice, swap, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapSymbol))

			// TotalSupply() *amount.Amount
			is, err = Exec(ctx, alice, swap, "TotalSupply", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			// Decimals() *big.Int
			is, err = Exec(ctx, alice, swap, "Decimals", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(big.NewInt(amount.FractionalCount)))
		})

		It("Transfer", func() {
			//Transfer(To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(ctx, swap, bob)).To(Equal(_TestAmount))
		})

		It("Approve", func() {
			//Approve(To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, swap, alice, bob)).To(Equal(_TestAmount))
		})

		It("IncreaseAllowance", func() {
			//IncreaseAllowance(spender common.Address, addAmount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "IncreaseAllowance",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, swap, alice, bob)).To(Equal(_TestAmount))
		})

		It("DecreaseAllowance", func() {
			//DecreaseAllowance(spender common.Address, subtractAmount *amount.Amount)

			tSeconds := ctx.LastTimestamp() / uint64(time.Second)

			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount.MulC(3)),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			tSeconds += step

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "DecreaseAllowance",
				Args:      bin.TypeWriteAll(bob, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, swap, alice, bob)).To(Equal(_TestAmount.MulC(2)))
		})

		It("TransferFrom", func() {
			//TransferFrom(From common.Address, To common.Address, Amount *amount.Amount)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(bob, _TestAmount.MulC(3)),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "TransferFrom",
				Args:      bin.TypeWriteAll(alice, charlie, _TestAmount),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(ctx, swap, alice, bob)).To(Equal(_TestAmount.MulC(2)))

			Expect(tokenBalanceOf(ctx, swap, alice)).To(Equal(_TotalSupply.Sub(_TestAmount)))
			Expect(tokenBalanceOf(ctx, swap, charlie)).To(Equal(_TestAmount))

		})

	})

	Describe("Stable", func() {

		BeforeEach(func() {
			beforeEachStable()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			Expect(err).To(Succeed())
			timestamp := uint64(86400) // uint64(time.Now().UnixNano()) / uint64(time.Second)
			ctx, _ = Sleep(cn, ctx, nil, timestamp, aliceKey)
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("Extype, Fee, AdminFee, NTokens, Rates, PrecisionMul, Tokens, Owner, WhiteList, GroupId", func() {
			//Extype() uint8
			is, err := Exec(ctx, alice, swap, "ExType", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(trade.STABLE))

			ctx, _ = setFees(cn, ctx, swap, uint64(1234), uint64(3456), uint64(67890), uint64(86400), aliceKey)

			// NTokens() uint8
			is, err = Exec(ctx, alice, swap, "NTokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint8)).To(Equal(uint8(N)))

			// RATES() []*big.Int
			is, err = Exec(ctx, alice, swap, "Rates", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]*big.Int)).To(Equal(_Rates))

			// PRECISION_MUL() []uint64
			is, err = Exec(ctx, alice, swap, "PrecisionMul", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]uint64)).To(Equal(_PrecisionMul))

			// Coins() []common.Address
			is, err = Exec(ctx, alice, swap, "Tokens", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].([]common.Address)).To(Equal(stableTokens))

			// Owner() common.Address
			is, err = Exec(ctx, alice, swap, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(alice))

			// Fee() uint64
			is, err = Exec(ctx, alice, swap, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(1234)))

			// AdminFee() uint64
			is, err = Exec(ctx, alice, swap, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(3456)))

			// WinnerFee() uint64
			is, err = Exec(ctx, alice, swap, "WinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(uint64(67890)))

			// WhiteList
			is, err = Exec(ctx, alice, swap, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(_WhiteList))

			// GroupId
			is, err = Exec(ctx, alice, swap, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(_GroupId))
		})

		It("InitialA, FutureA, RampA, InitialA, FutureA, InitialATime, FutureATime, StopRampA", func() {

			is, err := Exec(ctx, alice, swap, "InitialA", []interface{}{})
			Expect(err).To(Succeed())
			initial_A := DivC(is[0].(*big.Int), trade.A_PRECISION)
			future_time := ctx.LastTimestamp()/uint64(time.Second) + trade.MIN_RAMP_TIME + 5

			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RampA",
				Args:      bin.TypeWriteAll(MulC(initial_A, 2), future_time),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// InitialA() *big.Int
			is, err = Exec(ctx, alice, swap, "InitialA", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(MulC(initial_A, trade.A_PRECISION)))

			// FutureA() *big.Int
			is, err = Exec(ctx, alice, swap, "FutureA", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(MulC(MulC(initial_A, 2), trade.A_PRECISION)))

			// InitialATime() uint64
			is, err = Exec(ctx, alice, swap, "InitialATime", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(timestamp))

			// FutureATime() uint64
			is, err = Exec(ctx, alice, swap, "FutureATime", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(future_time))

			//StopRampA()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "StopRampA",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

		})

		It("CommitNewFee, AdminActionsDeadline, FutureFee, FutureAdminFee, ApplyNewFee, RevertNewParameters", func() {

			fee := uint64(rand.Intn(trade.MAX_FEE) + 1)
			admin_fee := uint64(rand.Intn(trade.MAX_ADMIN_FEE + 1))
			winner_fee := uint64(rand.Intn(trade.MAX_WINNER_FEE + 1))
			delay := uint64(3 * 86400)

			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "CommitNewFee",
				Args:      bin.TypeWriteAll(fee, admin_fee, winner_fee, delay),
			}

			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err := Exec(ctx, alice, swap, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureFee() uint64
			is, err = Exec(ctx, alice, swap, "FutureFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// FutureAdminFee() uint64
			is, err = Exec(ctx, alice, swap, "FutureAdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// FutureWinnerFee() uint64
			is, err = Exec(ctx, alice, swap, "FutureWinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "ApplyNewFee",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = Exec(ctx, alice, swap, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee() uint64
			is, err = Exec(ctx, alice, swap, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee() uint64
			is, err = Exec(ctx, alice, swap, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee() uint64
			is, err = Exec(ctx, alice, swap, "WinnerFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(winner_fee))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "CommitNewFee",
				Args:      bin.TypeWriteAll(uint64(1000), uint64(2000), uint64(3000), delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// RevertNewParameters()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RevertNewFee",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = Exec(ctx, alice, swap, "AdminActionsDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Fee() uint64
			is, err = Exec(ctx, alice, swap, "Fee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(fee))

			// AdminFee() uint64
			is, err = Exec(ctx, alice, swap, "AdminFee", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(admin_fee))

			// WinnerFee() uint64
			is, err = Exec(ctx, alice, swap, "WinnerFee", []interface{}{})
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
				To:        swap,
				Method:    "CommitNewWhiteList",
				Args:      bin.TypeWriteAll(wl, gId, delay),
			}

			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err := Exec(ctx, alice, swap, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureWhiteList() common.Address
			is, err = Exec(ctx, alice, swap, "FutureWhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// FutureGroupId() hash.Hash256
			is, err = Exec(ctx, alice, swap, "FutureGroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "ApplyNewWhiteList",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// WhiteListDeadline() uint64
			is, err = Exec(ctx, alice, swap, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = Exec(ctx, alice, swap, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// GroupId() hash.Hash256
			is, err = Exec(ctx, alice, swap, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "CommitNewWhiteList",
				Args:      bin.TypeWriteAll(_WhiteList, _GroupId, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// RevertNewWhiteList()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RevertNewWhiteList",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// AdminActionsDeadline() uint64
			is, err = Exec(ctx, alice, swap, "WhiteListDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// WhiteList() common.Address
			is, err = Exec(ctx, alice, swap, "WhiteList", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(wl))

			// GroupId() hash.Hash256
			is, err = Exec(ctx, alice, swap, "GroupId", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(hash.Hash256)).To(Equal(gId))
		})

		It("Owner, CommitTransferOwnerWinner, TransferOwnerWinnerDeadline, ApplyTransferOwnerWinner, FutureOwner, RevertTransferOwnerWinner", func() {

			delay := uint64(3 * 86400)

			// CommitTransferOwnerWinner( _owner common.Address)
			timestamp := ctx.LastTimestamp() / uint64(time.Second)
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "CommitTransferOwnerWinner",
				Args:      bin.TypeWriteAll(bob, charlie, delay),
			}

			ctx, err = Sleep(cn, ctx, tx, delay, aliceKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, err := Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(timestamp + delay))

			// FutureOwner() common.Address
			is, err = Exec(ctx, alice, swap, "FutureOwner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// FutureOwner() common.Address
			is, err = Exec(ctx, alice, swap, "FutureWinner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			// ApplyTransferOwnerWinner()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "ApplyTransferOwnerWinner",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, _ = Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner() common.Address
			is, err = Exec(ctx, alice, swap, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner() common.Address
			is, err = Exec(ctx, alice, swap, "Winner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "CommitTransferOwnerWinner",
				Args:      bin.TypeWriteAll(charlie, alice, delay),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// RevertTransferOwnerWinner()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "RevertTransferOwnerWinner",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// TransferOwnerWinnerDeadline() uint64
			is, _ = Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

			// Owner() common.Address
			is, err = Exec(ctx, alice, swap, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			// Owner() common.Address
			is, err = Exec(ctx, alice, swap, "Winner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(charlie))

		})

		It("Withdraw_admin_fees, Donate_admin_fees", func() {
			stableAddInitialLiquidity(ctx, alice)
			stableMint(ctx, bob)
			stableApprove(ctx, bob)
			ctx, _ = setFees(cn, ctx, swap, trade.MAX_FEE, trade.MAX_ADMIN_FEE, trade.MAX_WINNER_FEE, uint64(86400), aliceKey)

			// 0 -> 1 Exchange
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Exchange",
				Args:      bin.TypeWriteAll(uint8(0), uint8(1), _InitialAmounts[0], ZeroAmount, ZeroAddress),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// Withdraw_admin_fees()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "WithdrawAdminFees",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// 1 -> 0 Exchange
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "Exchange",
				Args:      bin.TypeWriteAll(uint8(1), uint8(0), _InitialAmounts[1], ZeroAmount, ZeroAddress),
			}
			ctx, err = Sleep(cn, ctx, tx, step, bobKey)
			Expect(err).To(Succeed())

			// Donate_admin_fees()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "DonateAdminFees",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())
		})

		It("KillMe, UnkillMe", func() {
			// KillMe()
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "KillMe",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())

			// UnkillMe()
			tx = &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: ctx.LastTimestamp(),
				To:        swap,
				Method:    "UnkillMe",
				Args:      bin.TypeWriteAll(),
			}
			ctx, err = Sleep(cn, ctx, tx, step, aliceKey)
			Expect(err).To(Succeed())
		})

	})
})
