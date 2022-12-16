package test

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/extern/test/util"
)

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

func TestDismantle(t *testing.T) {
	tc := util.NewTestContext()
	haddr, revealAddr, ss := makeRevealCont(tc, t)
	_, err := tc.SendTx(util.AdminKey, revealAddr, "dismantle", []interface{}{ss[0].(string)})
	if err != nil {
		t.Error(err)
		return
	}

	inf, _ := tc.ReadTx(util.AdminKey, revealAddr, "dismantleList", util.Admin)
	log.Println(inf[0])

	_, err = tc.SendTx(util.AdminKey, revealAddr, "dismantle", []interface{}{ss[1].(string), ss[2].(string), ss[3].(string)})
	if err != nil {
		t.Error(err)
		return
	}

	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "dismantleList", util.Admin)
	log.Println(inf[0])

	_, err = tc.SendTx(util.AdminKey, haddr, "transferFrom", util.Admin, util.Users[0], ss[4].(string))
	if err != nil {
		t.Error(err)
		return
	}

	_, err = tc.SendTx(util.AdminKey, revealAddr, "dismantle", []interface{}{ss[4].(string)})
	if err == nil {
		t.Error("expected not nft owner")
		return
	}

	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "dismantleList", util.Admin)
	log.Println(inf[0])

	_, err = tc.SendTx(util.AdminKey, revealAddr, "invalidDismantle", util.Admin, ss[3].(string))
	if err != nil {
		t.Error("expected not nft owner")
		return
	}

	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "dismantleList", util.Admin)
	log.Println(inf)
}

func TestMarge(t *testing.T) {
	tc := util.NewTestContext()
	_, revealAddr, ss := makeRevealCont(tc, t)
	_, err := tc.SendTx(util.AdminKey, revealAddr, "merge", ss[0], "1", ss[1], "1")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = tc.SendTx(util.AdminKey, revealAddr, "merge", ss[1], "1", ss[2], "1")
	if err == nil {
		t.Error("not owner")
		return
	}
	inf, err := tc.ReadTx(util.AdminKey, revealAddr, "slaveList", ss[0])
	if err != nil {
		t.Error(err)
		return
	}
	sl := inf[0].([]interface{})
	for _, slave := range sl {
		if slave != strings.Replace(ss[1].(string), "0x", "", -1) {
			t.Errorf("not slave %v %v %v", slave, ss[1], strings.Replace(ss[1].(string), "0x", "", -1))
			return
		}
	}
	_, err = tc.SendTx(util.AdminKey, revealAddr, "merge", ss[0], "2", ss[2], "2")
	if err != nil {
		t.Error(err)
		return
	}
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveList")
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveToMasterList", ss[0])
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveInfo", ss[1])
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveInfo", ss[2])
	log.Println(inf[0])

	_, err = tc.SendTx(util.AdminKey, revealAddr, "invalidMerge", ss[0], ss[2])
	if err != nil {
		t.Error(err)
		return
	}

	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveList")
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveToMasterList", ss[0])
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveInfo", ss[1])
	log.Println(inf[0])
	inf, _ = tc.ReadTx(util.AdminKey, revealAddr, "slaveInfo", ss[2])
	log.Println(inf[0])

	_, err = tc.SendTx(util.AdminKey, revealAddr, "merge", ss[0], "2", ss[2], "2")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = tc.SendTx(util.AdminKey, revealAddr, "merge", ss[0], "2", ss[3], "2")
	if err == nil {
		t.Error("expected invalid nft lv")
		return
	}

	_, err = tc.SendTx(util.AdminKey, revealAddr, "merge", ss[0], "3", ss[3], "3")
	if err != nil {
		t.Error(err)
		return
	}

}

func makeRevealCont(tc *util.TestContext, t *testing.T) (common.Address, common.Address, []interface{}) {
	egAddr := initEngin(tc)

	bAddr, err := makeNft(tc, egAddr)
	if err != nil {
		panic(err)
	}
	hAddr, err := makeNft(tc, egAddr)
	if err != nil {
		panic(err)
	}

	IDs := []string{"0x01", "0x02", "0x03", "0x04", "0x05"}
	for _, id := range IDs {
		tc.MustSendTx(util.AdminKey, bAddr, "mintWithID", id)
	}

	bs, err := ioutil.ReadFile("../../reveal.js")
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
	ids, err := tc.SendTx(util.AdminKey, revealAddr, "reveal", IDs)
	if err != nil {
		panic(err)
	}
	ss := ids[0].([]interface{})
	return hAddr, revealAddr, ss
}
