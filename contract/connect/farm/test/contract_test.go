package test

import (
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/connect/farm"
	"github.com/meverselabs/meverse/contract/connect/pool"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {

}

func TestExecuteContractTx(t *testing.T) {
	tc := util.NewTestContext()

	tokenContArgs := &token.TokenContractConstruction{
		Name:   "FarmToken",
		Symbol: "FARMTO",
		InitialSupplyMap: map[common.Address]*amount.Amount{
			util.AdminKey.PublicKey().Address(): amount.MustParseAmount("10000"),
		},
	}
	tokenContType := &token.TokenContract{}

	tokenAddr := tc.DeployContract(tokenContType, tokenContArgs)
	log.Println("farm token Addr", tokenAddr)

	farmContArgs := &farm.FarmContractConstruction{
		Owner:          util.Admin,
		FarmToken:      tokenAddr,
		OwnerReward:    100, //10%
		TokenMaxSupply: amount.MustParseAmount("2000000000"),
		TokenPerBlock:  amount.MustParseAmount("1"),
		StartBlock:     1,
	}
	farmType := &farm.FarmContract{}

	farmAddr := tc.DeployContract(farmType, farmContArgs)
	log.Println("farm Addr", farmAddr)

	poolContArgs := &pool.PoolContractConstruction{
		Gov:               util.Admin,
		Farm:              farmAddr,
		Want:              tokenAddr,
		FeeFundAddress:    common.HexToAddress("0x0000000000000000000000000000000000000fee"),
		RewardsAddress:    common.HexToAddress("0x000000000000000000000000000000000000ffee"),
		DepositFeeFactor:  uint16(10000),
		WithdrawFeeFactor: uint16(10000),
		RewardFeeFactor:   uint16(4000),
	}
	poolContType := &pool.PoolContract{}

	poolAddr := tc.DeployContract(poolContType, poolContArgs)
	log.Println("pool Addr", poolAddr)

	// 	- 팜토큰 민터로 팜 지정
	// .\sendtx.exe 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d SetMinter addr 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b bool true
	inf := tc.MustSendTx(util.AdminKey, tokenAddr, "SetMinter", farmAddr, true)
	log.Println(inf)
	// - 팜에 풀 등록
	// ; Add(cc, _allocPoint, _want, _withUpdate, _strat
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Add uint32 100 addr 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d bool false addr 0x3552a64ab4240c1e6e879273b2f4a2a3bd65c821
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Add", uint32(10), tokenAddr, false, poolAddr)

	// - user에게 mint
	// .\sendtx.exe 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d Approve addr 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b amt 1000000000
	_ = tc.MustSendTx(util.AdminKey, tokenAddr, "Mint", util.Users[0], amount.MustParseAmount("20000"))

	// - 팜에 approve 팜토큰
	// .\sendtx.exe 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d Approve addr 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b amt 1000000000
	_ = tc.MustSendTx(util.AdminKey, tokenAddr, "Approve", farmAddr, amount.MustParseAmount("10000000"))
	// - 팜에 approve 팜토큰
	// .\sendtx.exe 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d Approve addr 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b amt 1000000000
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.MustParseAmount("1000"))
	_ = tc.MustSendTx(util.UserKeys[0], tokenAddr, "Approve", farmAddr, amount.MustParseAmount("10000000000"))

	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "1")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "1")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "1")
	// - 팜에 팜토큰 deposit
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Deposit uint64 0 amt 100
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("7363.083090719219"))

	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "2")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "2")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "2")
	// - 팜에 팜토큰 Withdraw
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Withdraw uint64 0 amt 0
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("7363.083090719219"))

	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "3")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "3")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "3")
	// - 팜에 팜토큰 deposit
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Deposit uint64 0 amt 100
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("5522.312318039414"))

	// - 팜에 팜토큰 deposit
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Deposit uint64 0 amt 100
	_ = tc.MustSendTx(util.UserKeys[0], farmAddr, "Deposit", uint64(0), amount.MustParseAmount("15641"))

	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "4")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "4")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "4")

	_ = tc.MustSendTx(util.AdminKey, farmAddr, "SetHoldShares", tc.Ctx.TargetHeight())

	// - 팜에 팜토큰 Withdraw
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Withdraw uint64 0 amt 0
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("0"))
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "5")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "5")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "5")
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("0"))
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "6")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "6")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "6")
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("0"))
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "7")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "1")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "1")
	// 5.3521
	_ = tc.MustSendTx(util.UserKeys[0], farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("0"))
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "8")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "8")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "8")

	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("10000000000"))
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "9")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "9")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "9")

	inf, err := tc.MakeTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("10000000000"))
	log.Println(inf, err)
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "10")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "10")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "10")

	inf, err = tc.MakeTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("5423.34248643"))
	log.Println(inf, err)
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "11")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "11")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "11")

	inf, err = tc.MakeTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("10000000000"))
	log.Println(inf, err)
	printTotalSupply(tc, tokenAddr)
	printPoolInfo(tc, util.Admin, farmAddr, 0, "11")
	printUserInfo(tc, farmAddr, tokenAddr, util.Admin, "11")
	printUserInfo(tc, farmAddr, tokenAddr, util.Users[0], "11")

	// - 팜에 팜토큰 Withdraw
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Withdraw uint64 0 amt 0
	// _ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("99"))
	// printPoolInfo(tc, farmAddr, 0, "5")
}

func printPoolInfo(tc *util.TestContext, user, farmAddr common.Address, pid uint64, tag string) {
	inf, err := tc.SendTx(util.AdminKey, farmAddr, "PoolInfo", uint64(0))
	log.Println(tag, "inf", inf, err)
	strat := inf[4].(common.Address)
	inf1, _ := tc.SendTx(util.AdminKey, strat, "WantLockedTotal")
	inf2, _ := tc.SendTx(util.AdminKey, strat, "SharesTotal")
	log.Println(tag, "WantLockedTotal, SharesTotal", inf1, inf2, err)
}
func printUserInfo(tc *util.TestContext, farmAddr, tokenAddr, user common.Address, tag string) {
	inf3, _ := tc.SendTx(util.AdminKey, farmAddr, "UserInfo", uint64(0), user)
	amt1 := inf3[0].(*amount.Amount)
	amt2 := inf3[1].(*amount.Amount)
	inf4, _ := tc.SendTx(util.AdminKey, tokenAddr, "BalanceOf", user)
	amt3 := inf4[0].(*amount.Amount)
	log.Println(tag, "user Shares", amt1.String(), "rewardDabt", amt2.String(), " farmtoken", amt3.String())
}

func printTotalSupply(tc *util.TestContext, addr common.Address) {
	inf, err := tc.SendTx(util.AdminKey, addr, "TotalSupply")
	amt := inf[0].(*amount.Amount)
	log.Println("total supply", tc.Ctx.TargetHeight(), amt.String(), err)
}
