package main // import "github.com/fletaio/fleta"

import (
	"encoding/hex"
	"strconv"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
)

func main() {
	if err := test(); err != nil {
		panic(err)
	}
}

func test() error {
	//os.RemoveAll("./_data")

	obstrs := []string{
		"cd7cca6359869f4f58bb31aa11c2c4825d4621406f7b514058bc4dbe788c29be",
		"d8744df1e76a7b76f276656c48b68f1d40804f86518524d664b676674fccdd8a",
		"387b430fab25c03313a7e987385c81f4b027199304e2381561c9707847ec932d",
		"a99fa08114f41eb7e0a261cf11efdc60887c1d113ea6602aaf19eca5c3f5c720",
		"a9878ff3837700079fbf187c86ad22f1c123543a96cd11c53b70fedc3813c27b",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	ObserverKeys := make([]common.PublicHash, 0, len(obstrs))
	NetAddressMap := map[common.PublicHash]string{}
	FrNetAddressMap := map[common.PublicHash]string{}
	for i, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			pubhash := common.NewPublicHash(Key.PublicKey())
			ObserverKeys = append(ObserverKeys, pubhash)
			NetAddressMap[pubhash] = "localhost:1390" + strconv.Itoa(i+1)
			FrNetAddressMap[pubhash] = "localhost:1490" + strconv.Itoa(i+1)
		}
	}

	frstrs := []string{
		"67066852dd6586fa8b473452a66c43f3ce17bd4ec409f1fff036a617bb38f063",
		"a9878ff3837700079fbf187c86ad22f1c123543a96cd11c53b70fedc3813c27b",
	}

	frkeys := make([]key.Key, 0, len(frstrs))
	for _, v := range frstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			frkeys = append(frkeys, Key)
		}
	}

	MaxBlocksPerFormulator := uint32(10)

	obs := []*pof.ObserverNode{}
	for i, obkey := range obkeys {
		st, err := chain.NewStore("./_data/o"+strconv.Itoa(i+1), "FLEAT Mainnet", 0x0001, true)
		if err != nil {
			return err
		}
		cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
		app := &DApp{
			frkey:        frkeys[0],
			frkey2:       frkeys[1],
			adminAddress: common.NewAddress(0, 1, 0),
		}
		cn := chain.NewChain(cs, app, st)
		cn.MustAddProcess(vault.NewVault(1))
		cn.MustAddProcess(formulator.NewFormulator(2, app.adminAddress))
		if err := cn.Init(); err != nil {
			return err
		}
		ob := pof.NewObserverNode(obkey, NetAddressMap, cs, ObserverKeys[i])
		if err := ob.Init(); err != nil {
			return err
		}
		obs = append(obs, ob)
	}

	for i, ob := range obs {
		go ob.Run(
			"localhost:1390"+strconv.Itoa(i+1),
			"localhost:1490"+strconv.Itoa(i+1),
		)
	}

	frs := []*pof.FormulatorNode{}
	for i, frkey := range frkeys {
		st, err := chain.NewStore("./_data/f"+strconv.Itoa(i+1), "FLEAT Mainnet", 0x0001, true)
		if err != nil {
			return err
		}
		cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
		app := &DApp{
			frkey:        frkeys[0],
			frkey2:       frkeys[1],
			adminAddress: common.NewAddress(0, 1, 0),
		}
		cn := chain.NewChain(cs, app, st)
		cn.MustAddProcess(vault.NewVault(1))
		cn.MustAddProcess(formulator.NewFormulator(2, app.adminAddress))
		if err := cn.Init(); err != nil {
			return err
		}
		fr := pof.NewFormulator(&pof.FormulatorConfig{
			SeedNodes:  []string{},
			Formulator: common.NewAddress(0, 2+uint16(i), 0),
		}, frkey, FrNetAddressMap, cs)
		if err := fr.Init(); err != nil {
			return err
		}
		frs = append(frs, fr)
	}

	for _, fr := range frs {
		go fr.Run()
	}
	select {}
	return nil
}

// DApp is app
type DApp struct {
	sync.Mutex
	*types.ApplicationBase
	pm           types.ProcessManager
	cn           types.Provider
	frkey        key.Key
	frkey2       key.Key
	adminAddress common.Address
}

// Name returns the name of the application
func (app *DApp) Name() string {
	return "Test DApp"
}

// Version returns the version of the application
func (app *DApp) Version() string {
	return "v0.0.1"
}

// Init initializes the consensus
func (app *DApp) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	app.pm = pm
	app.cn = cn
	return nil
}

// InitGenesis initializes genesis data
func (app *DApp) InitGenesis(ctw *types.ContextWrapper) error {
	app.Lock()
	defer app.Unlock()

	if p, err := app.pm.ProcessByName("fleta.formulator"); err != nil {
		return err
	} else if fp, is := p.(*formulator.Formulator); !is {
		return types.ErrNotExistProcess
	} else {
		if err := fp.InitPolicy(ctw, &formulator.RewardPolicy{
			RewardPerBlock:        amount.NewCoinAmount(0, 500000000000000000),
			PayRewardEveryBlocks:  500,
			AlphaEfficiency1000:   1000,
			SigmaEfficiency1000:   1500,
			OmegaEfficiency1000:   2000,
			HyperEfficiency1000:   2500,
			StakingEfficiency1000: 500,
		}, &formulator.AlphaPolicy{
			AlphaCreationLimitHeight:  1000,
			AlphaCreationAmount:       amount.NewCoinAmount(1000, 0),
			AlphaUnlockRequiredBlocks: 1000,
		}, &formulator.SigmaPolicy{
			SigmaRequiredAlphaBlocks:  1000,
			SigmaRequiredAlphaCount:   4,
			SigmaUnlockRequiredBlocks: 1000,
		}, &formulator.OmegaPolicy{
			OmegaRequiredSigmaBlocks:  1000,
			OmegaRequiredSigmaCount:   2,
			OmegaUnlockRequiredBlocks: 1000,
		}, &formulator.HyperPolicy{
			HyperCreationAmount:         amount.NewCoinAmount(1000, 0),
			HyperUnlockRequiredBlocks:   1000,
			StakingUnlockRequiredBlocks: 1000,
		}); err != nil {
			return err
		}
	}
	if p, err := app.pm.ProcessByName("fleta.vault"); err != nil {
		return err
	} else if sp, is := p.(*vault.Vault); !is {
		return types.ErrNotExistProcess
	} else {
		acc := &vault.SingleAccount{
			Address_: common.NewAddress(0, 1, 0),
			Name_:    "admin",
			KeyHash:  common.NewPublicHash(app.frkey.PublicKey()),
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
		sp.AddBalance(ctw, acc.Address(), amount.NewCoinAmount(100000000000, 0))
	}
	if true {
		acc := &formulator.FormulatorAccount{
			Address_:       common.NewAddress(0, 2, 0),
			Name_:          "fleta.001",
			FormulatorType: formulator.AlphaFormulatorType,
			KeyHash:        common.NewPublicHash(app.frkey.PublicKey()),
			Amount:         amount.NewCoinAmount(1000, 0),
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
	}
	if true {
		acc := &formulator.FormulatorAccount{
			Address_:       common.NewAddress(0, 3, 0),
			Name_:          "fleta.002",
			FormulatorType: formulator.AlphaFormulatorType,
			KeyHash:        common.NewPublicHash(app.frkey2.PublicKey()),
			Amount:         amount.NewCoinAmount(1000, 0),
		}
		if err := ctw.CreateAccount(acc); err != nil {
			return err
		}
	}
	return nil
}
