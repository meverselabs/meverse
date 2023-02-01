package bloomservice

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/bridge"
	"github.com/meverselabs/meverse/contract/connect/depositpool"
	"github.com/meverselabs/meverse/contract/connect/farm"
	"github.com/meverselabs/meverse/contract/connect/imo"
	"github.com/meverselabs/meverse/contract/connect/pool"
	"github.com/meverselabs/meverse/contract/erc20wrapper"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/external/deployer"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/gateway"
	"github.com/meverselabs/meverse/contract/nft721"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"

	"github.com/pkg/errors"
)

var (
	Zero       = big.NewInt(0)
	ZeroAmount = amount.NewAmount(0, 0)
	MaxUint256 = ToAmount(Sub(Exp(big.NewInt(2), big.NewInt(256)), big.NewInt(1)))

	ZeroAddress = common.Address{}
)

// ToAmount converts big.Int to amount.Amount
func ToAmount(b *big.Int) *amount.Amount {
	return &amount.Amount{Int: b}
}

// Sub : sub
func Sub(a, b *big.Int) *big.Int {
	return big.NewInt(0).Sub(a, b)
}

// Exp : exponential
func Exp(a, b *big.Int) *big.Int {
	return big.NewInt(0).Exp(a, b, nil)
}

// removeChainData removes data directory which includes data files
func removeChainData(path string) error {
	// if _, err := os.Stat("/mnt/ramdisk"); !os.IsNotExist(err) {
	// 	dir = "/mnt/ramdisk/" + dir
	// }

	return os.RemoveAll(path)
}

// prepare return testBlockchain and results from Genesis
func prepare(path string, deletePath bool, chainID *big.Int, version uint16, chainAdmin common.Address, args []interface{}, genesisInitFunc func(*types.Context, map[string]uint64, []interface{}) ([]interface{}, error), cfg *initContextInfo) (*testBlockChain, []interface{}, error) {

	genesis := types.NewEmptyContext()
	classMap := RegisterContracts()

	ret, err := genesisInitFunc(genesis, classMap, args)
	if err != nil {
		return nil, nil, err
	}

	tb, err := NewTestBlockChain(path, deletePath, chainID, version, genesis, chainAdmin, cfg)
	if err != nil {
		//removeChainData(path)
		return nil, nil, err
	}

	return tb, ret, nil
}

// stringToKey converts private key string to Key struct
func stringToKey(chainID *big.Int, pkStr string) (*key.MemoryKey, error) {
	if strings.HasPrefix(pkStr, "0x") {
		pkStr = pkStr[2:]
	}
	h, err := hex.DecodeString(pkStr)
	if err != nil {
		return nil, err
	}
	return key.NewMemoryKeyFromBytes(chainID, h)
}

// getSigners gets signers which are same with hardhat node users
// in order to test in tandem
func getSingers(chainID *big.Int) ([]key.Key, error) {

	keyStrs := []string{
		"0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", //0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
		"0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d", //0x70997970c51812dc3a010c7d01b50e0d17dc79c8
		"0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a", //0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc
		"0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6", //0x90f79bf6eb2c4f870365e785982e1f101e93b906
		"0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a", //0x15d34aaf54267db7d7c367839aaf71a00a2c6a65
	}
	userKeys := []key.Key{}

	for _, keyStr := range keyStrs {
		k, err := stringToKey(chainID, keyStr)
		if err != nil {
			return nil, err
		}
		userKeys = append(userKeys, k)
	}
	return userKeys, nil
}

// GetCC gets ContractContext from Contex with given contract address and user address
func GetCC(ctx *types.Context, contAddr common.Address, user common.Address) (*types.ContractContext, error) {

	cont, err := ctx.Contract(contAddr)
	if err != nil {
		return nil, err
	}
	cc := ctx.ContractContext(cont, user)
	intr := types.NewInteractor(ctx, cont, cc, "000000000000", false)
	cc.Exec = intr.Exec

	return cc, nil
}

