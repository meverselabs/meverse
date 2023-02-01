package test

import (
	"errors"
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
	inf, err := tc.ReadTx(util.AdminKey, nftAddr, "balanceOf", util.Admin)
	if err != nil {
		panic(err)
	}
	infs := inf[0].([]interface{})
	bi, _ := big.NewInt(0).SetString(strings.ReplaceAll(infs[0].(string), "0x", ""), 16)
	if bi.Int64() != count {
		panic(errors.New("not expect nft count"))
	}
}
