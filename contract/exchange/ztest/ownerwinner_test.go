package test

import (
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// test_transfer_ownership.py

var _ = Describe("test_transfer_ownership.py", func() {
	var (
		cdx              int
		err              error
		cn               *chain.Chain
		ctx              *types.Context
		timestamp_second uint64
	)

	Describe("UniSwap", func() {

		BeforeEach(func() {
			beforeEachUni()
			cn, cdx, ctx, _ = initChain(genesis, admin)
			ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("test_commit", func() {
			delay := uint64(3 * 86400)
			_, err := Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})
			Expect(err).To(Succeed())

			is, err := Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(ctx.LastTimestamp()/uint64(time.Second) + delay))

			is, err = Exec(ctx, alice, pair, "FutureOwner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

		})

		It("test_commit_only_owner", func() {
			_, err := Exec(ctx, bob, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_commit_already_active", func() {
			_, err := Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})

			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			Expect(err).To(MatchError("Exchange: ACTIVE_TRANSFER"))
		})

		It("test_revert", func() {
			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			_, err = Exec(ctx, alice, pair, "RevertTransferOwnerWinner", []interface{}{})

			is, _ := Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

		})

		It("test_revert_only_owner", func() {
			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			_, err = Exec(ctx, bob, pair, "RevertTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_revert_without_commit", func() {
			_, err = Exec(ctx, bob, pair, "RevertTransferOwnerWinner", []interface{}{})

			is, _ := Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))
		})

		It("test_apply", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

			_, err = Exec(ctx, alice, pair, "ApplyTransferOwnerWinner", []interface{}{})

			is, err := Exec(ctx, alice, pair, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			is, _ = Exec(ctx, alice, pair, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))
		})

		It("test_apply_only_owner", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

			_, err = Exec(ctx, bob, pair, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_apply_insufficient_time", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, pair, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay-5, aliceKey)

			_, err = Exec(ctx, alice, pair, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_TIME"))
		})

		It("test_apply_no_active_transfer", func() {
			_, err = Exec(ctx, alice, pair, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: NO_ACTIVE_TRANSFER"))
		})

	})

	Describe("StableSwap", func() {

		BeforeEach(func() {
			genesis = types.NewEmptyContext()
			stableTokens = DeployTokens(genesis, classMap["Token"], N, admin)
			cn, cdx, ctx, _ = initChain(genesis, admin)
			timestamp_second = uint64(time.Now().UnixNano()) / uint64(time.Second)
			ctx, _ = Sleep(cn, genesis, nil, timestamp_second, aliceKey)
			swap, _ = stablebase(ctx, stableBaseContruction())
		})

		AfterEach(func() {
			RemoveChain(cdx)
			afterEach()
		})

		It("test_commit", func() {
			delay := uint64(3 * 86400)
			_, err := Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})
			Expect(err).To(Succeed())

			is, err := Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(uint64)).To(Equal(ctx.LastTimestamp()/uint64(time.Second) + delay))

			is, err = Exec(ctx, alice, swap, "FutureOwner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

		})

		It("test_commit_only_owner", func() {
			_, err := Exec(ctx, bob, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_commit_already_active", func() {
			_, err := Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})

			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			Expect(err).To(MatchError("Exchange: ACTIVE_TRANSFER"))
		})

		It("test_revert", func() {
			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			_, err = Exec(ctx, alice, swap, "RevertTransferOwnerWinner", []interface{}{})

			is, _ := Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))

		})

		It("test_revert_only_owner", func() {
			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, uint64(86400)})
			_, err = Exec(ctx, bob, swap, "RevertTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_revert_without_commit", func() {
			_, err = Exec(ctx, bob, swap, "RevertTransferOwnerWinner", []interface{}{})

			is, _ := Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))
		})

		It("test_apply", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

			_, err = Exec(ctx, alice, swap, "ApplyTransferOwnerWinner", []interface{}{})

			is, err := Exec(ctx, alice, swap, "Owner", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(common.Address)).To(Equal(bob))

			is, _ = Exec(ctx, alice, swap, "TransferOwnerWinnerDeadline", []interface{}{})
			Expect(is[0].(uint64)).To(Equal(uint64(0)))
		})

		It("test_apply_only_owner", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay, aliceKey)

			_, err = Exec(ctx, bob, swap, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))
		})

		It("test_apply_insufficient_time", func() {
			delay := uint64(3 * 86400)
			_, err = Exec(ctx, alice, swap, "CommitTransferOwnerWinner", []interface{}{bob, charlie, delay})

			ctx, _ = Sleep(cn, ctx, nil, delay-5, aliceKey)

			_, err = Exec(ctx, alice, swap, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: INSUFFICIENT_TIME"))
		})

		It("test_apply_no_active_transfer", func() {
			_, err = Exec(ctx, alice, swap, "ApplyTransferOwnerWinner", []interface{}{})
			Expect(err).To(MatchError("Exchange: NO_ACTIVE_TRANSFER"))
		})

	})
})
