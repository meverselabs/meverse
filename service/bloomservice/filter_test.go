package bloomservice

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/cmd/config"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
	"github.com/stretchr/testify/assert"
)

const bloomDataPath = "_bloombits"

var (
	bloomBitsBlocks = uint64(24)
	bloomConfirms   = uint64(10)

	transferHash     = hash.Hash([]byte("Transfer(address,address,uint256)"))
	approveHash      = hash.Hash([]byte("Approval(address,address,uint256)"))
	addLiquidityHash = hash.Hash([]byte("UniAddLiquidity(address,address,uint256,uint256,uint256,uint256)"))
	mintHash         = hash.Hash([]byte("Mint(address)"))
	balanceOfHash    = hash.Hash([]byte("BalanceOf(address)"))
)

// tsMock is the mock of txsearch service
// range로 검색할 경우 txsearch는 필요없기 때문에 mock으로 가능
type tsMock struct{}

func (t *tsMock) BlockHeight(bh hash.Hash256) (uint32, error) {
	return uint32(2), nil
}
func (t *tsMock) BlockList(index int) []*itxsearch.BlockInfo                     { return nil }
func (t *tsMock) Block(i uint32) (*types.Block, error)                           { return nil, nil }
func (t *tsMock) TxIndex(th hash.Hash256) (itxsearch.TxID, error)                { return itxsearch.TxID{}, nil }
func (t *tsMock) TxList(index, size int) ([]itxsearch.TxList, error)             { return nil, nil }
func (t *tsMock) Tx(height uint32, index uint16) (map[string]interface{}, error) { return nil, nil }
func (t *tsMock) AddressTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}
func (t *tsMock) TokenTxList(From common.Address, index, size int) ([]itxsearch.TxList, error) {
	return nil, nil
}
func (t *tsMock) Reward(cont, rewarder common.Address) (*amount.Amount, error) { return nil, nil }

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
	bs, err := NewBloomBitService(tb.chain, bloomDataPath, bloomBitsBlocks, bloomConfirms)
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

