package test

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/assert"

	emath "github.com/ethereum/go-ethereum/common/math"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/bloomservice"

	. "github.com/meverselabs/meverse/tests/lib"
)

func TestMutliTranactionEvents(t *testing.T) {

	chainID := big.NewInt(1337) // StorageWithEventContractCreation is made by chainID = 1337

	userKeys, err := GetSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	intialize := func(ctx *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {
		tRet, err := MevInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}

		dRet, err := DexInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}
		return append(tRet, dRet...), nil
	}

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, ret, err := Prepare(ChainDataPath, true, chainID, Version, alice, args, intialize, &InitContextInfo{})
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Close()

	provider := tb.Chain.Provider()
	//ctx := tb.NewContext()

	mev := NewTokenContract(ret[0].(*common.Address))
	router := NewRouterContract(ret[2].(*common.Address))
	token0, token1, pair := ret[4].(*common.Address), ret[5].(*common.Address), ret[6].(*common.Address)

	// block 1 - StorageWithEvent Contract 생성
	var storageWithEvent common.Address
	if tx, err := StorageWithEventContractCreation(tb); err != nil {
		t.Fatal(err)
	} else {
		tx0 := NewTxWithSigner(tx, aliceKey)
		b, err := tb.AddBlock([]*TxWithSigner{tx0})
		if err != nil {
			t.Fatal(err)
		}
		receipts, _ := provider.Receipts(b.Header.Height)
		storageWithEvent = receipts[0].ContractAddress

		// storageWithEvent = crypto.CreateAddress(alice, tx0.Tx.Seq)
		fmt.Println("storageWithEvent  Address", storageWithEvent.String())
	}

	// block 2
	var tx0 *TxWithSigner
	if tx, err := StorageWithEventSet(tb); err != nil {
		t.Fatal(err)
	} else {
		tx0 = NewTxWithSigner(tx, aliceKey)
	}

	transferAmount := amount.NewAmount(1, 0)
	tx1 := mev.TransferTx(aliceKey, provider, bob, transferAmount)

	tx2 := mev.ApproveTx(bobKey, provider, charlie, MaxUint256)

	token0Amount := amount.NewAmount(1, 0)
	token1Amount := amount.NewAmount(4, 0)
	tx3 := router.UniAddLiquidityTx(aliceKey, provider, *token0, *token1, token0Amount, token1Amount, amount.ZeroCoin, amount.ZeroCoin)

	b, err := tb.AddBlock([]*TxWithSigner{tx0, tx1, tx2, tx3})
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)
	var tIdx uint16
	var txBloom etypes.Bloom

	//tx 0
	tIdx = 0
	bHeight := provider.Height()
	receipts, err := provider.Receipts(bHeight)
	if err != nil {
		t.Fatal(err)
	}
	receipt := receipts[tIdx]
	txBloom = etypes.CreateBloom(etypes.Receipts{receipt})

	assertStorageWithEventSet := func(bloom etypes.Bloom) {
		// positive
		assert.Equal(bloom.Test(storageWithEvent[:]), true, "storageWithEvent :", storageWithEvent)
		eventFunc := "Set1(uint256,address,uint256,string)"
		assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1234))), true, "1234")
		assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)

		eventFunc = "Set2(string,address,uint256)"
		assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("set5678"))), true, "set5678")
		assert.Equal(bloom.Test(common.LeftPadBytes(common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc").Bytes(), 32)), true, "0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc : ")

		// negative : unindexed arguments
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(10000))), false, "10000")
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("abcd"))), false, "abcd")
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(20000))), false, "20000")
	}

	assertStorageWithEventSet(txBloom)

	// tx 1
	tIdx = 1
	tx := b.Body.Transactions[tIdx]
	//evs, err := FindTransactionsEvents(b.Body.Transactions, b.Body.Events, tIdx)
	evs, err := bloomservice.FindCallHistoryEvents(b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = bloomservice.CreateEventBloom(tb.Provider, evs)
	if err != nil {
		t.Fatal(err)
	}

	assertMevTransfer := func(bloom etypes.Bloom) {
		// positive
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(bob.Bytes(), 32)), true, "to : ", bob)
			assert.Equal(bloom.Test(emath.U256Bytes(transferAmount.Int)), true, "transferAmount : ", transferAmount)
		}

		// negative
		{
			assert.Equal(bloom.Test(crypto.Keccak256([]byte("Transfer(address,address)"))), false)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)
		}
	}

	assertMevTransfer(txBloom)

	// tx 2
	tIdx = 2
	tx = b.Body.Transactions[tIdx]
	evs, err = bloomservice.FindCallHistoryEvents(b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = bloomservice.CreateEventBloom(tb.Provider, evs)
	if err != nil {
		t.Fatal(err)
	}

	assertMevApprove := func(bloom etypes.Bloom) {

		// positive
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "Approval(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(charlie.Bytes(), 32)), true, "to : ", charlie)
			assert.Equal(bloom.Test(emath.U256Bytes(MaxUint256.Int)), true, "approveAmount : ", MaxUint256)
		}

		// negative
		{
			assert.Equal(bloom.Test(crypto.Keccak256([]byte("Approve(address,uint256)"))), false)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)
		}
	}

	assertMevApprove(txBloom)

	// tx 3
	tIdx = 3
	tx = b.Body.Transactions[tIdx]
	evs, err = bloomservice.FindCallHistoryEvents(b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = bloomservice.CreateEventBloom(tb.Provider, evs)
	if err != nil {
		t.Fatal(err)
	}

	assertUniswapAddLiquidity := func(bloom etypes.Bloom) {

		// positive
		// 0 : router.UniAddLiquidity
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "UniAddLiquidity(address,address,uint256,uint256,uint256,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(token0.Bytes(), 32)), true, "token0 : ", token0)
			assert.Equal(bloom.Test(common.LeftPadBytes(token1.Bytes(), 32)), true, "token1 : ", token1)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token0Amount : ", token0Amount)
			assert.Equal(bloom.Test(emath.U256Bytes(token1Amount.Int)), true, "token0Amount : ", token1Amount)
			assert.Equal(bloom.Test(emath.U256Bytes(ZeroAmount.Int)), true, "ZeroAmount : ", ZeroAmount)
		}
		// 1 : pair.Reserve
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Address.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(pair[:]), true, "to : ", pair)
			eventFunc := "Reserves()"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		}
		// 2 : TransferFrom router -> token0.Transferfrom(alice,pair,amt)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Address.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(token0[:]), true, "to : ", token0)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token0Amount : ", token0Amount)
		}
		// 3 : TransferFrom router -> token1.Transferfrom(alice,pair,amt)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Address.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(token1[:]), true, "to : ", token1)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token1Amount : ", token1Amount)
		}
		// 4 : pair.Mint
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Address.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(pair[:]), true, "to : ", pair)
			eventFunc := "Mint(address)" // "Mint(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
		}

		// 5 : BalanceOf pair : token0.BalanceOf(pair)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "from : ", pair)
			assert.Equal(bloom.Test(token0[:]), true, "to : ", token0)
			eventFunc := "BalanceOf(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
		}

		// 6 : BalanceOf pair : token1.BalanceOf(pair)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "from : ", pair)
			assert.Equal(bloom.Test(token1[:]), true, "to : ", token1)
			eventFunc := "BalanceOf(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
		}

		// negative
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("UniAddLiquiditi(address,address,uint256,uint256,uint256,uint256)"))), false)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("BalanceOg(address)"))), false)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("Transfer(address,address)"))), false)
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)

	}

	assertUniswapAddLiquidity(txBloom)

	blockBloom, err := bloomservice.BlockLogsBloom(tb.Chain, b)
	if err != nil {
		t.Fatal(err)
	}
	assertStorageWithEventSet(blockBloom)
	assertMevTransfer(blockBloom)
	assertMevApprove(blockBloom)
	assertUniswapAddLiquidity(blockBloom)

	RemoveChainData(tb.Path)
}

