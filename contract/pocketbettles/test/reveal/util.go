package test

import (
	"errors"
	"log"
	"math/big"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/extern/test/util"
)

func initEngin(tc *util.TestContext) (egAddr common.Address) {
	url := "https://meverse-engin.s3.ap-northeast-2.amazonaws.com/jsengin_v2_debug.so"

	ContArgs := &engin.EnginContractConstruction{}
	ContType := &engin.EnginContract{}
	egAddr = tc.DeployContract(ContType, ContArgs)

	_, err := tc.SendTx(util.AdminKey, egAddr, "AddEngin", "JSContractEngin", "javascript vm on meverse verseion 0.1.0", url)
	if err != nil {
		panic(err)
	}
	_, err = tc.ReadTx(util.AdminKey, egAddr, "EnginVersion", "JSContractEngin")
	if err != nil {
		panic("error not expect " + err.Error())
	}
	return
}

func deployToken(tc *util.TestContext, _name string, _symbol string) common.Address {
	ContArgs := &token.TokenContractConstruction{
		Name:   _name,
		Symbol: _symbol,
	}
	ContType := &token.TokenContract{}
	return tc.DeployContract(ContType, ContArgs)
}

func initFormulator(tc *util.TestContext) common.Address {
	ContArgs := &formulator.FormulatorContractConstruction{
		TokenAddress: tc.MainToken,
		FormulatorPolicy: formulator.FormulatorPolicy{
			AlphaAmount:    amount.NewAmount(200000, 0),
			SigmaCount:     4,
			SigmaBlocks:    10,
			OmegaCount:     2,
			OmegaBlocks:    15,
			HyperAmount:    amount.NewAmount(3000000, 0),
			MinStakeAmount: amount.NewAmount(100, 0),
		},
		RewardPolicy: formulator.RewardPolicy{
			RewardPerBlock:        amount.MustParseAmount("0.6341958396752917"),
			AlphaEfficiency1000:   1000,
			SigmaEfficiency1000:   1150,
			OmegaEfficiency1000:   1300,
			HyperEfficiency1000:   1300,
			StakingEfficiency1000: 700,
			CommissionRatio1000:   50,
			MiningFeeAddress:      util.Admin,
			MiningFee1000:         300,
		},
	}
	ContType := &formulator.FormulatorContract{}

	formulatorAddress := tc.DeployContract(ContType, ContArgs)

	_, err := tc.SendTx(util.AdminKey, tc.MainToken, "SetMinter", formulatorAddress, true)
	if err != nil {
		panic(err)
	}
	return formulatorAddress
}

func checkBalanceOf(tc *util.TestContext, nftAddr common.Address, count int64) {
	bal := getBalanceOf(tc, nftAddr, util.Admin)
	if bal != count {
		log.Println("get", bal, "want", count)
		panic(errors.New("not expect nft count"))
	}
}

func getTotalSupply(tc *util.TestContext, nftAddr common.Address) int64 {
	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "totalSupply")
	if err != nil {
		panic(err)
	}
	bi, _ := big.NewInt(0).SetString(strings.ReplaceAll(inf[0].(string), "0x", ""), 16)
	return bi.Int64()
}

func getMintCount(tc *util.TestContext, nftAddr common.Address) int64 {
	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "mintCount")
	if err != nil {
		panic(err)
	}
	bi, _ := big.NewInt(0).SetString(strings.ReplaceAll(inf[0].(string), "0x", ""), 16)
	return bi.Int64()
}

func getBalanceOf(tc *util.TestContext, nftAddr, user common.Address) int64 {
	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "balanceOf", user)
	if err != nil {
		panic(err)
	}
	bi, _ := big.NewInt(0).SetString(strings.ReplaceAll(inf[0].(string), "0x", ""), 16)
	return bi.Int64()
}
func balanceOf(tc *util.TestContext, cont, user common.Address) *amount.Amount {
	inf, err := tc.ReadTx(util.AdminKey, cont, "balanceOf", user)
	if err != nil {
		panic(err)
	}
	return inf[0].(*amount.Amount)
}

func getNFTList(tc *util.TestContext, nftAddr common.Address, start, end int) []string {
	return getNFTListByOwner(tc, nftAddr, util.Admin, start, end)
}
func getNFTListByOwner(tc *util.TestContext, nftAddr, owner common.Address, start, end int) []string {
	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "tokenOfOwnerByRange", owner, start, end)
	if err != nil {
		panic(err)
	}
	if inf == nil {
		return nil
	}
	ss := []string{}
	infs := inf[0].([]interface{})
	for _, i := range infs {
		ss = append(ss, i.(string))
	}
	return ss
}
