package localapp

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
	adminAddress := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	// adminAddress := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	generators := []common.Address{
		common.HexToAddress("0x4dD2bf28E72EA48f83d9d3F398a03bF8baa8cC26"),
		common.HexToAddress("0x7483cD4E2bf98aEc39dD839a2779e993327337ef"),
	}
	supplyAddress := adminAddress

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
		bob := common.HexToAddress("0x70997970c51812dc3a010c7d01b50e0d17dc79c8")
		charlie := common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc")

		arg := &token.TokenContractConstruction{
			Name:   "MEVerse",
			Symbol: "MEV",
			InitialSupplyMap: map[common.Address]*amount.Amount{
				supplyAddress: amount.NewAmount(1900000000, 0),
				common.HexToAddress("0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"): amount.NewAmount(100000000, 0), // alice
				common.HexToAddress("0x70997970c51812dc3a010c7d01b50e0d17dc79c8"): amount.NewAmount(100000000, 0), // bob
				common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc"): amount.NewAmount(100000000, 0), // charlie
				common.HexToAddress("0x90f79bf6eb2c4f870365e785982e1f101e93b906"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x15d34aaf54267db7d7c367839aaf71a00a2c6a65"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x9965507d1a55bcc2695c58ba16fb37d819b0a4dc"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x976ea74026e726554db657fa54763abd0c3a0aa9"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x14dc79964da2c08b23698b3d3cc7ca32193d9955"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x23618e81e3f5cdf7f54c3d65f7fbc0abf5b21e8f"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xa0ee7a142d267c1f36714e4a8f75612f20a79720"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xbcd4042de499d14e55001ccbb24a551f3b954096"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x71be63f3384f5fb98995898a86b02fb2426c5788"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xfabb0ac9d68b0b445fb7357272ff202c5651694a"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x1cbd3b2770909d4e10f157cabc84c7264073c9ec"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xdf3e18d64bc6a983f673ab319ccae4f1a57c7097"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xcd3b766ccdd6ae721141f452c550ca635964ce71"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x2546bcd3c84621e976d8185a91a922ae77ecec30"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xbda5747bfd65f08deb54cb465eb87d40e51b197e"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0xdd2fd4581271e230360230f9337d5c0430bf44c0"): amount.NewAmount(100000000, 0),
				common.HexToAddress("0x8626f6940e2eb28930efb4cef49b2d1f2c9c1199"): amount.NewAmount(100000000, 0),

				common.HexToAddress("0x494a598d5653996a2d802c264AB82655938C3885"): amount.NewAmount(100000000, 0),
				// //key.NewMemoryKeyFromBytes(chainID, []byte{1, byte(i),0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
				// common.HexToAddress("0xDC5b20847F43d67928F49Cd4f85D696b5A7617B5"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xcca18d832a5C4fA1235e6c1cEa7E4645cca00395"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x678658741a8A61B92EF6B5700397a83C92729d60"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x632B713ABAA2cBC9ef18B61678c5d65027a4d2f0"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xf3dA6Ce653D680EBAcC26873d38F91aCf33C56Ac"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x574656094b39CaB109CD43C9F60D1586ee7217e4"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x77d5126D91BD80893Fd558d900a9F89c7DC43Da9"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x704497c99D99599471bA2e158f039fD6f30c153C"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x29c08d7604622Ab24C8f693a57b2D8c7fa1cac8F"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xbfA777cb56e3f99e63A3726f22e52a4Bda073D34"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xF1B624d18c29990035a093c01a671bdda6C65b64"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x47F825E9585814DF7Ca58E6A702355dbBBc51882"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x360CD740329d24cE90c40626d8fE7F9Aa213cD8a"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0x5facfa694633e9D87B09713285386557e7C591c6"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xe5Bbd1E8ef78D28a96a12194D28CDf99206AE4B2"): amount.NewAmount(100000000, 0),
				// common.HexToAddress("0xDe18f42A08D59E7eD5ba749EcDD2a2b6E224939F"): amount.NewAmount(100000000, 0),

				bob:     amount.NewAmount(100000000, 0),
				charlie: amount.NewAmount(100000000, 0),
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
		fmt.Println("Token", tokenAddress.String()) // 0xa1f093A1d8D4Ed5a7cC8fE29586266C5609a23e8
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

		fmt.Println("Gateway", gatewayAddress.String()) //0x820Deee3D4fE0d417eC512a8Dd0002E7bd55F351
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

		fmt.Println("formulatorAddress", formulatorAddress.String()) //0xBaa3C856fbA6FFAda189D6bD0a89d5ef7959c75E
	}

	types.SetLegacyCheckHeight(10)
	// chain.SetVersion(15, 2)
	return genesis.Top()
}

func RegisterContractClass() map[string]uint64 {
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