// RewardPool contract uses mpl token (go Contract)
// Claim events : RewardPool Claim event +  Mpl Transfer Event
func TestEventLogFromEvm(t *testing.T) {

	userKeys, err := GetSingers(ChainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, _, err := Prepare(ChainDataPath, true, ChainID, Version, alice, args, MevInitialize, &InitContextInfo{})
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Close()

	provider := tb.Chain.Provider()
	//ctx := tb.NewContext()

	// block1
	mpl, err := NewMrc20Token(tb, aliceKey, "Meverse Play", "MPL")
	if err != nil {
		t.Fatal(err)
	}

	// rewardPool create tx
	var pool *RewardPool
	if cont, tx, err := NewRewardPoolTx(aliceKey, 0, provider, mpl.Address); err != nil {
		t.Fatal(err)
	} else {
		pool = cont
		b, err := tb.AddBlock([]*TxWithSigner{tx})
		if err != nil {
			t.Fatal(err)
		}
		receipts, _ := provider.Receipts(b.Header.Height)
		pool.SetAddress(&receipts[0].ContractAddress)
		log.Println("pool address = ", pool.Address)
	}

	// mpl approve
	if tx, err := mpl.ApproveTx(aliceKey, 0, pool.Address, MaxUint256.Int); err != nil {
		t.Fatal(err)
	} else {
		_, err = tb.AddBlock([]*TxWithSigner{tx})
		if err != nil {
			t.Fatal(err)
		}
	}

	// tx3: addReward to bob
	total := big.NewInt(123456)
	userReward := UserReward{
		User:   bob,
		Amount: total,
	}
	if tx, err := pool.AddRewardTx(aliceKey, 0, total, []UserReward{userReward}); err != nil {
		t.Fatal(err)
	} else {
		_, err = tb.AddBlock([]*TxWithSigner{tx})
		if err != nil {
			t.Fatal(err)
		}
	}

	// tx4 : clamin by bob
	var (
		txBloom etypes.Bloom
		logs    []*etypes.Log
		b       *types.Block
	)

	if tx, err := pool.ClaimTx(bobKey, 0); err != nil {
		t.Fatal(err)
	} else {
		b, err = tb.AddBlock([]*TxWithSigner{tx})
		if err != nil {
			t.Fatal(err)
		}
		receipts, _ := provider.Receipts(b.Header.Height)

		txBloom, logs, err = bloomservice.TxLogsBloom(tb.Chain, b, 0, receipts[0])
		if err != nil {
			t.Fatal(err)
		}
	}

	assert := assert.New(t)
	assertRewardPoolClaim := func(bloom etypes.Bloom, logs []*etypes.Log) {
		{
			//bloom
			assert.Equal(bloom.Test(mpl.Address[:]), true, "mpl.address :", mpl.Address)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(pool.Address.Bytes(), 32)), true, "pool.address : ", pool.Address)
			assert.Equal(bloom.Test(common.LeftPadBytes(bob.Bytes(), 32)), true, "bob address : ", bob)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(123456))), true, "claim amount : ", big.NewInt(123456))

			//log
			if logs != nil {
				assert.Equal(logs[0].Topics[0].Bytes(), crypto.Keccak256([]byte(eventFunc)), eventFunc)
				assert.Equal(logs[0].Topics[1].Bytes(), common.LeftPadBytes(pool.Address.Bytes(), 32), "pool.address : ", pool.Address)
				assert.Equal(logs[0].Topics[2].Bytes(), common.LeftPadBytes(bob.Bytes(), 32), "bob address : ", bob)
				assert.Equal(logs[0].Topics[3].Bytes(), emath.U256Bytes(big.NewInt(123456)), "claim amount : ", big.NewInt(123456))
			}

		}

		{
			//bloom
			assert.Equal(bloom.Test(pool.Address[:]), true, "pool.address :", pool.Address)
			eventFunc := "Claim(address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(bob.Bytes(), 32)), true, "bob address : ", bob)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(123456))), true, "claim amount : ", big.NewInt(123456))

			//log
			if logs != nil {
				assert.Equal(logs[1].Topics[0].Bytes(), crypto.Keccak256([]byte(eventFunc)), eventFunc)
				assert.Equal(logs[1].Topics[1].Bytes(), common.LeftPadBytes(bob.Bytes(), 32), "bob address : ", bob)
				assert.Equal(logs[1].Topics[2].Bytes(), emath.U256Bytes(big.NewInt(123456)), "claim amount : ", big.NewInt(123456))
			}
		}

	}

	assertRewardPoolClaim(txBloom, logs)

	blockBloom, err := bloomservice.BlockLogsBloom(tb.Chain, b)
	if err != nil {
		t.Fatal(err)
	}

	assertRewardPoolClaim(blockBloom, nil)

	// block bloom과 tx bloom이 서로 일치하는 지
	assert.Equal(txBloom, blockBloom, true, "bloom")

}

