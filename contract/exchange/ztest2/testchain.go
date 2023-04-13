package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"strings"
	"time"

	etypes "github.com/ethereum/go-ethereum/core/types"
	mtypes "github.com/meverselabs/meverse/ethereum/core/types"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/erc20wrapper"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"

	"github.com/meverselabs/meverse/contract/exchange/whitelist"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/gateway"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/chain/admin"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"

	"github.com/pkg/errors"
)

var (
	Zero       = big.NewInt(0)
	AmountZero = amount.NewAmount(0, 0)
	MaxUint256 = ToAmount(Sub(Exp(big.NewInt(2), big.NewInt(256)), big.NewInt(1)))

	AddressZero = common.Address{}

	ClassMap = map[string]uint64{}

	_BaseAmount = int64(1000000)

	_Fee30     = uint64(30000000)   // 30bp
	_Fee40     = uint64(40000000)   // 40bp
	_Fee5000   = uint64(5000000000) // 50%
	_AdminFee6 = uint64(1666666667) // 1/6 = uint64(1666666667)

	_MaxIter = 10
	_Step    = uint64(1000)
)

// removeChainData removes data directory which includes data files
func removeChainData(path string) error {
	return os.RemoveAll(path)
}

// prepare return testBlockchain and results from Genesis
func prepare(path string, deletePath bool, chainID *big.Int, version uint16, chainAdmin *common.Address, args []interface{}, genesisInitFunc func(*types.Context, []interface{}) ([]interface{}, error), cfg *initContextInfo) (*testBlockChain, []interface{}) {

	genesis := types.NewEmptyContext()
	RegisterContracts()

	ret, err := genesisInitFunc(genesis, args)
	if err != nil {
		panic(err)
	}

	tb, err := NewTestBlockChain(path, deletePath, chainID, version, genesis, chainAdmin, cfg)
	if err != nil {
		removeChainData(path)
		panic(err)
	}

	return tb, ret
}

