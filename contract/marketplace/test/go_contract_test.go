package test

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/extern/test/util"
)

var dataPath = "../market_data.js"
var operationPath = "../market_operation.js"

func _init() (tc *util.TestContext, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr common.Address, nft1IDs, nft2IDs []*big.Int) {
	tc, egAddr, dataAddr, marketAddr = setupMarketCont("200", "200", "100", dataPath, operationPath)
	tokenAddr = deployToken(tc, "STOKEN", "STC")
	nft1Addr = deployNFT(tc, egAddr, "FristNFT", "FNFT")
	nft2Addr = deployNFT(tc, egAddr, "SecondNFT", "SNFT")

	nft1IDs = mintNFT(tc, nft1Addr, util.Users[0], util.Users[1], util.Users[2])
	nft2IDs = mintNFT(tc, nft2Addr, util.Users[1], util.Users[2], util.Users[3])
	log.Printf("egAddr: %v marketAddr: %v nftAddr: %v tokenAddr : %v", egAddr, marketAddr, nft1Addr, tokenAddr) // 0x98C1C0Ea6A88E31983C1De3f5b91EfE9DAd8D2CB 0x69cE8cee401cA81D672E27aedA52dF6c3A805063 0xeAF7412ce6Ec9578d72c17b2216bA256EF01a460

	// await nftMarketData.setFoundationAdminAddress(pocketBattles.address, pocketAdmin, {from: owner });
	// await nftMarketData.setFoundationAdminAddress(kingOfPlanets.address, kopAdmin, {from: owner });
	inf, err := tc.MakeTx(util.AdminKey, dataAddr, "setFoundationAdminAddress", nft1Addr, util.Users[0])
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setFoundationAdminAddress", nft2Addr, util.Users[1])
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setBurnAddress", nft1Addr, common.HexToAddress("0xdead"))
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setBurnAddress", nft2Addr, common.HexToAddress("0xdead"))
	if err != nil {
		log.Println(inf)
		panic(err)
	}

	inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setERC20Contract", "MEV", tc.MainToken)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setERC20Contract", "USDT", tokenAddr)
	if err != nil {
		log.Println(inf)
		panic(err)
	}

	inf, err = tc.MakeTx(util.AdminKey, nft1Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[0], nft1Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[1], nft1Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[2], nft1Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.AdminKey, nft2Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[0], nft2Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[1], nft2Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	inf, err = tc.MakeTx(util.UserKeys[2], nft2Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		log.Println(inf)
		panic(err)
	}

	inf, err = tc.ReadTx(util.UserKeys[1], marketAddr, "abi", "getRoyaltyFee")
	if err != nil {
		log.Println(inf)
		panic(err)
	}
	if iss, ok := inf[0].([]interface{}); ok {
		log.Println(iss[0])
		// if strs, ok := iss[0].([]interface{}); ok {
		// 	ts := make([]string, len(strs))
		// 	for i, s := range strs {
		// 		ts[i] = s.(string)
		// 	}
		// }
	}

	return
}

