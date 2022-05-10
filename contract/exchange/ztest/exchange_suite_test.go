package test

import (
	"math"
	"math/big"
	"os"
	"sync"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	uniswap
	files :
		v2-core/test/UniswapV2ERC20.spec.ts
		v2-core/test/UniswapV2Factory.spec.ts
		v2-core/test/UniswapV2Pair.spec.ts
		v2-perphery/test/UniswapV2Router01.spec.ts


	stable : curve-contract
	files :
		tests/forked/test_gas.py
		tests/forked/test_insufficient_balances.py
		tests/pools/common/integration/test_curve.py
		tests/pools/common/integration/test_heavily_imbalanced.py
		tests/pools/common/integration/test_virtual_price_increases.py
		tests/pools/common/unitary/test_add_liquidity.py
		tests/pools/common/unitary/test_add_liquidity_initial.py
		tests/pools/common/unitary/test_claim_fees.py
		tests/pools/common/unitary/test_exchange.py
		tests/pools/common/unitary/test_exchange_reverts.py
		tests/pools/common/unitary/test_get_virtual_price.py
		tests/pools/common/unitary/test_kill.py
		tests/pools/common/unitary/test_modify_fees.py
		tests/pools/common/unitary/test_nonpayable.py
		tests/pools/common/unitary/test_ramp_A_precise.py
		tests/pools/common/unitary/test_remove_liquidity.py
		tests/pools/common/unitary/test_remove_liquidity_imbalance.py
		tests/pools/common/unitary/test_remove_liquidity_one_coin.py
		tests/pools/common/unitary/test_transfer_ownership.py
		tests/pools/common/unitary/test_xfer_to_contract.py

		tests/zaps  : skip
*/

func TestExchange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StableSwap Suite")
}

var (
	mutex    = &sync.Mutex{}
	classMap = map[string]uint64{}
	chainIdx = int(1)

	adminKey key.Key
	admin    common.Address
	usersKey []key.Key
	users    []common.Address

	genesis *types.Context

	//chain
	mainToken common.Address //gas

	// exchange
	factoryAddr, routerAddr      common.Address
	alice, bob, charlie, eve     common.Address // fixture : alice, bob, charlie, eve
	aliceKey, bobKey, charlieKey key.Key

	_Fee       = uint64(40000000)
	_AdminFee  = uint64(trade.MAX_ADMIN_FEE)
	_Fee30     = uint64(30000000)   // 30bp
	_AdminFee6 = uint64(1666666667) // 1/6 = uint64(1666666667)
	_WinnerFee = uint64(5000000000)

	_WhiteList common.Address
	_GroupId   = hash.BigToHash(big.NewInt(1))

	// token
	_TotalSupply = amount.NewAmount(10000, 0)
	_TestAmount  = amount.NewAmount(10, 0)

	// uniswap
	pair           common.Address
	uniTokens      []common.Address
	token0, token1 common.Address
	_ML            = amount.NewAmount(0, trade.MINIMUM_LIQUIDITY)
	_PairName      = "__UNI_NAME"
	_PairSymbol    = "__UNI_SYMBOL"
	_SupplyTokens  = []*amount.Amount{amount.NewAmount(500000, 0), amount.NewAmount(1000000, 0)}

	//stableswap
	stableTokens []common.Address  // fixture : wrapped_coins
	decimals     = []int{18, 6, 6} // fixture : wrapped_decimals, underlying_decimals
	//decimals        = []int{18, 18}
	N               = uint8(len(decimals)) // fixture : n_coins
	swap            common.Address         // fixture : swap
	_SwapName       = "__STABLE_NAME"
	_SwapSymbol     = "__STABLE_SYMBOL"
	_Amp            = int64(360 * 2)
	_PrecisionMul   []uint64
	_Rates          []*big.Int
	_BaseAmount     = int64(1000000) // fixture : base_amount
	_InitialAmounts []*amount.Amount // fixture : initial_amounts

	// test
	_MaxIter = 10
)

