package app

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/gateway"
	"github.com/fletaio/fleta/process/payment"
	"github.com/fletaio/fleta/process/vault"
)

// FletaApp is app
type FletaApp struct {
	*types.ApplicationBase
	pm      types.ProcessManager
	cn      types.Provider
	addrMap map[string]common.Address
}

// NewFletaApp returns a FletaApp
func NewFletaApp() *FletaApp {
	return &FletaApp{
		addrMap: map[string]common.Address{
			"fleta.gateway":    common.MustParseAddress("3CUsUpv9v"),
			"fleta.formulator": common.MustParseAddress("5PxjxeqJq"),
			"fleta.payment":    common.MustParseAddress("7bScSUkTk"),
		},
	}
}

// Name returns the name of the application
func (app *FletaApp) Name() string {
	return "FletaApp"
}

// Version returns the version of the application
func (app *FletaApp) Version() string {
	return "v1.0.0"
}

// Init initializes the consensus
func (app *FletaApp) Init(reg *types.Register, pm types.ProcessManager, cn types.Provider) error {
	app.pm = pm
	app.cn = cn
	return nil
}

// InitGenesis initializes genesis data
func (app *FletaApp) InitGenesis(ctw *types.ContextWrapper) error {
	rewardPolicy := &formulator.RewardPolicy{
		RewardPerBlock:        amount.NewCoinAmount(0, 951293759512937600), // 0.03%
		PayRewardEveryBlocks:  172800,                                      // 1 day
		AlphaEfficiency1000:   1000,                                        // 100%
		SigmaEfficiency1000:   1150,                                        // 115%
		OmegaEfficiency1000:   1300,                                        // 130%
		HyperEfficiency1000:   1300,                                        // 130%
		StakingEfficiency1000: 700,                                         // 70%
	}
	alphaPolicy := &formulator.AlphaPolicy{
		AlphaCreationLimitHeight:  5184000,                         // 30 days
		AlphaCreationAmount:       amount.NewCoinAmount(200000, 0), // 200,000 FLETA
		AlphaUnlockRequiredBlocks: 2592000,                         // 15 days
	}
	sigmaPolicy := &formulator.SigmaPolicy{
		SigmaRequiredAlphaBlocks:  5184000, // 30 days
		SigmaRequiredAlphaCount:   4,       // 4 Alpha (800,000 FLETA)
		SigmaUnlockRequiredBlocks: 2592000, // 15 days
	}
	omegaPolicy := &formulator.OmegaPolicy{
		OmegaRequiredSigmaBlocks:  5184000, // 30 days
		OmegaRequiredSigmaCount:   2,       // 2 Sigma (1,600,000 FLETA)
		OmegaUnlockRequiredBlocks: 2592000, // 15 days
	}
	hyperPolicy := &formulator.HyperPolicy{
		HyperCreationAmount:         amount.NewCoinAmount(5000000, 0), // 5,000,000 FLETA
		HyperUnlockRequiredBlocks:   2592000,                          // 15 days
		StakingUnlockRequiredBlocks: 2592000,                          // 15 days
	}

	if p, err := app.pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if ap, is := p.(*admin.Admin); !is {
		return types.ErrNotExistProcess
	} else {
		if err := ap.InitAdmin(ctw, app.addrMap); err != nil {
			return err
		}
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
	if p, err := app.pm.ProcessByName("fleta.payment"); err != nil {
		return err
	} else if pp, is := p.(*payment.Payment); !is {
		return types.ErrNotExistProcess
	} else {
		if err := pp.InitTopics(ctw, []string{
			"fleta.formulator.server.cost",
		}); err != nil {
			return err
		}
	}
	if p, err := app.pm.ProcessByName("fleta.gateway"); err != nil {
		return err
	} else if fp, is := p.(*gateway.Gateway); !is {
		return types.ErrNotExistProcess
	} else {
		if err := fp.InitPolicy(ctw,
			&gateway.Policy{
				WithdrawFee: amount.NewCoinAmount(30, 0),
			},
		); err != nil {
			return err
		}
	}
	if p, err := app.pm.ProcessByName("fleta.vault"); err != nil {
		return err
	} else if sp, is := p.(*vault.Vault); !is {
		return types.ErrNotExistProcess
	} else {
		totalSupply := amount.NewCoinAmount(2000000000, 0)
		alphaCreated := alphaPolicy.AlphaCreationAmount.MulC(189)
		sigmaCreated := alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount)).MulC(108)
		hyperCreated := hyperPolicy.HyperCreationAmount.MulC(6)
		totalDeligated := amount.NewCoinAmount(50585413, 290667405989600000)
		totalProvided := amount.NewCoinAmount(31076795, 184877310172010000)
		gatewaySupply := totalSupply.Sub(alphaCreated).Sub(sigmaCreated).Sub(hyperCreated).Sub(totalDeligated).Sub(totalProvided)
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4ArFPXfZF2MH7ZADqj8wD98cTuRSGXbqqNKpU3zPEgk"), common.MustParseAddress("3CUsUpv9v"), "fleta.gateway", gatewaySupply)
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3M227UoMP81Hp8bvpG8XpUqjHgcL7an2QEm44ENMJYi"), common.MustParseAddress("5PxjxeqJq"), "fleta.formulator", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4DBkqwYsmKFysTQFqsMxbtndDHx1V7YbKYQNZC4G33k"), common.MustParseAddress("7bScSUkTk"), "fleta.payment", amount.NewCoinAmount(0, 0))

		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("fv5X9PVeujGRCGNg9AkSFG8ZPFVXzCfMxrk61RxYv4"), common.MustParsePublicHash("4pciwh34bUy1tcHkjZA7B3zp4YzqpqY8jMxU8RzgffB"), common.MustParseAddress("385ujsGNZt"), "HashTower")
		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("UvixpAjKFckZZxu4gKoZvFGFTgC1CCXPztnTkS5kop"), common.MustParsePublicHash("2bhwWjkDVmxMxXxKFpk2Hij2NerxuykN8B4a2NZD9EP"), common.MustParseAddress("9nvUvJibL"), "Cosmostation")
		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("Sd3xmbKWTsAwRq4W2irF1LFGJBE6L8B4WM64YtxseV"), common.MustParsePublicHash("39woEcwAcX4wSJPyK3KC7nWxHMVY1mzF1knmF7DP2sU"), common.MustParseAddress("7bScSUoST"), "Bitsonic")
		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("38NorpBtMfe84EHcSJmwxH7WczEnUupidFUT1Sg74qa"), common.MustParsePublicHash("2m6kaF39t8nX9DL1u4jdWhFTMAH8Pq2Nb3umnKCtwni"), common.MustParseAddress("GPN6MnU3y"), "LikeLion")
		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("4EUPe8ccDeu2Dpwgu2h96EfogT65tHzSAodBpkPG9jb"), common.MustParsePublicHash("2ijQEDfn8pfLKQoDQBwKnfiCfz4NFSQHY59f4c9PRq1"), common.MustParseAddress("3EgMMJk82X"), "FOROUR")
		addHyperFormulator(sp, ctw, hyperPolicy, 0, common.MustParsePublicHash("npBJ6QbJETVMiBNqMmnj7NgwvTRDnuW2J8ndDXYE7u"), common.MustParsePublicHash("43urLQLsGr7uGhhQLLKtfqC3ML9mntL5gkwv57CFnhg"), common.MustParseAddress("3AHPcM6Him"), "WBL")

		setupSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy)
		setupAlphaFormulator(sp, ctw, alphaPolicy)
		setupSingleAccunt(sp, ctw)
	}
	if p, err := app.pm.ProcessByName("fleta.formulator"); err != nil {
		return err
	} else if fp, is := p.(*formulator.Formulator); !is {
		return types.ErrNotExistProcess
	} else {
		setupStaking(fp, ctw)

		HyperAddresses := []common.Address{
			common.MustParseAddress("385ujsGNZt"),
			common.MustParseAddress("9nvUvJibL"),
			common.MustParseAddress("7bScSUoST"),
			common.MustParseAddress("GPN6MnU3y"),
			common.MustParseAddress("3EgMMJk82X"),
			common.MustParseAddress("3AHPcM6Him"),
		}
		if err := fp.InitStakingMap(ctw, HyperAddresses); err != nil {
			return err
		}
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (app *FletaApp) OnLoadChain(loader types.LoaderWrapper) error {
	return nil
}

func addSingleAccount(sp *vault.Vault, ctw *types.ContextWrapper, KeyHash common.PublicHash, addr common.Address, name string, am *amount.Amount) {
	acc := &vault.SingleAccount{
		Address_: addr,
		Name_:    name,
		KeyHash:  KeyHash,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
	if !am.IsZero() {
		if err := sp.AddBalance(ctw, acc.Address(), am); err != nil {
			panic(err)
		}
	}
}

func addAlphaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, alphaPolicy *formulator.AlphaPolicy, PreHeight uint32, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.AlphaFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         alphaPolicy.AlphaCreationAmount,
		PreHeight:      PreHeight,
		UpdatedHeight:  0,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}

func addSigmaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, sigmaPolicy *formulator.SigmaPolicy, alphaPolicy *formulator.AlphaPolicy, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.SigmaFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount)),
		PreHeight:      0,
		UpdatedHeight:  0,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}

