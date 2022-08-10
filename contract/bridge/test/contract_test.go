package test

import (
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/bridge"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

var (
	bridgeAddr   common.Address
	banker       common.Address
	bankerKey    key.Key
	feeOwner     common.Address
	testUsers    []common.Address
	testUserKeys []key.Key
	testToken    common.Address
)

func init() {

}

func setupTest() *util.TestContext {
	tc := util.NewTestContext()

	banker = util.Users[0]
	bankerKey = util.UserKeys[0]
	feeOwner = util.Users[1]
	bridgeContArgs := &bridge.BridgeContractConstruction{
		Bank:         banker,
		FeeOwner:     feeOwner,
		MeverseToken: tc.MainToken,
	}
	contType := &bridge.BridgeContract{}

	bridgeAddr = tc.DeployContract(contType, bridgeContArgs)
	log.Println("bridge Addr", bridgeAddr)

	tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(1000000000, 0))

	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTransferFeeInfo", "ETHEREUM", amount.NewAmount(200, 0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTransferFeeInfo", "KLAYTN", amount.NewAmount(20, 0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTransferFeeInfo", "BSC", amount.NewAmount(20, 0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTokenFeeInfo", "ETHEREUM", uint16(0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTokenFeeInfo", "KLAYTN", uint16(0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTokenFeeInfo", "BSC", uint16(0))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTokenFeeInfo", "POLYGON", uint16(0))

	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "TransferTokenFeeOwnership", common.HexToAddress("0x0fee"))
	tc.MustSendTx(util.UserKeys[0], bridgeAddr, "SetTransferTokenFeeInfo", "ETHEREUM", uint16(10))

	testUsers = []common.Address{
		util.Users[2],
		util.Users[3],
		util.Users[4],
	}
	testUserKeys = []key.Key{
		util.UserKeys[2],
		util.UserKeys[3],
		util.UserKeys[4],
	}
	tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", testUsers[0], amount.NewAmount(10000, 0))
	tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", testUsers[1], amount.NewAmount(10000, 0))

	testToken = tc.MakeToken("Test Token", "TTEST", "10000")

	tc.MustSendTx(util.AdminKey, testToken, "Transfer", bridgeAddr, amount.NewAmount(10000, 0))
	tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", bridgeAddr, amount.NewAmount(10000, 0))

	return tc
}

func TestSendToGatewayTx(t *testing.T) {
	tc := setupTest()
	// token, amt, path []common.Address, toChain string, summary []byte
	_, err := tc.MakeTx(util.UserKeys[3], bridgeAddr, "SendToGateway", tc.MainToken, amount.NewAmount(10, 0), []common.Address{}, "POLYGON", []byte("MEVERSE/POLYGON/MEV"))
	if err == nil {
		t.Errorf("TestSendToGatewayTx error not approved but tx success")
	}

	_, err = tc.MakeTx(util.UserKeys[3], tc.MainToken, "Approve", bridgeAddr, amount.NewAmount(10000000000, 0))
	if err != nil {
		t.Errorf("TestSendToGatewayTx error = %+v", err)
	}

	mv, err := tc.Ctx.Contract(tc.MainToken)
	if err != nil {
		t.Errorf("contract load error")
	}
	tokenCont := mv.(*token.TokenContract)
	tcc := tc.Ctx.ContractContext(tokenCont, util.Admin)
	log.Println(tokenCont.TotalSupply(tcc))

	log.Println(tokenCont.BalanceOf(tcc, util.Users[3]).String())
	_, err = tc.MakeTx(util.UserKeys[3], bridgeAddr, "SendToGateway", tc.MainToken, amount.NewAmount(10, 0), []common.Address{}, "POLYGON", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf("TestSendToGatewayTx error = %+v", err)
	}
	log.Println(tokenCont.BalanceOf(tcc, util.Users[3]).String())

	bv, err := tc.Ctx.Contract(bridgeAddr)
	if err != nil {
		t.Errorf("contract load error")
	}

	cont := bv.(*bridge.BridgeContract)
	bcc := tc.Ctx.ContractContext(cont, util.Admin)
	fcont := cont.Front().(bridge.BridgeFront)
	log.Println(fcont.GetSequenceFrom(bcc, banker, "POLYGON"))
}

func TestSendToGatewayEthTx(t *testing.T) {
	tc := setupTest()
	// token, amt, path []common.Address, toChain string, summary []byte
	_, err := tc.MakeTx(util.UserKeys[3], bridgeAddr, "SendToGateway", tc.MainToken, amount.NewAmount(10, 0), []common.Address{}, "ETHEREUM", []byte("MEVERSE/ETHEREUM/MEV"))
	if err == nil {
		t.Errorf("TestSendToGatewayTx error not approved but tx success")
	}

	_, err = tc.MakeTx(util.UserKeys[3], tc.MainToken, "Approve", bridgeAddr, amount.NewAmount(10000000000, 0))
	if err != nil {
		t.Errorf("TestSendToGatewayTx error = %+v", err)
	}

	mv, err := tc.Ctx.Contract(tc.MainToken)
	if err != nil {
		t.Errorf("contract load error")
	}
	tokenCont := mv.(*token.TokenContract)
	tcc := tc.Ctx.ContractContext(tokenCont, util.Admin)
	log.Println(tokenCont.TotalSupply(tcc))

	log.Println(tokenCont.BalanceOf(tcc, util.Users[3]).String())
	_, err = tc.MakeTx(util.UserKeys[3], bridgeAddr, "SendToGateway", tc.MainToken, amount.NewAmount(100, 0), []common.Address{}, "ETHEREUM", []byte("MEVERSE/ETHEREUM/MEV"))
	if err != nil {
		t.Errorf("TestSendToGatewayTx error = %+v", err)
	}
	log.Println(tokenCont.BalanceOf(tcc, util.Users[3]).String())

	bv, err := tc.Ctx.Contract(bridgeAddr)
	if err != nil {
		t.Errorf("contract load error")
	}

	cont := bv.(*bridge.BridgeContract)
	bcc := tc.Ctx.ContractContext(cont, util.Admin)
	fcont := cont.Front().(bridge.BridgeFront)
	log.Println(fcont.GetSequenceFrom(bcc, banker, "POLYGON"))
}