// Exec calls ContractContext.Exec from Context
func Exec(ctx *types.Context, user common.Address, contAddr common.Address, methodName string, args []interface{}) ([]interface{}, error) {
	cc, err := GetCC(ctx, contAddr, user)
	if err != nil {
		return nil, err
	}
	is, err := cc.Exec(cc, contAddr, methodName, args)
	return is, err
}

// TokenApprove call token.Apporve(to,Amount) from Context
func TokenApprove(ctx *types.Context, token, owner, spender common.Address) error {
	cc, err := GetCC(ctx, token, owner)
	if err != nil {
		return err
	}
	return tokenApprove(cc, token, spender, MaxUint256.Int)
}

// tokenApprove call token.Apporve(to,Amount) from ContractContext
func tokenApprove(cc *types.ContractContext, token, to common.Address, am *big.Int) error {
	_, err := cc.Exec(cc, token, "Approve", []interface{}{to, ToAmount(am)})
	return err
}

// registerContractClass register class item
func registerContractClass(cont types.Contract, className string, ClassMap map[string]uint64) {
	ClassID, err := types.RegisterContractType(cont)
	if err != nil {
		panic(err)
	}
	ClassMap[className] = ClassID
}

// RegisterContracts creates classmap for deploying contracts ususally in genesis
func RegisterContracts() map[string]uint64 {
	ClassMap := map[string]uint64{}

	registerContractClass(&token.TokenContract{}, "Token", ClassMap)
	registerContractClass(&formulator.FormulatorContract{}, "Formulator", ClassMap)
	registerContractClass(&gateway.GatewayContract{}, "Gateway", ClassMap)
	registerContractClass(&factory.FactoryContract{}, "Factory", ClassMap)
	registerContractClass(&router.RouterContract{}, "Router", ClassMap)
	registerContractClass(&trade.UniSwap{}, "UniSwap", ClassMap)
	registerContractClass(&trade.StableSwap{}, "StableSwap", ClassMap)
	registerContractClass(&bridge.BridgeContract{}, "Bridge", ClassMap)
	registerContractClass(&farm.FarmContract{}, "ConnectFarm", ClassMap)
	registerContractClass(&pool.PoolContract{}, "ConnectPool", ClassMap)
	registerContractClass(&whitelist.WhiteListContract{}, "WhiteList", ClassMap)
	registerContractClass(&imo.ImoContract{}, "IMO", ClassMap)
	registerContractClass(&depositpool.DepositPoolContract{}, "DepositUSDT", ClassMap)
	registerContractClass(&nft721.NFT721Contract{}, "NFT721", ClassMap)
	registerContractClass(&engin.EnginContract{}, "Engin", ClassMap)
	registerContractClass(&deployer.DeployerContract{}, "EnginDeployer", ClassMap)
	registerContractClass(&erc20wrapper.Erc20WrapperContract{}, "Erc20Wrapper", ClassMap)

	return ClassMap
}

// testBlockChain is blockchain mock for testing
type testBlockChain struct {
	chainID  *big.Int // hardhat 1337
	version  uint16
	path     string // 화일저장 디렉토리
	chain    *chain.Chain
	obKeys   []key.Key
	frKeyMap map[common.Address]key.Key
}

// initContextInfo struct is parameters for meverse chain with non-zero initheight
type initContextInfo struct {
	InitGenesisHash string
	InitHash        string
	InitHeight      uint32
	InitTimestamp   uint64
}

// non-evmtype transaction with signer key
type txWithSigner struct {
	tx     *types.Transaction
	signer key.Key
}