var _ = BeforeSuite(func() {

	classID, _ := types.RegisterContractType(&token.TokenContract{})
	classMap["Token"] = classID
	classID, _ = types.RegisterContractType(&factory.FactoryContract{})
	classMap["Factory"] = classID
	classID, _ = types.RegisterContractType(&trade.UniSwap{})
	classMap["UniSwap"] = classID
	classID, _ = types.RegisterContractType(&trade.StableSwap{})
	classMap["StableSwap"] = classID
	classID, _ = types.RegisterContractType(&router.RouterContract{})
	classMap["Router"] = classID
	classID, _ = types.RegisterContractType(&whitelist.WhiteListContract{})
	classMap["WhiteList"] = classID

	adminKey, admin, usersKey, users, _ = Accounts()
	alice, bob, charlie, eve = users[0], users[1], users[2], users[3]
	aliceKey, bobKey, charlieKey = usersKey[0], usersKey[1], usersKey[2]

	//1e6 of each coin - used to make an even initial deposit in many test setups
	_InitialAmounts = MakeAmountSlice(N)
	for k := uint8(0); k < N; k++ {
		_InitialAmounts[k].Set(Mul(Exp(big.NewInt(10), big.NewInt(int64(decimals[k]))), big.NewInt(_BaseAmount)))
	}
	// _load_pool_data in curve-contract/brownie_hooks.py
	_PrecisionMul = make([]uint64, N, N)
	_Rates = make([]*big.Int, N, N)
	for k := uint8(0); k < N; k++ {
		_PrecisionMul[k] = uint64(trade.PRECISION / int64(math.Pow10(decimals[k])))
		_Rates[k] = MulC(big.NewInt(int64(_PrecisionMul[k])), trade.PRECISION)
	}
})

// 디렉토리 전체 삭제
var _ = AfterSuite(func() {
	dir := "chain"
	if _, err := os.Stat("/mnt/ramdisk"); !os.IsNotExist(err) {
		dir = "/mnt/ramdisk/" + dir
	}
	err := os.RemoveAll(dir)
	Expect(err).To(Succeed())
})

func beforeEachWithoutTokens() error {
	genesis = types.NewEmptyContext()

	// factory
	factoryConstrunction := &factory.FactoryContractConstruction{Owner: admin}
	bs, _, err := bin.WriterToBytes(factoryConstrunction)
	if err != nil {
		return err
	}
	v, err := genesis.DeployContract(admin, classMap["Factory"], bs)
	if err != nil {
		return err
	}
	factoryAddr = v.(*factory.FactoryContract).Address()

	// router
	routerConstrunction := &router.RouterContractConstruction{Factory: factoryAddr}
	bs, _, err = bin.WriterToBytes(routerConstrunction)
	if err != nil {
		return err
	}
	v, err = genesis.DeployContract(admin, classMap["Router"], bs)
	if err != nil {
		return err
	}
	routerAddr = v.(*router.RouterContract).Address()

	// whitelist
	whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
	bs, _, err = bin.WriterToBytes(whitelistConstrunction)
	if err != nil {
		return err
	}
	v, err = genesis.DeployContract(admin, classMap["WhiteList"], bs)
	if err != nil {
		return err
	}
	_WhiteList = v.(*whitelist.WhiteListContract).Address()

	return nil
}

func beforeEach() error {
	if err := beforeEachWithoutTokens(); err != nil {
		return err
	}
	deployInitialTokens()
	return nil
}

func deployInitialTokens() {
	var err error
	uniTokens = DeployTokens(genesis, classMap["Token"], 2, admin)
	token0, token1, err = trade.SortTokens(uniTokens[0], uniTokens[1])
	if err != nil {
		panic(err)
	}
	stableTokens = DeployTokens(genesis, classMap["Token"], N, admin)
}

