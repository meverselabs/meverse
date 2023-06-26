package testlib

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
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
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/apiserver/metamaskrelay"
	"github.com/meverselabs/meverse/service/apiserver/viewchain"
	"github.com/meverselabs/meverse/service/bloomservice"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
	"github.com/meverselabs/meverse/tests/formulator/notformulator"
)

// getSigners gets signers which are same with hardhat node users
// in order to test in tandem
func GetSingers(chainID *big.Int) ([]key.Key, error) {

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
		"0x8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba", //0x9965507d1a55bcc2695c58ba16fb37d819b0a4dc
		"0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e", //0x976ea74026e726554db657fa54763abd0c3a0aa9
		"0x4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356", //0x14dc79964da2c08b23698b3d3cc7ca32193d9955
		"0xdbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97", //0x23618e81e3f5cdf7f54c3d65f7fbc0abf5b21e8f
		"0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6", //0xa0ee7a142d267c1f36714e4a8f75612f20a79720
		"0xf214f2b2cd398c806f84e317254e0f0b801d0643303237d97a22a48e01628897", //0xbcd4042de499d14e55001ccbb24a551f3b954096
		"0x701b615bbdfb9de65240bc28bd21bbc0d996645a3dd57e7b12bc2bdf6f192c82", //0x71be63f3384f5fb98995898a86b02fb2426c5788
		"0xa267530f49f8280200edf313ee7af6b827f2a8bce2897751d06a843f644967b1", //0xfabb0ac9d68b0b445fb7357272ff202c5651694a
		"0x47c99abed3324a2707c28affff1267e45918ec8c3f20b8aa892e8b065d2942dd", //0x1cbd3b2770909d4e10f157cabc84c7264073c9ec
		"0xc526ee95bf44d8fc405a158bb884d9d1238d99f0612e9f33d006bb0789009aaa", //0xdf3e18d64bc6a983f673ab319ccae4f1a57c7097
		"0x8166f546bab6da521a8369cab06c5d2b9e46670292d85c875ee9ec20e84ffb61", //0xcd3b766ccdd6ae721141f452c550ca635964ce71
		"0xea6c44ac03bff858b476bba40716402b03e41b8e97e276d1baec7c37d42484a0", //0x2546bcd3c84621e976d8185a91a922ae77ecec30
		"0x689af8efa8c651a91ad287602527f3af2fe9f6501a7ac4b061667b5a93e037fd", //0xbda5747bfd65f08deb54cb465eb87d40e51b197e
		"0xde9be858da4a475276426320d5e9262ecfc3ba460bfac56360bfa6c4c28b4ee0", //0xdd2fd4581271e230360230f9337d5c0430bf44c0
		"0xdf57089febbacf7ba0bc227dafbffa9fc08a93fdc68e1e42411a14efcf23656e", //0x8626f6940e2eb28930efb4cef49b2d1f2c9c1199

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
	// tokenApprove call token.Apporve(to,Amount) from ContractContext
	tokenApprove := func(cc *types.ContractContext, token, to common.Address, am *big.Int) error {
		_, err := cc.Exec(cc, token, "Approve", []interface{}{to, ToAmount(am)})
		return err
	}

	cc, err := GetCC(ctx, token, owner)
	if err != nil {
		return err
	}
	return tokenApprove(cc, token, spender, MaxUint256.Int)
}

// RegisterContracts creates classmap for deploying contracts ususally in genesis
func registerContracts() map[string]uint64 {
	ClassMap := map[string]uint64{}

	// registerContractClass register class item
	registerContractClass := func(cont types.Contract, className string, ClassMap map[string]uint64) {
		ClassID, err := types.RegisterContractType(cont)
		if err != nil {
			panic(err)
		}
		ClassMap[className] = ClassID
	}

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
	registerContractClass(&notformulator.NotFormulatorContract{}, "NotFormulatorContract", ClassMap)

	return ClassMap
}

// MevInitialize deploy mev Token, mint to sigers and register mainToken as mev
func MevInitialize(ctx *types.Context, classMap map[string]uint64, owner common.Address, initialSupplyMap map[common.Address]*amount.Amount) (*common.Address, error) {
	arg := &token.TokenContractConstruction{
		Name:             "MEVerse",
		Symbol:           "MEV",
		InitialSupplyMap: initialSupplyMap,
	}
	bs, _, _ := bin.WriterToBytes(arg)
	v, _ := ctx.DeployContract(owner, classMap["Token"], bs)
	mev := v.(*token.TokenContract).Address()

	ctx.SetMainToken(mev)

	fmt.Println("mev   Address", mev.String())
	return &mev, nil
}

// Prepare return testBlockchain and results from Genesis

// TestBlockChain is blockchain mock for testing
type TestBlockChain struct {
	ChainID         *big.Int //
	Version         uint16
	Path            string // 화일저장 디렉토리
	Chain           *chain.Chain
	Store           *chain.Store
	Provider        types.Provider
	obKeys          []key.Key
	FrKeyMap        map[common.Address]key.Key
	ClassMap        map[string]uint64
	StepMiliSeconds uint64 // interval from previous transaction
	rpcapi          *apiserver.APIServer
	Ts              itxsearch.ITxSearch
	Bs              *bloomservice.BloomBitService
}