// getSigners gets signers which are same with hardhat node users in order to test in tandem
func getSingers(chainID *big.Int) ([]key.Key, error) {

	// stringToKey converts private key string to Key struct
	stringToKey := func(chainID *big.Int, pkStr string) (*key.MemoryKey, error) {
		if strings.HasPrefix(pkStr, "0x") {
			pkStr = pkStr[2:]
		}
		h, err := hex.DecodeString(pkStr)
		if err != nil {
			return nil, err
		}
		return key.NewMemoryKeyFromBytes(chainID, h)
	}

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
func getCC(ctx *types.Context, contAddr, user common.Address) (*types.ContractContext, error) {

	if cont, err := ctx.Contract(contAddr); err == nil {
		cc := ctx.ContractContext(cont, user)
		intr := types.NewInteractor(ctx, cont, cc, "000000000000", false)
		cc.Exec = intr.Exec

		return cc, nil
	} else {
		statedb := types.NewStateDB(ctx)
		if statedb.IsEvmContract(contAddr) {
			cc := ctx.ContractContextFromAddress(contAddr, user)
			intr := types.NewInteractor2(ctx, cc, "000000000000", false)
			cc.Exec = intr.Exec
			return cc, nil
		} else {
			return nil, err
		}
	}
}

// tokenTotalSupply return token's totalSupply method
func tokenTotalSupply(ctx *types.Context, token common.Address) *amount.Amount {
	cc, err := getCC(ctx, token, AddressZero)
	if err != nil {
		panic(err)
	}

	is, err := cc.Exec(cc, token, "TotalSupply", []interface{}{})
	if err != nil {
		panic(err)
	}
	return is[0].(*amount.Amount)
}

// tokenBalanceOf return token's balanceOf method
func tokenBalanceOf(ctx *types.Context, token, from common.Address) *amount.Amount {
	cc, err := getCC(ctx, token, AddressZero)
	if err != nil {
		panic(err)
	}

	is, err := cc.Exec(cc, token, "BalanceOf", []interface{}{from})
	if err != nil {
		panic(err)
	}
	return is[0].(*amount.Amount)
}

// tokenAllowance return token's allowance method
func tokenAllowance(ctx *types.Context, token, owner, spender common.Address) *amount.Amount {
	cc, err := getCC(ctx, token, AddressZero)
	if err != nil {
		panic(err)
	}
	is, err := cc.Exec(cc, token, "Allowance", []interface{}{owner, spender})
	if err != nil {
		panic(err)
	}
	return is[0].(*amount.Amount)
}

// tokenIsMinter return token's isMinter method
func tokenIsMinter(ctx *types.Context, token, minter common.Address) bool {
	cc, err := getCC(ctx, token, AddressZero)
	if err != nil {
		panic(err)
	}
	is, err := cc.Exec(cc, token, "IsMinter", []interface{}{minter})
	if err != nil {
		panic(err)
	}
	return is[0].(bool)
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
func RegisterContracts() {

	registerContractClass(&token.TokenContract{}, "Token", ClassMap)
	registerContractClass(&formulator.FormulatorContract{}, "Formulator", ClassMap)
	registerContractClass(&gateway.GatewayContract{}, "Gateway", ClassMap)
	registerContractClass(&factory.FactoryContract{}, "Factory", ClassMap)
	registerContractClass(&router.RouterContract{}, "Router", ClassMap)
	registerContractClass(&trade.UniSwap{}, "UniSwap", ClassMap)
	registerContractClass(&trade.StableSwap{}, "StableSwap", ClassMap)
	registerContractClass(&whitelist.WhiteListContract{}, "WhiteList", ClassMap)
	registerContractClass(&erc20wrapper.Erc20WrapperContract{}, "Erc20Wrapper", ClassMap)
}

// testBlockChain is blockchain mock for testing
type testBlockChain struct {
	chainID  *big.Int // hardhat 1337
	version  uint16
	path     string // 화일저장 디렉토리
	chain    *chain.Chain
	obKeys   []key.Key
	ctx      *types.Context
	frKeyMap map[common.Address]key.Key
	step     uint64 // chain default forward time for each block : units - miliseconds
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
func NewTestBlockChain(path string, deletePath bool, chainID *big.Int, version uint16, genesis *types.Context, admin *common.Address, cfg *initContextInfo) (*testBlockChain, error) {

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

	// observer : 5개
	obStrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000001",
		"b000000000000000000000000000000000000000000000000000000000000002",
		"b000000000000000000000000000000000000000000000000000000000000003",
		"b000000000000000000000000000000000000000000000000000000000000004",
		"b000000000000000000000000000000000000000000000000000000000000005",
	}
	obKeys := []key.Key{}
	ObserverKeys := []common.PublicKey{}
	for _, v := range obStrs {
		if bs, err := hex.DecodeString(v); err != nil {
			return nil, err
		} else if Key, err := key.NewMemoryKeyFromBytes(chainID, bs); err != nil {
			return nil, err
		} else {
			obKeys = append(obKeys, Key)
			ObserverKeys = append(ObserverKeys, Key.PublicKey())
		}
	}

	// formulator : 1개
	frStrs := []string{
		"a000000000000000000000000000000000000000000000000000000000000001",
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
	if err := genesis.SetAdmin(*admin, true); err != nil {
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
		ctx:      cn.NewContext(),
		step:     _Step,
	}

	return tb, nil
}

// newContext calls chain.NewContext()
func (tb *testBlockChain) newContext() *types.Context {
	return tb.chain.NewContext()
}

// nextTimestamp forwards blockchain time (tx, block)
func (tb *testBlockChain) nextTimestamp() uint64 {
	timestamp := tb.chain.Provider().LastTimestamp()
	if timestamp == 0 {
		timestamp = uint64(time.Now().UnixNano())
	}
	return timestamp + tb.step*uint64(time.Millisecond)
}

// sleep forwards block time by amount sleeptime - miliseconds
func (tb *testBlockChain) sleep(sleepTime uint64) {
	tb.step = sleepTime
	_, err := tb.addBlock(nil)
	if err != nil {
		panic(err)
	}
	tb.step = _Step
}

// addBlock adds a block containing txs
func (tb *testBlockChain) addBlock(txs []*txWithSigner) (*types.Block, error) {
	TimeoutCount := uint32(0)
	Generator, err := tb.chain.TopGenerator(TimeoutCount)

	bc := chain.NewBlockCreator(tb.chain, tb.newContext(), Generator, TimeoutCount, tb.nextTimestamp(), 0)
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

	// ctx 재지정
	tb.ctx = tb.newContext()

	return b, nil
}

// view calls ContractContext.Exec from Context
func (tb *testBlockChain) view(contAddr common.Address, methodName string, args ...any) ([]interface{}, error) {
	cc, err := getCC(tb.ctx, contAddr, AddressZero)
	if err != nil {
		panic(err)
	}

	return cc.Exec(cc, contAddr, methodName, args)
}

// view calls ContractContext.Exec from Context
func (tb *testBlockChain) viewFrom(from common.Address, contAddr common.Address, methodName string, args ...any) ([]interface{}, error) {
	cc, err := getCC(tb.ctx, contAddr, from)
	if err != nil {
		panic(err)
	}

	return cc.Exec(cc, contAddr, methodName, args)
}

// amountView return first amount type result
func (tb *testBlockChain) viewAmount(contAddr common.Address, methodName string, args ...any) *amount.Amount {

	is, err := tb.view(contAddr, methodName, args...)
	if err != nil {
		panic(err)
	}
	return is[0].(*amount.Amount)
}

// amountView return first amount type result
func (tb *testBlockChain) viewAddress(contAddr common.Address, methodName string, args ...any) common.Address {

	is, err := tb.view(contAddr, methodName, args...)
	if err != nil {
		panic(err)
	}
	return is[0].(common.Address)
}

// txAddBlock executes matched-tx and connect block to the chain
func (tb *testBlockChain) txCall(senderKey key.Key, cont common.Address, method string, args ...any) (*types.Block, error) {
	provider := tb.chain.Provider()

	txws := &txWithSigner{&types.Transaction{
		ChainID:     provider.ChainID(),
		Timestamp:   tb.nextTimestamp(),
		Seq:         provider.AddrSeq(senderKey.PublicKey().Address()),
		To:          cont,
		Method:      method,
		GasPrice:    big.NewInt(10000000),
		UseSeq:      true,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bin.TypeWriteAll(args...),
	}, senderKey}

	b, err := tb.addBlock([]*txWithSigner{txws})
	if err != nil {
		return nil, err
	}

	return b, nil
}

// txCall return results after txCall
func (tb *testBlockChain) call(senderKey key.Key, cont common.Address, method string, args ...any) ([]interface{}, error) {
	b, err := tb.txCall(senderKey, cont, method, args...)
	if err != nil {
		return nil, err
	}

	return getResults(b)
}

// getResults return block's results
func getResults(b *types.Block) ([]interface{}, error) {
	for i := 0; i < len(b.Body.Events); i++ {
		en := b.Body.Events[i]
		if en.Type == ctypes.EventTagTxMsg {
			ins, err := bin.TypeReadAll(en.Result, 1)
			if err != nil {
				return nil, err
			}
			return ins, nil
		}
	}
	return nil, nil
}

// Close calls chain.Close()
func (tb *testBlockChain) Close() {
	tb.chain.Close()
}

// mevInitialize deploy mev Token, mint to sigers and register mainToken as mev
func mevInitialize(ctx *types.Context, args []interface{}) ([]interface{}, error) {
	//alice(admin)
	alice, ok := args[0].(common.Address)
	if !ok {
		return nil, ErrArgument
	}

	initialSupplyMap := map[common.Address]*amount.Amount{}
	for _, arg := range args {
		initialSupplyMap[arg.(common.Address)] = amount.NewAmount(uint64(_BaseAmount), 0)
	}

	arg := &token.TokenContractConstruction{
		Name:             "MEVerse",
		Symbol:           "MEV",
		InitialSupplyMap: initialSupplyMap,
	}
	bs, _, _ := bin.WriterToBytes(arg)
	v, err := ctx.DeployContract(alice, ClassMap["Token"], bs)
	if err != nil {
		return nil, err
	}
	mev := v.(*token.TokenContract).Address()

	ctx.SetMainToken(mev)

	//fmt.Println("mev   Address", mev.String())
	return []interface{}{mev}, nil
}

// erc20TokenWrapperCreationTx deploy Erc20TokenWrapper contract by chainadmin (only possible)
func erc20TokenWrapperCreationTx(tb *testBlockChain, owner, erc20Token common.Address) (*types.Transaction, error) {

	erc20WrapperArgs := &erc20wrapper.ERC20WrapperContractConstruction{Erc20Token: erc20Token}
	bs, _, _ := bin.WriterToBytes(erc20WrapperArgs)
	arg := chain.DeployContractData{
		Owner:   owner,
		ClassID: ClassMap["Erc20Wrapper"],
		Args:    bs,
	}

	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}
	return tx, nil
}

// tokenCreationTx deploy Erc20TokenWrapper contract by chainadmin (only possible)
func tokenCreationTx(tb *testBlockChain, owner common.Address, name, symbol string) (*types.Transaction, error) {

	erc20WrapperArgs := &token.TokenContractConstruction{
		Name:   name,
		Symbol: symbol,
	}
	bs, _, _ := bin.WriterToBytes(erc20WrapperArgs)
	arg := chain.DeployContractData{
		Owner:   owner,
		ClassID: ClassMap["Token"],
		Args:    bs,
	}

	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}
	return tx, nil
}

// dexInitialize deploy dex contracts and mint neccesary tokens to sigers
func dexInitialize(genesis *types.Context, args []interface{}) ([]interface{}, error) {
	//alice(admin), bob, charlie
	alice, ok := args[0].(common.Address)
	if !ok {
		return nil, ErrArgument
	}

	// factory
	factoryConstrunction := &factory.FactoryContractConstruction{Owner: alice}
	bs, _, err := bin.WriterToBytes(factoryConstrunction)
	if err != nil {
		return nil, err
	}
	v, err := genesis.DeployContract(alice, ClassMap["Factory"], bs)
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
	v, err = genesis.DeployContract(alice, ClassMap["Router"], bs)
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
	v, err = genesis.DeployContract(alice, ClassMap["WhiteList"], bs)
	if err != nil {
		return nil, err
	}
	whiteList := v.(*whitelist.WhiteListContract).Address()

	fmt.Println("factory   Address", factory.String())
	fmt.Println("router    Address", router.String())
	fmt.Println("whitelist Address", whiteList.String())

	return []interface{}{factory, router, whiteList}, nil
}

// Erc20TokenContractCreation deploy Erc20Token contract with initialSupply(= 1 ether)
// source code : evm-client/contracts/ERC20Token.sol
func Erc20TokenContractCreationTx(tb *testBlockChain, senderKey key.Key, initialSupply *amount.Amount) (*types.Transaction, error) {

	path, _ := os.Getwd() // /home/khzhao/prj/meverse/fleta2.0/exchange2/contract/exchange/ztest2
	b, err := os.ReadFile(path + "/ERC20Token.json")
	if err != nil {
		return nil, ErrFileRead
	}
	var abiJson map[string]interface{}
	if err := json.Unmarshal(b, &abiJson); err != nil {
		return nil, err
	}

	bytecode := strings.Replace(abiJson["bytecode"].(string), "0x", "", -1) + fmt.Sprintf("%064x", initialSupply.Int)
	data, err := hex.DecodeString(bytecode)
	if err != nil {
		return nil, err
	}

	provider := tb.chain.Provider()

	signer := mtypes.MakeSigner(tb.chainID, provider.Height())

	txSigner := senderKey.PublicKey().Address()
	nonce := provider.AddrSeq(txSigner)
	etx := etypes.NewTx(&etypes.DynamicFeeTx{
		ChainID:   tb.chainID,
		Nonce:     nonce,
		Gas:       0x1DCD6500, // GasLimit = 500000000
		Data:      data,
		GasTipCap: big.NewInt(0x0),        // maxPriorityFeePerGas
		GasFeeCap: big.NewInt(0x2c9d07ea), // maxFeePerGas = 748488682
		Value:     big.NewInt(0),
	})

	signedTx, err := etypes.SignTx(etx, signer, senderKey.PrivateKey())
	if err != nil {
		return nil, err
	}
	bs, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		Seq:         tb.ctx.AddrSeq(senderKey.PublicKey().Address()),
		To:          AddressZero,
		Method:      "",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        bs,
	}
	return tx, nil
}