func addHyperFormulator(sp *vault.Vault, ctw *types.ContextWrapper, hyperPolicy *formulator.HyperPolicy, Commission1000 uint32, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.HyperFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         hyperPolicy.HyperCreationAmount,
		PreHeight:      0,
		UpdatedHeight:  0,
		StakingAmount:  amount.NewCoinAmount(0, 0),
		Policy: &formulator.ValidatorPolicy{
			CommissionRatio1000: Commission1000,
			MinimumStaking:      amount.NewCoinAmount(100, 0),
			PayOutInterval:      1,
		},
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}

func addStaking(fp *formulator.Formulator, ctw *types.ContextWrapper, HyperAddress common.Address, StakingAddress common.Address, am *amount.Amount) {
	if has, err := ctw.HasAccount(StakingAddress); err != nil {
		panic(err)
	} else if !has {
		panic(types.ErrNotExistAccount)
	}
	if acc, err := ctw.Account(HyperAddress); err != nil {
		panic(err)
	} else if frAcc, is := acc.(*formulator.FormulatorAccount); !is {
		panic(formulator.ErrInvalidFormulatorAddress)
	} else if frAcc.FormulatorType != formulator.HyperFormulatorType {
		panic(formulator.ErrNotHyperFormulator)
	} else {
		frAcc.StakingAmount = frAcc.StakingAmount.Add(am)
	}
	fp.AddStakingAmount(ctw, HyperAddress, StakingAddress, am)
}