func TestRegisterMarketItemBuyNowWithToken(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("50000")

	inf, err := tc.MakeTx(util.UserKeys[1], marketAddr, "registerMarketItem", nft1Addr, nft1IDs[1], buyp, "USDT", "0", amount.MustParseAmount("10000"))
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println("registerMarketItem", inf)

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket._settle: is not sufficient balance" {
		t.Errorf("error expect but err is nil or not start with 'NftMarket._settle' %+v", err)
		return
	}
	log.Println(inf)

	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[2], amount.MustParseAmount("50000"))
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket._settle: is not allowanced" {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println(inf)

	inf, err = tc.MakeTx(util.UserKeys[2], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println(inf)

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, _ = tc.ReadTx(util.UserKeys[2], tokenAddr, "balanceOf", util.Users[2])
	am := inf[0].(*amount.Amount)
	if am.String() != "0" {
		t.Errorf("expect zero " + am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", util.Users[1])
	am = inf[0].(*amount.Amount)
	if am.String() != "47500" {
		t.Errorf("expect 47500 " + am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", dataAddr)
	am = inf[0].(*amount.Amount)
	if am.String() != "1000" {
		t.Errorf("expect 1000 " + am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", util.Users[0])
	am = inf[0].(*amount.Amount)
	if am.String() != "1000" {
		t.Errorf("expect 1000 " + am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", common.HexToAddress("0xdead"))
	am = inf[0].(*amount.Amount)
	if am.String() != "500" {
		t.Errorf("expect 500 " + am.String())
		return
	}
}

func TestRegisterMarketItemNftTransferBuyNowWithTokenRegisterMarketItemBuyNowWithToken(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("50000")

	inf, err := tc.MakeTx(util.UserKeys[1], marketAddr, "registerMarketItem", nft1Addr, nft1IDs[1], buyp, "USDT", "0", "100000")
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println("registerMarketItem", inf)

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket._settle: is not sufficient balance" {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println(inf)

	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[2], amount.MustParseAmount("50000"))
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket._settle: is not allowanced" {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println(inf)

	inf, err = tc.MakeTx(util.UserKeys[2], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	log.Println(inf)

	inf, err = tc.MakeTx(util.UserKeys[1], nft1Addr, "approve", util.Admin, nft1IDs[1])
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}
	inf, err = tc.MakeTx(util.AdminKey, nft1Addr, "transferFrom", util.Users[1], util.Admin, nft1IDs[1])
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket._settle: seller is Not owner" {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nft1Addr, "setApprovalForAll", marketAddr, true)
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}
	inf, err = tc.MakeTx(util.AdminKey, marketAddr, "registerMarketItem", nft1Addr, nft1IDs[1], buyp, "USDT", "0", "100000")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "buyNowWithToken", nft1Addr, nft1IDs[1], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, _ = tc.ReadTx(util.UserKeys[2], tokenAddr, "balanceOf", util.Users[2])
	am := inf[0].(*amount.Amount)
	if am.String() != "0" {
		t.Errorf("expect zero")
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", util.Admin)
	am = inf[0].(*amount.Amount)
	if am.String() != "47500" {
		t.Errorf("expect 47500 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", dataAddr)
	am = inf[0].(*amount.Amount)
	if am.String() != "1000" {
		t.Errorf("expect 1000 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", util.Users[0])
	am = inf[0].(*amount.Amount)
	if am.String() != "1000" {
		t.Errorf("expect 1000 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", common.HexToAddress("0xdead"))
	am = inf[0].(*amount.Amount)
	if am.String() != "500" {
		t.Errorf("expect 500 " + am.String())
		return
	}
}

func TestSuggestItemToBuyCancelItemToBuyWithSuggester(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("2500")

	inf, err := tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket.suggestItemToBuy: is not allowanced" {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	_, err = tc.MakeTx(util.UserKeys[1], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket.suggestItemToBuy: is not sufficient balance" {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[1], amount.MustParseAmount("3500"))
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[1], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[1], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 1 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket.suggestItemToBuy: not valid suggestBiddingPrice" {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	_, err = tc.MakeTx(util.UserKeys[1], tokenAddr, "approve", marketAddr, amount.MustParseAmount("3500"))
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], amount.MustParseAmount("3500"), "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[1], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[1], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 2 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "cancelItemToBuyWithSuggester", nft2Addr, nft2IDs[1], "USDT", amount.MustParseAmount("2500"))
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[1], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[1], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 1 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "cancelItemToBuyWithSuggester", nft2Addr, nft2IDs[1], "USDT", amount.MustParseAmount("3500"))
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[1], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[1], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 0 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
		return
	}

}

func TestSuggestItemToBuyCancelItem(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("2500")

	_, err := tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[1], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	_, err = tc.MakeTx(util.UserKeys[1], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err := tc.MakeTx(util.UserKeys[1], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "cancelItemToBuyWithSuggester", nft2Addr, nft2IDs[1], "USDT", buyp)
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}
}

func TestSuggestItemToBuyAcceptItemToBuyWithSeller(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("2500")
	_, err := tc.MakeTx(util.UserKeys[0], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[0], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err := tc.MakeTx(util.UserKeys[0], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "acceptItemToBuyWithSeller", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}
}

func TestSuggestItemToBuyChangeNftOwnerAcceptItemToBuyWithSeller(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}
	buyp := amount.MustParseAmount("2500")
	_, err := tc.MakeTx(util.UserKeys[0], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[0], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	inf, err := tc.MakeTx(util.UserKeys[0], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[0], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[0], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 1 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}

	_, err = tc.MakeTx(util.UserKeys[1], nft2Addr, "approve", util.Admin, nft2IDs[0])
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, nft2Addr, "transferFrom", util.Users[1], util.Admin, nft2IDs[0])
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "acceptItemToBuyWithSeller", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err == nil || err.Error() != "NftMarket.acceptItemToBuy: seller is not owner" {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	_, err = tc.MakeTx(util.AdminKey, nft2Addr, "approve", util.Users[1], nft2IDs[0])
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.UserKeys[1], nft2Addr, "transferFrom", util.Admin, util.Users[1], nft2IDs[0])
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[0], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[0], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 1 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "acceptItemToBuyWithSeller", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.ReadTx(util.UserKeys[0], marketAddr, "getItemSuggestionInfos", nft2Addr, nft2IDs[0], "USDT")
	if strings.Count(fmt.Sprintln(inf), "SuggestItem") != 0 {
		t.Errorf("error not expect len 1 %v, err %v", strings.Count(fmt.Sprintln(inf), "SuggestItem"), err)
	}
}

func TestMultiSuggestItemToBuyAcceptItemToBuyWithSeller(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("2500")
	buyp2 := amount.MustParseAmount("3000")
	_, err := tc.MakeTx(util.UserKeys[0], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[0], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.UserKeys[2], tokenAddr, "approve", marketAddr, buyp2)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[2], buyp2)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err := tc.MakeTx(util.UserKeys[0], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[0], buyp2, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], marketAddr, "acceptItemToBuyWithSeller", nft2Addr, nft2IDs[0], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}
}

func TestMultiSuggestItemToBuyLowAcceptItemToBuyWithSeller(t *testing.T) {
	tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs := _init()
	if false {
		log.Println(tc, egAddr, dataAddr, marketAddr, tokenAddr, nft1Addr, nft2Addr, nft1IDs, nft2IDs)
	}

	buyp := amount.MustParseAmount("2500")
	buyp2 := amount.MustParseAmount("3000")
	_, err := tc.MakeTx(util.UserKeys[3], tokenAddr, "approve", marketAddr, buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[3], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Mint", util.Users[4], buyp)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.UserKeys[4], tokenAddr, "approve", marketAddr, buyp2)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Mint", util.Users[4], buyp2)
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return
	}

	inf, err := tc.MakeTx(util.UserKeys[3], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[4], marketAddr, "suggestItemToBuyWithSuggester", nft2Addr, nft2IDs[1], buyp2, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[2], marketAddr, "acceptItemToBuyWithSeller", nft2Addr, nft2IDs[1], buyp2, "USDT")
	if err != nil {
		t.Errorf("error not expect %v %+v", inf, err)
		return
	}

	inf, _ = tc.ReadTx(util.UserKeys[4], tokenAddr, "balanceOf", util.Users[4])
	am := inf[0].(*amount.Amount)
	if am.String() != "0" {
		t.Errorf("expect zero")
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[2], tokenAddr, "balanceOf", util.Users[2])
	am = inf[0].(*amount.Amount)
	if am.String() != "2850" {
		t.Errorf("expect 2850 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[2], tokenAddr, "balanceOf", dataAddr)
	am = inf[0].(*amount.Amount)
	if am.String() != "60" {
		t.Errorf("expect 60 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[2], tokenAddr, "balanceOf", util.Users[1])
	am = inf[0].(*amount.Amount)
	if am.String() != "60" {
		t.Errorf("expect 60 %v", am.String())
		return
	}
	inf, _ = tc.ReadTx(util.UserKeys[0], tokenAddr, "balanceOf", common.HexToAddress("0xdead"))
	am = inf[0].(*amount.Amount)
	if am.String() != "30" {
		t.Errorf("expect 30 " + am.String())
		return
	}
}