// test for the events of token contract : Mint, Burn, Approve, Tansfer, TransferFrom method
func TestTokenContractFilter(t *testing.T) {
	path := "_data"
	chainID := big.NewInt(1337)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := (aliceKey).PublicKey().Address(), (bobKey).PublicKey().Address(), (charlieKey).PublicKey().Address()

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, ret, err := prepare(path, true, chainID, version, alice, args, mevInitialize, &initContextInfo{})
	if err != nil {
		t.Fatal(err)
	}

	ctx := tb.newContext()

	mev := ret[0].(common.Address)

	assert := assert.New(t)

	mintAmount := amount.NewAmount(1, 0)
	burnAmount := amount.NewAmount(2, 0)
	approveAmount := amount.NewAmount(10, 0)
	transferAmount := amount.NewAmount(4, 0)
	transferFromAmount := amount.NewAmount(5, 0)

	testCases := []struct {
		tx     types.Transaction
		key    key.Key
		topics [4][]byte
	}{
		// Mint
		{
			types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Mint",
				Args:      bin.TypeWriteAll(bob, mintAmount),
			},
			aliceKey,
			[4][]byte{
				crypto.Keccak256([]byte("Transfer(address,address,uint256)")),
				common.LeftPadBytes([]byte{}, 32),
				common.LeftPadBytes(bob.Bytes(), 32),
				math.U256Bytes(mintAmount.Int),
			},
		},
		// Burn
		{
			types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Burn",
				Args:      bin.TypeWriteAll(burnAmount),
			},
			bobKey,
			[4][]byte{
				crypto.Keccak256([]byte("Transfer(address,address,uint256)")),
				common.LeftPadBytes(bob.Bytes(), 32),
				common.LeftPadBytes([]byte{}, 32),
				math.U256Bytes(burnAmount.Int),
			},
		},
		// Approve
		{
			types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Approve",
				Args:      bin.TypeWriteAll(charlie, approveAmount),
			},
			aliceKey,
			[4][]byte{
				crypto.Keccak256([]byte("Approval(address,address,uint256)")),
				common.LeftPadBytes(alice.Bytes(), 32),
				common.LeftPadBytes(charlie.Bytes(), 32),
				math.U256Bytes(approveAmount.Int),
			},
		},
		//Transfer
		{
			types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "Transfer",
				Args:      bin.TypeWriteAll(alice, transferAmount),
			},
			charlieKey,
			[4][]byte{
				crypto.Keccak256([]byte("Transfer(address,address,uint256)")),
				common.LeftPadBytes(charlie.Bytes(), 32),
				common.LeftPadBytes(alice.Bytes(), 32),
				math.U256Bytes(transferAmount.Int),
			},
		},
		//TransferFrom
		{
			types.Transaction{
				ChainID:   ctx.ChainID(),
				Timestamp: uint64(time.Now().UnixNano()),
				To:        mev,
				Method:    "TransferFrom",
				Args:      bin.TypeWriteAll(alice, bob, transferFromAmount),
			},
			charlieKey,
			[4][]byte{
				crypto.Keccak256([]byte("Transfer(address,address,uint256)")),
				common.LeftPadBytes(alice.Bytes(), 32),
				common.LeftPadBytes(bob.Bytes(), 32),
				math.U256Bytes(transferFromAmount.Int),
			},
		},
	}

	for _, test := range testCases {

		b, err := tb.addBlock(ctx, []*txWithSigner{{&test.tx, test.key}})
		if err != nil {
			t.Fatal(err)
		}

		_, logs, err := TxLogsBloom(tb.chain, b, 0, nil)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(logs[0].Topics)
		assert.Equal(len(logs[0].Topics), 4)
		for j := 0; j < 4; j++ {
			assert.Equal(logs[0].Topics[j], common.BytesToHash(test.topics[j]))
		}
		ctx = types.NewContext(tb.chain.Store())
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
	defer tb.Close()

	mev := ret[0].(common.Address)

	ctx := tb.newContext()

	if err := os.RemoveAll(bloomDataPath); err != nil {
		t.Fatal(err)
	}
	bs, err := NewBloomBitService(tb.chain, bloomDataPath, bloomBitsBlocks, bloomConfirms)
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
		if err = os.RemoveAll(bloomDataPath); err != nil {
			t.Fatal(err)
		}

		bs, err := NewBloomBitService(tb.chain, bloomDataPath, bloomBitsBlocks, bloomConfirms)
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

// StorageWithEventContractCreation deploy StorageWithEvent contract by alice with nonce = 0
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventContractCreation(tb *testBlockChain) (*types.Transaction, error) {
	rlp := "0x02f90cca8205398080842c9d07ea841dcd65008080b90c72608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610c12806100606000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632ecb20d31461005157806342526e4e1461008157806360fe47b1146100b15780636d4ce63c146100cd575b600080fd5b61006b60048036038101906100669190610710565b6100eb565b604051610078919061085e565b60405180910390f35b61009b600480360381019061009691906106a6565b61045a565b6040516100a891906107df565b60405180910390f35b6100cb60048036038101906100c691906106e7565b610503565b005b6100d561060a565b6040516100e29190610843565b60405180910390f35b60007f30000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156101cb57507f39000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610206577f300000000000000000000000000000000000000000000000000000000000000060f81c826101ff91906109ba565b9050610455565b7f61000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156102e457507f66000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b1561032b577f610000000000000000000000000000000000000000000000000000000000000060f81c82600a61031a9190610935565b61032491906109ba565b9050610455565b7f41000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161015801561040957507f46000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610450577f410000000000000000000000000000000000000000000000000000000000000060f81c82600a61043f9190610935565b61044991906109ba565b9050610455565b600090505b919050565b600080600090506000805b60288160ff1610156104f85760108361047e919061096c565b92506104d2858260ff16815181106104bf577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b602001015160f81c60f81b60f81c6100eb565b60ff16915081836104e391906108eb565b925080806104f090610a9b565b915050610465565b508192505050919050565b80600181905550600061052d604051806060016040528060288152602001610bb56028913961045a565b905060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166104d27f0ca6778676cf982a32c6c02a2aa57966620f6581c6f016a2837cb60c2e9f5c5361271060405161059a91906107fa565b60405180910390a38073ffffffffffffffffffffffffffffffffffffffff166040516105c5906107ca565b60405180910390207f1658817f2f384793b2fe883b68c3df91c0122e6b4ed86878f4c10e494082d8cc614e206040516105fe9190610828565b60405180910390a35050565b6000600154905090565b60006106276106228461089e565b610879565b90508281526020810184848401111561063f57600080fd5b61064a848285610a5b565b509392505050565b600082601f83011261066357600080fd5b8135610673848260208601610614565b91505092915050565b60008135905061068b81610b86565b92915050565b6000813590506106a081610b9d565b92915050565b6000602082840312156106b857600080fd5b600082013567ffffffffffffffff8111156106d257600080fd5b6106de84828501610652565b91505092915050565b6000602082840312156106f957600080fd5b60006107078482850161067c565b91505092915050565b60006020828403121561072257600080fd5b600061073084828501610691565b91505092915050565b610742816109ee565b82525050565b61075181610a37565b82525050565b61076081610a49565b82525050565b60006107736004836108cf565b915061077e82610b34565b602082019050919050565b60006107966007836108e0565b91506107a182610b5d565b600782019050919050565b6107b581610a20565b82525050565b6107c481610a2a565b82525050565b60006107d582610789565b9150819050919050565b60006020820190506107f46000830184610739565b92915050565b600060408201905061080f6000830184610748565b818103602083015261082081610766565b905092915050565b600060208201905061083d6000830184610757565b92915050565b600060208201905061085860008301846107ac565b92915050565b600060208201905061087360008301846107bb565b92915050565b6000610883610894565b905061088f8282610a6a565b919050565b6000604051905090565b600067ffffffffffffffff8211156108b9576108b8610af4565b5b6108c282610b23565b9050602081019050919050565b600082825260208201905092915050565b600081905092915050565b60006108f682610a00565b915061090183610a00565b92508273ffffffffffffffffffffffffffffffffffffffff0382111561092a57610929610ac5565b5b828201905092915050565b600061094082610a2a565b915061094b83610a2a565b92508260ff0382111561096157610960610ac5565b5b828201905092915050565b600061097782610a00565b915061098283610a00565b92508173ffffffffffffffffffffffffffffffffffffffff04831182151516156109af576109ae610ac5565b5b828202905092915050565b60006109c582610a2a565b91506109d083610a2a565b9250828210156109e3576109e2610ac5565b5b828203905092915050565b60006109f982610a00565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600060ff82169050919050565b6000610a4282610a20565b9050919050565b6000610a5482610a20565b9050919050565b82818337600083830152505050565b610a7382610b23565b810181811067ffffffffffffffff82111715610a9257610a91610af4565b5b80604052505050565b6000610aa682610a2a565b915060ff821415610aba57610ab9610ac5565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b7f6162636400000000000000000000000000000000000000000000000000000000600082015250565b7f7365743536373800000000000000000000000000000000000000000000000000600082015250565b610b8f81610a20565b8114610b9a57600080fd5b50565b610ba681610a2a565b8114610bb157600080fd5b5056fe33633434636464646236613930306661326235383564643239396530336431326661343239336263a26469706673582212200227ddd9f59da8a18fa533bb272780c8c899d4fa64a28dc408b7d65c3430c50364736f6c63430008040033c080a01d6bb58ee614b7906d7903798b4e8dc635627f9e1b4840e781dc814491cea11ea004a4e1dae0b64bfbbd27a9966192c1d0d23eb756692d38e4a36a26a2103b156c"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         0,
		To:          ZeroAddress,
		Method:      "",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}

// StorageWithEventSet returns StorageWithEvent's Set function by alice with nonce = 1
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventSet(tb *testBlockChain) (*types.Transaction, error) {
	rlp := "0x02f88e8205390180842c9d07ea841dcd6500945fbdb2315678afecb367f032d93f642f64180aa380a460fe47b1000000000000000000000000000000000000000000000000000000003ade68b1c001a06e1ab9f8e0b00a24946a95475d04adae3fd4cd5d92baf14178a5112247b554d1a01d30e5a83a548f34543b8a986404332bf34bb0cdce3d9a2bc0522194406c669f"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         1,
		To:          common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Method:      "",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}
