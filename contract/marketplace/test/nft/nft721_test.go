package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/contract/nft721/nft721receiver"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {
}

const (
	None uint8 = 1 + iota
)

func TestName(t *testing.T) {
	tc := util.NewTestContext()
	// TAG := "TESTNAME"
	_name := "TESTNFT"
	_symbol := "TNFT"

	// dat, err := os.ReadFile("./engin/jsengin.so")
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}

	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "name")
	if err != nil {
		t.Errorf("error not expected %+v", err)
	} else if iss, ok := inf[0].([]interface{}); !ok {
		t.Error("Name", "invalid readTx Type")
		return
	} else if len(iss) == 0 {
		t.Error("Name", "invalid readTx result value")
		return
	} else if name, ok := iss[0].(string); !ok {
		t.Error("Name", "name not returns string")
		return
	} else if name != _name {
		t.Error("Name", "name not matchd init name")
		return
	}

	inf, err = tc.ReadTx(util.AdminKey, nftAddr, "symbol")
	if err != nil {
		t.Errorf("error not expected %+v", err)
	} else if iss, ok := inf[0].([]interface{}); !ok {
		t.Error("Name", "invalid readTx Type")
		return
	} else if len(iss) == 0 {
		t.Error("Name", "invalid readTx result value")
		return
	} else if symbol, ok := iss[0].(string); !ok {
		t.Error("Name", "symbol not returns string")
		return
	} else if symbol != _symbol {
		t.Error("Name", "symbol not matchd init symbol")
		return
	}
}

func deployNFT(tc *util.TestContext, t *testing.T, _name string, _symbol string) (common.Address, error) {
	url := "https://meverse-engin.s3.ap-northeast-2.amazonaws.com/jsengin_v2_debug.so"

	ContArgs := &engin.EnginContractConstruction{}
	ContType := &engin.EnginContract{}
	egAddr := tc.DeployContract(ContType, ContArgs)

	_, err := tc.MakeTx(util.AdminKey, egAddr, "AddEngin", "JSContractEngin", "javascript vm on meverse verseion 0.1.0", url)
	if err != nil {
		t.Errorf("error not expect")
		return common.Address{}, nil
	}
	inf, err := tc.ReadTx(util.AdminKey, egAddr, "EnginVersion", "JSContractEngin")
	if err != nil {
		t.Errorf("error not expect %+v", err)
		return common.Address{}, nil
	}

	var nftAddr common.Address
	{
		bs, err := ioutil.ReadFile("../../nft721.js")
		if err != nil {
			t.Error(err)
			return common.Address{}, nil
		}

		inf, err = tc.MakeTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
			util.Admin.String(),
			_name,
			_symbol,
		}, true)
		if err != nil {
			t.Errorf("error not expected %+v", err)
			return common.Address{}, nil
		}
		var ok bool
		if nftAddr, ok = inf[0].(common.Address); !ok {
			t.Errorf("nftAddr is not set")
			return common.Address{}, nil
		}
	}
	return nftAddr, err
}

func _testMint3NFT(t *testing.T, i int) (nftAddr common.Address, tc *util.TestContext) {
	TAG := "MINTTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc = util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	inf, err := tc.MakeTx(util.UserKeys[0], nftAddr, "mint", big.NewInt(1))
	if err == nil {
		t.Error(TAG, "minted not owner", err, inf)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(1))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 1 {
		t.Error(TAG, "not match mint result")
		return
	}
	str := is[0].(string)
	nftid, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id", is[0], ":")
		return
	}
	// bs := hex.DecodeString(nftid)
	// hs := hash.HexToHash(nftid)
	checkMap := map[*big.Int]bool{}
	checkMap[nftid] = true

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(2))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok = inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 2 {
		t.Error(TAG, "not match mint result")
		return
	}
	for _, i := range is {
		str := i.(string)
		nftid, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			t.Error(TAG, "not nft id", nftid, "2")
		}
		// hs := hash.HexToHash(nftid)
		checkMap[nftid] = true
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	str = inf[0].(string)
	bi, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	} else if bi.Cmp(big.NewInt(int64(len(checkMap)))) != 0 {
		t.Errorf(TAG, "not expect totalSupply want 3 get %v", bi.Int64())
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	str = inf[0].(string)
	adminBal, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	}

	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect totalSupply and admin balanceOf difference")
		return
	}

	return
}