func beforeEachUni() error {
	if err := beforeEach(); err != nil {
		return err
	}
	//pair
	is, err := Exec(genesis, admin, factoryAddr, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], ZeroAddress, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, _WhiteList, _GroupId, classMap["UniSwap"]})
	if err != nil {
		return err
	}
	pair = is[0].(common.Address)

	return nil
}

func beforeEachStable() error {
	if err := beforeEach(); err != nil {
		return err
	}
	var err error
	swap, err = stablebase(genesis, stableBaseContruction())
	if err != nil {
		return err
	}
	return nil
}

func afterEach() {
	cleanUp()
}

func cleanUp() {
	genesis = nil
}
func stableBaseContruction() *trade.StableSwapConstruction {
	// swap
	return &trade.StableSwapConstruction{
		Name:         _SwapName,
		Symbol:       _SwapSymbol,
		Factory:      ZeroAddress,
		NTokens:      uint8(N),
		Tokens:       stableTokens,
		PayToken:     ZeroAddress,
		Owner:        alice,
		Winner:       charlie,
		Fee:          _Fee,
		AdminFee:     _AdminFee,
		WinnerFee:    _WinnerFee,
		WhiteList:    _WhiteList,
		GroupId:      _GroupId,
		Amp:          big.NewInt(_Amp),
		PrecisionMul: _PrecisionMul,
		Rates:        _Rates,
	}
}

func stablebase(ctx *types.Context, sbc *trade.StableSwapConstruction) (common.Address, error) {
	bs, _, err := bin.WriterToBytes(sbc)
	if err != nil {
		return ZeroAddress, err
	}
	v, err := ctx.DeployContract(admin, classMap["StableSwap"], bs)
	if err != nil {
		return ZeroAddress, err
	}
	stableBaseContract := v.(*trade.StableSwap)
	return stableBaseContract.Address(), nil
}

/////////// fixtures  ///////////

// token mint : minter = admin
func tokenMint(ctx *types.Context, token, to common.Address, amt *amount.Amount) error {
	_, err := Exec(ctx, admin, token, "Mint", []interface{}{to, amt})
	if err != nil {
		return err
	}
	return nil
}
func uniMint(ctx *types.Context, to common.Address) error {
	if err := tokenMint(ctx, token0, to, _SupplyTokens[0]); err != nil {
		return err
	}
	if err := tokenMint(ctx, token1, to, _SupplyTokens[1]); err != nil {
		return err
	}
	return nil
}
func stableMint(ctx *types.Context, to common.Address) error {
	for k := uint8(0); k < N; k++ {
		if err := tokenMint(ctx, stableTokens[k], to, _InitialAmounts[k]); err != nil {
			return err
		}
	}
	return nil
}

