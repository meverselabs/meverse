package test

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"

	"github.com/labstack/echo"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/contract/external/deployer"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {

}

func TestGoContractTx(t *testing.T) {
	tc := util.NewTestContext()

	e := echo.New()
	e.Static("/bin", "bin")

	url := "http://localhost:3000/bin/jsengin_debug.so"
	// dat, err := os.ReadFile("./engin/jsengin.so")

	ContArgs := &engin.EnginContractConstruction{}
	ContType := &engin.EnginContract{}
	util.RegisterContractClass(ContType, "EnginContract")
	util.RegisterContractClass(&deployer.DeployerContract{}, "DeployerContract")

	egAddr := tc.DeployContract(ContType, ContArgs)
	log.Println("engin Addr", egAddr)

	inf, err := tc.MakeTx(util.AdminKey, egAddr, "AddEngin", "JSContractEngin", "javascript vm on meverse verseion 0.1.0", url)
	log.Println(inf, err)
	if err != nil {
		t.Errorf("error not expect")
		return
	}
	inf, err = tc.ReadTx(util.AdminKey, egAddr, "EnginVersion", "JSContractEngin")
	log.Println("engin version", inf, err)
	if err != nil {
		t.Errorf("error not expect")
		return
	}

	bs, err := ioutil.ReadFile("./contract/Sample1.js")
	if err != nil {
		t.Error(err)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, egAddr, "DeploryContract", "JSContractEngin", "1", bs, []interface{}{
		util.Admin.String(),
	}, true)
	log.Println(inf, err)
	if err != nil {
		t.Errorf("error not expect")
		return
	}
	jsAddr, ok := inf[0].(common.Address)
	if !ok {
		t.Errorf("deplory contract not retruned address %v", inf)
		return
	}

	firstData := "data is Set"
	inf, err = tc.MakeTx(util.AdminKey, jsAddr, "SetData", firstData)
	if err != nil {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return
	}
	res, err := readData(tc, jsAddr, t, "1", "")
	if err != nil {
		t.Errorf("readData error result: %v, err: %v", inf, err)
		return
	}
	if res != firstData {
		t.Errorf("res is not matched SetData (%v, %v)", inf, firstData)
		return
	}

	bs, err = ioutil.ReadFile("./contract/Sample2.js")
	if err != nil {
		t.Error(err)
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, jsAddr, "Update", "JSContractEngin", "1", bs)
	log.Println(inf, err)
	if err != nil {
		t.Errorf("error not expect")
		return
	}

	res, err = readData(tc, jsAddr, t, "1", "")
	if err != nil {
		t.Errorf("readData error result: %v, err: %v", inf, err)
		return
	}
	if res != firstData {
		t.Errorf("readData not support ragacy func (%v, %v)", inf, firstData)
		return
	}

	secondData := "data is override"
	inf, err = tc.MakeTx(util.AdminKey, jsAddr, "SetData", secondData)
	if err != nil {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return
	}
	res, err = readData(tc, jsAddr, t, "2", "0")
	if err != nil {
		t.Errorf("readData error result: %v, err: %v", inf, err)
		return
	}
	if res != secondData {
		t.Errorf("res is not matched SetData (%v, %v)", inf, firstData)
		return
	}
}

func readData(tc *util.TestContext, jsAddr common.Address, t *testing.T, version string, index string) (string, error) {
	var inf interface{}
	var err error
	if version == "1" {
		inf, err = tc.ReadTx(util.AdminKey, jsAddr, "GetData")
	} else if version == "2" {
		inf, err = tc.ReadTx(util.AdminKey, jsAddr, "GetData", index)
	}
	if err != nil {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return "", err
	}
	if is, ok := inf.([]interface{}); !ok {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return "", errors.New("response is not array")
	} else if len(is) == 0 {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return "", errors.New("response is empty array")
	} else if resis, ok := is[0].([]interface{}); !ok {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return "", errors.New("response is not array")
	} else if len(resis) == 0 {
		t.Errorf("setdata error result: %v, err: %v", inf, err)
		return "", errors.New("response is empty array")
	} else {
		res, ok := resis[0].(string)
		if !ok {
			t.Errorf("getdata is not string (%v, err: %v)", resis[0], err)
			return "", errors.New("response is not string")
		}
		return res, nil
	}
}