func TestMint(t *testing.T) {
	_testMint3NFT(t, 1)
}

func TestBurn(t *testing.T) {
	TAG := "BURNTEST"

	nftAddr, tc := _testMint3NFT(t, 2)

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "tokenByIndex", big.NewInt(0))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	str := inf[0].(string)
	nftid, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id")
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "burn", nftid)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "tokenByIndex", big.NewInt(0))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	str = inf[0].(string)
	nftid2, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id")
		return
	}

	if nftid.String() == nftid2.String() {
		t.Error(TAG, "expect different nftid but same")
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	str = inf[0].(string)
	bi, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	} else if bi.Cmp(big.NewInt(2)) != 0 {
		t.Errorf(TAG, "not expect totalSupply want 2 get %v", bi.Int64())
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	str = inf[0].(string)
	adminBal, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	}

	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect totalSupply and admin balanceOf difference")
	}

	for i := bi.Int64(); i > 0; i-- {
		inf, err = tc.MakeTx(util.AdminKey, nftAddr, "tokenByIndex", big.NewInt(i-1))
		if err != nil {
			t.Error(TAG, err, inf)
			return
		}

		str = inf[0].(string)
		nftid, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			t.Error(TAG, "no nft")
			return
		}

		inf, err = tc.MakeTx(util.AdminKey, nftAddr, "burn", nftid)
		if err != nil {
			t.Error(TAG, err, inf)
			return
		}
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	str = inf[0].(string)
	bi, ok = big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	} else if bi.Cmp(big.NewInt(0)) != 0 {
		t.Errorf(TAG, "not expect totalSupply want 0 get %v", bi.Int64())
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	str = inf[0].(string)
	adminBal, ok = big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "no bigint")
		return
	}
	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect totalSupply and admin balanceOf difference")
		return
	}

}

func TestApprove(t *testing.T) {
	TAG := "ApproveTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(1))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 1 {
		t.Error(TAG, "not match mint result")
		return
	}
	str := is[0].(string)
	tokenID, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id")
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "approve", util.Users[0], tokenID)
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "approve", util.Users[0], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenID)
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "transferFrom", util.Users[1], util.Users[0], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
}

func TestSetApprovalForAll(t *testing.T) {
	TAG := "SetApprovalForAllTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error(TAG, err)
		return
	}
	log.Println(TAG, nftAddr)

	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 10 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]*big.Int, 10)
	for i, h := range is {
		str := h.(string)
		var ok bool
		tokenIDs[i], ok = big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			panic(str + " is not bigInt")
		}
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "setApprovalForAll", util.Users[0], true)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[1])
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[2])
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[3])
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "setApprovalForAll", util.Users[0], false)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "transferFrom", util.Admin, util.Users[1], tokenIDs[4])
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)
}

func TestSafeTransferFrom(t *testing.T) {
	TAG := "SAFETRANSFERTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	util.RegisterContractClass(&nft721receiver.NFT721ReceiverContract{}, "NFT721Receiver")

	receiverAddr := tc.DeployContract(&nft721receiver.NFT721ReceiverContract{}, &nft721receiver.NFT721ReceiverContractConstruction{})

	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}
	_, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))
	if err != nil {
		t.Error(TAG, "not expact error", err)
		return
	}

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 10 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]*big.Int, 10)
	for i, h := range is {
		str := h.(string)
		var ok bool
		tokenIDs[i], ok = big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			panic(str + " is not bigInt")
		}
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "setApprovalForAll", util.Users[0], true)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "safeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "safeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "totalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "balanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "safeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err == nil {
		t.Error(TAG, err, inf)
		return
	} else {
		log.Println(TAG, err)
	}
}

