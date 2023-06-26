package test

import (
	"log"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	"github.com/stretchr/testify/assert"

	. "github.com/meverselabs/meverse/tests/lib"
)

func TestBlockGasUsedAndReceiptsGas(t *testing.T) {

	userKeys, err := GetSingers(ChainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	var mevAddress *common.Address
	var routerAddress, token0, token1 *common.Address
	intialize := func(ctx *types.Context, classMap map[string]uint64) error {
		initSupplyMap := map[common.Address]*amount.Amount{
			alice:   amount.NewAmount(100000000, 0),
			bob:     amount.NewAmount(100000000, 0),
			charlie: amount.NewAmount(100000000, 0),
		}
		mevAddress, err = MevInitialize(ctx, classMap, alice, initSupplyMap)
		if err != nil {
			return err
		}

		_, _, routerAddress, token0, token1, _, err = DexInitialize(ctx, classMap, alice, charlie, initSupplyMap)
		if err != nil {
			return err
		}
		return nil
	}

	// alice(admin), bob, charlie
	//args := []interface{}{alice, bob, charlie}
	tb := NewTestBlockChain(ChainDataPath, true, ChainID, Version, alice, intialize, DefaultInitContextInfo)
	defer tb.Close()

	mev := BindTokenContract(mevAddress, tb.Provider)
	router := BindRouterContract(routerAddress, tb.Provider)
	//token0, token1 := ret[4].(*common.Address), ret[5].(*common.Address)

	provider := tb.Provider
	// mev Transfer : alice -> bob
	tx0 := mev.TransferTx(aliceKey, bob, amount.NewAmount(1, 0))
	// mev Approve : bob -> charlie
	tx1 := mev.ApproveTx(bobKey, charlie, MaxUint256)
	// UniAddLiquidity : alice
	tx2 := router.UniAddLiquidityTx(aliceKey, *token0, *token1, amount.NewAmount(1, 0), amount.NewAmount(4, 0), amount.ZeroCoin, amount.ZeroCoin)

	// mpl mrc20 deploy
	tx3, err := DeployTokenTx(tb, aliceKey, "Meverse Play", "MPL",
		map[common.Address]*amount.Amount{
			alice: amount.NewAmount(100000000, 0),
		})
	if err != nil {
		t.Fatal(err)
	}

	b1 := tb.MustAddBlock([]*TxWithSigner{tx0, tx1, tx2, tx3})

	var mplAddress common.Address
	for _, event := range b1.Body.Events {
		if event.Index == 3 && event.Type == ctypes.EventTagTxMsg {
			mplAddress = common.BytesToAddress(event.Result)
		}
	}
	mpl, err := NewMrc20TokenFromAddress(&mplAddress, provider)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("mpl address = ", mpl.Address)

	jc := NewJsonClient(tb)
	assert := assert.New(t)
	{
		b := b1
		gasUsed, err := strconv.ParseUint(strings.Replace(jc.GetBlockByNumber(b.Header.Height, true)["gasUsed"].(string), "0x", "", -1), 16, 64)
		if err != nil {
			t.Fatal(err)
		}

		cumulativeGasUsed := uint64(0)
		for i := 0; i < len(b.Body.Transactions); i++ {
			json := jc.GetTransactionReceipt(b.Body.Transactions[i].HashSig())
			gasUsed, _ := strconv.ParseUint(strings.Replace(json["gasUsed"].(string), "0x", "", -1), 16, 64)
			cumulativeGasUsed_tx, _ := strconv.ParseUint(strings.Replace(json["cumulativeGasUsed"].(string), "0x", "", -1), 16, 64)
			cumulativeGasUsed += gasUsed
			assert.Equal(cumulativeGasUsed_tx, cumulativeGasUsed, "cumulativeGasUsed")
		}
		assert.Equal(gasUsed, cumulativeGasUsed)
	}

	var pool *RewardPool
	if cont, tx, err := NewRewardPoolTx(aliceKey, 0, provider, mpl.Address); err != nil {
		t.Fatal(err)
	} else {
		pool = cont
		b := tb.MustAddBlock([]*TxWithSigner{tx})
		receipts, _ := provider.Receipts(b.Header.Height)
		pool.SetAddress(&receipts[0].ContractAddress)
		log.Println("pool address = ", pool.Address)
	}

	// tx5 : mpl approve
	tx5, err := mpl.ApproveTx(aliceKey, 0, pool.Address, MaxUint256.Int)
	if err != nil {
		t.Fatal(err)
	}

	// tx6: addReward to bob
	total := big.NewInt(123456)
	userReward := UserReward{
		User:   bob,
		Amount: total,
	}
	userRewards := []UserReward{userReward}
	tx6, err := pool.AddRewardTx(aliceKey, 1, total, userRewards)

	if err != nil {
		t.Fatal(err)
	}

	// tx7: clamim by bob
	tx7, err := pool.ClaimTx(bobKey, 0)
	if err != nil {
		t.Fatal(err)
	}

	b2 := tb.MustAddBlock([]*TxWithSigner{tx5, tx6, tx7})

	{
		b := b2

		gasUsed, err := strconv.ParseUint(strings.Replace(jc.GetBlockByNumber(b.Header.Height, true)["gasUsed"].(string), "0x", "", -1), 16, 64)
		if err != nil {
			t.Fatal(err)
		}

		cumulativeGasUsed := uint64(0)
		for i := 0; i < len(b.Body.Transactions); i++ {
			json := jc.GetTransactionReceipt(b.Body.Transactions[i].HashSig())
			gasUsed, _ := strconv.ParseUint(strings.Replace(json["gasUsed"].(string), "0x", "", -1), 16, 64)
			cumulativeGasUsed_tx, _ := strconv.ParseUint(strings.Replace(json["cumulativeGasUsed"].(string), "0x", "", -1), 16, 64)
			cumulativeGasUsed += gasUsed
			assert.Equal(cumulativeGasUsed_tx, cumulativeGasUsed, "cumulativeGasUsed")
		}
		assert.Equal(gasUsed, cumulativeGasUsed)
	}
}
