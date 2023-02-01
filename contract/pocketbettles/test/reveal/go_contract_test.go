package test

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/nft721"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/extern/test/util"
)

func TestConverter(t *testing.T) {
	tc := util.NewTestContext()

	token1Addr := tc.MakeToken("TestToken1", "TEST1", "10000")
	log.Println("Test Token Addr", token1Addr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf
	token2Addr := tc.MakeToken("TestToken2", "TEST2", "10000")
	log.Println("Test Token Addr", token2Addr) //

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
		token1Addr, token2Addr, token1Addr, "testlp", "TESTLP", util.Admin,
		util.Admin, FEE, ADMINFEE, WINNERFEE, whiteListAddr, hash.Hash256{}, UniClassID)
	log.Println("TestExecuteContractTx", inf)

	inf = tc.MustSendTx(util.AdminKey, token1Addr, "Approve", routerAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve MainToken", inf)
	inf = tc.MustSendTx(util.AdminKey, token2Addr, "Approve", routerAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve tokenAddr", inf)

	tc.Sleep(1, nil, nil)

	inf = tc.MustSendTx(util.AdminKey, routerAddr, "UniAddLiquidity", token1Addr, token2Addr, amount.NewAmount(1000, 0), amount.NewAmount(1000, 0), amount.NewAmount(0, 0), amount.NewAmount(0, 0))
	log.Println("UniAddLiquidity", inf)

	egAddr := initEngin(tc)

	bs, err := ioutil.ReadFile("../converter.js")
	if err != nil {
		panic(err)
	}

	inf = tc.MustSendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		routerAddr.String(),
		util.Admin.String(),
		token1Addr.String(),
		token2Addr.String(),
	}, true)
	converterAddr, ok := inf[0].(common.Address)
	if !ok {
		panic(err)
	}
	log.Println("SetMinter", converterAddr.String())
	inf = tc.MustSendTx(util.AdminKey, token2Addr, "SetMinter", converterAddr, true)
	log.Println("converterAddr", inf)

	inf, err = tc.ReadTx(util.AdminKey, routerAddr, "GetAmountsOut", amount.MustParseAmount("12"), []common.Address{token1Addr, token2Addr})
	log.Println(routerAddr.String(), inf, err)

	inf = tc.MustSendTx(util.AdminKey, token1Addr, "Approve", converterAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve MainToken", inf)
	inf = tc.MustSendTx(util.AdminKey, token2Addr, "Approve", converterAddr, amount.NewAmount(1000000000, 0))
	log.Println("Approve tokenAddr", inf)

	log.Println("admin", util.Admin.String())
	inf, err = tc.ReadTx(util.AdminKey, converterAddr, "getConverterRatio")
	log.Println(inf, err)

	inf, err = tc.SendTx(util.AdminKey, converterAddr, "converter", amount.MustParseAmount("12"))
	log.Println(inf, err)

	inf, err = tc.ReadTx(util.AdminKey, token1Addr, "BalanceOf", converterAddr)
	log.Println("balt", inf, err)
	inf, err = tc.ReadTx(util.AdminKey, token1Addr, "BalanceOf", util.Admin)
	log.Println("balt", inf, err)

	inf, err = tc.SendTx(util.AdminKey, converterAddr, "call", amount.MustParseAmount("11"))
	log.Println("call", inf, err)

	inf, err = tc.ReadTx(util.AdminKey, token1Addr, "BalanceOf", converterAddr)
	log.Println("balta", inf, err)
	inf, err = tc.ReadTx(util.AdminKey, token1Addr, "BalanceOf", util.Admin)
	log.Println("balta", inf, err)
}

func mintGenesis(tc *util.TestContext) (genesisAddr common.Address, err error) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	_name := "TestNFT"
	_symbol := "TNFT"

	genesisAddr = tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Owner:  util.Admin,
		Name:   _name,
		Symbol: _symbol,
	})

	_, err = tc.SendTx(util.AdminKey, genesisAddr, "Name")
	return

}

func minterInit(tc *util.TestContext, formulatorAddr common.Address) (common.Address, common.Address, common.Address, error) {
	egAddr := initEngin(tc)

	// testFormulatorAddr := common.HexToAddress("0xBaa3C856fbA6FFAda189D6bD0a89d5ef7959c75E")
	genesisAddr, err := mintGenesis(tc)
	if err != nil {
		return common.Address{}, common.Address{}, common.Address{}, err
	}

	nftAddr, err := makeNft(tc, egAddr)
	if err != nil {
		return common.Address{}, common.Address{}, common.Address{}, err
	}

	bs, err := ioutil.ReadFile("../minting.js")
	if err != nil {
		return common.Address{}, common.Address{}, common.Address{}, err
	}

	// owner, nftAddr, payToken, unitPrice, formulatorAddr, genesisAddr
	// owner, nftAddr, payToken, unitPrice, formulatorAddr, genesisAddr
	inf := tc.MustSendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		util.Admin.String(),
		nftAddr.String(),
		tc.MainToken.String(),
		amount.MustParseAmount("5000"),
		formulatorAddr.String(),
		genesisAddr.String(),
	}, true)

	minterAddr, ok := inf[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, common.Address{}, errors.New("addr invalid")
	}

	return minterAddr, nftAddr, genesisAddr, nil
}

