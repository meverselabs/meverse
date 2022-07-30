package test

import (
	"log"
	"math/big"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/nft721"
	"github.com/meverselabs/meverse/contract/nft721/nft721receiver"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {
}

const (
	None uint8 = 1 + iota
)

func TestName(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "TESTNAME"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
	})
	log.Println(TAG, nftAddr)

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Name")
	if err != nil {
		t.Error("Name", err, inf)
	}
	name, ok := inf[0].(string)
	if !ok {
		t.Error("Name", "name not returns string")
	}
	if name != _name {
		t.Error("Name", "name not matchd init name")
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Symbol")
	if err != nil {
		t.Error("Symbol", err, inf)
	}
	symbol, ok := inf[0].(string)
	if !ok {
		t.Error("Symbol", "symbol not returns string")
	}
	if symbol != _symbol {
		t.Error("Symbol", "symbol not matchd init symbol")
	}
}

func _testMint3NFT(t *testing.T, i int) (nftAddr common.Address, tc *util.TestContext) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "MINTTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc = util.NewTestContext()
	nftAddr = tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	inf, err := tc.MakeTx(util.UserKeys[0], nftAddr, "Mint", big.NewInt(1))
	if err == nil {
		t.Error(TAG, "minted not owner", err, inf)
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(1))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 1 {
		t.Error(TAG, "not match mint result")
		return
	}
	hs, ok := is[0].(hash.Hash256)
	if !ok {
		t.Error(TAG, "not nft id")
	}
	checkMap := map[hash.Hash256]bool{}
	checkMap[hs] = true

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(2))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok = inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 2 {
		t.Error(TAG, "not match mint result")
		return
	}
	for _, i := range is {
		hs, ok = i.(hash.Hash256)
		if !ok {
			t.Error(TAG, "not nft id")
		}
		checkMap[hs] = true
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
	}
	bi, ok := inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	} else if bi.Cmp(big.NewInt(int64(len(checkMap)))) != 0 {
		t.Errorf(TAG, "not expect TotalSupply want 3 get %v", bi.Int64())
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
	}
	adminBal, ok := inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	}

	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect TotalSupply and admin balanceOf difference")
	}

	return
}

func TestMint(t *testing.T) {
	_testMint3NFT(t, 1)
}

func TestBurn(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "BURNTEST"

	nftAddr, tc := _testMint3NFT(t, 2)

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "TokenByIndex", big.NewInt(0))
	if err != nil {
		t.Error(TAG, err, inf)
	}

	nftID, ok := inf[0].(hash.Hash256)
	if !ok {
		t.Error(TAG, "no nft")
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Burn", nftID)
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "TokenByIndex", big.NewInt(0))
	if err != nil {
		t.Error(TAG, err, inf)
	}

	nftID2, ok := inf[0].(hash.Hash256)
	if !ok {
		t.Error(TAG, "no nft")
	}

	if nftID == nftID2 {
		t.Error(TAG, "expect different nftid but same")
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
	}
	bi, ok := inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	} else if bi.Cmp(big.NewInt(2)) != 0 {
		t.Errorf(TAG, "not expect TotalSupply want 2 get %v", bi.Int64())
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
	}
	adminBal, ok := inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	}

	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect TotalSupply and admin balanceOf difference")
	}

	for i := bi.Int64(); i > 0; i-- {
		inf, err = tc.MakeTx(util.AdminKey, nftAddr, "TokenByIndex", big.NewInt(i-1))
		if err != nil {
			t.Error(TAG, err, inf)
		}

		nftID, ok = inf[0].(hash.Hash256)
		if !ok {
			t.Error(TAG, "no nft")
		}

		inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Burn", nftID)
		if err != nil {
			t.Error(TAG, err, inf)
		}
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	if err != nil {
		t.Error(TAG, err, inf)
	}
	bi, ok = inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	} else if bi.Cmp(big.NewInt(0)) != 0 {
		t.Errorf(TAG, "not expect TotalSupply want 0 get %v", bi.Int64())
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
	}
	adminBal, ok = inf[0].(*big.Int)
	if !ok {
		t.Error(TAG, "no bigint")
	}
	if adminBal.Cmp(bi) != 0 {
		t.Errorf(TAG, "not expect TotalSupply and admin balanceOf difference")
	}

}