func TestMintBatch(t *testing.T) {
	TAG := "MINTBATCH"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mintBatch", []common.Address{util.Users[0], util.Users[1], util.Users[2]})
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != 3 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]*big.Int, len(is))
	for i, h := range is {
		str := h.(string)
		var ok bool
		tokenIDs[i], ok = big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			panic(str + " is not bigInt")
		}
	}

	for i, tokenID := range tokenIDs {
		inf, err := tc.MakeTx(util.AdminKey, nftAddr, "ownerOf", tokenID)
		if err != nil {
			t.Error(TAG, err, inf)
		}
		addrStr := inf[0].(string)
		addr := common.HexToAddress(addrStr)
		if addr != util.Users[i] {
			t.Error(TAG, addr, "!=", util.Users[i])
		}
	}
}

func TestBurnAndMint(t *testing.T) {
	TAG := "BurnAndMintTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	var mintCount int64 = 1

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(mintCount))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != int(mintCount) {
		t.Error(TAG, "not match mint result")
		return
	}

	str := is[0].(string)
	bi, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id")
		return
	}
	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "burn", bi)
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(mintCount))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok = inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != int(mintCount) {
		t.Error(TAG, "not match mint result")
		return
	}
	// tc.ReadTx(util.AdminKey, nftAddr, "PrintContractData", util.Admin)

	str = is[0].(string)
	bi2, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
	if !ok {
		t.Error(TAG, "not nft id")
		return
	}

	if bi.String() == bi2.String() {
		t.Error("duplicate mint nft")
		return
	}

}

func TestBurnAll(t *testing.T) {
	TAG := "BurnAllTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr, err := deployNFT(tc, t, _name, _symbol)
	if err != nil {
		t.Error("Name", err)
		return
	}
	log.Println(TAG, nftAddr)

	var mintCount int64 = 10

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(mintCount))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
		return
	}
	if len(is) != int(mintCount) {
		t.Error(TAG, "not match mint result")
		return
	}

	p := &pError{
		t: t,
	}
	firstI := p.getInt(tc.ReadTx(util.AdminKey, nftAddr, "totalSupply"))

	for _, h := range is {
		str := h.(string)
		bi, ok := big.NewInt(0).SetString(strings.Replace(str, "0x", "", -1), 16)
		if !ok {
			t.Error(TAG, "not nft id")
			return
		}
		inf, err := tc.MakeTx(util.AdminKey, nftAddr, "burn", bi)
		if err != nil {
			t.Error(TAG, err, inf)
			return
		}
	}

	afterBurnI := p.getInt(tc.ReadTx(util.AdminKey, nftAddr, "totalSupply"))

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "mint", big.NewInt(mintCount))
	if err != nil {
		t.Error(TAG, err, inf)
		return
	}

	mintAgainI := p.getInt(tc.ReadTx(util.AdminKey, nftAddr, "totalSupply"))

	if firstI != int(mintCount) || mintAgainI != int(mintCount) {
		t.Error(TAG, "not matched mint count")
		return
	}
	if afterBurnI != 0 {
		t.Error(TAG, "burn not working")
		return
	}
}

type pError struct {
	t   *testing.T
	TAG string
}

func (p *pError) getInt(inf interface{}, err error) int {
	if err != nil {
		p.t.Error(p.TAG, err, inf)
		panic(err)
	}

	if i, err := strconv.Atoi(fmt.Sprintf("%v", inf)); err == nil {
		return i
	}

	is, ok := inf.([]interface{})
	if !ok {
		p.t.Error(p.TAG, err, inf)
		panic(err)
	}
	if len(is) == 0 {
		p.t.Error(p.TAG, "no result")
		panic("no result")
	}

	bi, ok := big.NewInt(0).SetString(strings.Replace(fmt.Sprintf("%v", is[0]), "0x", "", -1), 16)
	if !ok {
		p.t.Error(p.TAG, "no result")
		panic("no result")
	}
	return int(bi.Int64())
}