// factoryDeploy deploy Factory contract
func factoryDeploy(tb *testBlockChain, deployerKey key.Key) (*common.Address, error) {

	owner := deployerKey.PublicKey().Address()
	factoryConstrunction := &factory.FactoryContractConstruction{
		Owner: owner,
	}
	bs, _, _ := bin.WriterToBytes(factoryConstrunction)
	arg := chain.DeployContractData{
		Owner:   owner,
		ClassID: ClassMap["Factory"],
		Args:    bs,
	}
	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}
	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	factory := ret[0].(common.Address)

	return &factory, nil
}

// routerDeploy deploy router contract
func routerDeploy(tb *testBlockChain, deployerKey key.Key, factory *common.Address) (*common.Address, error) {
	routerConstrunction := &router.RouterContractConstruction{Factory: *factory}
	bs, _, _ := bin.WriterToBytes(routerConstrunction)
	arg := chain.DeployContractData{
		Owner:   deployerKey.PublicKey().Address(),
		ClassID: ClassMap["Router"],
		Args:    bs,
	}
	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}
	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	router := ret[0].(common.Address)
	return &router, nil
}

// routerDeploy deploy router contract
func whiteListDeploy(tb *testBlockChain, deployerKey key.Key) (*common.Address, error) {

	whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
	bs, _, _ := bin.WriterToBytes(whitelistConstrunction)
	arg := chain.DeployContractData{
		Owner:   deployerKey.PublicKey().Address(),
		ClassID: ClassMap["WhiteList"],
		Args:    bs,
	}

	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}

	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	whiteList := ret[0].(common.Address)
	return &whiteList, nil
}

