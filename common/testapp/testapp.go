package testapp

import (
	"fmt"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/types"
)

func Genesis() *types.ContextData {
	adminAddress := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	generators := []common.Address{
		common.HexToAddress("0x4dD2bf28E72EA48f83d9d3F398a03bF8baa8cC26"),
		common.HexToAddress("0x7483cD4E2bf98aEc39dD839a2779e993327337ef"),
	}
	supplyAddress := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")

	alphaOwners := []common.Address{
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0xe4eD9b1b1E94E14771620263876Fc1027a0B6f4e"),
	}

	sigmaOwners := []common.Address{
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0xe4eD9b1b1E94E14771620263876Fc1027a0B6f4e"),
	}

	omegaOwners := []common.Address{
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x505d734566f2a5C7BC43e63d2D7448A8fd30482a"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0x1175e5Ad142037E3cb50477c7319C2c7Bd188626"),
		common.HexToAddress("0xe4eD9b1b1E94E14771620263876Fc1027a0B6f4e"),
	}

	stakingMap := map[common.Address]map[common.Address]*amount.Amount{
		common.HexToAddress("0x4dD2bf28E72EA48f83d9d3F398a03bF8baa8cC26"): {
			common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F"): amount.NewAmount(100, 0),
		},
	}

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

	genesis := types.NewEmptyContext()
	var tokenAddress common.Address
	var formulatorAddress common.Address
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
			Name:   "Test",
			Symbol: "TEST",
			InitialSupplyMap: map[common.Address]*amount.Amount{
				supplyAddress: amount.NewAmount(2000000000, 0),
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
		arg := &token.TokenContractConstruction{
			Name:   "Test2",
			Symbol: "TEST2",
			InitialSupplyMap: map[common.Address]*amount.Amount{
				supplyAddress: amount.NewAmount(2000000000, 0),
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
		tokenAddress := cont.Address()

		fmt.Println("Token2", tokenAddress.String())
	}

	if true {
		arg := &formulator.FormulatorContractConstruction{
			TokenAddress: tokenAddress,
			FormulatorPolicy: formulator.FormulatorPolicy{
				AlphaAmount:    amount.NewAmount(200000, 0),
				SigmaCount:     4,
				SigmaBlocks:    0,
				OmegaCount:     2,
				OmegaBlocks:    0,
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
		intr := types.NewInteractor(genesis, cont, cc, "000000000000", false)
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
