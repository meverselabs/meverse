package bloomservice

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/cmd/config"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/params"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
	"github.com/stretchr/testify/assert"
)

var (
	transferHash     = hash.Hash([]byte("Transfer(address,address,uint256)"))
	approveHash      = hash.Hash([]byte("Approval(address,address,uint256)"))
	addLiquidityHash = hash.Hash([]byte("UniAddLiquidity(address,address,uint256,uint256,uint256,uint256)"))
	mintHash         = hash.Hash([]byte("Mint(address)"))
	balanceOfHash    = hash.Hash([]byte("BalanceOf(address)"))
)

// tsMock is the mock of txsearch service
// range로 검색할 경우 txsearch는 필요없기 때문에 mock으로 가능
type tsMock struct{}

func (*tsMock) BlockHeight(bh hash.Hash256) (uint32, error) {
	return uint32(2), nil
}
func (*tsMock) BlockList(index int) []*itxsearch.BlockInfo                     { return nil }
func (*tsMock) Block(i uint32) (*types.Block, error)                           { return nil, nil }
func (*tsMock) TxIndex(th hash.Hash256) (itxsearch.TxID, error)                { return itxsearch.TxID{}, nil }
func (*tsMock) TxList(index, size int) ([]itxsearch.TxList, error)             { return nil, nil }
func (*tsMock) Tx(height uint32, index uint16) (map[string]interface{}, error) { return nil, nil }
func (*tsMock) AddressTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}
func (*tsMock) TokenTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}
func (*tsMock) Reward(cont, rewarder common.Address) (*amount.Amount, error) { return nil, nil }

