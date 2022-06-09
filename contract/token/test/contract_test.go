package test

import (
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {
}

func getContClassID(rt reflect.Type) uint64 {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		pkgPath = strings.Replace(pkgPath, "meverselabs/meverse", "fletaio/fleta_v2", -1)
		name = pkgPath + "." + name
	}
	h := hash.Hash([]byte(name))
	return bin.Uint64(h[len(h)-8:])
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

	inf, err := tc.MakeTx(util.UserKeys[0], tokenAddr, "SwapToMainToken", amount.NewAmount(10, 0))
	log.Println("SwapToMainToken", inf, ":", err)
}