// StorageWithEventContractCreation deploy StorageWithEvent contract by alice with nonce = 0
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventContractCreation(tb *TestBlockChain) (*types.Transaction, error) {
	rlp := "0x02f90cca8205398080842c9d07ea841dcd65008080b90c72608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610c12806100606000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632ecb20d31461005157806342526e4e1461008157806360fe47b1146100b15780636d4ce63c146100cd575b600080fd5b61006b60048036038101906100669190610710565b6100eb565b604051610078919061085e565b60405180910390f35b61009b600480360381019061009691906106a6565b61045a565b6040516100a891906107df565b60405180910390f35b6100cb60048036038101906100c691906106e7565b610503565b005b6100d561060a565b6040516100e29190610843565b60405180910390f35b60007f30000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156101cb57507f39000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610206577f300000000000000000000000000000000000000000000000000000000000000060f81c826101ff91906109ba565b9050610455565b7f61000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156102e457507f66000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b1561032b577f610000000000000000000000000000000000000000000000000000000000000060f81c82600a61031a9190610935565b61032491906109ba565b9050610455565b7f41000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161015801561040957507f46000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610450577f410000000000000000000000000000000000000000000000000000000000000060f81c82600a61043f9190610935565b61044991906109ba565b9050610455565b600090505b919050565b600080600090506000805b60288160ff1610156104f85760108361047e919061096c565b92506104d2858260ff16815181106104bf577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b602001015160f81c60f81b60f81c6100eb565b60ff16915081836104e391906108eb565b925080806104f090610a9b565b915050610465565b508192505050919050565b80600181905550600061052d604051806060016040528060288152602001610bb56028913961045a565b905060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166104d27f0ca6778676cf982a32c6c02a2aa57966620f6581c6f016a2837cb60c2e9f5c5361271060405161059a91906107fa565b60405180910390a38073ffffffffffffffffffffffffffffffffffffffff166040516105c5906107ca565b60405180910390207f1658817f2f384793b2fe883b68c3df91c0122e6b4ed86878f4c10e494082d8cc614e206040516105fe9190610828565b60405180910390a35050565b6000600154905090565b60006106276106228461089e565b610879565b90508281526020810184848401111561063f57600080fd5b61064a848285610a5b565b509392505050565b600082601f83011261066357600080fd5b8135610673848260208601610614565b91505092915050565b60008135905061068b81610b86565b92915050565b6000813590506106a081610b9d565b92915050565b6000602082840312156106b857600080fd5b600082013567ffffffffffffffff8111156106d257600080fd5b6106de84828501610652565b91505092915050565b6000602082840312156106f957600080fd5b60006107078482850161067c565b91505092915050565b60006020828403121561072257600080fd5b600061073084828501610691565b91505092915050565b610742816109ee565b82525050565b61075181610a37565b82525050565b61076081610a49565b82525050565b60006107736004836108cf565b915061077e82610b34565b602082019050919050565b60006107966007836108e0565b91506107a182610b5d565b600782019050919050565b6107b581610a20565b82525050565b6107c481610a2a565b82525050565b60006107d582610789565b9150819050919050565b60006020820190506107f46000830184610739565b92915050565b600060408201905061080f6000830184610748565b818103602083015261082081610766565b905092915050565b600060208201905061083d6000830184610757565b92915050565b600060208201905061085860008301846107ac565b92915050565b600060208201905061087360008301846107bb565b92915050565b6000610883610894565b905061088f8282610a6a565b919050565b6000604051905090565b600067ffffffffffffffff8211156108b9576108b8610af4565b5b6108c282610b23565b9050602081019050919050565b600082825260208201905092915050565b600081905092915050565b60006108f682610a00565b915061090183610a00565b92508273ffffffffffffffffffffffffffffffffffffffff0382111561092a57610929610ac5565b5b828201905092915050565b600061094082610a2a565b915061094b83610a2a565b92508260ff0382111561096157610960610ac5565b5b828201905092915050565b600061097782610a00565b915061098283610a00565b92508173ffffffffffffffffffffffffffffffffffffffff04831182151516156109af576109ae610ac5565b5b828202905092915050565b60006109c582610a2a565b91506109d083610a2a565b9250828210156109e3576109e2610ac5565b5b828203905092915050565b60006109f982610a00565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600060ff82169050919050565b6000610a4282610a20565b9050919050565b6000610a5482610a20565b9050919050565b82818337600083830152505050565b610a7382610b23565b810181811067ffffffffffffffff82111715610a9257610a91610af4565b5b80604052505050565b6000610aa682610a2a565b915060ff821415610aba57610ab9610ac5565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b7f6162636400000000000000000000000000000000000000000000000000000000600082015250565b7f7365743536373800000000000000000000000000000000000000000000000000600082015250565b610b8f81610a20565b8114610b9a57600080fd5b50565b610ba681610a2a565b8114610bb157600080fd5b5056fe33633434636464646236613930306661326235383564643239396530336431326661343239336263a26469706673582212200227ddd9f59da8a18fa533bb272780c8c899d4fa64a28dc408b7d65c3430c50364736f6c63430008040033c080a01d6bb58ee614b7906d7903798b4e8dc635627f9e1b4840e781dc814491cea11ea004a4e1dae0b64bfbbd27a9966192c1d0d23eb756692d38e4a36a26a2103b156c"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.ChainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         0,
		To:          ZeroAddress,
		Method:      "",
		GasPrice:    GasPrice,
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}

// StorageWithEventSet returns StorageWithEvent's Set function by alice with nonce = 1
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventSet(tb *TestBlockChain) (*types.Transaction, error) {
	rlp := "0x02f88e8205390180842c9d07ea841dcd6500945fbdb2315678afecb367f032d93f642f64180aa380a460fe47b1000000000000000000000000000000000000000000000000000000003ade68b1c001a06e1ab9f8e0b00a24946a95475d04adae3fd4cd5d92baf14178a5112247b554d1a01d30e5a83a548f34543b8a986404332bf34bb0cdce3d9a2bc0522194406c669f"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.ChainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         1,
		To:          common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Method:      "",
		GasPrice:    GasPrice,
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}
