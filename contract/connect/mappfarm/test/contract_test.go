package test

import (
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/connect/mappfarm"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {

}

func TestExecuteContractTx(t *testing.T) {
	tc := util.NewTestContext()

	tokenAddr := deployToken(tc, "FarmToken", "FARMTO")
	log.Println("farm token Addr1", tokenAddr)
	wantAddr := deployToken(tc, "WantToken", "WantTO")
	log.Println("want token Addr2", wantAddr)

	farmContArgs := &mappfarm.FarmContractConstruction{
		Owner:         util.Admin,
		Banker:        util.Users[0],
		FarmToken:     tokenAddr,
		WantToken:     wantAddr,
		OwnerReward:   0, //10%
		TokenPerBlock: amount.MustParseAmount("01"),
		StartBlock:    1,
	}
	farmType := &mappfarm.FarmContract{}

	farmAddr := tc.DeployContract(farmType, farmContArgs)
	log.Println("farm Addr", farmAddr)
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.MustParseAmount("1000"))
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.MustParseAmount("1000"))
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[2], amount.MustParseAmount("1000"))
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[3], amount.MustParseAmount("1000"))
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[4], amount.MustParseAmount("1000"))

	// - user에게 mint
	// .\sendtx.exe 0x698da8d6b3382cabcfcde3fc514bd21c1f5d866d Approve addr 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b amt 1000000000
	_ = tc.MustSendTx(util.AdminKey, tokenAddr, "Mint", util.Users[0], amount.MustParseAmount("2000"))
	_ = tc.MustSendTx(util.UserKeys[0], tokenAddr, "Approve", farmAddr, amount.MustParseAmount("100000000"))

	_ = tc.MustSendTx(util.AdminKey, tokenAddr, "Approve", farmAddr, amount.MustParseAmount("100000000"))

	// - 팜에 팜토큰 deposit
	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Deposit uint64 0 amt 100
	_ = tc.MustSendTx(util.AdminKey, wantAddr, "Approve", farmAddr, amount.MustParseAmount("1000000000000"))
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("7363.083090719219"))

	// .\sendtx.exe 0xc30f8c78d4ac1a1f2401c51eacfb987c2c3d216b Withdraw uint64 0 amt 0
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("7363.083090719219"))
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("5522.312318039414"))
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("0"))
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Withdraw", uint64(0), amount.MustParseAmount("10000000"))
	_ = tc.MustSendTx(util.AdminKey, farmAddr, "Deposit", uint64(0), amount.MustParseAmount("100"))
	log.Println(getBalanceOf(tc, tokenAddr, util.Admin), getBalanceOf(tc, wantAddr, util.Admin))
	log.Println(getBalanceOf(tc, tokenAddr, farmAddr), getBalanceOf(tc, wantAddr, farmAddr))
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[0]), getBalanceOf(tc, wantAddr, util.Users[0]))

	_ = tc.MustSendTx(util.AdminKey, wantAddr, "Mint", util.Users[1], amount.MustParseAmount("100"))
	_ = tc.MustSendTx(util.AdminKey, wantAddr, "Mint", util.Users[2], amount.MustParseAmount("200"))
	_ = tc.MustSendTx(util.AdminKey, wantAddr, "Mint", util.Users[3], amount.MustParseAmount("300"))
	_ = tc.MustSendTx(util.AdminKey, wantAddr, "Mint", util.Users[4], amount.MustParseAmount("400"))
	_ = tc.MustSendTx(util.UserKeys[1], wantAddr, "Approve", farmAddr, amount.MustParseAmount("1000000000000"))
	_ = tc.MustSendTx(util.UserKeys[2], wantAddr, "Approve", farmAddr, amount.MustParseAmount("1000000000000"))
	_ = tc.MustSendTx(util.UserKeys[3], wantAddr, "Approve", farmAddr, amount.MustParseAmount("1000000000000"))
	_ = tc.MustSendTx(util.UserKeys[4], wantAddr, "Approve", farmAddr, amount.MustParseAmount("1000000000000"))

	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[2], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[3], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[4], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[2], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[3], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[4], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[2], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[3], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[4], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
	})
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[2], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[3], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[4], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[3], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
		{util.UserKeys[4], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
	})
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[2], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[3], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[4], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[4], farmAddr, "Deposit", []interface{}{uint64(0), amount.MustParseAmount("100")}},
	})
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[2], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[3], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
		{util.UserKeys[4], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("0")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())
	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[1], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("1000000")}},
		{util.UserKeys[2], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("1000000")}},
		{util.UserKeys[3], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("1000000")}},
		{util.UserKeys[4], farmAddr, "Withdraw", []interface{}{uint64(0), amount.MustParseAmount("1000000")}},
	})
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[1]).String(), getBalanceOf(tc, wantAddr, util.Users[1]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[2]).String(), getBalanceOf(tc, wantAddr, util.Users[2]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[3]).String(), getBalanceOf(tc, wantAddr, util.Users[3]).String())
	log.Println(getBalanceOf(tc, tokenAddr, util.Users[4]).String(), getBalanceOf(tc, wantAddr, util.Users[4]).String())

	tc.MustSendTxs([]*util.TxCase{
		{util.UserKeys[4], farmAddr, "Deposit", []interface{}{uint64(1), amount.MustParseAmount("100")}},
	})

}

func getBalanceOf(tc *util.TestContext, tokenAddr, to common.Address) *amount.Amount {
	inf, err := tc.ReadTx(util.AdminKey, tokenAddr, "BalanceOf", to)
	if err != nil {
		panic(err)
	}
	return inf[0].(*amount.Amount)
}

func deployToken(tc *util.TestContext, name, symbol string) common.Address {
	tokenContArgs := &token.TokenContractConstruction{
		Name:   name,
		Symbol: symbol,
		InitialSupplyMap: map[common.Address]*amount.Amount{
			util.AdminKey.PublicKey().Address(): amount.MustParseAmount("10000"),
		},
	}
	tokenContType := &token.TokenContract{}

	tokenAddr := tc.DeployContract(tokenContType, tokenContArgs)
	return tokenAddr
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
