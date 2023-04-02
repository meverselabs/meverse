package test

import (
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {
}

func TestSendTx(t *testing.T) {
	tc := util.NewTestContext()

	is, err := tc.ReadTx(util.AdminKey, tc.MainToken, "BalanceOf", util.Admin)
	am1 := is[0].(*amount.Amount)
	log.Println(am1.String(), err)

	tokenAddr := tc.MakeToken("TestToken", "TEST", "10000")
	log.Println("Test Token Addr", tokenAddr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf
	is, err = tc.SendTx(util.AdminKey, tokenAddr, "Transfer", util.Users[0], amount.MustParseAmount("1"))
	log.Println(is, err)

	is, err = tc.ReadTx(util.AdminKey, tc.MainToken, "BalanceOf", util.Admin)
	am2 := is[0].(*amount.Amount)
	log.Println(am1.Sub(am2).String(), err)

	log.Println("Test Token Addr", tokenAddr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf
	is, err = tc.SendTx(util.AdminKey, tokenAddr, "Transfer", util.Users[0], amount.MustParseAmount("1"))
	log.Println(is, err)

	is, err = tc.ReadTx(util.AdminKey, tc.MainToken, "BalanceOf", util.Admin)
	am3 := is[0].(*amount.Amount)
	log.Println(am2.Sub(am3).String(), err)

	log.Println("Test Token Addr", tokenAddr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf
	is, err = tc.SendTx(util.AdminKey, tokenAddr, "Transfer", util.Users[0], amount.MustParseAmount("1"))
	log.Println(is, err)

	is, err = tc.ReadTx(util.AdminKey, tc.MainToken, "BalanceOf", util.Admin)
	am4 := is[0].(*amount.Amount)
	log.Println(am3.Sub(am4).String(), err)
}

func TestSwapMaintokenTx(t *testing.T) {
	tc := util.NewTestContext()

	tokenAddr := tc.MakeToken("TestToken", "TEST", "10000")
	log.Println("Test Token Addr", tokenAddr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf

	FactoryClassID := util.RegisterContractClass(&factory.FactoryContract{}, "")
	UniClassID := util.RegisterContractClass(&trade.UniSwap{}, "UniSwap")
	StableClassID := util.RegisterContractClass(&trade.StableSwap{}, "StableSwap")
	log.Println(FactoryClassID, UniClassID, StableClassID)

	factoryAddr := tc.DeployContract(&factory.FactoryContract{}, &factory.FactoryContractConstruction{
		Owner: util.Admin,
	})
	log.Println("factoryAddr", factoryAddr) // 0x4b33385C9138d0Ec546F60043F1A9b99F8B6019d

	routerAddr := tc.DeployContract(&router.RouterContract{}, &router.RouterContractConstruction{
		Factory: factoryAddr,
	})
	log.Println("routerAddr", routerAddr) // 0x29c5b439356A8E2a89E25DdF1B7271D21Ef423bc

	whiteListAddr := tc.DeployContract(&whitelist.WhiteListContract{}, &whitelist.WhiteListContractConstruction{})
	log.Println("whiteListAddr", whiteListAddr) // 0x5575351cB7A4Add01e1E844EC67081Aa2b8c936D

	FEE := uint64(30000000)
	ADMINFEE := uint64(30000000)
	WINNERFEE := uint64(30000000)

	// tokenA, tokenB, payToken common.Address, name, symbol string, owner, winner common.Address, fee, adminFee, winnerFee uint64, whiteList common.Address, groupId hash.Hash256, classID uint64
	inf := tc.MustSendTx(util.AdminKey, factoryAddr, "CreatePairUni",
		tokenAddr, tc.MainToken, tc.MainToken, "testlp", "TESTLP", util.Admin,
		util.Admin, FEE, ADMINFEE, WINNERFEE, whiteListAddr, hash.Hash256{}, UniClassID)
	log.Println("TestExecuteContractTx", inf)

	inf = tc.MustSendTx(util.AdminKey, tc.MainToken, "Approve", routerAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve MainToken", inf)
	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "Approve", routerAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve tokenAddr", inf)

	tc.Sleep(5, nil, nil)

	inf = tc.MustSendTx(util.AdminKey, routerAddr, "UniAddLiquidity", tc.MainToken, tokenAddr, amount.NewAmount(1000, 0), amount.NewAmount(1000, 0), amount.NewAmount(0, 0), amount.NewAmount(0, 0))
	log.Println("UniAddLiquidity", inf)

	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "Transfer", util.Users[0], amount.NewAmount(100, 0))
	log.Println("Transfer", inf)

	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "SetRouter", routerAddr, []common.Address{tokenAddr, tc.MainToken})
	log.Println("SetRouter", inf)

	inf, err := tc.SendTx(util.UserKeys[0], tokenAddr, "SwapToMainToken", amount.NewAmount(10, 0))
	log.Println("SwapToMainToken", inf, ":", err)
}

func TestApprove(t *testing.T) {
	tc := util.NewTestContext()

	mt := tc.MainToken

	tc.MustSendTx(util.AdminKey, mt, "SetVersion", "1")

	tc.MustSendTx(util.AdminKey, mt, "Transfer", util.Users[0], amount.MustParseAmount("10000"))
	tc.MustSendTx(util.AdminKey, mt, "Transfer", util.Users[1], amount.MustParseAmount("2"))
	tc.MustSendTx(util.UserKeys[0], mt, "Approve", util.Users[1], amount.MustParseAmount("5000"))

	inf, err := tc.ReadTx(util.UserKeys[0], mt, "Allowance", util.Users[0], util.Users[1])
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}
	log.Println(inf)

	tc.MustSendTx(util.UserKeys[1], mt, "TransferFrom", util.Users[0], util.Users[2], amount.MustParseAmount("1000"))

	inf, err = tc.ReadTx(util.UserKeys[0], mt, "Allowance", util.Users[0], util.Users[1])
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}
	log.Println(inf)

}