func NewTestBlockChain(path string, deletePath bool, chainID *big.Int, version uint16, chainAdmin common.Address, genesisInitFunc func(*types.Context, map[string]uint64) error, cfg *InitContextInfo) *TestBlockChain {

	genesis := types.NewEmptyContext()
	classMap := registerContracts()

	err := genesisInitFunc(genesis, classMap)
	if err != nil {
		panic(err)
	}

	tb, err := NewTestBlockChainWithGenesis(path, deletePath, chainID, version, genesis, chainAdmin, cfg, classMap)
	if err != nil {
		//removeChainData(path)
		panic(err)
	}

	return tb
}

// InitContextInfo struct is parameters for meverse chain with non-zero initheight
type InitContextInfo struct {
	InitGenesisHash string
	InitHash        string
	InitHeight      uint32
	InitTimestamp   uint64
}

var DefaultInitContextInfo = &InitContextInfo{}

// non-evmtype transaction with signer key
type TxWithSigner struct {
	Tx     *types.Transaction
	Signer key.Key
}

// NewTxWithSigner makes a new TxWithSigner struct
func NewTxWithSigner(tx *types.Transaction, signer key.Key) *TxWithSigner {
	return &TxWithSigner{
		Tx:     tx,
		Signer: signer,
	}
}

// NewTestBlockChain makes new test blockchain
func NewTestBlockChainWithGenesis(path string, deletePath bool, chainID *big.Int, version uint16, genesis *types.Context, admin common.Address, cfg *InitContextInfo, classMap map[string]uint64) (*TestBlockChain, error) {

	if deletePath {
		err := RemoveChainData(path)
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

	// txsearch
	ts := NewTsMocK(cn.Provider())
	// bloomservice : 디렉토리 삭제 후 생성
	if err := os.RemoveAll(BloomDataPath); err != nil {
		return nil, err
	}
	bs, err := bloomservice.NewBloomBitService(cn, BloomDataPath, BloomBitsBlocks, BloomConfirms)
	if err != nil {
		return nil, err
	}
	// onBlockConnected
	cn.MustAddService(ts)
	cn.MustAddService(bs)

	if cfg.InitHeight == 0 {
		if err := cn.Init(genesis.Top()); err != nil {
			return nil, err
		}
	} else {
		if err := cn.InitWith(hash.HexToHash(cfg.InitGenesisHash), hash.HexToHash(cfg.InitHash), cfg.InitHeight, cfg.InitTimestamp); err != nil {
			return nil, err
		}
	}

	// rpc
	rpcapi := apiserver.NewAPIServer()
	metamaskrelay.NewMetamaskRelay(rpcapi, ts, bs, cn, nil)
	viewchain.NewViewchain(rpcapi, ts, cn, st, bs, nil)
	tb := &TestBlockChain{
		Path:            path,
		ChainID:         chainID,
		Version:         version,
		obKeys:          obKeys,
		FrKeyMap:        frKeyMap,
		Chain:           cn,
		Store:           cn.Store(),
		Provider:        cn.Provider(),
		ClassMap:        classMap,
		StepMiliSeconds: uint64(1000),
		rpcapi:          rpcapi,
		Ts:              ts,
		Bs:              bs,
	}

	return tb, nil
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

// AddBlock adds a block containing txs
// The block time forwarded by tb.StepMiliSeconds
func (tb *TestBlockChain) AddBlock(txs []*TxWithSigner) (*types.Block, error) {
	TimeoutCount := uint32(0)
	Generator, err := tb.Chain.TopGenerator(TimeoutCount)
	ctx := types.NewContext(tb.Chain.Store())
	Timestamp := uint64(time.Now().UnixNano()) + tb.StepMiliSeconds*1000000

	bc := chain.NewBlockCreator(tb.Chain, ctx, Generator, TimeoutCount, Timestamp, 0)
	var receipts = types.Receipts{}

	for _, tx := range txs {
		sig, err := tx.Signer.Sign(tx.Tx.Message())
		if err != nil {
			return nil, err
		}
		if receipt, err := bc.AddTx(tx.Tx, sig); err != nil {
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

	pk := tb.FrKeyMap[Generator]
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

	err = tb.Chain.ConnectBlock(b, nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// MustAddBlock adds a block containing txs without err
// The block time forwarded by tb.StepMiliSeconds
func (tb *TestBlockChain) MustAddBlock(txs []*TxWithSigner) *types.Block {
	b, err := tb.AddBlock(txs)
	if err != nil {
		panic(err)
	}
	return b
}

// newContext calls chain.NewContext()
func (tb *TestBlockChain) HandleJRPC(req *apiserver.JRPCRequest) interface{} {
	return tb.rpcapi.HandleJRPC(req)
}

// Close calls chain.Close()
func (tb *TestBlockChain) Close() {
	tb.Chain.Close()
	RemoveChainData(tb.Path)

	// bloobits 디렉토리 삭제
	os.RemoveAll(BloomDataPath)
}

// RemoveChainData removes data directory which includes data files
func RemoveChainData(path string) error {
	// if _, err := os.Stat("/mnt/ramdisk"); !os.IsNotExist(err) {
	// 	dir = "/mnt/ramdisk/" + dir
	// }

	return os.RemoveAll(path)
}