// NewTestBlockChain makes new test blockchain
func NewTestBlockChain(path string, deletePath bool, chainID *big.Int, version uint16, genesis *types.Context, admin common.Address, cfg *initContextInfo) (*testBlockChain, error) {

	if deletePath {
		err := removeChainData(path)
		if err != nil {
			return nil, err
		}
	}

	cdb, err := piledb.Open(path+"/chain", hash.HexToHash(cfg.InitHash), cfg.InitHeight, cfg.InitTimestamp)
	if err != nil {
		return nil, err
	}

	cdb.SetSyncMode(true)
	st, err := chain.NewStore(path+"/context", cdb, chainID, version)
	if err != nil {
		return nil, err
	}

	obKeys := []key.Key{}
	ObserverKeys := []common.PublicKey{}
	for i := 0; i < 5; i++ {
		pk, err := key.NewMemoryKeyFromBytes(chainID, []byte{1, 1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			return nil, err
		}
		obKeys = append(obKeys, pk)

		ObserverKeys = append(ObserverKeys, pk.PublicKey())
	}

	// formulator : 1개
	frStrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000010",
	}
	frkeys := []key.Key{}
	frKeyMap := map[common.Address]key.Key{}
	for _, v := range frStrs {
		if bs, err := hex.DecodeString(v); err != nil {
			return nil, err
		} else if Key, err := key.NewMemoryKeyFromBytes(chainID, bs); err != nil {
			return nil, err
		} else {
			frkeys = append(frkeys, Key)
			frKeyMap[Key.PublicKey().Address()] = Key
		}
	}

	cn := chain.NewChain(ObserverKeys, st, "main")
	if err := genesis.SetAdmin(admin, true); err != nil {
		return nil, err
	}

	for _, v := range frkeys {
		if err := genesis.SetGenerator(v.PublicKey().Address(), true); err != nil {
			return nil, err
		}
	}

	if cfg.InitHeight == 0 {
		if err := cn.Init(genesis.Top()); err != nil {
			return nil, err
		}
	} else {
		// RegisterContracts()
		// log.Println(cfg)
		if err := cn.InitWith(hash.HexToHash(cfg.InitGenesisHash), hash.HexToHash(cfg.InitHash), cfg.InitHeight, cfg.InitTimestamp); err != nil {
			return nil, err
		}
	}

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Cause(err) == chain.ErrStoreClosed {
			return nil, err
		}
		return nil, err
	}

	tb := &testBlockChain{
		path:     path,
		chainID:  chainID,
		version:  version,
		obKeys:   obKeys,
		frKeyMap: frKeyMap,
		chain:    cn,
	}

	return tb, nil
}

// newContext calls chain.NewContext()
func (tb *testBlockChain) newContext() *types.Context {
	return tb.chain.NewContext()
}

