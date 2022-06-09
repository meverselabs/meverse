package test

import (
	"math/big"

	"github.com/meverselabs/meverse/common/amount"

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

var _ = Describe("Token", func() {

	var err error

	Describe("Uniswap", func() {

		BeforeEach(func() {
			beforeEachUni()
			lpTokenMint(genesis, pair, alice, _TotalSupply) // initial Token Supply
		})

		AfterEach(func() {
			afterEach()
		})

		It("name, symbol, decimals, totalSupply, balanceOf", func() {

			//Name
			is, err := Exec(genesis, alice, pair, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_PairName))

			//Symbol
			is, err = Exec(genesis, alice, pair, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_PairSymbol))

			//Decimals
			is, err = Exec(genesis, alice, pair, "Decimals", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(big.NewInt(18)))

			//Total Supply
			is, err = Exec(genesis, alice, pair, "TotalSupply", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			//BalanceOf  alice
			is, err = Exec(genesis, alice, pair, "BalanceOf", []interface{}{alice})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			//BalanceOf  bob
			balance, err := tokenBalanceOf(genesis, pair, bob)
			Expect(err).To(Succeed())
			Expect(balance.Cmp(big.NewInt(0)) == 0).To(BeTrue())
		})

		It("SetName, onlyOwner", func() {

			name := "__New_Name"

			is, err := Exec(genesis, alice, pair, "SetName", []interface{}{name})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, pair, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(name))

			is, err = Exec(genesis, bob, pair, "SetName", []interface{}{name})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetSymbol, onlyOwner", func() {

			symbol := "__New_Symbol"

			is, err := Exec(genesis, alice, pair, "SetSymbol", []interface{}{symbol})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, pair, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(symbol))

			is, err = Exec(genesis, bob, pair, "SetSymbol", []interface{}{symbol})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		Describe("Approve", func() {

			It("allowance without approve", func() {
				Expect(tokenAllowance(genesis, pair, alice, alice)).NotTo(Equal(_TestAmount))
			})

			It("test_initial_approval_is_zero", func() {
				for idx := 0; idx < 5; idx++ {
					Expect(tokenAllowance(genesis, pair, alice, users[idx])).To(Equal(ZeroAmount))
				}
			})

			It("test_approve", func() {
				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(_TestAmount))
			})

			It("test_modify_approve_nonzero", func() {
				/* 조건 삭제
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, amount.NewAmount(1, 0)})
				Expect(err).To(MatchError("LPToken: APPROVE_ALREADY_SET"))
				*/
			})

			It("test_modify_approve_zero_nonzero", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, ZeroAmount})
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, amount.NewAmount(0, 123456)})

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(amount.NewAmount(0, 123456)))
			})

			It("test_revoke_approve", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, ZeroAmount})

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(ZeroAmount))
			})

			It("test_approve_self", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{alice, _TestAmount})

				Expect(tokenAllowance(genesis, pair, alice, alice)).To(Equal(_TestAmount))
			})

			It("test_only_affects_target", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})

				Expect(tokenAllowance(genesis, pair, bob, alice)).To(Equal(ZeroAmount))
			})

			It("test_approval_event_fires", func() {
				// no event in Meverse
			})

			It("test_infinite_approval", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, MaxUint256})

				_, err = Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(MaxUint256))
			})

			It("test_increase_allowance", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, pair, "IncreaseAllowance", []interface{}{bob, amount.NewAmount(0, 403)})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(_TestAmount.Add(amount.NewAmount(0, 403))))
			})

			It("test_decrease_allowance", func() {
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, pair, "DecreaseAllowance", []interface{}{bob, amount.NewAmount(0, 403)})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(_TestAmount.Sub(amount.NewAmount(0, 403))))
			})

			It("test_apporve_negative_tokens", func() {
				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: APPROVE_NEGATIVE_AMOUNT"))
			})

			It("test_increase_negative_tokens", func() {
				_, err := Exec(genesis, alice, pair, "IncreaseAllowance", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: INCREASEALLOWANCE_NEGATIVE_AMOUNT"))
			})

			It("test_decrease_negative_tokens", func() {
				_, err := Exec(genesis, alice, pair, "DecreaseAllowance", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: DECREASEALLOWANCE_NEGATIVE_AMOUNT"))
			})
		})

		Describe("Mint", func() {
			// 외부에서 mint, burn을 할 수 없다
			// trade/z_lpToken_test.go 참조
		})

		Describe("Burn", func() {
			// 외부에서 mint, burn을 할 수 없다
			// trade/z_lpToken_test.go 참조
		})

		Describe("Transfer", func() {
			It("tranfer", func() {
				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(_TotalSupply.Sub(_TestAmount)))
				Expect(tokenBalanceOf(genesis, pair, bob)).To(Equal(_TestAmount))
			})

			It("test_sender_balance_decreases", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_receiver_balance_increases", func() {
				receiver_balance, _ := tokenBalanceOf(genesis, pair, bob)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, bob)).To(Equal(receiver_balance.Add(_TestAmount)))
			})

			It("test_total_supply_not_affected", func() {
				total_supply, _ := tokenTotalSupply(genesis, pair)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenTotalSupply(genesis, pair)).To(Equal(total_supply))
			})

			It("test_returns_true", func() {
				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
			})

			It("test_transfer_full_balance", func() {
				amount, _ := tokenBalanceOf(genesis, pair, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, pair, bob)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, amount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(genesis, pair, bob)).To(Equal(receiver_balance.Add(amount)))

			})

			It("test_transfer_zero_tokens", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, pair, bob)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, ZeroAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, pair, bob)).To(Equal(receiver_balance))

			})

			It("test_transfer_to_self", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{alice, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance))
			})

			It("test_insufficient_balance", func() {
				balance, _ := tokenBalanceOf(genesis, pair, bob)

				_, err := Exec(genesis, bob, pair, "Transfer", []interface{}{alice, Add(balance.Int, big.NewInt(1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))

				_, err = Exec(genesis, bob, pair, "Transfer", []interface{}{alice, amount.NewAmount(0, 1)})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))
			})

			It("test_transfer_event_fires", func() {
				// no event in Meverse
			})

			It("test_transfer_negative_tokens", func() {
				_, err := Exec(genesis, alice, pair, "Transfer", []interface{}{bob, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_NEGATIVE_AMOUNT"))
			})
		})

		Describe("TransferFrom", func() {
			It("test_sender_balance_decreases", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_receiver_balance_increases", func() {
				receiver_balance, _ := tokenBalanceOf(genesis, pair, charlie)

				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, charlie)).To(Equal(receiver_balance.Add(_TestAmount)))
			})

			It("test_caller_balance_not_affected", func() {
				caller_balance, _ := tokenBalanceOf(genesis, pair, bob)

				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, pair, bob)).To(Equal(caller_balance))
			})

			It("test_caller_approval_affected", func() {
				approval_amount, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, alice, pair, "Approve", []interface{}{bob, approval_amount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, pair, alice, bob)).To(Equal(approval_amount.Sub(_TestAmount)))
			})

			It("test_receiver_approval_not_affected", func() {
				approval_amount, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, approval_amount})
				Exec(genesis, alice, pair, "Approve", []interface{}{charlie, approval_amount})

				Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})

				Expect(tokenAllowance(genesis, pair, alice, charlie)).To(Equal(approval_amount))
			})

			It("test_returns_true", func() {
				amount, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, amount})
				_, err := Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, _TestAmount})

				Expect(err).To(Succeed())

			})

			It("test_transfer_full_balance", func() {
				amount, _ := tokenBalanceOf(genesis, pair, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, pair, charlie)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, amount})
				Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, amount})

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(genesis, pair, charlie)).To(Equal(receiver_balance.Add(amount)))
			})

			It("test_transfer_zero_tokens", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, pair, charlie)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, sender_balance})
				Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, ZeroAmount})

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, pair, charlie)).To(Equal(receiver_balance))
			})

			It("test_transfer_zero_tokens_without_approval", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, pair, charlie)

				Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, ZeroAmount})

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, pair, charlie)).To(Equal(receiver_balance))
			})

			It("test_insufficient_balance", func() {
				balance, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, balance.Add(amount.NewAmount(0, 1))})
				_, err := Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, balance.Add(amount.NewAmount(0, 1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))
			})

			It("test_insufficient_approval", func() {
				balance, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, balance.Sub(amount.NewAmount(0, 1))})
				_, err := Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_no_approval", func() {
				balance, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_revoked_approval", func() {
				balance, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{bob, balance})
				Exec(genesis, alice, pair, "Approve", []interface{}{bob, ZeroAmount})

				_, err := Exec(genesis, bob, pair, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_transfer_to_self", func() {
				sender_balance, _ := tokenBalanceOf(genesis, pair, alice)

				Exec(genesis, alice, pair, "Approve", []interface{}{alice, sender_balance})

				Exec(genesis, alice, pair, "TransferFrom", []interface{}{alice, alice, _TestAmount})

				Expect(tokenBalanceOf(genesis, pair, alice)).To(Equal(sender_balance))
				Expect(tokenAllowance(genesis, pair, alice, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_transfer_to_self_no_approval", func() {
				amount, _ := tokenBalanceOf(genesis, pair, alice)

				_, err := Exec(genesis, alice, pair, "TransferFrom", []interface{}{alice, alice, amount})

				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_transfer_event_fires", func() {
				// no event in Merverse
			})

			It("test_transfer_negative_tokens", func() {
				_, err := Exec(genesis, alice, pair, "TransferFrom", []interface{}{alice, alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_NEGATIVE_AMOUNT"))
			})
		})
	})

	Describe("StableSwap", func() {

		BeforeEach(func() {
			beforeEachStable()
			lpTokenMint(genesis, swap, alice, _TotalSupply) // initial Token Supply
		})

		AfterEach(func() {
			afterEach()
		})

		It("name, symbol, decimals, totalSupply, balanceOf", func() {

			//Name
			is, err := Exec(genesis, alice, swap, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapName))

			//Symbol
			is, err = Exec(genesis, alice, swap, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(_SwapSymbol))

			//Decimals
			is, err = Exec(genesis, alice, swap, "Decimals", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*big.Int)).To(Equal(big.NewInt(18)))

			//Total Supply
			is, err = Exec(genesis, alice, swap, "TotalSupply", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			//BalanceOf  alice
			is, err = Exec(genesis, alice, swap, "BalanceOf", []interface{}{alice})
			Expect(err).To(Succeed())
			Expect(is[0].(*amount.Amount)).To(Equal(_TotalSupply))

			//BalanceOf  bob
			balance, err := tokenBalanceOf(genesis, swap, bob)
			Expect(err).To(Succeed())
			Expect(balance.Cmp(big.NewInt(0)) == 0).To(BeTrue())
		})

		It("SetName, onlyOwner", func() {

			name := "__New_Name"

			is, err := Exec(genesis, alice, swap, "SetName", []interface{}{name})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, swap, "Name", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(name))

			is, err = Exec(genesis, bob, swap, "SetName", []interface{}{name})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		It("SetSymbol, onlyOwner", func() {

			symbol := "__New_Symbol"

			is, err := Exec(genesis, alice, swap, "SetSymbol", []interface{}{symbol})
			Expect(err).To(Succeed())

			is, err = Exec(genesis, alice, swap, "Symbol", []interface{}{})
			Expect(err).To(Succeed())
			Expect(is[0].(string)).To(Equal(symbol))

			is, err = Exec(genesis, bob, swap, "SetSymbol", []interface{}{symbol})
			Expect(err).To(MatchError("Exchange: FORBIDDEN"))

		})

		Describe("Approve", func() {

			It("allowance without approve", func() {
				Expect(tokenAllowance(genesis, swap, alice, alice)).NotTo(Equal(_TestAmount))
			})

			It("test_initial_approval_is_zero", func() {
				for idx := 0; idx < 5; idx++ {
					Expect(tokenAllowance(genesis, swap, alice, users[idx])).To(Equal(ZeroAmount))
				}
			})

			It("test_approve", func() {
				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(_TestAmount))
			})

			It("test_modify_approve_nonzero", func() {
				/* 조건 삭제
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, amount.NewAmount(1, 0)})
				Expect(err).To(MatchError("LPToken: APPROVE_ALREADY_SET"))
				*/
			})

			It("test_modify_approve_zero_nonzero", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, ZeroAmount})
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, amount.NewAmount(0, 123456)})

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(amount.NewAmount(0, 123456)))
			})

			It("test_revoke_approve", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, ZeroAmount})

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(ZeroAmount))
			})

			It("test_approve_self", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{alice, _TestAmount})

				Expect(tokenAllowance(genesis, swap, alice, alice)).To(Equal(_TestAmount))
			})

			It("test_only_affects_target", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})

				Expect(tokenAllowance(genesis, swap, bob, alice)).To(Equal(ZeroAmount))
			})

			It("test_approval_event_fires", func() {
				// no event in Meverse
			})

			It("test_infinite_approval", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, MaxUint256})

				_, err = Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(MaxUint256))
			})

			It("test_increase_allowance", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, swap, "IncreaseAllowance", []interface{}{bob, amount.NewAmount(0, 403)})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(_TestAmount.Add(amount.NewAmount(0, 403))))
			})

			It("test_decrease_allowance", func() {
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})

				_, err := Exec(genesis, alice, swap, "DecreaseAllowance", []interface{}{bob, amount.NewAmount(0, 403)})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(_TestAmount.Sub(amount.NewAmount(0, 403))))
			})

			It("test_apporve_negative_tokens", func() {
				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: APPROVE_NEGATIVE_AMOUNT"))
			})

			It("test_increase_negative_tokens", func() {
				_, err := Exec(genesis, alice, swap, "IncreaseAllowance", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: INCREASEALLOWANCE_NEGATIVE_AMOUNT"))
			})

			It("test_decrease_negative_tokens", func() {
				_, err := Exec(genesis, alice, swap, "DecreaseAllowance", []interface{}{alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: DECREASEALLOWANCE_NEGATIVE_AMOUNT"))
			})
		})

		Describe("Mint", func() {
			// 외부에서 mint, burn을 할 수 없다
			// trade/z_lpToken_test.go 참조
		})

		Describe("Burn", func() {
			// 외부에서 mint, burn을 할 수 없다
			// trade/z_lpToken_test.go 참조
		})

		Describe("Transfer", func() {
			It("tranfer", func() {
				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(_TotalSupply.Sub(_TestAmount)))
				Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(_TestAmount))
			})

			It("test_sender_balance_decreases", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_receiver_balance_increases", func() {
				receiver_balance, _ := tokenBalanceOf(genesis, swap, bob)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(receiver_balance.Add(_TestAmount)))
			})

			It("test_total_supply_not_affected", func() {
				total_supply, _ := tokenTotalSupply(genesis, swap)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenTotalSupply(genesis, swap)).To(Equal(total_supply))
			})

			It("test_returns_true", func() {
				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
			})

			It("test_transfer_full_balance", func() {
				amount, _ := tokenBalanceOf(genesis, swap, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, swap, bob)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, amount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(receiver_balance.Add(amount)))

			})

			It("test_transfer_zero_tokens", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, swap, bob)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, ZeroAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(receiver_balance))

			})

			It("test_transfer_to_self", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{alice, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance))
			})

			It("test_insufficient_balance", func() {
				balance, _ := tokenBalanceOf(genesis, swap, bob)

				_, err := Exec(genesis, bob, swap, "Transfer", []interface{}{alice, Add(balance.Int, big.NewInt(1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))

				_, err = Exec(genesis, bob, swap, "Transfer", []interface{}{alice, amount.NewAmount(0, 1)})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))
			})

			It("test_transfer_event_fires", func() {
				// no event in Meverse
			})

			It("test_transfer_negative_tokens", func() {
				_, err := Exec(genesis, alice, swap, "Transfer", []interface{}{bob, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_NEGATIVE_AMOUNT"))
			})
		})

		Describe("TransferFrom", func() {
			It("test_sender_balance_decreases", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_receiver_balance_increases", func() {
				receiver_balance, _ := tokenBalanceOf(genesis, swap, charlie)

				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, charlie)).To(Equal(receiver_balance.Add(_TestAmount)))
			})

			It("test_caller_balance_not_affected", func() {
				caller_balance, _ := tokenBalanceOf(genesis, swap, bob)

				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, _TestAmount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenBalanceOf(genesis, swap, bob)).To(Equal(caller_balance))
			})

			It("test_caller_approval_affected", func() {
				approval_amount, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, alice, swap, "Approve", []interface{}{bob, approval_amount})
				Expect(err).To(Succeed())
				_, err = Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})
				Expect(err).To(Succeed())

				Expect(tokenAllowance(genesis, swap, alice, bob)).To(Equal(approval_amount.Sub(_TestAmount)))
			})

			It("test_receiver_approval_not_affected", func() {
				approval_amount, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, approval_amount})
				Exec(genesis, alice, swap, "Approve", []interface{}{charlie, approval_amount})

				Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})

				Expect(tokenAllowance(genesis, swap, alice, charlie)).To(Equal(approval_amount))
			})

			It("test_returns_true", func() {
				amount, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, amount})
				_, err := Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, _TestAmount})

				Expect(err).To(Succeed())

			})

			It("test_transfer_full_balance", func() {
				amount, _ := tokenBalanceOf(genesis, swap, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, swap, charlie)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, amount})
				Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, amount})

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(ZeroAmount))
				Expect(tokenBalanceOf(genesis, swap, charlie)).To(Equal(receiver_balance.Add(amount)))
			})

			It("test_transfer_zero_tokens", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, swap, charlie)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, sender_balance})
				Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, ZeroAmount})

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, swap, charlie)).To(Equal(receiver_balance))
			})

			It("test_transfer_zero_tokens_without_approval", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)
				receiver_balance, _ := tokenBalanceOf(genesis, swap, charlie)

				Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, ZeroAmount})

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance))
				Expect(tokenBalanceOf(genesis, swap, charlie)).To(Equal(receiver_balance))
			})

			It("test_insufficient_balance", func() {
				balance, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, balance.Add(amount.NewAmount(0, 1))})
				_, err := Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, balance.Add(amount.NewAmount(0, 1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_BALANCE"))
			})

			It("test_insufficient_approval", func() {
				balance, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, balance.Sub(amount.NewAmount(0, 1))})
				_, err := Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_no_approval", func() {
				balance, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_revoked_approval", func() {
				balance, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{bob, balance})
				Exec(genesis, alice, swap, "Approve", []interface{}{bob, ZeroAmount})

				_, err := Exec(genesis, bob, swap, "TransferFrom", []interface{}{alice, charlie, balance})
				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_transfer_to_self", func() {
				sender_balance, _ := tokenBalanceOf(genesis, swap, alice)

				Exec(genesis, alice, swap, "Approve", []interface{}{alice, sender_balance})

				Exec(genesis, alice, swap, "TransferFrom", []interface{}{alice, alice, _TestAmount})

				Expect(tokenBalanceOf(genesis, swap, alice)).To(Equal(sender_balance))
				Expect(tokenAllowance(genesis, swap, alice, alice)).To(Equal(sender_balance.Sub(_TestAmount)))
			})

			It("test_transfer_to_self_no_approval", func() {
				amount, _ := tokenBalanceOf(genesis, swap, alice)

				_, err := Exec(genesis, alice, swap, "TransferFrom", []interface{}{alice, alice, amount})

				Expect(err).To(MatchError("LPToken: TRANSFER_EXCEED_ALLOWANCE"))
			})

			It("test_transfer_event_fires", func() {
				// no event in Merverse
			})

			It("test_transfer_negative_tokens", func() {
				_, err := Exec(genesis, alice, swap, "TransferFrom", []interface{}{alice, alice, ToAmount(big.NewInt(-1))})
				Expect(err).To(MatchError("LPToken: TRANSFER_NEGATIVE_AMOUNT"))
			})
		})
	})
})
