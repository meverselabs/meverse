package util

import (
	"fmt"
	"log"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/bridge"
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
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/types"
)

var (
	ChainID  = big.NewInt(1)
	frKeyMap = map[common.Address]key.Key{}
	obkeys   = []key.Key{}

	Admin       = common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	AdminKey, _ = key.NewMemoryKeyFromString(ChainID, "a000000000000000000000000000000000000000000000000000000000000999")
	Users       []common.Address
	UserKeys    []key.Key
	Obstrs      = [5]string{
		"c000000000000000000000000000000000000000000000000000000000000000",
		"c000000000000000000000000000000000000000000000000000000000000001",
		"c000000000000000000000000000000000000000000000000000000000000002",
		"c000000000000000000000000000000000000000000000000000000000000003",
		"c000000000000000000000000000000000000000000000000000000000000004",
	}
	ObserverKeys = []common.PublicKey{}

	Frstrs = [10]string{
		"b000000000000000000000000000000000000000000000000000000000000010",
		"b000000000000000000000000000000000000000000000000000000000000011",
		"b000000000000000000000000000000000000000000000000000000000000012",
		"b000000000000000000000000000000000000000000000000000000000000013",
		"b000000000000000000000000000000000000000000000000000000000000014",
		"b000000000000000000000000000000000000000000000000000000000000015",
		"b000000000000000000000000000000000000000000000000000000000000016",
		"b000000000000000000000000000000000000000000000000000000000000017",
		"b000000000000000000000000000000000000000000000000000000000000018",
		"b000000000000000000000000000000000000000000000000000000000000019",
	}
	frkeys = []key.Key{}
	// Ctx = types.NewEmptyContext()
	// Cn  *chain.Chain

)

var ClassMap map[string]uint64

func init() {
	err := RemoveTestData()
	if err != nil {
		panic(err)
	}

	ClassMap = map[string]uint64{}
	RegisterContractClass(&token.TokenContract{}, "Token")
	RegisterContractClass(&farm.FarmContract{}, "Farm")
	RegisterContractClass(&pool.PoolContract{}, "Pool")
	RegisterContractClass(&formulator.FormulatorContract{}, "Formulator")
	RegisterContractClass(&gateway.GatewayContract{}, "Gateway")
	RegisterContractClass(&factory.FactoryContract{}, "Factory")
	RegisterContractClass(&router.RouterContract{}, "Router")
	RegisterContractClass(&trade.UniSwap{}, "UniSwap")
	RegisterContractClass(&trade.StableSwap{}, "StableSwap")
	RegisterContractClass(&bridge.BridgeContract{}, "Bridge")
	RegisterContractClass(&whitelist.WhiteListContract{}, "WhiteList")
	RegisterContractClass(&imo.ImoContract{}, "IMO")
	RegisterContractClass(&engin.EnginContract{}, "EnginContract")
	RegisterContractClass(&deployer.DeployerContract{}, "DeployerContract")
	RegisterContractClass(&mappfarm.FarmContract{}, "MappFarm")

	for i := 0; i < 5; i++ {
		pk, err := key.NewMemoryKeyFromString(ChainID, Obstrs[i])
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pk.PublicKey())
		obkeys = append(obkeys, pk)
	}
	for i := 0; i < 10; i++ {
		pk, err := key.NewMemoryKeyFromString(ChainID, Frstrs[i])
		if err != nil {
			panic(err)
		}
		frkeys = append(frkeys, pk)
		frKeyMap[pk.PublicKey().Address()] = pk
	}

	UserKeys = []key.Key{}
	for i := 998; i > 988; i-- {
		pk := fmt.Sprintf("a000000000000000000000000000000000000000000000000000000000000%3v", i)
		log.Println(pk)
		k, err := key.NewMemoryKeyFromString(ChainID, pk)
		if err != nil {
			panic(err)
		}
		UserKeys = append(UserKeys, k)
	}
	Users = []common.Address{}
	for i, k := range UserKeys {
		addr := k.PublicKey().Address()
		Users = append(Users, addr)
		log.Println("userAddr", i, " : ", addr)
	}
}

func RegisterContractClass(cont types.Contract, className string) uint64 {
	ClassID, err := types.RegisterContractType(cont)
	if err != nil {
		panic(err)
	}
	ClassMap[className] = ClassID
	return ClassID
}
