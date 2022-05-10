package trade

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Token", func() {

	var (
		cc    *types.ContractContext
		token *LPToken

		_TOTAL_SUPPLY   = amount.NewAmount(10000, 0)
		_TEST_AMOUNT    = big.NewInt(12345678)
		_TEST_ADDRESSES = []common.Address{
			common.HexToAddress("0x1000000000000000000000000000000000000000"),
			common.HexToAddress("0x2000000000000000000000000000000000000000")}
	)

	BeforeEach(func() {
		genesis = types.NewEmptyContext()

		token = &LPToken{}

		stableBaseContruction := &StableSwapConstruction{
			Name:         "__STABLE_NAME",
			Symbol:       "__STABLE_SYMBOL",
			NTokens:      uint8(2),
			Tokens:       _TEST_ADDRESSES,
			PayToken:     _TEST_ADDRESSES[1],
			Owner:        admin,
			Winner:       bob,
			Fee:          uint64(30000000),
			AdminFee:     uint64(2000000000),
			Amp:          big.NewInt(10000),
			PrecisionMul: []uint64{10000, 1000000},
			Rates:        []*big.Int{big.NewInt(10000), big.NewInt(1000000)},
		}

		bs, _, err := bin.WriterToBytes(stableBaseContruction)
		Expect(err).To(Succeed())
		v, err := genesis.DeployContract(admin, classMap["StableSwap"], bs)
		Expect(err).To(Succeed())

		stableBaseContract := v.(*StableSwap)
		cc = genesis.ContractContext(stableBaseContract, admin)
		intr := types.NewInteractor(genesis, stableBaseContract, cc, "000000000000", false)
		cc.Exec = intr.Exec

		// initial Tokken Supply
		tagTokenTotalSupply := byte(0x03)
		tagTokenAmount := byte(0x04)

		cc.SetAccountData(alice, []byte{tagTokenAmount}, _TOTAL_SUPPLY.Bytes())
		cc.SetContractData([]byte{tagTokenTotalSupply}, _TOTAL_SUPPLY.Bytes())

	})

	AfterEach(func() {
		cleanUp()
	})

	Describe("Mint", func() {
		It("test_assumptions", func() {
			Expect(token.totalSupply(cc)).To(Equal(token.balanceOf(cc, alice)))
			Expect(token.balanceOf(cc, bob)).To(Equal(Zero))
		})

		It("test_set_minter", func() {
			// minter 따로 없음
		})

		It("test_only_minter", func() {
			// minter 따로 없음
		})

		It("test_transferFrom_without_approval", func() {
			// minter 따로 없음
		})

		It("test_mint_negative", func() {
			err := token._mint(cc, bob, big.NewInt(-1))
			Expect(err).To(MatchError("LPToken: MINT_NEGATIVE_AMOUNT"))
		})

		It("test_mint_affects_balance", func() {
			token._mint(cc, bob, _TEST_AMOUNT)

			Expect(token.balanceOf(cc, bob)).To(Equal(_TEST_AMOUNT))
		})

		It("test_mint_affects_totalSupply", func() {
			total_supply := token.totalSupply(cc)
			token._mint(cc, bob, _TEST_AMOUNT)

			Expect(token.totalSupply(cc)).To(Equal(Add(total_supply, _TEST_AMOUNT)))
		})

		It("test_mint_overflow", func() {
			// golang big.Int has no overflow
		})

		It("test_mint_not_minter", func() {
			// minter 따로 없음
		})

		It("test_mint_zero_address", func() {
			// zero_address 에 _mint 가능 : uniswap에서는 zero_address로 mint 함
		})
	})

	Describe("Burn", func() {
		It("test_assumptions", func() {
			Expect(token.totalSupply(cc)).To(Equal(token.balanceOf(cc, alice)))
			Expect(token.balanceOf(cc, bob)).To(Equal(Zero))
		})

		It("test_burn_negative", func() {
			err := token._burn(cc, bob, big.NewInt(-1))
			Expect(err).To(MatchError("LPToken: BURN_NEGATIVE_AMOUNT"))
		})

		It("test_burn_affects_totalSupply", func() {
			total_supply := token.totalSupply(cc)
			token._burn(cc, alice, _TEST_AMOUNT)

			Expect(token.totalSupply(cc)).To(Equal(Sub(total_supply, _TEST_AMOUNT)))
		})

		It("test_burn_underflow", func() {
			total_supply := token.totalSupply(cc)
			err := token._burn(cc, alice, Add(total_supply, big.NewInt(1)))

			Expect(err).To(MatchError("LPToken: BURN_EXCEED_BALANCE"))
		})

		It("test_burn_not_minter", func() {
			// minter 따로 없음
		})

		It("test_burn_zero_address", func() {
			// zero_address 에 _burn 제약 없음
		})

		It("test_burn_returns_true", func() {
			err := token._burn(cc, alice, _TEST_AMOUNT)
			Expect(err).To(Succeed())
		})

		It("test_mint_returns_true", func() {
			// _mint no return
		})
	})
})
