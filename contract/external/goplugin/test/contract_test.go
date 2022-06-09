package test

import (
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/external/goplugin"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/test/util"
)

func TestMain(t *testing.T) {
	util.RegisterContractClass(&goplugin.PluginContract{}, "PluginContract")

	tc := util.NewTestContext()

	dat, err := os.ReadFile("./token.so")
	if err != nil {
		log.Fatal(err)
	}

	tokenContArgs := &goplugin.PluginContractConstruction{
		Bin: dat,
	}
	tokenContType := &goplugin.PluginContract{}

	piAddr := tc.DeployContract(tokenContType, tokenContArgs)
	log.Println("jscont Addr", piAddr)

	inf, err := tc.MakeTx(util.AdminKey, piAddr, "ContractInvoke", "Mint", []interface{}{util.Admin.String(), amount.NewAmount(30, 0).Int})
	log.Println(inf, err)
	inf, err = tc.MakeTx(util.AdminKey, piAddr, "ContractInvoke", "Transfer", []interface{}{util.Users[0].String(), amount.NewAmount(10, 0).Int})
	log.Println(inf, err)
	inf, err = tc.MakeTx(util.AdminKey, piAddr, "ContractInvoke", "BalanceOf", []interface{}{util.Admin.String()})
	log.Println(inf, err)

	ChainID := big.NewInt(1)

	txs := []*types.Transaction{}
	ks := []key.Key{}
	for i := 0; i < 7000; i++ {
		tx := &types.Transaction{
			ChainID:   ChainID,
			Timestamp: tc.Ctx.LastTimestamp(),
			To:        piAddr,
			Method:    "ContractInvoke",
		}

		tx.Args = bin.TypeWriteAll("BalanceOf", []interface{}{util.Users[0].String()})
		txs = append(txs, tx)
		ks = append(ks, util.AdminKey)
		// inf, err = tc.MakeTx(util.AdminKey, piAddr, "ContractInvoke", "BalanceOf", []interface{}{util.Users[0].String()})
	}
	start := time.Now()
	err = tc.Sleep(10, txs, ks)
	log.Println("inf", err)

	du := time.Since(start)
	log.Println(inf, err, du)
}