func TestPackBoolAndNumber(t *testing.T) {
	tests := []struct {
		value  reflect.Value
		packed []byte
	}{
		{reflect.ValueOf(false), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(true), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},

		// Protocol limits
		{reflect.ValueOf(0), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(1), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},
		{reflect.ValueOf(-1), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},

		// Type corner cases
		{reflect.ValueOf(uint8(math.MaxUint8)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000ff")},
		{reflect.ValueOf(uint16(math.MaxUint16)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000ffff")},
		{reflect.ValueOf(uint32(math.MaxUint32)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000ffffffff")},
		{reflect.ValueOf(uint64(math.MaxUint64)), common.Hex2Bytes("000000000000000000000000000000000000000000000000ffffffffffffffff")},

		{reflect.ValueOf(int8(math.MaxInt8)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000007f")},
		{reflect.ValueOf(int16(math.MaxInt16)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000007fff")},
		{reflect.ValueOf(int32(math.MaxInt32)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000007fffffff")},
		{reflect.ValueOf(int64(math.MaxInt64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},

		{reflect.ValueOf(int8(math.MinInt8)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80")},
		{reflect.ValueOf(int16(math.MinInt16)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000")},
		{reflect.ValueOf(int32(math.MinInt32)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000")},
		{reflect.ValueOf(int64(math.MinInt64)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000")},

		{reflect.ValueOf(big.NewInt(int64(math.MaxInt64))), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},
		{reflect.ValueOf(amount.NewAmount(0, uint64(math.MaxInt64))), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},
	}
	for i, tt := range tests {
		packed, err := packElement(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(packed, tt.packed) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, packed, tt.packed)
		}
	}
}

func TestPack(t *testing.T) {
	tests := []struct {
		value  reflect.Value
		packed []byte
	}{
		{reflect.ValueOf(common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")), common.Hex2Bytes("000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266")},
		{reflect.ValueOf("a"), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000016100000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([2]string{"a", "b"}), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([]string{"a", "b"}), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([2]byte{10, 20}), common.Hex2Bytes("0a14000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([]byte{10, 20}), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000020a14000000000000000000000000000000000000000000000000000000000000")},
	}
	for i, tt := range tests {
		packed, err := packElement(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(packed, tt.packed) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, packed, tt.packed)
		}
	}
}

func TestMutliTranactionFilter(t *testing.T) {
	path := "_data"
	chainID := big.NewInt(1337)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := (aliceKey).PublicKey().Address(), (bobKey).PublicKey().Address(), (charlieKey).PublicKey().Address()

	intialize := func(ctx *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {
		tRet, err := mevInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}

		dRet, err := dexInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}
		return append(tRet, dRet...), nil
	}

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, ret, err := prepare(path, true, chainID, version, alice, args, intialize, &initContextInfo{})
	if err != nil {
		t.Fatal(err)
	}

	stepMiliSeconds := uint64(1000)
	ctx := tb.newContext()

	mev := ret[0].(common.Address)
	_, router, _, token0, token1, _ := ret[1].(common.Address), ret[2].(common.Address), ret[3].(common.Address), ret[4].(common.Address), ret[5].(common.Address), ret[6].(common.Address)

	// block 1 - StorageWithEvent Contract 생성
	tx0, err := StorageWithEventContractCreation(tb)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err = tb.addBlockAndSleep(ctx, []*txWithSigner{{tx0, aliceKey}}, stepMiliSeconds)
	if err != nil {
		t.Fatal(err)
	}

	storageWithEvent := crypto.CreateAddress(alice, tx0.Seq)
	fmt.Println("storageWithEvent  Address", storageWithEvent.String())

	txs := []*txWithSigner{}
	//signers := []key.Key{}
	transferAmount := amount.NewAmount(0, 10000)

	// block 2
	from := 2
	tx0, err = StorageWithEventSet(tb)
	if err != nil {
		t.Fatal(err)
	}
	txs = append(txs, &txWithSigner{tx0, aliceKey})

	singersLen := 2
	for i := 0; i < singersLen; i++ {
		for j := 0; j < singersLen; j++ {
			if i == j {
				continue
			}
			txApprove := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(userKeys[j].PublicKey().Address(), MaxUint256),
			}
			txs = append(txs, &txWithSigner{txApprove, userKeys[i]})

			txTransfer := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(userKeys[i].PublicKey().Address(), transferAmount),
			}
			txs = append(txs, &txWithSigner{txTransfer, userKeys[i]})

			txTransferFrom := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "TransferFrom",
				Args:      bin.TypeWriteAll(userKeys[i].PublicKey().Address(), userKeys[(i+1)%singersLen].PublicKey().Address(), transferAmount),
			}
			txs = append(txs, &txWithSigner{txTransferFrom, userKeys[j]})
		}
	}

	token0Amount := amount.NewAmount(1, 0)
	token1Amount := amount.NewAmount(4, 0)
	tx3 := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        router,
		Method:    "UniAddLiquidity",
		Args:      bin.TypeWriteAll(token0, token1, token0Amount, token1Amount, amount.ZeroCoin, amount.ZeroCoin),
	}

	txs = append(txs, &txWithSigner{tx3, aliceKey})

	ctx, err = tb.addBlockAndSleep(ctx, txs, stepMiliSeconds)
	if err != nil {
		t.Fatal(err)
	}

	provider := tb.chain.Provider()
	b, err := provider.Block(provider.Height())
	if err != nil {
		t.Fatal(err)
	}
	bHash := bin.MustWriterToHash(&b.Header)

	testCases := []struct {
		crit  FilterQuery
		count int
	}{
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{transferHash}}}, 6},
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{approveHash}}}, 2},
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{transferHash}, {approveHash}}}, 0},
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{transferHash, approveHash}}}, 8},
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{addLiquidityHash, mintHash, balanceOfHash}}}, 4},
		{FilterQuery{FromBlock: big.NewInt(int64(from)), Topics: [][]common.Hash{{addLiquidityHash}, {mintHash}}}, 0},
		{FilterQuery{BlockHash: &bHash}, 15}, // 2 + 6 + 7
		{FilterQuery{Addresses: []common.Address{router}, BlockHash: &bHash}, 1},
	}

	ts := &tsMock{}
	bs, err := NewBloomBitService(tb.chain, "_bloombits", params.BloomBitsBlocks, params.BloomConfirms)
	if err != nil {
		panic(err)
	}
	for i, test := range testCases {
		logs, err := FilterLogs(tb.chain, ts, bs, test.crit)
		if err != nil {
			t.Errorf("filter query for case %d : err %v", i, err)
		}
		if len(logs) != test.count {
			t.Errorf("filter query for case %d : got %d, have %d", i, len(logs), test.count)
		}
	}

	removeChainData(path)
}

func TestFilterIndexed(t *testing.T) {
	path := "_data"
	chainID := big.NewInt(1337)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, ret, err := prepare(path, true, chainID, version, alice, args, mevInitialize, &initContextInfo{})
	if err != nil {
		t.Fatal(err)
	}
	mev := ret[0].(common.Address)

	ctx := tb.newContext()

	bs, err := NewBloomBitService(tb.chain, "_bloombits", params.BloomBitsBlocks, params.BloomConfirms)
	if err != nil {
		t.Fatal(err)
	}
	transferAmount := amount.NewAmount(0, 1)

	for i := 0; i < 100; i++ {
		var block *types.Block
		if i%2 == 0 {
			block, err = tb.addBlock(ctx, []*txWithSigner{})
			if err != nil {
				t.Fatal(err)
			}
		} else {
			tx := &types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(bob, transferAmount),
			}
			block, err = tb.addBlock(ctx, []*txWithSigner{{tx, aliceKey}})
			if err != nil {
				t.Fatal(err)
			}
		}
		bs.OnBlockConnected(block, nil)
		ctx = types.NewContext(tb.chain.Store())
	}

	crit := FilterQuery{FromBlock: big.NewInt(0), Topics: [][]common.Hash{{transferHash}}}

	assert := assert.New(t)

	ts := &tsMock{}
	{
		logs, err := FilterLogs(tb.chain, ts, bs, crit)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(len(logs), 50, "logs :", len(logs))
	}
}

func TestFilterInitWith(t *testing.T) {

	chainID := big.NewInt(1337)
	version := uint16(1)
	workPath := "./_data"
	path := "./_data" + strconv.Itoa(int(chainID.Uint64()))

	userKeys, err := getSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()
	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}

	bloomBitsBlocks := uint64(24)
	bloomConfirms := uint64(10)

	tests := []struct {
		initHeight         int
		storedSections     uint64 // Number of sections successfully indexed into the database
		knownSections      uint64 // Number of sections known to be complete (block wise)
		checkpointSections uint64 // Number of sections covered by the checkpoint
	}{
		{0, 3, 3, 0},
		{12, 4, 4, 0},
		{23, 4, 4, 1},
		{24, 4, 4, 1},
		{40, 5, 5, 1},
		{47, 5, 5, 2},
		{48, 5, 5, 2},
		{49, 5, 5, 2},
		{111, 8, 8, 4},
	}

	intialize := mevInitialize
	blocks := 100
	assert := assert.New(t)
	var transferLogHeights []uint64
	for i, test := range tests {
		err := createContextFile(path, workPath, chainID, version, args, intialize, test.initHeight)
		if err != nil {
			t.Fatal(err)
		}

		cfg := initContextInfo{}
		if test.initHeight != 0 {
			if err := config.LoadFile(path+"/config"+strconv.Itoa(test.initHeight)+".toml", &cfg); err != nil {
				t.Fatal(err)
			}
		}

		tb, ret, err := prepare(workPath, false, chainID, version, alice, args, mevInitialize, &cfg)
		if err != nil {
			t.Fatal(err)
		}
		mev := ret[0].(common.Address)
		ctx := tb.newContext()

		// 디렉토리 삭제
		dir := "_bloombits"
		if err = os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}

		bs, err := NewBloomBitService(tb.chain, dir, bloomBitsBlocks, bloomConfirms)
		if err != nil {
			t.Fatal(err)
		}

		transferAmount := amount.NewAmount(0, 1)
		for i := 0; i < blocks; i++ {
			var block *types.Block
			if i%2 == 0 {
				block, err = tb.addBlock(ctx, []*txWithSigner{})
				if err != nil {
					t.Fatal(err)
				}
			} else {
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: uint64(time.Now().UnixNano()),
					To:        mev,
					Method:    "Transfer",
					Args:      bin.TypeWriteAll(bob, transferAmount),
				}
				block, err = tb.addBlock(ctx, []*txWithSigner{{tx, aliceKey}})
				if err != nil {
					t.Fatal(err)
				}
			}
			bs.OnBlockConnected(block, nil)
			ctx = types.NewContext(tb.chain.Store())
		}

		height := tb.chain.Provider().Height()

		crit := FilterQuery{FromBlock: big.NewInt(0), Topics: [][]common.Hash{{transferHash}}}
		ts := &tsMock{}

		logs, err := FilterLogs(tb.chain, ts, bs, crit)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(bs.indexer.storedSections, test.storedSections, "storedSections : ", bs.indexer.storedSections, test.storedSections)
		assert.Equal(bs.indexer.knownSections, test.knownSections, "knownSections : ", bs.indexer.knownSections, test.knownSections)
		assert.Equal(bs.indexer.checkpointSections, test.checkpointSections, "initHeight :", test.initHeight, "checkpointSections : ", bs.indexer.checkpointSections, test.checkpointSections)
		assert.Equal(height, uint32(test.initHeight+blocks), "initHeight :", test.initHeight, "height : ", height, test.initHeight+blocks)
		assert.Equal(len(logs), blocks/2, "logs :", len(logs))

		// log의 numbering 이 제대로 되는지
		if i == 0 {
			transferLogHeights = make([]uint64, len(logs))
			for j, l := range logs {
				transferLogHeights[j] = l.BlockNumber
			}
		} else {
			for j, l := range logs {
				assert.Equal(l.BlockNumber, transferLogHeights[j]+uint64(test.initHeight), "log BlockNumber", l.BlockHash, transferLogHeights[j]+uint64(test.initHeight))
			}
		}

		tb.Close()
	}

}

