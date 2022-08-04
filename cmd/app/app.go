package app

import (
	"fmt"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/contract/bridge"
	"github.com/meverselabs/meverse/contract/connect/depositpool"
	"github.com/meverselabs/meverse/contract/connect/farm"
	"github.com/meverselabs/meverse/contract/connect/imo"
	"github.com/meverselabs/meverse/contract/connect/mappfarm"
	"github.com/meverselabs/meverse/contract/connect/pool"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/external/deployer"
	"github.com/meverselabs/meverse/contract/external/engin"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/gateway"
	"github.com/meverselabs/meverse/contract/nft721"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/types"
)

func Genesis() *types.ContextData {
	adminAddress := common.HexToAddress("0x24D5da1B198c5049016c513099916498ceE9ccf5")
	generators := []common.Address{
		common.HexToAddress("0x7abD922630E41765d6674784C65D794015c1676B"),
		common.HexToAddress("0xDDB9c2B2198daA32f70A4E50B817C0f4bb0BA6b2"),
		common.HexToAddress("0x3E229C2165f3ce8fA3712a779c366603753Fcd1C"),
		common.HexToAddress("0x8cC1b16aAd2d9baAefd03110b5E76B89E66CF8cC"),
		common.HexToAddress("0x6aDed46ff1dfb0263C2b8c8ec6618Ebec19682CD"),
		common.HexToAddress("0x0CBb5713E066c2c432A588Fb1C37BB907E99cd7B"),
		common.HexToAddress("0x81779750Fc6EdCbeBd7c0CCE1F4DEAabb8774A7e"),
		common.HexToAddress("0x60F53D40C32ec5f09A5e12f4F2e6d464bE9F91ba"),
		common.HexToAddress("0x3c6E849cA24b0c8A62697e1870Eb74F5190735b5"),
		common.HexToAddress("0x36104d32d97DE4e11E3590A1c0bDfb715521F256"),
	}
	supplyAddress := common.HexToAddress("0x24D5da1B198c5049016c513099916498ceE9ccf5")

	alphaOwners := genesisAlphas()
	sigmaOwners := genesisSigmas()
	omegaOwners := genesisOmegas()

	stakingMap := map[common.Address]map[common.Address]*amount.Amount{}

	ClassMap := RegisterContractClass()

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
			Name:             "MEVerse",
			Symbol:           "MEV",
			InitialSupplyMap: initialSupply(supplyAddress),
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
		// Token 0xeF3432F1D54eC559613f44879FEF8a7866dA3e07
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
		intr := types.NewInteractor(genesis, cont, cc, "000000000000", false)
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

		fmt.Println("Gateway", gatewayAddress.String())
		// Gateway 0x0eA20dAcf567eB840f1B2463F366650086fb995c
	}

	if true {
		arg := &formulator.FormulatorContractConstruction{
			TokenAddress: tokenAddress,
			FormulatorPolicy: formulator.FormulatorPolicy{
				AlphaAmount:    amount.NewAmount(200000, 0),
				SigmaCount:     4,
				SigmaBlocks:    5184000,
				OmegaCount:     2,
				OmegaBlocks:    5184000,
				HyperAmount:    amount.NewAmount(0, 0),
				MinStakeAmount: amount.NewAmount(100, 0),
			},
			RewardPolicy: formulator.RewardPolicy{
				RewardPerBlock:        amount.MustParseAmount("0.4756468797564688"),
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

		fmt.Println("formulatorAddress", formulatorAddress.String())
		// formulatorAddress 0x75A098f86bAe71039217a879f064d034c59C3766
	}
	return genesis.Top()
}

func RegisterContractClass() map[string]uint64 {
	types.SetLegacyCheckHeight(25298976)
	ClassMap := map[string]uint64{}
	registerContractClass(&token.TokenContract{}, "Token", ClassMap)
	registerContractClass(&formulator.FormulatorContract{}, "Formulator", ClassMap)
	registerContractClass(&gateway.GatewayContract{}, "Gateway", ClassMap)

	registerContractClass(&factory.FactoryContract{}, "Factory", ClassMap)
	registerContractClass(&router.RouterContract{}, "Router", ClassMap)
	registerContractClass(&trade.UniSwap{}, "UniSwap", ClassMap)
	registerContractClass(&trade.StableSwap{}, "StableSwap", ClassMap)

	registerContractClass(&bridge.BridgeContract{}, "Bridge", ClassMap)

	registerContractClass(&farm.FarmContract{}, "ConnectFarm", ClassMap)
	registerContractClass(&pool.PoolContract{}, "ConnectPool", ClassMap)
	registerContractClass(&whitelist.WhiteListContract{}, "WhiteList", ClassMap)
	registerContractClass(&imo.ImoContract{}, "IMO", ClassMap)

	registerContractClass(&depositpool.DepositPoolContract{}, "DepositUSDT", ClassMap)

	registerContractClass(&nft721.NFT721Contract{}, "NFT721", ClassMap)

	registerContractClass(&engin.EnginContract{}, "Engin", ClassMap)
	registerContractClass(&deployer.DeployerContract{}, "EnginDeployer", ClassMap)

	registerContractClass(&mappfarm.FarmContract{}, "MappFarm", ClassMap)
	return ClassMap
}
func registerContractClass(cont types.Contract, className string, ClassMap map[string]uint64) {
	ClassID, err := types.RegisterContractType(cont)
	if err != nil {
		panic(err)
	}
	ClassMap[className] = ClassID
}
