package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/service/bloomservice"
	. "github.com/meverselabs/meverse/tests/lib"
)

// RewardPool  evm contract에서 go contract 호출 : 두가지의 log 찾아 보고 검색해 본다.
func TestFilterRewardPool(t *testing.T) {

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

	//ctx := tb.NewContext()
	provider := tb.Provider

	// mpl mrc20 token
	mpl, err := NewMrc20Token(tb, aliceKey, "Meverse Play", "MPL")
	if err != nil {
		t.Fatal(err)
	}

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

	// addReward 변수 userRewards
	total := big.NewInt(123456)
	userReward := UserReward{
		User:   bob,
		Amount: total,
	}
	userRewards := []UserReward{userReward}

	rewardAddedHash := hash.Hash([]byte("RewardAdded(uint256)"))
	transferHash := hash.Hash([]byte("Transfer(address,address,uint256)"))

	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			_, err = tb.AddBlock([]*TxWithSigner{})
			if err != nil {
				t.Fatal(err)
			}
		} else {
			// tx: addReward to bob
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

	testCases := []struct {
		crit  bloomservice.FilterQuery
		count int
	}{
		{bloomservice.FilterQuery{FromBlock: big.NewInt(0), Addresses: []common.Address{*mpl.Address}}, 51},
		{bloomservice.FilterQuery{FromBlock: big.NewInt(0), Topics: [][]common.Hash{{transferHash}}}, 50},
		{bloomservice.FilterQuery{FromBlock: big.NewInt(0), Topics: [][]common.Hash{{rewardAddedHash}}}, 50},
		{bloomservice.FilterQuery{FromBlock: big.NewInt(0), Addresses: []common.Address{*pool.Address}}, 51},
	}

	for i, test := range testCases {
		logs, err := bloomservice.FilterLogs(tb.Chain, tb.Ts, tb.Bs, test.crit)
		if err != nil {
			t.Errorf("filter query for case %d : err %v", i, err)
		}
		if len(logs) != test.count {
			t.Errorf("filter query for case %d : got %d, have %d", i, len(logs), test.count)
		}
	}
}