// LP token mint - 외부에서 mint하는 함수가 없어 직접 db에 기입
func lpTokenMint(ctx *types.Context, token, to common.Address, am *amount.Amount) error {
	cc, err := GetCC(ctx, token, admin)
	if err != nil {
		return err
	}
	totalSupply, err := TokenTotalSupply(cc, token)
	if err != nil {
		return err
	}
	balance, err := TokenBalanceOf(cc, token, to)
	if err != nil {
		return err
	}

	tagTokenTotalSupply := byte(0x03)
	tagTokenAmount := byte(0x04)

	cc.SetAccountData(to, []byte{tagTokenAmount}, ToAmount(balance).Add(am).Bytes())
	cc.SetContractData([]byte{tagTokenTotalSupply}, ToAmount(totalSupply).Add(am).Bytes())
	return nil
}
func tokenTotalSupply(ctx *types.Context, token common.Address) (*amount.Amount, error) {
	cc, err := GetCC(ctx, token, admin)
	if err != nil {
		return nil, err
	}

	result, err := TokenTotalSupply(cc, token)
	return ToAmount(result), err
}
func tokenBalanceOf(ctx *types.Context, token, from common.Address) (*amount.Amount, error) {
	cc, err := GetCC(ctx, token, admin)
	if err != nil {
		return nil, err
	}

	result, err := TokenBalanceOf(cc, token, from)
	return ToAmount(result), err
}
func tokenAllowance(ctx *types.Context, token, owner, spender common.Address) (*amount.Amount, error) {
	cc, err := GetCC(ctx, token, admin)
	if err != nil {
		return nil, err
	}
	result, err := TokenAllowance(cc, token, owner, spender)
	return ToAmount(result), err
}
func safeTransfer(ctx *types.Context, signer, token, to common.Address, am *amount.Amount) error {
	cc, err := GetCC(ctx, token, signer)
	if err != nil {
		return err
	}
	return SafeTransfer(cc, token, to, am.Int)
}
func safeTransferFrom(ctx *types.Context, signer, token, from, to common.Address, am *amount.Amount) error {
	cc, err := GetCC(ctx, token, signer)
	if err != nil {
		return err
	}
	err = SafeTransferFrom(cc, token, from, to, am.Int)
	return err
}
func stableTokenBalances(ctx *types.Context, from common.Address) ([]*amount.Amount, error) {
	balances := MakeAmountSlice(N)
	for k, token := range stableTokens {
		bal, err := tokenBalanceOf(ctx, token, from)
		if err != nil {
			return nil, err
		}
		balances[k].Set(bal.Int)
	}
	return balances, nil
}

func tokenApprove(ctx *types.Context, token, owner, spender common.Address) error {
	cc, err := GetCC(ctx, token, owner)
	if err != nil {
		return err
	}
	return TokenApprove(cc, token, spender, MaxUint256.Int)
}

func uniApprove(ctx *types.Context, owner common.Address) error {
	if err := tokenApprove(ctx, token0, owner, routerAddr); err != nil {
		return err
	}
	if err := tokenApprove(ctx, token1, owner, routerAddr); err != nil {
		return err
	}
	return nil
}

func stableApprove(ctx *types.Context, owner common.Address) error {
	for k := uint8(0); k < N; k++ {
		if err := tokenApprove(ctx, stableTokens[k], owner, swap); err != nil {
			return err
		}
	}
	return nil
}

func uniAddLiquidity(ctx *types.Context, owner common.Address, amount0Desired, amount1Desired *amount.Amount) (*amount.Amount, *amount.Amount, *amount.Amount, common.Address, error) {
	is, err := Exec(ctx, owner, routerAddr, "UniAddLiquidity", []interface{}{token0, token1, amount0Desired, amount1Desired, ZeroAmount, ZeroAmount})
	if err != nil {
		return nil, nil, nil, ZeroAddress, err
	}

	return is[0].(*amount.Amount), is[1].(*amount.Amount), is[2].(*amount.Amount), is[3].(common.Address), nil
}

func stableAddLiquidity(ctx *types.Context, _owner, _swap common.Address, _amounts []*amount.Amount) (*amount.Amount, error) {
	is, err := Exec(ctx, _owner, _swap, "AddLiquidity", []interface{}{_amounts, amount.NewAmount(0, 0)})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount), nil
}

func uniAddInitialLiquidity(ctx *types.Context, _owner common.Address) (*amount.Amount, *amount.Amount, *amount.Amount, common.Address, error) {
	if err := uniMint(ctx, _owner); err != nil {
		return nil, nil, nil, ZeroAddress, err
	}
	if err := uniApprove(ctx, _owner); err != nil {
		return nil, nil, nil, ZeroAddress, err
	}
	return uniAddLiquidity(ctx, _owner, _SupplyTokens[0], _SupplyTokens[1])
}

func stableAddInitialLiquidity(ctx *types.Context, _owner common.Address) (*amount.Amount, error) {
	if err := stableMint(ctx, _owner); err != nil {
		return nil, err
	}
	if err := stableApprove(ctx, _owner); err != nil {
		return nil, err
	}
	return stableAddLiquidity(ctx, _owner, swap, _InitialAmounts)
}