// addBlock adds a block containing txs
func (tb *testBlockChain) addBlock(ctx *types.Context, txs []*txWithSigner) (*types.Block, error) {
	TimeoutCount := uint32(0)
	Generator, err := tb.chain.TopGenerator(TimeoutCount)

	bc := chain.NewBlockCreator(tb.chain, ctx, Generator, TimeoutCount, uint64(time.Now().UnixNano()), 0)
	var receipts = types.Receipts{}
	for _, tx := range txs {
		sig, err := tx.signer.Sign(tx.tx.Message())
		if err != nil {
			return nil, err
		}
		if receipt, err := bc.AddTx(tx.tx, sig); err != nil {
			return nil, err
		} else {
			receipts = append(receipts, receipt)
		}
	}

	b, err := bc.Finalize(0, receipts)
	if err != nil {
		return nil, err
	}

	HeaderHash := bin.MustWriterToHash(&b.Header)
	//LastHash := HeaderHash

	pk := tb.frKeyMap[Generator]
	if pk == nil {
		return nil, errors.New("Generator pk is nil")
	}
	GenSig, err := pk.Sign(HeaderHash)
	if err != nil {
		return nil, err
	}

	b.Body.BlockSignatures = append(b.Body.BlockSignatures, GenSig)

	blockSign := &types.BlockSign{
		HeaderHash:         HeaderHash,
		GeneratorSignature: GenSig,
	}

	BlockSignHash := bin.MustWriterToHash(blockSign)

	idxes := rand.Perm(len(tb.obKeys))
	for i := 0; i < len(tb.obKeys)/2+1; i++ {
		pk := tb.obKeys[idxes[i]]
		ObSig, err := pk.Sign(BlockSignHash)
		if err != nil {
			return nil, err
		}
		b.Body.BlockSignatures = append(b.Body.BlockSignatures, ObSig)
	}

	err = tb.chain.ConnectBlock(b, nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// addBlock adds a new block and forward the next block time
// 문제점
// addBlock 으로 전달되는 ctx 는 아래와 같이 ctx 의 interval 만큼 시간을 전진 시킬 수 있다.
//    timestamp := ctx.LastTimestamp() + interval*uint64(time.Millisecond)
//    ctx.NextContext(lastHash, timestamp)
// 하지만 ConnectBlock에서 사용하는 ctx의 genTimestamp는 기존 store에 저장되어 있는 timestamp를 이용한다.
// ctx = types.NewContext(tb.chain.Store())
// 보통의 transaction에서는 문제가 없으나, uniswap의 pair contract._update 함수와 같이 시간을 저장해야하는 경우
// blockTimestamp := cc.LastTimestamp() / uint64(time.Second) 차이가 발생할 수 있다.
// ex. key = PairContract + 0x43(=tagBlockTimestampLast)의 값이 서로 달라진다
// 위의 값이 다르면, ctx.Top() 의 data가 달라지므로 Header.ContextHash = bc.ctx.Hash()의 값이 서로 달라져서 아래 에러체크에서 걸린다.
// chain.ConnectBlock ->
//   chain.connectBlockWithContext(b *types.Block, ctx *types.Context, receipts types.Receipts) error {
// 		if b.Header.ContextHash != ctx.Hash() {
// 		log.Println("CONNECT", ctx.Hash(), b.Header.ContextHash, ctx.Dump())
// 		panic("")
// 		return errors.WithStack(ErrInvalidContextHash)
// 	}
// 비교
// // NextContext returns the next Context of the Context
// func (ctx *Context) NextContext(LastHash hash.Hash256, Timestamp uint64) *Context {
// 	ctx.Top().isTop = false
// 	nctx := NewContext(ctx)
// 	nctx.genTargetHeight = ctx.genTargetHeight + 1
// 	nctx.genPrevHash = LastHash
// 	nctx.genTimestamp = Timestamp
// 	return nctx
// }
// // NewContext returns a Context
// func NewContext(loader internalLoader) *Context {
// 	ctx := &Context{
// 		loader:          loader,
// 		genTargetHeight: loader.TargetHeight(),
// 		genPrevHash:     loader.PrevHash(),
// 		genTimestamp:    loader.LastTimestamp(),
// 	}
// 	ctx.cache = newContextCache(ctx)
// 	ctx.stack = []*ContextData{NewContextData(ctx.cache, nil)}
// 	return ctx
// }
func (tb *testBlockChain) addBlockAndSleep(ctx *types.Context, txs []*txWithSigner, miliSeconds uint64) (*types.Context, error) {

	_, err := tb.addBlock(ctx, txs)
	if err != nil {
		return nil, err
	}
	// lastHash := bin.MustWriterToHash(&b.Header)
	// timestamp := ctx.LastTimestamp() + miliSeconds*uint64(time.Millisecond)
	// return ctx.NextContext(lastHash, timestamp), nil
	return types.NewContext(tb.chain.Store()), nil
}

// Close calls chain.Close()
func (tb *testBlockChain) Close() {
	tb.chain.Close()
}

// mevInitialize deploy mev Token, mint to sigers and register mainToken as mev
func mevInitialize(ctx *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {
	//alice(admin), bob, charlie
	alice, ok := args[0].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	bob, ok := args[1].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	charlie, ok := args[2].(common.Address)
	if !ok {
		return nil, ErrArgument
	}

	arg := &token.TokenContractConstruction{
		Name:   "MEVerse",
		Symbol: "MEV",
		InitialSupplyMap: map[common.Address]*amount.Amount{
			alice:   amount.NewAmount(100000000, 0),
			bob:     amount.NewAmount(100000000, 0),
			charlie: amount.NewAmount(100000000, 0),
		},
	}
	bs, _, _ := bin.WriterToBytes(arg)
	v, _ := ctx.DeployContract(alice, classMap["Token"], bs)
	mev := v.(*token.TokenContract).Address()

	ctx.SetMainToken(mev)

	fmt.Println("mev   Address", mev.String())
	return []interface{}{mev}, nil
}

// dexInitialize deploy dex contracts and mint neccesary tokens to sigers
func dexInitialize(genesis *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {

	//alice(admin), bob, charlie
	alice, ok := args[0].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	bob, ok := args[1].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	charlie, ok := args[2].(common.Address)
	if !ok {
		return nil, ErrArgument
	}

	// factory
	factoryConstrunction := &factory.FactoryContractConstruction{Owner: alice}
	bs, _, err := bin.WriterToBytes(factoryConstrunction)
	if err != nil {
		return nil, err
	}
	v, err := genesis.DeployContract(alice, classMap["Factory"], bs)
	if err != nil {
		return nil, err
	}
	factory := v.(*factory.FactoryContract).Address()

	// router
	routerConstrunction := &router.RouterContractConstruction{Factory: factory}
	bs, _, err = bin.WriterToBytes(routerConstrunction)
	if err != nil {
		return nil, err
	}
	v, err = genesis.DeployContract(alice, classMap["Router"], bs)
	if err != nil {
		return nil, err
	}
	router := v.(*router.RouterContract).Address()

	// whitelist
	whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
	bs, _, err = bin.WriterToBytes(whitelistConstrunction)
	if err != nil {
		return nil, err
	}
	v, err = genesis.DeployContract(alice, classMap["WhiteList"], bs)
	if err != nil {
		return nil, err
	}
	whiteList := v.(*whitelist.WhiteListContract).Address()

	// tokens
	size := uint8(2)
	uniTokens := make([]common.Address, size, size)
	for k := uint8(0); k < size; k++ {
		tokenConstrunction := &token.TokenContractConstruction{
			Name:   "Token" + strconv.Itoa(int(k)),
			Symbol: "TOKEN" + strconv.Itoa(int(k)),
			InitialSupplyMap: map[common.Address]*amount.Amount{
				alice:   amount.NewAmount(100000000, 0),
				bob:     amount.NewAmount(100000000, 0),
				charlie: amount.NewAmount(100000000, 0),
			},
		}

		bs, _, _ := bin.WriterToBytes(tokenConstrunction)
		v, _ := genesis.DeployContract(alice, classMap["Token"], bs)
		uniTokens[k] = v.(*token.TokenContract).Address()
	}

	// pair
	_PairName := "__UNI_NAME"
	_PairSymbol := "__UNI_SYMBOL"
	_Fee := uint64(40000000)
	_AdminFee := uint64(trade.MAX_ADMIN_FEE)
	_WinnerFee := uint64(5000000000)
	_GroupId := hash.BigToHash(big.NewInt(1))

	is, err := Exec(genesis, alice, factory, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], common.Address{}, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, whiteList, _GroupId, classMap["UniSwap"]})
	if err != nil {
		return nil, err
	}
	pair := is[0].(common.Address)

	token0, token1, err := trade.SortTokens(uniTokens[0], uniTokens[1])
	if err != nil {
		return nil, err
	}

	// approve
	for _, token := range []common.Address{token0, token1} {
		for _, signer := range []common.Address{alice, bob, charlie} {
			TokenApprove(genesis, token, signer, router)
		}
	}

	fmt.Println("factory   Address", factory.String())
	fmt.Println("router    Address", router.String())
	fmt.Println("whitelist Address", whiteList.String())
	fmt.Println("token0    Address", token0.String())
	fmt.Println("token1    Address", token1.String())
	fmt.Println("pair      Address", pair.String())

	return []interface{}{factory, router, whiteList, token0, token1, pair}, nil
}
