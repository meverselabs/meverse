package test

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	. "github.com/meverselabs/meverse/tests/lib"
)

// getLogs test
func TestGetLogs(t *testing.T) {

	userKeys, err := GetSingers(ChainID)
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
	tb, ret, err := Prepare(ChainDataPath, true, ChainID, Version, alice, args, intialize, &InitContextInfo{})
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Close()

	mev := NewTokenContract(ret[0].(*common.Address))
	router := NewRouterContract(ret[2].(*common.Address))
	token0, token1 := ret[4].(*common.Address), ret[5].(*common.Address)

	provider := tb.Provider

	var mpl *Mrc20Token
	{
		// mev Transfer : alice -> bob
		tx0 := mev.TransferTx(aliceKey, provider, bob, amount.NewAmount(1, 0))
		// mev Approve : bob -> charlie
		tx1 := mev.ApproveTx(bobKey, provider, charlie, MaxUint256)
		// UniAddLiquidity : alice
		tx2 := router.UniAddLiquidityTx(aliceKey, provider, *token0, *token1, amount.NewAmount(1, 0), amount.NewAmount(4, 0), amount.ZeroCoin, amount.ZeroCoin)

		// mpl mrc20 deploy
		tx3, err := NewTokenTx(tb, aliceKey, "Meverse Play", "MPL",
			map[common.Address]*amount.Amount{
				alice: amount.NewAmount(100000000, 0),
			})
		if err != nil {
			t.Fatal(err)
		}

		b, err := tb.AddBlock([]*TxWithSigner{tx0, tx1, tx2, tx3}) //block 1
		if err != nil {
			t.Fatal(err)
		}

		var mplAddress common.Address
		for _, event := range b.Body.Events {
			if event.Index == 3 && event.Type == ctypes.EventTagTxMsg {
				mplAddress = common.BytesToAddress(event.Result)
			}
		}

		mpl, err = NewMrc20TokenFromAddress(&mplAddress, provider)
		if err != nil {
			t.Fatal(err)
		}
		log.Println("mpl address = ", mpl.Address)
	}

	var pool *RewardPool
	if cont, tx, err := NewRewardPoolTx(aliceKey, 0, provider, mpl.Address); err != nil {
		t.Fatal(err)
	} else {
		pool = cont
		b, err := tb.AddBlock([]*TxWithSigner{tx}) // block 2
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
		_, err = tb.AddBlock([]*TxWithSigner{tx}) // block 3
		if err != nil {
			t.Fatal(err)
		}
	}

	// 100 blocks to activate bloomservice
	total := big.NewInt(123456)
	userReward := UserReward{
		User:   bob,
		Amount: total,
	}
	userRewards := []UserReward{userReward}
	for i := 0; i < 100; i++ { // block 4 -> 103
		if i%2 == 0 {
			_, err = tb.AddBlock([]*TxWithSigner{})
			if err != nil {
				t.Fatal(err)
			}
		} else {
			if tx, err := pool.AddRewardTx(aliceKey, 0, total, userRewards); err != nil {
				t.Fatal(err)
			} else {
				_, err = tb.AddBlock([]*TxWithSigner{tx})
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	// clamim by bob
	if tx, err := pool.ClaimTx(bobKey, 0); err != nil {
		t.Fatal(err)
	} else {
		_, err = tb.AddBlock([]*TxWithSigner{tx}) // block 104 = 0x68
		if err != nil {
			t.Fatal(err)
		}
	}

	jc := NewJsonClient(tb)

	transferHash := hash.Hash([]byte("Transfer(address,address,uint256)")) //0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
	log.Println("tranferhash", transferHash.String())
	rewardAddedHash := hash.Hash([]byte("RewardAdded(uint256)"))
	log.Println("rewardAddedHash", rewardAddedHash.String())

	b1Hash, _ := provider.Hash(1)

	testCases := []struct {
		filterMap map[string]interface{}
		count     int
	}{
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(0)),
			"address":   mpl.Address.String(),
		}, 52},
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(0)),
			"topics":    []interface{}{[]interface{}{transferHash.String()}},
		}, 54},
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(0)),
			"topics":    []interface{}{[]interface{}{rewardAddedHash.String()}},
		}, 50},
		{map[string]interface{}{
			"blockHash": b1Hash.String(),
		}, 9},

		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(1)),
			"toBlock":   fmt.Sprintf("0x%x", big.NewInt(1)),
		}, 9},
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("%d", big.NewInt(1)),
			"toBlock":   fmt.Sprintf("%d", big.NewInt(1)),
		}, 9},
		{map[string]interface{}{
			"fromBlock": 1,
			"toBlock":   1,
		}, 9},
		//topics OR
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(0)),
			"topics":    []interface{}{[]interface{}{transferHash.String(), rewardAddedHash.String()}},
		}, 104},
		// topics AND
		{map[string]interface{}{
			"fromBlock": fmt.Sprintf("0x%x", big.NewInt(0)),
			"topics":    []interface{}{[]interface{}{transferHash.String()}, []interface{}{rewardAddedHash.String()}},
		}, 0},
	}

	for i, test := range testCases {
		logs := jc.GetLogs(test.filterMap)

		// for i, rlog := range logs {
		// 	b, err := json.MarshalIndent(rlog, "", "  ")
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	log.Printf("logs[%v] = %v", i, string(b))
		// }

		if len(logs) != test.count {
			t.Errorf("filter query for case %d : got %d, have %d", i, len(logs), test.count)
		}
	}
}