// erc20WrapperCreate create erc20Token contract and erc20TokenWrapper contract
func erc20TokenDeploy(tb *testBlockChain, deployerKey key.Key, initialSupply *amount.Amount) (*common.Address, error) {

	tx, err := Erc20TokenContractCreationTx(tb, deployerKey, initialSupply)
	if err != nil {
		return nil, err
	}
	_, err = tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	provider := tb.chain.Provider()
	receipts, err := provider.Receipts(provider.Height())
	if err != nil {
		return nil, err
	}
	receipt := receipts[0]
	erc20Token := receipt.ContractAddress
	return &erc20Token, nil
}

// erc20WrapperCreate create erc20Token contract and erc20TokenWrapper contract
func erc20WrapperDeploy(tb *testBlockChain, deployerKey key.Key, initialSupply *amount.Amount) (*common.Address, *common.Address, error) {

	erc20Token, err := erc20TokenDeploy(tb, deployerKey, initialSupply)

	tx, err := erc20TokenWrapperCreationTx(tb, deployerKey.PublicKey().Address(), *erc20Token)
	if err != nil {
		return nil, nil, err
	}

	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, nil, err
	}
	erc20Wrapper := ret[0].(common.Address)
	return &erc20Wrapper, erc20Token, nil
}

// tokenCreate create Token contract
func tokenDeploy(tb *testBlockChain, deployerKey key.Key, name, symbol string) (*common.Address, error) {

	tx, err := tokenCreationTx(tb, deployerKey.PublicKey().Address(), name, symbol)
	if err != nil {
		return nil, err
	}
	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	token := ret[0].(common.Address)
	return &token, nil
}