// createContextFile creates context File upto height and config<height>.toml
func createContextFile(path string, workPath string, chainID *big.Int, version uint16, args []interface{}, genesisInitFunc func(*types.Context, map[string]uint64, []interface{}) ([]interface{}, error), height int) error {

	chainAdmin := args[0].(common.Address)
	tb, _, err := prepare(path, true, chainID, version, chainAdmin, args, genesisInitFunc, &initContextInfo{})
	if err != nil {
		return err
	}

	ctx := tb.newContext()

	for i := 0; i < height; i++ {
		ctx, err = tb.addBlockAndSleep(ctx, []*txWithSigner{}, 1000)
		if err != nil {
			return err
		}
	}

	// _config.toml 화일생성
	configFile, err := os.Create(path + "/config" + strconv.Itoa(height) + ".toml")
	if err != nil {
		return err
	}
	defer configFile.Close()

	provider := tb.chain.Provider()
	InitGenesisHash, err := provider.Hash(0)
	if err != nil {
		return err
	}
	InitHash, err := provider.Hash(uint32(height))
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	buffer.WriteString("InitGenesisHash = \"")
	buffer.WriteString(InitGenesisHash.String())
	buffer.WriteString("\"\n")
	buffer.WriteString("InitHeight = ")
	buffer.WriteString(strconv.Itoa(height))
	buffer.WriteString("\n")
	buffer.WriteString("InitHash = \"")
	buffer.WriteString(InitHash.String())
	buffer.WriteString("\"\n")
	buffer.WriteString("InitTimestamp = ")
	buffer.WriteString(strconv.FormatUint(provider.LastTimestamp(), 10))
	buffer.WriteString("\n")

	if _, err = configFile.Write(buffer.Bytes()); err != nil {
		return err
	}

	tb.Close()

	// chain 디렉토리 삭제, ontext 화일 복사
	err = os.RemoveAll(path + "/chain")
	if err != nil {
		return err
	}

	original, err := os.Open(path + "/context")
	if err != nil {
		return err
	}

	copy1, err := os.Create(path + "/context" + strconv.Itoa(height))
	if err != nil {
		return err
	}

	if _, err := io.Copy(copy1, original); err != nil {
		return err
	}
	original.Close()
	copy1.Close()

	// _data 삭제 후 재생성, context화일 복사
	err = os.RemoveAll(workPath)
	if err != nil {
		return err
	}
	original, err = os.Open(path + "/context")
	if err != nil {
		return err
	}
	if err = os.Mkdir(workPath, os.ModePerm); err != nil {
		return err
	}
	copy2, err := os.Create(workPath + "/context")
	if err != nil {
		return err
	}

	if _, err := io.Copy(copy2, original); err != nil {
		return err
	}
	original.Close()
	copy2.Close()

	return nil
}