func makeNft(tc *util.TestContext, egAddr common.Address) (common.Address, error) {
	var nftAddr common.Address
	{
		bs, err := ioutil.ReadFile("../../../marketplace/nft721.js")
		if err != nil {
			return common.Address{}, err
		}
		inf, err := tc.SendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
			util.Admin.String(),
			"_name",
			"_symbol",
		}, true)
		if err != nil {
			return common.Address{}, err
		}
		var ok bool
		if nftAddr, ok = inf[0].(common.Address); !ok {
			return common.Address{}, errors.New("addr invalid")
		}
	}
	return nftAddr, nil
}

func TestMintingRound1(t *testing.T) {
	tc := util.NewTestContext()
	formulatorAddr := initFormulator(tc)
	minterAddr, nftAddr, _, err := minterInit(tc, formulatorAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "20", "99", "100", "199", "200", "299", "100", "2500", "2150")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("is not approved %v, %v", inf, err)
		return
	}

	// log.Println(tc.MainToken, minterAddr)
	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minterAddr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("this is must invalid round height 10 %v, %v", inf, err)
		return
	}

	tc.SleepBlocks(11)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	tc.SendTx(util.AdminKey, tc.MainToken, "Approve", formulatorAddr, amount.MustParseAmount("1000000000000000000"))

	inf, _ = tc.SendTx(util.AdminKey, formulatorAddr, "CreateAlphaBatch", 5)
	addrs := inf[0].([]common.Address)
	inf, err = tc.SendTx(util.AdminKey, formulatorAddr, "CreateSigma", addrs[:4])
	if err == nil {
		t.Errorf("expect err create sigma %v, %v", inf, err)
		log.Println(inf, err)
		return
	}

	inf, _ = tc.ReadTx(util.AdminKey, formulatorAddr, "BalanceOf", util.Admin)
	infs := inf[0].([]interface{})
	if infs[0].(uint32) != 5 {
		t.Errorf("invalid create alpha")
		return
	}

	tc.SleepBlocks(20)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "CreateSigma", addrs[:4])
	if err != nil {
		t.Errorf("invalid create sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err not mint permission %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, nftAddr, "setMinter", minterAddr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "4")
	if err == nil {
		t.Errorf("expect err You can only mint up to 3 heroes in one transaction %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "0")
	if err == nil {
		t.Errorf("expect err You can only mint up to 3 heroes in one transaction %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 4)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "Revoke", addrs[0])
	if err != nil {
		t.Errorf("invalid Revoke sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("expect err Only Sigma, Omega Formulator holders can participate %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 4)
}

func TestMintingRound2(t *testing.T) {
	tc := util.NewTestContext()
	formulatorAddr := initFormulator(tc)
	minterAddr, nftAddr, genesisAddr, err := minterInit(tc, formulatorAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "20", "99", "100", "199", "200", "299", "100", "2500", "2150")
	log.Println(inf, err)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("is not approved %v, %v", inf, err)
		return
	}

	log.Println(tc.MainToken, minterAddr)
	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minterAddr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("this is must invalid round height 10 %v, %v", inf, err)
		return
	}

	tc.SleepBlocks(11)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	tc.SendTx(util.AdminKey, tc.MainToken, "Approve", formulatorAddr, amount.MustParseAmount("1000000000000000000"))

	inf, _ = tc.SendTx(util.AdminKey, formulatorAddr, "CreateAlphaBatch", 5)
	addrs := inf[0].([]common.Address)
	inf, err = tc.SendTx(util.AdminKey, formulatorAddr, "CreateSigma", addrs[:4])
	log.Println(inf, err)

	inf, _ = tc.ReadTx(util.AdminKey, formulatorAddr, "BalanceOf", util.Admin)
	infs := inf[0].([]interface{})
	if infs[0].(uint32) != 5 {
		t.Errorf("invalid create alpha")
		return
	}

	tc.SleepBlocks(20)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "CreateSigma", addrs[:4])
	if err != nil {
		t.Errorf("invalid create sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err not mint permission %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, nftAddr, "setMinter", minterAddr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "4")
	if err == nil {
		t.Errorf("expect err You can only mint up to 3 heroes in one transaction %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "0")
	if err == nil {
		t.Errorf("expect err You can only mint up to 3 heroes in one transaction %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 4)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "Revoke", addrs[0])
	if err != nil {
		t.Errorf("invalid Revoke sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("expect err Only Sigma, Omega Formulator holders can participate %v, %v", inf, err)
		return
	}

	tc.SleepBlocks(51)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 5)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "Revoke", addrs[4])
	if err != nil {
		t.Errorf("invalid Revoke sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("expect err %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, genesisAddr, "Mint", "1")
	if err != nil {
		t.Errorf("expect err %v, %v", inf, err)
		return
	}
	gaddr := inf[0].([]interface{})[0].(hash.Hash256)
	log.Println(gaddr)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("expect err %v, %v", inf, err)
		return
	}
	checkBalanceOf(tc, nftAddr, 6)

	inf, err = tc.SendTx(util.AdminKey, genesisAddr, "Burn", gaddr)
	if err != nil {
		t.Errorf("expect err %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("expect err Alpha, Sigma, Omega Formulator holders %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "addCommunityEventAddr", []common.Address{util.Admin})
	if err != nil {
		t.Errorf("expect err Alpha, Sigma, Omega Formulator holders %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("expect err %v, %v", inf, err)
		return
	}
	checkBalanceOf(tc, nftAddr, 7)

	log.Println(inf)
}

func TestMintingRound3(t *testing.T) {
	tc := util.NewTestContext()
	formulatorAddr := initFormulator(tc)
	minterAddr, nftAddr, _, err := minterInit(tc, formulatorAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "20", "99", "100", "199", "200", "299", "100", "2500", "2150")
	log.Println(inf, err)

	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minterAddr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, nftAddr, "setMinter", minterAddr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	tc.SleepBlocks(200)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)
}

func TestMintover(t *testing.T) {
	tc := util.NewTestContext()
	formulatorAddr := initFormulator(tc)
	minterAddr, nftAddr, _, err := minterInit(tc, formulatorAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "1", "2", "3", "4", "5", "400", "1", "1", "100")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minterAddr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, nftAddr, "setMinter", minterAddr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	for i := 0; i < 34; i++ {
		inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
		if err != nil {
			t.Errorf("expect err not in round %v, %v", inf, err)
			return
		}
	}

	inf, err = tc.ReadTx(util.AdminKey, nftAddr, "totalSupply")
	log.Println(inf, err)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err mint max value has been reached %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 102)
}

func TestMintingFailBetweenRound(t *testing.T) {
	tc := util.NewTestContext()
	formulatorAddr := initFormulator(tc)
	minterAddr, nftAddr, _, err := minterInit(tc, formulatorAddr)
	if err != nil {
		t.Error(err)
		return
	}

	inf, err := tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minterAddr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}
	tc.SendTx(util.AdminKey, tc.MainToken, "Approve", formulatorAddr, amount.MustParseAmount("1000000000000000000"))

	inf, _ = tc.SendTx(util.AdminKey, formulatorAddr, "CreateAlphaBatch", 5)
	addrs := inf[0].([]common.Address)
	tc.SleepBlocks(20)

	_, err = tc.SendTx(util.AdminKey, formulatorAddr, "CreateSigma", addrs[:4])
	if err != nil {
		t.Errorf("invalid create sigma")
		return
	}

	inf, err = tc.SendTx(util.AdminKey, nftAddr, "setMinter", minterAddr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())
		t.Errorf("expect err round not setup%v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "40", "60", "80", "100", "120", "140", "100", "2500", "2150")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())
		t.Errorf("expect err round not started%v, %v", inf, err)
		return
	}

	tc.SleepBlocks(4)

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err != nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}
	log.Println("round", inf)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}
	checkBalanceOf(tc, nftAddr, 3)

	tc.SleepBlocks(20)

	log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err == nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)

	tc.SleepBlocks(20)

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err != nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}
	log.Println("round", inf)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}
	checkBalanceOf(tc, nftAddr, 6)

	tc.SleepBlocks(20)

	log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err == nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 6)

	tc.SleepBlocks(19)

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err != nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}
	log.Println("round", inf)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}
	checkBalanceOf(tc, nftAddr, 9)

	tc.SleepBlocks(20)

	log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err == nil {
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getRound")
	if err == nil {
		log.Println(inf, err)
		t.Errorf("not expect err round1 end %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 9)
}

func TestMarge(t *testing.T) {
	tc := util.NewTestContext()
	egAddr := initEngin(tc)

	nftAddr, err := makeNft(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}

	IDs := []string{"0x01", "0x02", "0x03", "0x04", "0x05"}
	for _, id := range IDs {
		tc.MustSendTx(util.AdminKey, nftAddr, "mintWithID", id)
	}

	bs, err := ioutil.ReadFile("../../reveal.js")
	if err != nil {
		t.Errorf("%v", err)
	}
	inf, err := tc.SendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		util.Admin.String(),
		common.Address{}.String(),
		nftAddr.String(),
	}, true)
	if err != nil {
		t.Errorf("%v %v", inf, err)
	}
	var ok bool
	var revealAddr common.Address
	if revealAddr, ok = inf[0].(common.Address); !ok {
		t.Errorf("%v %v", inf, err)
	}
	log.Println(revealAddr)
}