type pairContractConstruction struct {
	TokenA, TokenB, PayToken common.Address
	Name, Symbol             string
	Owner, Winner            common.Address
	Fee, AdminFee, WinnerFee uint64
	Factory, WhiteList       common.Address
	GroupId                  hash.Hash256
}

// pairCreate create Uniswap pair contract
func pairCreate(tb *testBlockChain, senderKey key.Key, p *pairContractConstruction) (*common.Address, error) {

	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		Seq:         tb.ctx.AddrSeq(senderKey.PublicKey().Address()),
		To:          p.Factory,
		Method:      "CreatePairUni",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bin.TypeWriteAll(p.TokenA, p.TokenB, p.PayToken, p.Name, p.Symbol, p.Owner, p.Winner, p.Fee, p.AdminFee, p.WinnerFee, p.WhiteList, p.GroupId, ClassMap["UniSwap"]),
	}
	b, err := tb.addBlock([]*txWithSigner{{tx, senderKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	pair := ret[0].(common.Address)

	return &pair, nil
}

// pairCreate create Uniswap pair contract
func swapDeploy(tb *testBlockChain, deployerKey key.Key, p *trade.StableSwapConstruction) (*common.Address, error) {

	bs, _, _ := bin.WriterToBytes(p)
	arg := chain.DeployContractData{
		Owner:   deployerKey.PublicKey().Address(),
		ClassID: ClassMap["StableSwap"],
		Args:    bs,
	}

	bs2, _, _ := bin.WriterToBytes(&arg)
	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   tb.nextTimestamp(),
		To:          AddressZero,
		Method:      admin.ContractDeploy,
		GasPrice:    big.NewInt(748488682),
		UseSeq:      false,
		IsEtherType: false,
		VmType:      types.Go,
		Args:        bs2,
	}

	b, err := tb.addBlock([]*txWithSigner{{tx, deployerKey}})
	if err != nil {
		return nil, err
	}
	ret, err := getResults(b)
	if err != nil {
		return nil, err
	}
	swap := ret[0].(common.Address)

	return &swap, nil
}

// setFees set Uniswap pair or Stableswap swap's fees
func setFees(tb *testBlockChain, senderKey key.Key, ex common.Address, fee, admin_fee, winner_fee uint64) {
	delay := uint64(86400)
	//tb.step = delay * 1000
	_, err := tb.call(senderKey, ex, "CommitNewFee", fee, admin_fee, winner_fee, delay)
	if err != nil {
		panic(err)
	}
	//tb.step = _Step
	tb.sleep(delay*1000 - 1000)
	_, err = tb.call(senderKey, ex, "ApplyNewFee")
	if err != nil {
		panic(err)
	}
}

func insufficientBalanceErrorCheck(err error) {
	errString := err.Error()
	if k := strings.Index(errString, ":"); k >= 0 {
		errString = errString[0:k]
	}
	if errString != "execution reverted" && errString != "the token holding quantity is insufficient balance" {
		panic(err)
	}
}
