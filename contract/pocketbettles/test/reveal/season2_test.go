package test

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/test/util"
)

func TestSeason2HeroReveal(t *testing.T) {
	tc := util.NewTestContext()
	egAddr := initEngin(tc)
	minterAddr, nftAddr, err := minter2Init(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "1", "2", "3", "4", "20", "2990", "0", "0", "3911")
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

	for i := 0; i < 12; i++ {
		tc.Sleep(1, nil, nil)
	}

	log.Println("tc.Ctx.TargetHeight()", tc.Ctx.TargetHeight())

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err == nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
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

	hAddr, err := makeNft(tc, egAddr)
	if err != nil {
		panic(err)
	}

	revealAddr := makeReveal2Cont(tc, nftAddr, hAddr, egAddr, "reveal.js")
	ts := getNFTList(tc, nftAddr, 0, 10)
	_, err = tc.SendTx(util.AdminKey, revealAddr, "reveal", ts)
	if err != nil {
		panic(err)
	}
	hs := getNFTList(tc, hAddr, 0, 10)
	log.Println(hs)

	checkBalanceOf(tc, nftAddr, 0)

	minter2Addr, nft2Addr, err := minter2Init(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minter2Addr, "roundSetting", "1", "2", "3", "4", "30", "4000", "0", "0", "3000")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}
	// log.Println(tc.MainToken, minterAddr)
	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minter2Addr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}
	inf, err = tc.SendTx(util.AdminKey, nft2Addr, "setMinter", minter2Addr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minter2Addr, "mint", "3")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nft2Addr, 3)

	reveal2Addr := makeReveal2Cont(tc, nft2Addr, hAddr, egAddr, "reveals2.js")

	{
		ts2 := getNFTList(tc, nft2Addr, 0, 10)
		_, err = tc.SendTx(util.AdminKey, reveal2Addr, "reveal", ts2)
		if err != nil {
			panic(err)
		}
		hs2 := getNFTList(tc, hAddr, 0, 10)
		log.Println(hs2)
	}
	for i := 0; i < 9; i++ {
		txs := []*types.Transaction{}
		keys := []key.Key{}
		for j := 0; j < 111; j++ {
			tx, err := tc.MakeTx(minter2Addr, "mint", "3")
			if err != nil {
				panic(err)
			}
			txs = append(txs, tx)
			keys = append(keys, util.AdminKey)
		}

		inf, err := tc.MultiSendTx(txs, keys)
		if err != nil {
			t.Errorf("not expect err %v, %v", inf, err)
			return
		}
	}

	checkBalanceOf(tc, nft2Addr, 2997)

	for {
		ts2 := getNFTList(tc, nft2Addr, 0, 10)
		if len(ts2) == 0 {
			break
		}
		_, err = tc.SendTx(util.AdminKey, reveal2Addr, "reveal", ts2)
		if err != nil {
			log.Println(err)
			break
		}
		c := getBalanceOf(tc, hAddr, util.Admin)
		log.Println("bal:", c)
	}

	iter := 0
	idCount1 := 0
	idCount2 := 0
	for {
		ts2 := getNFTList(tc, hAddr, iter, iter+9)
		if len(ts2) == 0 {
			break
		}
		for _, id := range ts2 {
			if id > "0x0f47" {
				idCount2++
			} else {
				idCount1++
			}
		}
		iter += 10
	}

	checkBalanceOf(tc, nft2Addr, 0)
	log.Println(idCount1, idCount2)
}

