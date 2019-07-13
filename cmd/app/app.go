package app

import (
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
)

// FletaApp is app
type FletaApp struct {
	sync.Mutex
	*types.ApplicationBase
	pm           types.ProcessManager
	cn           types.Provider
	adminAddress common.Address
}

// NewApp returns a FletaApp
func NewFletaApp() *FletaApp {
	return &FletaApp{}
}

// Name returns the name of the application
func (app *FletaApp) Name() string {
	return "FletaApp"
}

// Version returns the version of the application
func (app *FletaApp) Version() string {
	return "v0.0.1"
}

// AdminAddress returns the admin address of the application
func (app *FletaApp) AdminAddress() common.Address {
	return app.adminAddress
}

// Init initializes the consensus
func (app *FletaApp) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	app.pm = pm
	app.cn = cn
	app.adminAddress = common.NewAddress(0, 37, 0)
	return nil
}

// InitGenesis initializes genesis data
func (app *FletaApp) InitGenesis(ctw *types.ContextWrapper) error {
	app.Lock()
	defer app.Unlock()

	rewardPolicy := &formulator.RewardPolicy{
		RewardPerBlock:        amount.NewCoinAmount(0, 500000000000000000),
		PayRewardEveryBlocks:  50,
		AlphaEfficiency1000:   1000,
		SigmaEfficiency1000:   1500,
		OmegaEfficiency1000:   2000,
		HyperEfficiency1000:   2500,
		StakingEfficiency1000: 500,
	}
	alphaPolicy := &formulator.AlphaPolicy{
		AlphaCreationLimitHeight:  1000,
		AlphaCreationAmount:       amount.NewCoinAmount(1000, 0),
		AlphaUnlockRequiredBlocks: 1000,
	}
	sigmaPolicy := &formulator.SigmaPolicy{
		SigmaRequiredAlphaBlocks:  1000,
		SigmaRequiredAlphaCount:   4,
		SigmaUnlockRequiredBlocks: 1000,
	}
	omegaPolicy := &formulator.OmegaPolicy{
		OmegaRequiredSigmaBlocks:  1000,
		OmegaRequiredSigmaCount:   2,
		OmegaUnlockRequiredBlocks: 1000,
	}
	hyperPolicy := &formulator.HyperPolicy{
		HyperCreationAmount:         amount.NewCoinAmount(1000, 0),
		HyperUnlockRequiredBlocks:   1000,
		StakingUnlockRequiredBlocks: 1000,
	}

	if p, err := app.pm.ProcessByName("fleta.formulator"); err != nil {
		return err
	} else if fp, is := p.(*formulator.Formulator); !is {
		return types.ErrNotExistProcess
	} else {
		if err := fp.InitPolicy(ctw,
			rewardPolicy,
			alphaPolicy,
			sigmaPolicy,
			omegaPolicy,
			hyperPolicy,
		); err != nil {
			return err
		}
	}
	if p, err := app.pm.ProcessByName("fleta.vault"); err != nil {
		return err
	} else if sp, is := p.(*vault.Vault); !is {
		return types.ErrNotExistProcess
	} else {
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("gDGAcf9V9i8oWLTeayoKC8bdAooNVaFnAeQKy4CsUB"), common.MustParsePublicHash("gDGAcf9V9i8oWLTeayoKC8bdAooNVaFnAeQKy4CsUB"), common.MustParseAddress("3CUsUpvEK"), "fleta.io.fr00001")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4m6XsJbq6EFb5bqhZuKFc99SmF86ymcLcRPwrWyToHQ"), common.MustParsePublicHash("4m6XsJbq6EFb5bqhZuKFc99SmF86ymcLcRPwrWyToHQ"), common.MustParseAddress("5PxjxeqTd"), "fleta.io.fr00002")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("o1rVoXHFuz5EtwLwCLcrmHpqPdugAnWHEVVMtnCb32"), common.MustParsePublicHash("o1rVoXHFuz5EtwLwCLcrmHpqPdugAnWHEVVMtnCb32"), common.MustParseAddress("7bScSUkgw"), "fleta.io.fr00003")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("47NZ8oadY4dCAM3ZrGFrENPn99L1SLSqzpR4DFPUpk5"), common.MustParsePublicHash("47NZ8oadY4dCAM3ZrGFrENPn99L1SLSqzpR4DFPUpk5"), common.MustParseAddress("9nvUvJfvF"), "fleta.io.fr00004")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4TaHVFSzcrNPktRiNdpPitoUgLXtZzrVmkxE3GmcYjG"), common.MustParsePublicHash("4TaHVFSzcrNPktRiNdpPitoUgLXtZzrVmkxE3GmcYjG"), common.MustParseAddress("BzQMQ8b9Z"), "fleta.io.fr00005")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2wqsb4J47T4JkNUp1Bma1HkjpCyei7sZinLmNprpdtY"), common.MustParsePublicHash("2wqsb4J47T4JkNUp1Bma1HkjpCyei7sZinLmNprpdtY"), common.MustParseAddress("EBtDsxWNs"), "fleta.io.fr00006")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2a1CirwCHSYYpLqpbi1b7Rpr4BAJZvydbDA1bGjJ7FG"), common.MustParsePublicHash("2a1CirwCHSYYpLqpbi1b7Rpr4BAJZvydbDA1bGjJ7FG"), common.MustParseAddress("GPN6MnRcB"), "fleta.io.fr00007")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2KnMHH973ZLicENxcsJbARdeTUiYZmN3WnBzbZqvvEx"), common.MustParsePublicHash("2KnMHH973ZLicENxcsJbARdeTUiYZmN3WnBzbZqvvEx"), common.MustParseAddress("JaqxqcLqV"), "fleta.io.fr00008")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4fyTmraz8x3NKWnj4nWgPWKy8qCBF1hyqVJQeyupHAe"), common.MustParsePublicHash("4fyTmraz8x3NKWnj4nWgPWKy8qCBF1hyqVJQeyupHAe"), common.MustParseAddress("LnKqKSG4o"), "fleta.io.fr00009")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2V1zboMnJbJdeLvRBRFVPvVqs8CCmjxToBpGJSNScu2"), common.MustParsePublicHash("2V1zboMnJbJdeLvRBRFVPvVqs8CCmjxToBpGJSNScu2"), common.MustParseAddress("NyohoGBJ7"), "fleta.io.fr00010")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3pEYkEgXoPUm4vdcGBXP46q1BpMj215uVQdAg6P4g74"), common.MustParsePublicHash("3pEYkEgXoPUm4vdcGBXP46q1BpMj215uVQdAg6P4g74"), common.MustParseAddress("RBHaH66XR"), "fleta.io.fr00011")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("rsUoPRfVgXJFuV6wYcy4M4kntvr3tooeXzcRhrjBq6"), common.MustParsePublicHash("rsUoPRfVgXJFuV6wYcy4M4kntvr3tooeXzcRhrjBq6"), common.MustParseAddress("TNmSkv1kj"), "fleta.io.fr00012")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4UMYzaBeXEKcm6hnDDEMqYRR5NLwGndCLksryVj98Fw"), common.MustParsePublicHash("4UMYzaBeXEKcm6hnDDEMqYRR5NLwGndCLksryVj98Fw"), common.MustParseAddress("VaFKEjvz3"), "fleta.io.fr00013")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3h2Lt2uYFMqVQKFgKszLJzwaLhQ5kt1nMcg8M758aLh"), common.MustParsePublicHash("3h2Lt2uYFMqVQKFgKszLJzwaLhQ5kt1nMcg8M758aLh"), common.MustParseAddress("XmjBiZrDM"), "fleta.io.fr00014")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4NkvvfPdHHvpo9YTkAQBrGxpnnML2pVRXHdLgzB2EYe"), common.MustParsePublicHash("4NkvvfPdHHvpo9YTkAQBrGxpnnML2pVRXHdLgzB2EYe"), common.MustParseAddress("ZyD4CPmSf"), "fleta.io.fr00015")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3ae9sCuM75vAheVLNp3DjQqDiD3TaxY5HYduHvsgzYZ"), common.MustParsePublicHash("3ae9sCuM75vAheVLNp3DjQqDiD3TaxY5HYduHvsgzYZ"), common.MustParseAddress("cAgvgDgfy"), "fleta.io.fr00016")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2bR5L2ZSqKLUFQzdhzWV6e4BUupHPGDFtnZUNrZBZbZ"), common.MustParsePublicHash("2bR5L2ZSqKLUFQzdhzWV6e4BUupHPGDFtnZUNrZBZbZ"), common.MustParseAddress("eNAoA3buH"), "fleta.io.fr00017")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("BPqzvcrYi364mm6GyraHHqJHrvEfqjwo1jEC8crTxZ"), common.MustParsePublicHash("BPqzvcrYi364mm6GyraHHqJHrvEfqjwo1jEC8crTxZ"), common.MustParseAddress("gZefdsX8b"), "fleta.io.fr00018")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2vtYXNUAtBtt4fF6DEbVKNc7bGhA7yBbatTA6Ye9kMT"), common.MustParsePublicHash("2vtYXNUAtBtt4fF6DEbVKNc7bGhA7yBbatTA6Ye9kMT"), common.MustParseAddress("im8Y7hSMu"), "fleta.io.fr00019")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("42TUBLNb1natk7s7qsHNqxHwn7Pb3pNmTfTnd1sDQnb"), common.MustParsePublicHash("42TUBLNb1natk7s7qsHNqxHwn7Pb3pNmTfTnd1sDQnb"), common.MustParseAddress("kxcQbXMbD"), "fleta.io.fr00020")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2yng1DwwBqMixjCnjx6Pdf9o5AkgEzkumxJySr8Qe6C"), common.MustParsePublicHash("2yng1DwwBqMixjCnjx6Pdf9o5AkgEzkumxJySr8Qe6C"), common.MustParseAddress("oA6H5MGpX"), "fleta.io.fr00021")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3PNrAwb7FrvKeB1hCxYADwNxqWuYmaqoc8E8VjdBC"), common.MustParsePublicHash("3PNrAwb7FrvKeB1hCxYADwNxqWuYmaqoc8E8VjdBC"), common.MustParseAddress("qMa9ZBC3q"), "fleta.io.fr00022")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2eZAofvjk5AHUpaUyC7EDx3K8KAHUQNXMynHG7ZYFfn"), common.MustParsePublicHash("2eZAofvjk5AHUpaUyC7EDx3K8KAHUQNXMynHG7ZYFfn"), common.MustParseAddress("sZ42317H9"), "fleta.io.fr00023")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4QT4FGpoaFkPiRaZQCKDfrANWJ6EAqavqkQfGr6g4oG"), common.MustParsePublicHash("4QT4FGpoaFkPiRaZQCKDfrANWJ6EAqavqkQfGr6g4oG"), common.MustParseAddress("ukXtWq2WT"), "fleta.io.fr00024")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2nPZHDpFavW2VjnZGs7ZeQyFM19y517ZTQaTgqe3G69"), common.MustParsePublicHash("2nPZHDpFavW2VjnZGs7ZeQyFM19y517ZTQaTgqe3G69"), common.MustParseAddress("wx1kzewjm"), "fleta.io.fr00025")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("bB88uMhpM4vjUHpV5WZqfQBh4kyi6wnnKCtVF4AE2D"), common.MustParsePublicHash("bB88uMhpM4vjUHpV5WZqfQBh4kyi6wnnKCtVF4AE2D"), common.MustParseAddress("z9VdUUry5"), "fleta.io.fr00026")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("2ZLEXwQ9pqvaATFttkkNWY2CGDHdJFa5V3GNapKeqtx"), common.MustParsePublicHash("2ZLEXwQ9pqvaATFttkkNWY2CGDHdJFa5V3GNapKeqtx"), common.MustParseAddress("22LyVxJnCP"), "fleta.io.fr00027")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4M2KFgmWSKu8JyjhkmVJ8U4hjtn9MX4rsch4ZoE1i32"), common.MustParsePublicHash("4M2KFgmWSKu8JyjhkmVJ8U4hjtn9MX4rsch4ZoE1i32"), common.MustParseAddress("24YTNS8hRh"), "fleta.io.fr00028")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("XG9nFJsdMpo6D6wYxYSyH5zAtnvsMjySFHp1XjCouY"), common.MustParsePublicHash("XG9nFJsdMpo6D6wYxYSyH5zAtnvsMjySFHp1XjCouY"), common.MustParseAddress("26jwEuxcf1"), "fleta.io.fr00029")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3uW4bb1kAx35ndj4ZVLMF8xWYercS2RfP7moxZvUm8Y"), common.MustParsePublicHash("3uW4bb1kAx35ndj4ZVLMF8xWYercS2RfP7moxZvUm8Y"), common.MustParseAddress("28wR7PnXtK"), "fleta.io.fr00030")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4mY5G1BZuZaeHR5cH1K4sUNmccPa11JkHtjv5ctde3K"), common.MustParsePublicHash("4mY5G1BZuZaeHR5cH1K4sUNmccPa11JkHtjv5ctde3K"), common.MustParseAddress("2B8tyscT7d"), "fleta.io.fr00031")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3oocpeXtqUZeaut1A71fbCMBQefMFMCBt2BpamNZfA9"), common.MustParsePublicHash("3oocpeXtqUZeaut1A71fbCMBQefMFMCBt2BpamNZfA9"), common.MustParseAddress("2DLNrMSNLw"), "fleta.io.fr00032")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4wknRQ86rTcN1cQbXZfbCMkqXcS1FsYG8ihAYFhmxF"), common.MustParsePublicHash("4wknRQ86rTcN1cQbXZfbCMkqXcS1FsYG8ihAYFhmxF"), common.MustParseAddress("2FXriqGHaF"), "fleta.io.fr00033")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("3mT9SNvGscpwmDjHnojnVysd9pXUvg1fenVyiBFYTDs"), common.MustParsePublicHash("3mT9SNvGscpwmDjHnojnVysd9pXUvg1fenVyiBFYTDs"), common.MustParseAddress("2HjLbK6CoZ"), "fleta.io.fr00034")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("24zn1BgQBmMD8dWap9XbBHdZAivDppVhnYxzZ4ftZw4"), common.MustParsePublicHash("24zn1BgQBmMD8dWap9XbBHdZAivDppVhnYxzZ4ftZw4"), common.MustParseAddress("2KvpTnv82s"), "fleta.io.fr00035")
		addAlphaFormulator(sp, ctw, alphaPolicy, common.MustParsePublicHash("4TKCbNqM68vKmmXiMsjdb7qND8Qy1DCJKvFge7Dhw16"), common.MustParsePublicHash("4TKCbNqM68vKmmXiMsjdb7qND8Qy1DCJKvFge7Dhw16"), common.MustParseAddress("2N8JLGk3GB"), "fleta.io.fr00036")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3Zmc4bGPP7TuMYxZZdUhA9kVjukdsE2S8Xpbj4Laovv"), common.NewAddress(0, 37, 0), "fleta.io")
	}
	return nil
}

func addSingleAccount(sp *vault.Vault, ctw *types.ContextWrapper, KeyHash common.PublicHash, addr common.Address, name string) {
	acc := &vault.SingleAccount{
		Address_: addr,
		Name_:    name,
		KeyHash:  KeyHash,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
	sp.AddBalance(ctw, acc.Address(), amount.NewCoinAmount(100000000000, 0))
}

func addAlphaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, alphaPolicy *formulator.AlphaPolicy, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.AlphaFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         alphaPolicy.AlphaCreationAmount,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}
