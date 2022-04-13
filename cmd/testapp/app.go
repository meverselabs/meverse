package testapp

import (
	"fmt"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/fletaio/fleta_v2/common/bin"
	"github.com/fletaio/fleta_v2/contract/formulator"
	"github.com/fletaio/fleta_v2/contract/gateway"
	"github.com/fletaio/fleta_v2/contract/token"
	"github.com/fletaio/fleta_v2/core/types"
)

func Genesis() *types.ContextData {
	adminAddress := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	generators := []common.Address{
		common.HexToAddress("0x4dD2bf28E72EA48f83d9d3F398a03bF8baa8cC26"),
		common.HexToAddress("0x7483cD4E2bf98aEc39dD839a2779e993327337ef"),
		common.HexToAddress("0x8A782c8393eF1aF3BD3E92445C34e52BEcfCc84D"),
		common.HexToAddress("0x8592E0BDD7D50eFFceC1886f676424e1397b0293"),
		common.HexToAddress("0xB0523c0B1152Aa695937bA2E6817D6B512a02D95"),
		common.HexToAddress("0xa813D1B7a7718816559B171990E4fC359FD798F7"),
		common.HexToAddress("0xEd293303faD7b420D074285743C4b87D62d95189"),
		common.HexToAddress("0x19F8FCf57312a66eD1Bf2dccB9305475F169f412"),
		common.HexToAddress("0x618Ea511A7040FBdEa96E7Ab3B647c3C90391e82"),
		common.HexToAddress("0xfaa9E0070469B6Ba5D4918238D4BA87f0CF7Eb85"),
	}
	supplyAddress := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")

	alphaOwners := genesisAlphas()

	sigmaOwners := genesisSigmas()

	omegaOwners := genesisOmegas()

	stakingMap := map[common.Address]map[common.Address]*amount.Amount{}

	ClassMap := map[string]uint64{}
	if true {
		ClassID, err := types.RegisterContractType(&token.TokenContract{})
		if err != nil {
			panic(err)
		}
		ClassMap["Token"] = ClassID
	}
	if true {
		ClassID, err := types.RegisterContractType(&formulator.FormulatorContract{})
		if err != nil {
			panic(err)
		}
		ClassMap["Formulator"] = ClassID
	}
	if true {
		ClassID, err := types.RegisterContractType(&gateway.GatewayContract{})
		if err != nil {
			panic(err)
		}
		ClassMap["Gateway"] = ClassID
	}

	genesis := types.NewEmptyContext()
	var tokenAddress common.Address
	var formulatorAddress common.Address
	var gatewayAddress common.Address
	if true {
		if err := genesis.SetAdmin(adminAddress, true); err != nil {
			panic(err)
		}
		for _, v := range generators {
			if err := genesis.SetGenerator(v, true); err != nil {
				panic(err)
			}
		}
	}
	if true {
		arg := &token.TokenContractConstruction{
			Name:   "MEVerse",
			Symbol: "MEV",
			InitialSupplyMap: map[common.Address]*amount.Amount{
				supplyAddress: amount.NewAmount(1900000000, 0),
				common.HexToAddress("0x494a598d5653996a2d802c264AB82655938C3885"): amount.NewAmount(100000000, 0),
			},
		}
		bs, _, err := bin.WriterToBytes(arg)
		if err != nil {
			panic(err)
		}
		cont, err := genesis.DeployContract(adminAddress, ClassMap["Token"], bs)
		if err != nil {
			panic(err)
		}
		tokenAddress = cont.Address()
		genesis.SetMainToken(tokenAddress)

		fmt.Println("Token", tokenAddress.String())
	}
	if true {
		arg := &gateway.GatewayContractConstruction{
			TokenAddress: tokenAddress,
		}
		bs, _, err := bin.WriterToBytes(arg)
		if err != nil {
			panic(err)
		}
		v, err := genesis.DeployContract(adminAddress, ClassMap["Gateway"], bs)
		if err != nil {
			panic(err)
		}
		cont := v.(*gateway.GatewayContract)
		gatewayAddress = cont.Address()

		cc := genesis.ContractContext(cont, adminAddress)
		intr := types.NewInteractor(genesis, cont, cc)
		cc.Exec = intr.Exec

		PlatformList := map[string]*amount.Amount{
			"MEVerse":   amount.MustParseAmount("0.1"),
			"ETH":       amount.MustParseAmount("300"),
			"BSC":       amount.MustParseAmount("20"),
			"Klaytn":    amount.MustParseAmount("5"),
			"Polygon":   amount.MustParseAmount("5"),
			"Tomochain": amount.MustParseAmount("50"),
		}
		for platform, am := range PlatformList {
			if err := cont.AddPlatform(cc, platform, am); err != nil {
				panic(err)
			}
		}
		intr.Distroy()

		fmt.Println("Gateway", gatewayAddress.String())
	}

	if true {
		arg := &formulator.FormulatorContractConstruction{
			TokenAddress: tokenAddress,
			FormulatorPolicy: formulator.FormulatorPolicy{
				AlphaAmount:    amount.NewAmount(200000, 0),
				SigmaCount:     4,
				SigmaBlocks:    200,
				OmegaCount:     2,
				OmegaBlocks:    300,
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
				MiningFeeAddress:      adminAddress,
				MiningFee1000:         300,
			},
		}
		bs, _, err := bin.WriterToBytes(arg)
		if err != nil {
			panic(err)
		}
		v, err := genesis.DeployContract(adminAddress, ClassMap["Formulator"], bs)
		if err != nil {
			panic(err)
		}
		cont := v.(*formulator.FormulatorContract)
		formulatorAddress = cont.Address()

		if true {
			v, err := genesis.Contract(tokenAddress)
			if err != nil {
				panic(err)
			}
			cont := v.(*token.TokenContract)
			cc := genesis.ContractContext(cont, adminAddress)
			if err := cont.SetMinter(cc, formulatorAddress, true); err != nil {
				panic(err)
			}
		}

		cc := genesis.ContractContext(cont, cont.Address())
		intr := types.NewInteractor(genesis, cont, cc)
		cc.Exec = intr.Exec
		for _, addr := range alphaOwners {
			if _, err := cont.CreateGenesisAlpha(cc, addr); err != nil {
				panic(err)
			}
		}
		for _, addr := range sigmaOwners {
			if _, err := cont.CreateGenesisSigma(cc, addr); err != nil {
				panic(err)
			}
		}
		for _, addr := range omegaOwners {
			if _, err := cont.CreateGenesisOmega(cc, addr); err != nil {
				panic(err)
			}
		}
		for hyper, mp := range stakingMap {
			for addr, am := range mp {
				if err := cont.AddGenesisStakingAmount(cc, hyper, addr, am); err != nil {
					panic(err)
				}
			}
		}
		intr.Distroy()

		fmt.Println("formulatorAddress", formulatorAddress.String())
	}
	return genesis.Top()
}