func stableGetAdminBalances(ctx *types.Context) ([]*amount.Amount, error) {
	is, err := Exec(ctx, admin, swap, "Reserves", []interface{}{})
	if err != nil {
		return nil, err
	}
	swap_reserves := is[0].([]*amount.Amount)

	admin_balances := MakeSlice(N)

	for k := uint8(0); k < N; k++ {
		tBalance, err := tokenBalanceOf(ctx, stableTokens[k], swap)
		if err != nil {
			return nil, err
		}
		admin_balances[k].Set(Sub(tBalance.Int, swap_reserves[k].Int))
	}
	return ToAmounts(admin_balances), nil
}

/////////// functions  ///////////
func initChain(ctx *types.Context, adm common.Address) (*chain.Chain, int, *types.Context, error) {
	mutex.Lock()
	cdx := chainIdx
	chainIdx++
	mutex.Unlock()

	mainTokenAndMint(ctx)

	cn, err := Chain(cdx, ctx, adm)
	if err != nil {
		RemoveChain(cdx)
		return nil, 0, nil, err
	}
	nctx := cn.NewContext()

	return cn, cdx, nctx, nil
}

// mainToken Deploy and Mint
func mainTokenAndMint(ctx *types.Context) error {
	tokenConstrunction := &token.TokenContractConstruction{
		Name:   "MainToken",
		Symbol: "MAINTOKEN",
	}
	bs, _, _ := bin.WriterToBytes(tokenConstrunction)
	v, _ := ctx.DeployContract(admin, classMap["Token"], bs)
	mainToken = v.(*token.TokenContract).Address()

	ctx.SetMainToken(mainToken)
	if _, err := Exec(ctx, admin, mainToken, "Mint", []interface{}{alice, amount.NewAmount(uint64(_BaseAmount), 0)}); err != nil {
		return err
	}
	if _, err := Exec(ctx, admin, mainToken, "Mint", []interface{}{bob, amount.NewAmount(uint64(_BaseAmount), 0)}); err != nil {
		return err
	}
	if _, err := Exec(ctx, admin, mainToken, "Mint", []interface{}{charlie, amount.NewAmount(uint64(_BaseAmount), 0)}); err != nil {
		return err
	}
	return nil
}

func setFees(cn *chain.Chain, ctx *types.Context, ex common.Address, fee, admin_fee, winner_fee, delay uint64, signer key.Key) (*types.Context, error) {
	is, err := Exec(ctx, admin, ex, "Owner", []interface{}{})
	if err != nil {
		return nil, err
	}
	owner := is[0].(common.Address)

	if _, err := Exec(ctx, owner, ex, "CommitNewFee", []interface{}{fee, admin_fee, winner_fee, delay}); err != nil {
		return nil, err
	}
	ctx, err = Sleep(cn, ctx, nil, delay, signer)
	if err != nil {
		return nil, err
	}
	_, err = Exec(ctx, owner, ex, "ApplyNewFee", []interface{}{})
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func _min_max(_ctx *types.Context, _swap common.Address, user common.Address) (uint8, uint8, *big.Int, *big.Int) {
	is, _ := Exec(_ctx, user, _swap, "Reserves", []interface{}{})
	reserves := ToBigInts(is[0].([]*amount.Amount))
	min_idx, min_amount := MinInArray(reserves)
	max_idx, max_amount := MaxInArray(reserves)
	if min_idx == max_idx {
		min_idx = int(Abs(SubC(big.NewInt(int64(min_idx)), 1)).Int64()) // abs(min_idx -1)
	}
	return uint8(min_idx), uint8(max_idx), min_amount, max_amount
}

func getFee(b *big.Int, fee uint64) *big.Int {
	return MulDivCC(b, int64(fee), trade.FEE_DENOMINATOR)
}
