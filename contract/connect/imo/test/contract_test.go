package test

import (
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/connect/imo"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {

}

func TestExecuteContractTx(t *testing.T) {
	tc := util.NewTestContext()

	// 프로잭트 토큰 생성
	tokenContArgs := &token.TokenContractConstruction{
		Name:   "ProjevtToken",
		Symbol: "PROJTKN",
		InitialSupplyMap: map[common.Address]*amount.Amount{
			util.AdminKey.PublicKey().Address(): amount.MustParseAmount("10000"),
		},
	}
	tokenContType := &token.TokenContract{}
	tokenAddr := tc.DeployContract(tokenContType, tokenContArgs)
	log.Println("Projevt Token Addr", tokenAddr)

	//
	imoContArgs := &imo.ImoContractConstruction{
		ProjectOwner:     util.Admin,
		PayToken:         tc.MainToken,
		ProjectToken:     tokenAddr,
		ProjectOffering:  amount.MustParseAmount("1000"),
		ProjectRaising:   amount.MustParseAmount("1000"),
		PayLimit:         amount.MustParseAmount("1000"),
		StartBlock:       10,
		EndBlock:         100,
		HarvestFeeFactor: 10000, //max 10000,
		WhiteListAddress: common.ZeroAddr,
		WhiteListGroupId: hash.Hash256{},
	}
	imoType := &imo.ImoContract{}

	imoAddr := tc.DeployContract(imoType, imoContArgs)
	log.Println("imo Addr", imoAddr)

	// - 프로잭트 토큰 민트
	_ = tc.MustSendTx(util.AdminKey, tokenAddr, "Mint", imoAddr, amount.NewAmount(1000, 0))

	// - imo에 approve
	_ = tc.MustSendTx(util.AdminKey, tc.MainToken, "Approve", imoAddr, amount.MustParseAmount("1000"))

	var err error
	// - imo에 deposit
	_, err = tc.MakeTx(util.AdminKey, imoAddr, "Deposit", amount.MustParseAmount("1000"))
	if err == nil {
		t.Errorf("ExecuteContractTx wantErr %v but not occur the error", true)
	}

	log.Println("current height:", tc.Cn.Provider().Height())
	tc.MustSkipBlock(5)
	// - imo에 deposit
	_ = tc.MustSendTx(util.AdminKey, imoAddr, "Deposit", amount.MustParseAmount("1000"))
	log.Println("current height:", tc.Cn.Provider().Height())

	// - imo에 harvest
	_, err = tc.MakeTx(util.AdminKey, imoAddr, "Harvest")
	if err == nil {
		t.Errorf("ExecuteContractTx wantErr %v but not occur the error", true)
	}
	tc.MustSkipBlock(88)
	log.Println(tc.Ctx.TargetHeight())

	_ = tc.MustSendTx(util.AdminKey, imoAddr, "Harvest")
	log.Println(tc.Ctx.TargetHeight())
}
