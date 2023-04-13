package test

import (
	"bytes"
	"log"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// evm-client/hardhat/test/erc20token-test.js 와 동일하게 테스트

// test : ginkgo
//        ginkgo -v  (verbose mode)
// skip : It("...", func() {
//           if(condition)  {
//			 	Skip("생략이유")
//           }
//         })
var _ = Describe("Erc20Token", func() {
	path := "_data_t"
	chainID := big.NewInt(65535)
	version := uint16(1)

	userKeys, _ := getSingers(chainID)
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	args := []interface{}{alice, bob, charlie} // alice(admin), bob, charlie

	// 체인생성
	tb, _ := prepare(path, true, chainID, version, &alice, args, mevInitialize, &initContextInfo{})
	defer removeChainData(tb.path)

	var provider types.Provider

	var token common.Address
	var initialSupply = amount.NewAmount(1, 0)

	// 1. Erc20Token Contract Deploy
	// 2. Erc20TokenWrapper Contract Deploy
	BeforeEach(func() {

		tx1, err := Erc20TokenContractCreationTx(tb, aliceKey, initialSupply)
		if err != nil {
			panic(err)
		}

		_, err = tb.addBlock([]*txWithSigner{{tx1, aliceKey}})
		if err != nil {
			panic(err)
		}

		provider = tb.chain.Provider()
		receipts, err := provider.Receipts(provider.Height())
		if err != nil {
			panic(err)
		}
		receipt := receipts[0]
		token = receipt.ContractAddress
	})

	// AfterEach(func() {
	// 	removeChainData(path)
	// })

	It("name, symbol, decimals, totalSupply, balanceOf", func() {

		//Name
		is, err := tb.view(token, "Name")
		Expect(err).To(Succeed())
		Expect(is[0].(string)).To(Equal("Gold"))

		//Symbol
		is, err = tb.view(token, "Symbol")
		Expect(err).To(Succeed())
		Expect(is[0].(string)).To(Equal("GLD"))

		//Decimals
		is, err = tb.view(token, "Decimals")
		Expect(err).To(Succeed())
		Expect(is[0].(*big.Int)).To(Equal(big.NewInt(18)))

		//Total Supply
		is, err = tb.view(token, "TotalSupply")
		Expect(err).To(Succeed())
		Expect(is[0].(*amount.Amount)).To(Equal(initialSupply))

		//BalanceOf  alice
		is, err = tb.view(token, "BalanceOf", alice)
		Expect(err).To(Succeed())
		Expect(is[0].(*amount.Amount)).To(Equal(initialSupply))

		//BalanceOf  bob
		is, err = tb.view(token, "BalanceOf", bob)
		Expect(err).To(Succeed())
		Expect(is[0].(*amount.Amount).Cmp(AmountZero.Int)).To(Equal(0))
	})

	Describe("approve", func() {
		It("intial approval is zero", func() {
			Expect(tokenAllowance(tb.ctx, token, alice, bob).Cmp(AmountZero.Int)).To(Equal(0))
		})

		It("approve", func() {
			amt := amount.NewAmount(10, 0)

			_, err := tb.call(aliceKey, token, "Approve", bob, amt)
			log.Printf("%+v", err)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(amt))
			Expect(tokenAllowance(tb.ctx, token, token, bob).Cmp(AmountZero.Int)).To(Equal(0))
		})

		It("modify approve", func() {
			amt := amount.NewAmount(0, 12345678)
			_, err := tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(10, 0))
			Expect(err).To(Succeed())
			_, err = tb.call(aliceKey, token, "Approve", bob, amt)
			Expect(err).To(Succeed())

			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(amt))
		})

		It("revoke approve", func() {
			tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(10, 0))
			tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(0, 0))

			Expect(tokenAllowance(tb.ctx, token, alice, bob).Cmp(AmountZero.Int)).To(Equal(0))
		})

		It("approve self", func() {
			amt := amount.NewAmount(10, 0)

			tb.call(aliceKey, token, "Approve", alice, amt)

			Expect(tokenAllowance(tb.ctx, token, alice, alice)).To(Equal(amt))
		})

		It("only affects target", func() {
			tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(10, 0))

			Expect(tokenAllowance(tb.ctx, token, bob, alice).Cmp(AmountZero.Int)).To(Equal(0))
		})

		It("approve with event", func() {
			amt := amount.NewAmount(10, 0)
			_, err := tb.call(aliceKey, token, "Approve", bob, amt)
			Expect(err).To(Succeed())

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[1].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("Approve"))
			Expect(mc.Args[0].(common.Address)).To(Equal(bob))
			Expect(mc.Args[1].(*amount.Amount)).To(Equal(amt))
		})

		It("infinite approval", func() {
			tb.call(aliceKey, token, "Approve", bob, MaxUint256)
			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(MaxUint256))

			tb.call(bobKey, token, "TransferFrom", alice, bob, amount.NewAmount(1, 0))
			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(MaxUint256))
		})
	})

	Describe("transfer", func() {
		It("sender balance decrease", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			amt := senderBalance.DivC(4)

			_, err := tb.call(aliceKey, token, "Transfer", bob, amt)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance.Sub(amt)))
		})

		It("receiver balance increase", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, bob)
			amt := senderBalance.DivC(4)

			_, err := tb.call(aliceKey, token, "Transfer", bob, amt)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(receiverBalance.Add(amt)))
		})

		It("total supply not affected", func() {
			totalSupply := tokenTotalSupply(tb.ctx, token)
			balance := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Transfer", bob, balance)

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(totalSupply))
		})

		It("transfer full balance", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, bob)

			tb.call(aliceKey, token, "Transfer", bob, senderBalance)

			Expect(tokenBalanceOf(tb.ctx, token, alice).Cmp(AmountZero.Int)).To(Equal(0))
			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(receiverBalance.Add(senderBalance)))
		})

		It("transfer zero token", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, bob)

			tb.call(aliceKey, token, "Transfer", bob, AmountZero)

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance))
			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(receiverBalance))
		})

		It("transfer to self", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			amt := senderBalance.DivC(4)

			tb.call(aliceKey, token, "Transfer", alice, amt)

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance))
		})

		It("fail if insufficient balance", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			_, err := tb.call(aliceKey, token, "Transfer", bob, senderBalance.Add(amount.NewAmount(0, 1)))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("transfer with event", func() {
			amt := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Transfer", bob, amt)

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[1].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("Transfer"))
			Expect(mc.Args[0].(common.Address)).To(Equal(bob))
			Expect(mc.Args[1].(*amount.Amount)).To(Equal(amt))
		})
	})

	Describe("transferFrom", func() {
		It("sender balance decrease", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			amt := senderBalance.DivC(4)

			_, err := tb.call(aliceKey, token, "Approve", bob, amt)
			Expect(err).To(Succeed())

			_, err = tb.call(bobKey, token, "TransferFrom", alice, charlie, amt)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance.Sub(amt)))
		})

		It("receiver balance increase", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			amt := senderBalance.DivC(4)
			receiverBalance := tokenBalanceOf(tb.ctx, token, charlie)

			tb.call(aliceKey, token, "Approve", bob, amt)

			_, err := tb.call(bobKey, token, "TransferFrom", alice, charlie, amt)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, charlie)).To(Equal(receiverBalance.Add(amt)))
		})

		It("caller balance not affected", func() {
			callerBalance := tokenBalanceOf(tb.ctx, token, bob)
			amt := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, amt)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, amt)

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(callerBalance))
		})

		It("caller approval affected", func() {
			approvalAmount := tokenBalanceOf(tb.ctx, token, alice)
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			transferAmount := senderBalance.DivC(4)

			tb.call(aliceKey, token, "Approve", bob, approvalAmount)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, transferAmount)

			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(approvalAmount.Sub(transferAmount)))

		})

		It("receiver approval not affected", func() {
			approvalAmount := tokenBalanceOf(tb.ctx, token, alice)
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			transferAmount := senderBalance.DivC(4)

			tb.call(aliceKey, token, "Approve", bob, approvalAmount)
			tb.call(aliceKey, token, "Approve", charlie, approvalAmount)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, transferAmount)

			Expect(tokenAllowance(tb.ctx, token, alice, charlie)).To(Equal(approvalAmount))
		})

		It("total supply not affected", func() {
			totalSupply := tokenBalanceOf(tb.ctx, token, bob)
			amt := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, amt)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, amt)

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(totalSupply))
		})

		It("transfer full balance", func() {
			amt := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, charlie)

			tb.call(aliceKey, token, "Approve", bob, amt)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, amt)

			Expect(tokenBalanceOf(tb.ctx, token, alice).Cmp(AmountZero.Int)).To(Equal(0))
			Expect(tokenBalanceOf(tb.ctx, token, charlie)).To(Equal(receiverBalance.Add(amt)))
		})

		It("transfer zero tokens", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, charlie)

			tb.call(aliceKey, token, "Approve", bob, senderBalance)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, AmountZero)

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance))
			Expect(tokenBalanceOf(tb.ctx, token, charlie)).To(Equal(receiverBalance))
		})

		It("transfer zero tokens without approval", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			receiverBalance := tokenBalanceOf(tb.ctx, token, charlie)

			tb.call(bobKey, token, "TransferFrom", alice, charlie, AmountZero)

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance))
			Expect(tokenBalanceOf(tb.ctx, token, charlie)).To(Equal(receiverBalance))
		})

		It("fail if insufficient balance", func() {
			balance := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, balance.Add(amount.NewAmount(0, 1)))
			_, err := tb.call(bobKey, token, "TransferFrom", alice, charlie, balance.Add(amount.NewAmount(0, 1)))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail if insufficient approval", func() {
			balance := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, balance.Sub(amount.NewAmount(0, 1)))

			_, err := tb.call(bobKey, token, "TransferFrom", alice, charlie, balance)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail if no approval", func() {
			balance := tokenBalanceOf(tb.ctx, token, alice)
			_, err := tb.call(bobKey, token, "TransferFrom", alice, charlie, balance)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail if revoke approval", func() {
			balance := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, balance)
			tb.call(aliceKey, token, "Approve", bob, AmountZero)

			_, err := tb.call(bobKey, token, "TransferFrom", alice, charlie, balance)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("transfer to self", func() {
			senderBalance := tokenBalanceOf(tb.ctx, token, alice)
			amt := senderBalance.DivC(4)

			tb.call(aliceKey, token, "Approve", alice, senderBalance)
			tb.call(aliceKey, token, "TransferFrom", alice, alice, amt)

			Expect(tokenBalanceOf(tb.ctx, token, alice)).To(Equal(senderBalance))
			Expect(tokenAllowance(tb.ctx, token, alice, alice)).To(Equal(senderBalance.Sub(amt)))
		})

		It("fail if transfer to self no approval", func() {
			amt := tokenBalanceOf(tb.ctx, token, alice)

			_, err := tb.call(aliceKey, token, "TransferFrom", alice, alice, amt)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("transferFrom with event", func() {
			balance := tokenBalanceOf(tb.ctx, token, alice)

			tb.call(aliceKey, token, "Approve", bob, balance)
			tb.call(bobKey, token, "TransferFrom", alice, charlie, balance)

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[1].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("TransferFrom"))
			Expect(mc.Args[0].(common.Address)).To(Equal(alice))
			Expect(mc.Args[1].(common.Address)).To(Equal(charlie))
			Expect(mc.Args[2].(*amount.Amount)).To(Equal(balance))
		})
	})

	Describe("mint burn", func() {
		minter := charlie
		minterKey := charlieKey
		BeforeEach(func() {
			minter = charlie
			_, err := tb.call(aliceKey, token, "SetMinter", minter, true)
			Expect(err).To(Succeed())
		})

		It("assumptions", func() {
			Expect(tokenTotalSupply(tb.ctx, token)).To(Equal(tokenBalanceOf(tb.ctx, token, alice)))

			Expect(tokenBalanceOf(tb.ctx, token, bob).Cmp(AmountZero.Int)).To(Equal(0))
			Expect(tokenBalanceOf(tb.ctx, token, minter).Cmp(AmountZero.Int)).To(Equal(0))
		})

		It("setMinter true", func() {
			_, err := tb.call(aliceKey, token, "SetMinter", bob, true)
			Expect(err).To(Succeed())
		})

		It("fail if set minter true twice", func() {
			_, err := tb.call(aliceKey, token, "SetMinter", minter, true)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("setMinter false", func() {
			_, err := tb.call(aliceKey, token, "SetMinter", minter, false)
			Expect(err).To(Succeed())
		})

		It("fail set minter false if the given address is not a minter", func() {
			_, err := tb.call(aliceKey, token, "SetMinter", bob, false)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("isMinter true", func() {
			Expect(tokenIsMinter(tb.ctx, token, minter)).To(BeTrue())
		})

		It("isMinter false", func() {
			Expect(tokenIsMinter(tb.ctx, token, bob)).To(BeFalse())
		})

		It("fail if caller doesn't have admin role", func() {
			_, err := tb.call(bobKey, token, "SetMinter", minter, true)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("mint affects balance", func() {
			amt := amount.NewAmount(0, 12345678)
			_, err := tb.call(minterKey, token, "Mint", bob, amt)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(amt))
		})

		It("mint affects totalSupply", func() {
			totalSupply := tokenTotalSupply(tb.ctx, token)
			amt := amount.NewAmount(0, 12345678)
			_, err := tb.call(minterKey, token, "Mint", bob, amt)
			Expect(err).To(Succeed())

			Expect(tokenTotalSupply(tb.ctx, token)).To(Equal(totalSupply.Add(amt)))
		})

		It("fail if mint overflow", func() {
			amt := MaxUint256.Sub(tokenTotalSupply(tb.ctx, token)).Add(amount.NewAmount(0, 1))

			_, err := tb.call(minterKey, token, "Mint", bob, amt)
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail mint without minter role", func() {
			_, err := tb.call(bobKey, token, "Mint", bob, amount.NewAmount(0, 1))
			Expect(err).To(MatchError("execution reverted"))

		})

		It("fail if mint to zero address", func() {
			_, err := tb.call(minterKey, token, "Mint", AddressZero, amount.NewAmount(0, 1))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail mint after set minter false", func() {
			tb.call(aliceKey, token, "SetMinter", minter, false)

			_, err := tb.call(minterKey, token, "Mint", bob, amount.NewAmount(0, 1))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("mint with event", func() {
			mintAmount := amount.NewAmount(0, 12345678)

			tb.call(minterKey, token, "Mint", bob, mintAmount)

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[0].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("Mint"))
			Expect(mc.Args[0].(common.Address)).To(Equal(bob))
			Expect(mc.Args[1].(*amount.Amount)).To(Equal(mintAmount))
		})

		It("burn affects balance", func() {
			burnAmount := amount.NewAmount(0, 1000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			balance := tokenBalanceOf(tb.ctx, token, bob)
			_, err := tb.call(bobKey, token, "Burn", burnAmount)
			Expect(err).To(Succeed())

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(balance.Sub(burnAmount)))
		})

		It("burn affects totalSupply", func() {
			totalSupply := tokenTotalSupply(tb.ctx, token)
			burnAmount := amount.NewAmount(0, 1000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			_, err := tb.call(bobKey, token, "Burn", burnAmount)
			Expect(err).To(Succeed())

			Expect(tokenTotalSupply(tb.ctx, token)).To(Equal(totalSupply.Sub(burnAmount)))
		})

		It("fail if burn underflow", func() {
			burnAmount := amount.NewAmount(0, 1000)
			tb.call(aliceKey, token, "Transfer", bob, burnAmount)

			_, err := tb.call(bobKey, token, "Burn", burnAmount.Add(amount.NewAmount(0, 1)))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("burn with event", func() {
			burnAmount := amount.NewAmount(0, 10000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			tb.call(bobKey, token, "Burn", burnAmount)

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[0].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("Burn"))
			Expect(mc.Args[0].(*amount.Amount)).To(Equal(burnAmount))
		})

		It("burnFrom affects balance", func() {
			burnAmount := amount.NewAmount(0, 10000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			balance := tokenBalanceOf(tb.ctx, token, bob)

			tb.call(minterKey, token, "BurnFrom", bob, burnAmount)

			Expect(tokenBalanceOf(tb.ctx, token, bob)).To(Equal(balance.Sub(burnAmount)))
		})

		It("burnFrom affects totalSupply", func() {
			totalSupply := tokenTotalSupply(tb.ctx, token)
			burnAmount := amount.NewAmount(0, 1000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			_, err := tb.call(minterKey, token, "BurnFrom", bob, burnAmount)
			Expect(err).To(Succeed())

			Expect(tokenTotalSupply(tb.ctx, token)).To(Equal(totalSupply.Sub(burnAmount)))
		})

		It("fail if burn underflow", func() {
			burnAmount := amount.NewAmount(0, 1000)
			tb.call(aliceKey, token, "Transfer", bob, burnAmount)

			_, err := tb.call(minterKey, token, "BurnFrom", bob, burnAmount.Add(amount.NewAmount(0, 1)))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail if burnFrom is not minter", func() {
			_, err := tb.call(bobKey, token, "BurnFrom", bob, amount.NewAmount(0, 1))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("fail if burnFrom zero address", func() {
			_, err := tb.call(bobKey, token, "BurnFrom", AddressZero, amount.NewAmount(0, 1))
			Expect(err).To(MatchError("execution reverted"))
		})

		It("burnFrom with event", func() {
			burnAmount := amount.NewAmount(0, 10000)

			tb.call(aliceKey, token, "Transfer", bob, burnAmount)
			tb.call(minterKey, token, "BurnFrom", bob, burnAmount)

			b, err := provider.Block(provider.Height())
			Expect(err).To(Succeed())

			mc := &ctypes.MethodCallEvent{}
			_, err = mc.ReadFrom(bytes.NewReader(b.Body.Events[0].Result))
			Expect(err).To(Succeed())

			Expect(mc.Method).To(Equal("BurnFrom"))
			Expect(mc.Args[0].(common.Address)).To(Equal(bob))
			Expect(mc.Args[1].(*amount.Amount)).To(Equal(burnAmount))
		})
	})

	FDescribe("not exist abi method", func() {

		It("increase allowance", func() {
			tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(0, 100))
			ret, err := tb.call(aliceKey, token, "IncreaseAllowance", bob, amount.NewAmount(0, 403))
			Expect(err).To(Succeed())

			Expect(ret[0].([]byte)).To(Equal(math.PaddedBigBytes(ecommon.Big1, 32)))
			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(amount.NewAmount(0, 503)))
		})

		It("decrease allowance", func() {
			tb.call(aliceKey, token, "Approve", bob, amount.NewAmount(0, 100))
			ret, err := tb.call(aliceKey, token, "DecreaseAllowance", bob, amount.NewAmount(0, 34))
			Expect(err).To(Succeed())

			Expect(ret[0].([]byte)).To(Equal(math.PaddedBigBytes(ecommon.Big1, 32)))
			Expect(tokenAllowance(tb.ctx, token, alice, bob)).To(Equal(amount.NewAmount(0, 66)))
		})

		It("not exist function", func() {
			_, err := tb.call(aliceKey, token, "NotExistFunction", bob, amount.NewAmount(0, 34))
			Expect(err).To(MatchError("execution reverted"))
		})

	})

})