func TestSend1Mev(t *testing.T) {
	tc := setupTest()

	TAG := "TestSend1Mev "

	balBefore := getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	_, err := tc.MakeTx(util.AdminKey, bridgeAddr, "SendFromGateway", testToken, util.Users[9], amount.NewAmount(10, 0), []common.Address{}, "POLYGON", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
	}
	balAfter := getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	if balBefore.Cmp(balAfter.Int) != 0 {
		t.Errorf(TAG+"not expact maintoken value change %v -> %v", balBefore, balAfter)
	}

	store := util.Users[4]
	storeKey := util.UserKeys[4]
	// tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", store, amount.NewAmount(10000, 0))
	// tc.MustSendTx(util.UserKeys[4], tc.MainToken, "Approve", bridgeAddr, amount.NewAmount(10000000, 0))

	// SetSendMaintoken(cc *types.ContractContext, store common.Address, fromChains []string, overthens, amts []*amount.Amount)

	sendMevBalance := amount.NewAmount(1, 0)
	_, err = tc.MakeTx(util.AdminKey, bridgeAddr, "SetSendMaintoken", store,
		[]string{
			"ETHEREUM",
			"BSC",
			"KLAYTN",
			"POLYGON",
		},
		[]*amount.Amount{
			amount.NewAmount(0, 0),
			amount.NewAmount(0, 0),
			amount.NewAmount(0, 0),
			amount.NewAmount(100, 0),
		},
		[]*amount.Amount{
			sendMevBalance,
			sendMevBalance,
			sendMevBalance,
			sendMevBalance,
		},
	)
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
	}
	_, err = tc.MakeTx(util.UserKeys[1], bridgeAddr, "SetSendMaintoken", store,
		[]string{
			"ETHEREUM",
			"BSC",
			"KLAYTN",
			"POLYGON",
		},
		[]*amount.Amount{
			amount.NewAmount(0, 0),
			amount.NewAmount(0, 0),
			amount.NewAmount(1, 0),
			amount.NewAmount(100, 0),
		},
		[]*amount.Amount{
			sendMevBalance,
			sendMevBalance,
			sendMevBalance,
			sendMevBalance,
		},
	)
	if err == nil {
		t.Errorf(TAG+"expact error %v", err)
	}

	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", store, amount.MustParseAmount("1.1"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
	}

	bal, _ := tc.MakeTx(util.AdminKey, tc.MainToken, "BalanceOf", store)
	log.Println(bal)
	_, err = tc.MakeTx(storeKey, tc.MainToken, "Approve", bridgeAddr, amount.NewAmount(10000000000, 0))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
	}

	balBefore = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	_, err = tc.MakeTx(util.AdminKey, bridgeAddr, "SendFromGateway", testToken, util.Users[9], amount.NewAmount(10, 0), []common.Address{}, "POLYGON", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
	}
	balAfter = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	if balBefore.Cmp(balAfter.Int) != 0 {
		t.Errorf(TAG+"not expact maintoken value change %v -> %v", balBefore, balAfter)
	}

	balBefore = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	_, err = tc.MakeTx(util.AdminKey, bridgeAddr, "SendFromGateway", testToken, util.Users[9], amount.NewAmount(100, 0), []common.Address{}, "POLYGON", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}
	balAfter = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	if balAfter.Sub(balBefore).Cmp(amount.NewAmount(1, 0).Int) != 0 {
		t.Errorf(TAG+"not expact maintoken balance %v\n", balAfter)
	}
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}

	balBefore = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	_, err = tc.MakeTx(util.AdminKey, bridgeAddr, "SendFromGateway", testToken, util.Users[9], amount.NewAmount(10, 0), []common.Address{}, "ETHEREUM", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}
	balAfter = getBal(tc, tc.MainToken, util.Users[9], t, TAG)
	if balAfter.Cmp(balBefore.Int) != 0 {
		t.Errorf(TAG+"not expact maintoken balance %v\n", balAfter.String(), balBefore.String())
	}
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}

	balBefore = getBal(tc, tc.MainToken, util.Users[8], t, TAG)
	_, err = tc.MakeTx(util.AdminKey, bridgeAddr, "SendFromGateway", testToken, util.Users[8], amount.NewAmount(10, 0), []common.Address{}, "ETHEREUM", []byte("MEVERSE/POLYGON/MEV"))
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}
	balAfter = getBal(tc, tc.MainToken, util.Users[8], t, TAG)
	if balAfter.Cmp(balBefore.Int) != 0 {
		t.Errorf(TAG+"not expact maintoken balance after %v before %v\n", balAfter.String(), balBefore.String())
	}
	if err != nil {
		t.Errorf(TAG+"not expact error %v", err)
		return
	}

}

func getBal(tc *util.TestContext, testToken, user common.Address, t *testing.T, TAG string) *amount.Amount {
	inf, err := tc.MakeTx(util.AdminKey, testToken, "BalanceOf", user)
	if err != nil {
		t.Errorf(TAG+"TestSendToGatewayTx", err, inf)
	}
	bal := inf[0].(*amount.Amount)
	return bal
}