func TestApprove(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "MINTTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(1))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 1 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenID, ok := is[0].(hash.Hash256)
	if !ok {
		t.Error(TAG, "not nft id")
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "Approve", util.Users[0], tokenID)
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Approve", util.Users[0], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenID)
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
	}
	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "TransferFrom", util.Users[1], util.Users[0], tokenID)
	if err != nil {
		t.Error(TAG, err, inf)
	}
}

func TestSetApprovalForAll(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "MINTTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 10 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]hash.Hash256, 10)
	for i, h := range is {
		tokenIDs[i] = h.(hash.Hash256)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "SetApprovalForAll", util.Users[0], true)
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[0])
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[1])
	if err != nil {
		t.Error(TAG, err, inf)
	}
	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[2])
	if err != nil {
		t.Error(TAG, err, inf)
	}
	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[3])
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "SetApprovalForAll", util.Users[0], false)
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "TransferFrom", util.Admin, util.Users[1], tokenIDs[4])
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)
}

func TestSafeTransferFrom(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "SAFETRANSFERTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	util.RegisterContractClass(&nft721receiver.NFT721ReceiverContract{}, "NFT721Receiver")

	receiverAddr := tc.DeployContract(&nft721receiver.NFT721ReceiverContract{}, &nft721receiver.NFT721ReceiverContractConstruction{})

	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(10, 0))

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 10 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]hash.Hash256, 10)
	for i, h := range is {
		tokenIDs[i] = h.(hash.Hash256)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ := tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "SetApprovalForAll", util.Users[0], true)
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.MakeTx(util.UserKeys[1], nftAddr, "SafeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "SafeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "TotalSupply")
	Admininf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Admin)
	Users0inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[0])
	Users1inf, _ = tc.MakeTx(util.AdminKey, nftAddr, "BalanceOf", util.Users[1])
	log.Println(TAG, inf, Admininf, Users0inf, Users1inf)

	inf, err = tc.MakeTx(util.UserKeys[0], nftAddr, "SafeTransferFrom", util.Admin, receiverAddr, tokenIDs[0], []byte{0})
	if err == nil {
		t.Error(TAG, err, inf)
	} else {
		log.Println(TAG, err)
	}
}

func TestMintBatch(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "MINTBATCH"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "MintBatch", []common.Address{util.Users[0], util.Users[1], util.Users[2]})
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 3 {
		t.Error(TAG, "not match mint result")
		return
	}
	tokenIDs := make([]hash.Hash256, len(is))
	for i, h := range is {
		tokenIDs[i] = h.(hash.Hash256)
	}

	for i, tokenID := range tokenIDs {
		inf, err := tc.MakeTx(util.AdminKey, nftAddr, "OwnerOf", tokenID)
		if err != nil {
			t.Error(TAG, err, inf)
		}
		addr := inf[0].(common.Address)
		if addr != util.Users[i] {
			t.Error(TAG, addr, "!=", util.Users[i])
		}
	}
}

func TestBurnAll(t *testing.T) {
	util.RegisterContractClass(&nft721.NFT721Contract{}, "NFT721")
	TAG := "MINTTEST"

	_name := "TestNFT"
	_symbol := "TNFT"

	tc := util.NewTestContext()
	nftAddr := tc.DeployContract(&nft721.NFT721Contract{}, &nft721.NFT721ContractConstruction{
		Name:   _name,
		Symbol: _symbol,
		Owner:  util.Admin,
	})

	inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
	}
	is, ok := inf[0].([]interface{})
	if !ok {
		t.Error(TAG, "no result")
	}
	if len(is) != 10 {
		t.Error(TAG, "not match mint result")
		return
	}
	_, ok = is[0].(hash.Hash256)
	if !ok {
		t.Error(TAG, "not nft id")
	}

	for i, h := range is {
		hs := h.(hash.Hash256)
		if !ok {
			t.Error(TAG, "not nft id")
		}
		inf, err := tc.MakeTx(util.AdminKey, nftAddr, "Burn", hs)
		if err != nil {
			t.Error(TAG, err, inf)
		}
		inf, err = tc.ReadTx(util.AdminKey, nftAddr, "PrintContractData", util.Admin)
		if err != nil {
			t.Error(TAG, err, inf)
		}
		log.Println("***-******", i)
	}
	inf, err = tc.MakeTx(util.AdminKey, nftAddr, "Mint", big.NewInt(10))
	if err != nil {
		t.Error(TAG, err, inf)
	}

	inf, err = tc.ReadTx(util.AdminKey, nftAddr, "PrintContractData", util.Admin)
	if err != nil {
		t.Error(TAG, err, inf)
	}

}