func TestInherit(t *testing.T) {
	tc := util.NewTestContext()
	egAddr := initEngin(tc)
	minterAddr, nftAddr, err := minter2Init(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "1", "2", "3", "4", "20", "2990", "0", "0", "3911")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}

	// log.Println(tc.MainToken, minterAddr)
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

	tc.SleepBlocks(16)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)

	hAddr, err := makeNft(tc, egAddr)
	if err != nil {
		panic(err)
	}

	revealAddr := makeReveal2Cont(tc, nftAddr, hAddr, egAddr, "reveal.js")
	ts := getNFTList(tc, nftAddr, 0, 10)
	_, err = tc.SendTx(util.AdminKey, revealAddr, "reveal", ts)
	if err != nil {
		panic(err)
	}
	hs := getNFTList(tc, hAddr, 0, 10)
	log.Println(hs)

	checkBalanceOf(tc, nftAddr, 0)

	minter2Addr, nft2Addr, err := minter2Init(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minter2Addr, "roundSetting", "1", "2", "3", "4", "30", "4000", "0", "0", "3000")
	if err != nil {
		log.Println(inf, err)
		t.Error(err)
		return
	}
	// log.Println(tc.MainToken, minterAddr)
	inf, err = tc.SendTx(util.AdminKey, tc.MainToken, "Approve", minter2Addr, "100000000000000")
	if err != nil {
		t.Errorf("not expect err %v %v", inf, err)
		return
	}
	inf, err = tc.SendTx(util.AdminKey, nft2Addr, "setMinter", minter2Addr)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minter2Addr, "mint", "3")
	if err != nil {
		t.Errorf("not expect err %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nft2Addr, 3)

	reveal2Addr := makeReveal2Cont(tc, nft2Addr, hAddr, egAddr, "reveals2.js")

	ts2 := getNFTList(tc, nft2Addr, 0, 10)
	_, err = tc.SendTx(util.AdminKey, reveal2Addr, "reveal", ts2)
	if err != nil {
		panic(err)
	}

	hs2 := getNFTList(tc, hAddr, 0, 10)
	log.Println(hs2)

	token1Addr := tc.MakeToken("TestToken1", "TEST1", "10000")
	log.Println("Test Token Addr", token1Addr) // 0xadCAdf65B8A05e5Fbc0cfB0dEe8De2d2BAa16bDf

	_, err = tc.SendTx(util.AdminKey, token1Addr, "approve", revealAddr, amount.NewAmount(10000, 0))
	if err != nil {
		panic(err)
	}

	am := balanceOf(tc, token1Addr, util.Admin)
	log.Println("token1Addr amt", am.String())

	_, err = tc.SendTx(util.AdminKey, revealAddr, "setPktAddr", token1Addr.String())
	if err != nil {
		panic(err)
	}

	_, err = tc.SendTx(util.AdminKey, revealAddr, "dismantle", []string{"inherit", hs2[0], hs2[1], "2"})
	if err != nil {
		panic(err)
	}
	am = balanceOf(tc, token1Addr, util.Admin)
	log.Println("token1Addr amt", am.String())

	hs2 = getNFTList(tc, hAddr, 0, 10)
	hs3 := getNFTListByOwner(tc, hAddr, revealAddr, 0, 10)
	log.Println(hs2, hs3)

	inf, err = tc.ReadTx(util.AdminKey, revealAddr, "inheritList", util.Admin)
	if err != nil {
		panic(err)
	}
	log.Println(inf)

	inf, err = tc.ReadTx(util.AdminKey, revealAddr, "inheritInfo", hs2[1])
	if err != nil {
		panic(err)
	}

	log.Println(inf)
}

func minter2Init(tc *util.TestContext, egAddr common.Address) (common.Address, common.Address, error) {
	nftAddr, err := makeNft(tc, egAddr)
	if err != nil {
		return common.Address{}, common.Address{}, err
	}

	bs, err := ioutil.ReadFile("../../minting2.js")
	if err != nil {
		return common.Address{}, common.Address{}, err
	}

	dead := "0x000000000000000000000000000000000000dead"

	// owner, nftAddr, payToken, unitPrice, formulatorAddr, genesisAddr
	// owner, nftAddr, payToken, unitPrice, formulatorAddr, genesisAddr
	inf := tc.MustSendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		util.Admin.String(),
		nftAddr.String(),
		tc.MainToken.String(),
		amount.MustParseAmount("5000"),
		dead,
		dead,
	}, true)

	minterAddr, ok := inf[0].(common.Address)
	if !ok {
		return common.Address{}, common.Address{}, errors.New("addr invalid")
	}

	return minterAddr, nftAddr, nil
}

func makeReveal2Cont(tc *util.TestContext, bAddr, hAddr, egAddr common.Address, contName string) common.Address {

	bs, err := ioutil.ReadFile("../../" + contName)
	if err != nil {
		panic(err)
	}
	inf, err := tc.SendTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		util.Admin.String(),
		bAddr.String(),
		hAddr.String(),
	}, true)
	if err != nil {
		panic(err)
	}
	var ok bool
	var revealAddr common.Address
	if revealAddr, ok = inf[0].(common.Address); !ok {
		panic("reveal not maked")
	}
	log.Println(revealAddr)
	tc.MustSendTx(util.AdminKey, bAddr, "setApprovalForAll", revealAddr.String(), true)
	tc.MustSendTx(util.AdminKey, bAddr, "setMinter", revealAddr.String(), true)
	tc.MustSendTx(util.AdminKey, hAddr, "setApprovalForAll", revealAddr.String(), true)
	tc.MustSendTx(util.AdminKey, hAddr, "setMinter", revealAddr.String(), true)
	return revealAddr
}

func TestMintingS3(t *testing.T) {
	tc := util.NewTestContext()
	egAddr := initEngin(tc)
	minterAddr, nftAddr, err := minter2Init(tc, egAddr)
	if err != nil {
		t.Error(err)
		return
	}
	inf, err := tc.SendTx(util.AdminKey, minterAddr, "roundSetting", "1", "2", "3", "4", "10", "20000", "0", "0", "2700")
	log.Println(inf, err)

	tc.SleepBlocks(10)

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

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "addEventInfo", 1, 5, 3)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}
	inf, err = tc.SendTx(util.AdminKey, minterAddr, "addEventInfo", 2, 10, 3)
	if err != nil {
		t.Errorf("not expect must mint nft %v, %v", inf, err)
		return
	}

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 3)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "2")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 6)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 7)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "3")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 10)

	inf, err = tc.SendTx(util.AdminKey, minterAddr, "mint", "1")
	if err != nil {
		t.Errorf("expect err not in round %v, %v", inf, err)
		return
	}

	checkBalanceOf(tc, nftAddr, 12)
	log.Println(getTotalSupply(tc, nftAddr), getMintCount(tc, minterAddr))

	for i := 0; i < 5; i++ {

		tc.MustSendTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[i], amount.MustParseAmount("100000"))

		inf, err = tc.SendTx(util.UserKeys[i], tc.MainToken, "Approve", minterAddr, "100000000000000")
		if err != nil {
			t.Errorf("not expect err %v %v", inf, err)
			return
		}
		inf, err = tc.SendTx(util.UserKeys[i], minterAddr, "mint", "5")
		if err != nil {
			t.Errorf("expect err not in round %v, %v", inf, err)
			return
		}

		log.Println(i, getBalanceOf(tc, nftAddr, util.Users[i]))
	}

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "getEventInfos")
	log.Println(inf, err)

	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "eventSendCount", 1)
	log.Println(inf, err)
	inf, err = tc.ReadTx(util.AdminKey, minterAddr, "eventSendCount", 2)
	log.Println(inf, err)
}
