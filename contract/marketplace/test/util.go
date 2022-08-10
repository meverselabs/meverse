package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

func initEngin() (tc *util.TestContext, egAddr common.Address) {
	tc = util.NewTestContext()
	url := "https://meverse-engin.s3.ap-northeast-2.amazonaws.com/jsengin_v2_debug.so"

	ContArgs := &engin.EnginContractConstruction{}
	ContType := &engin.EnginContract{}
	egAddr = tc.DeployContract(ContType, ContArgs)

	_, err := tc.MakeTx(util.AdminKey, egAddr, "AddEngin", "JSContractEngin", "javascript vm on meverse verseion 0.1.0", url)
	if err != nil {
		panic(err)
	}
	_, err = tc.ReadTx(util.AdminKey, egAddr, "EnginVersion", "JSContractEngin")
	if err != nil {
		panic("error not expect " + err.Error())
	}
	return
}

func deployNFT(tc *util.TestContext, egAddr common.Address, _name string, _symbol string) common.Address {
	var nftAddr common.Address
	{
		bs, err := ioutil.ReadFile("../nft721.js")
		if err != nil {
			panic(err)
		}

		inf, err := tc.MakeTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
			util.Admin.String(),
			_name,
			_symbol,
		}, true)
		if err != nil {
			panic(err)
		}
		var ok bool
		if nftAddr, ok = inf[0].(common.Address); !ok {
			panic(err)
		}
	}
	return nftAddr
}

func deployToken(tc *util.TestContext, _name string, _symbol string) common.Address {
	ContArgs := &token.TokenContractConstruction{
		Name:   _name,
		Symbol: _symbol,
	}
	ContType := &token.TokenContract{}
	return tc.DeployContract(ContType, ContArgs)
}

func setupMarketCont(marketFeeStr, royaltyFeeStr, burnFee, dataPath, operationPath string) (tc *util.TestContext, egAddr, dataAddr, marketAddr common.Address) {
	tc, egAddr = initEngin()

	{
		bs, err := ioutil.ReadFile(dataPath)
		if err != nil {
			panic(err)
		}

		inf, err := tc.MakeTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{util.Admin.String()}, true)
		if err != nil {
			panic(err)
		}
		var ok bool
		if dataAddr, ok = inf[0].(common.Address); !ok {
			panic("dataAddr is not set")
		}

		_, err = tc.ReadTx(util.AdminKey, dataAddr, "getMarketOperationAddress")
		if err != nil {
			panic(err)
		}
	}

	{
		bs, err := ioutil.ReadFile(operationPath)
		if err != nil {
			panic(err)
		}

		inf, err := tc.MakeTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
			util.Admin.String(),
			marketFeeStr, royaltyFeeStr,
		}, true)
		log.Println(inf, err)
		if err != nil {
			panic("error not expect")
		}
		var ok bool
		marketAddr, ok = inf[0].(common.Address)
		if !ok {
			panic(fmt.Sprintf("deplory contract not retruned address %v", inf))
		}
		log.Println(marketAddr)

		inf, err = tc.MakeTx(util.AdminKey, marketAddr, "setMandatoryMarketDataContract", dataAddr)
		if err != nil {
			panic(fmt.Sprintf("error not expect %+v", err))
		}
		log.Println(inf)
		inf, err = tc.MakeTx(util.AdminKey, dataAddr, "setMandatoryInitContract", marketAddr)
		if err != nil {
			panic(fmt.Sprintf("error not expect %+v", err))
		}
		log.Println(inf)
		inf, err = tc.MakeTx(util.AdminKey, marketAddr, "setBurnFee", burnFee)
		if err != nil {
			log.Println(inf)
			panic(err)
		}
	}
	return
}

func mintNFT(tc *util.TestContext, nftAddr common.Address, addrs ...common.Address) []*big.Int {
	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mintBatch", addrs)
	if err != nil {
		panic(err)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		panic("no result")
	}
	if len(is) != len(addrs) {
		panic("not match mint result")
	}
	tokenIDs := make([]*big.Int, len(addrs))
	for i, h := range is {
		_, err := tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", addrs[i], amount.MustParseAmount("100"))
		if err != nil {
			panic(err)
		}
		str := h.(string)
		str = strings.Replace(str, "0x", "", -1)
		tokenIDs[i], ok = big.NewInt(0).SetString(str, 16)
		if !ok {
			panic(str + " is can not conv big.Int")
		}
	}
	return tokenIDs
}
