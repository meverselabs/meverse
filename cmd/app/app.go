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
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2nRdYjkWxJtjQ9NgVn9C68yhPR52iFxxs6AKqA7A9ZE"), common.MustParseAddress("4kbaAVnrij"), "fleta.io")

		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2A3pNbqA5jt299VMavzgZMT7bmno876sRg3LtsPCsYS"), common.MustParsePublicHash("1MrZjju9BLsqNFyzyJzPNyHUGqekQ6WmP7FKo26uRX"), common.MustParseAddress("9uWvXk9fhk"), "validator.001")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2A3pNbqA5jt299VMavzgZMT7bmno876sRg3LtsPCsYS"), common.MustParseAddress("99ZK89dJox"), "account.001")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2LX7DLdn7NHBjU8YSAFfwSr1Jm5gYfQEsaQtGk3dC1M"), common.MustParsePublicHash("3NNt3dxrMGge22arNYmdr1guojWvst2cjTg578gytin"), common.MustParseAddress("9sKSfGKkdG"), "validator.002")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2LX7DLdn7NHBjU8YSAFfwSr1Jm5gYfQEsaQtGk3dC1M"), common.MustParseAddress("97MqFfoPjU"), "account.002")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3yqo1AwfmacJvqg2pVB1jvz9SKTBnJPzoajmkqiktDL"), common.MustParsePublicHash("4MgDGVDLxZSR7c7eFj3r9bCkzU7n7r56oAT6BFY5fWR"), common.MustParseAddress("9q7xnnVqYn"), "validator.003")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3yqo1AwfmacJvqg2pVB1jvz9SKTBnJPzoajmkqiktDL"), common.MustParseAddress("95AMPByUez"), "account.003")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("h76W4erjTbxxSfhfnBgvcKoJzLz4VMg9rtuBaRmHnd"), common.MustParsePublicHash("2Vm1UaB5prNZMgNwKX7Rm9AMxvrJGggLjJHKwJPYnWs"), common.MustParseAddress("9nvUvJfvUJ"), "validator.004")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("h76W4erjTbxxSfhfnBgvcKoJzLz4VMg9rtuBaRmHnd"), common.MustParseAddress("92xsWi9ZaW"), "account.004")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4EeM7NxwhPHXH7nTfeCjTTLh4jGcSmiSNnYaCn2HJmK"), common.MustParsePublicHash("25snTEMQb6r893y7CN112rqp9QqddrPNWbipaGji36Z"), common.MustParseAddress("A4Jr1fTLcz"), "validator.005")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4EeM7NxwhPHXH7nTfeCjTTLh4jGcSmiSNnYaCn2HJmK"), common.MustParseAddress("8zmPeEKeW2"), "account.005")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("21GXLfmhAbvJaCzFAQzeftK1jCVnkC1ea3SSjmFdpwp"), common.MustParsePublicHash("zzYW68M9GsqpLM2XJuxFfeKjsn3CTQ5zQDivt29K5g"), common.MustParseAddress("A27N9BdRYW"), "validator.006")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("21GXLfmhAbvJaCzFAQzeftK1jCVnkC1ea3SSjmFdpwp"), common.MustParseAddress("8xZumkVjRY"), "account.006")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("eMkLX9UvxfhSM8bKGng5KWjUQfmgfTW8B8fn8DR9aC"), common.MustParsePublicHash("2n5oFoaQwo9ypEheRtAF9teJy2REFGXVTHv8mqTJqHD"), common.MustParseAddress("9yutGhoWU2"), "validator.007")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("eMkLX9UvxfhSM8bKGng5KWjUQfmgfTW8B8fn8DR9aC"), common.MustParseAddress("8vNRuGfpM4"), "account.007")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("Tghne5GQwmYnEKAyssBbQ4RiFrekN8aAdHnhvi62k2"), common.MustParsePublicHash("487WkiKx8ejKtUSvGAUprmzmQH5m7wk52tMaKXgLupT"), common.MustParseAddress("9wiQQDybPY"), "validator.008")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("Tghne5GQwmYnEKAyssBbQ4RiFrekN8aAdHnhvi62k2"), common.MustParseAddress("8tAx2nquGa"), "account.008")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3KqQQoafABoKw3Q5YfwnT8CWTS3pqBHazUc4oDNMuWw"), common.MustParsePublicHash("4k2wKJDJCJxe1t3YxQEK5bkSEhuVSzjVi2R59sF6AJi"), common.MustParseAddress("AD6mVam1YE"), "validator.009")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3KqQQoafABoKw3Q5YfwnT8CWTS3pqBHazUc4oDNMuWw"), common.MustParseAddress("9T9A5zEeeS"), "account.009")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3rH8de1D3ntHvzFqVCdQ5amF1rhoew9aKw9ygNbSWSi"), common.MustParsePublicHash("35jS6NRVBBWchRBydb4FTpiQodUeEeemG5bmtNnoYmb"), common.MustParseAddress("AAuHd6w6Tk"), "validator.010")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3rH8de1D3ntHvzFqVCdQ5amF1rhoew9aKw9ygNbSWSi"), common.MustParseAddress("9QwgDWQjZx"), "account.010")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2Exx9ZdQ6yynanY19exDPLvxJySfRkYfEy9TgiQcQkN"), common.MustParsePublicHash("3NLQHUvSqCByR9hq5E3JbjXyhtd47ENo1ifFouLQ3GL"), common.MustParseAddress("A8hokd7BPG"), "validator.011")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2Exx9ZdQ6yynanY19exDPLvxJySfRkYfEy9TgiQcQkN"), common.MustParseAddress("9NkCM2apVU"), "account.011")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3a13TkcDX4kyuvXvLZ11ip34i6YXEAokAbfeRgNC4LH"), common.MustParsePublicHash("Q8fqbdoCbQgb6DCqMTSwB89e4dTXDPvkoG5XaDbVT7"), common.MustParseAddress("A6WKt9HGJn"), "validator.012")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3a13TkcDX4kyuvXvLZ11ip34i6YXEAokAbfeRgNC4LH"), common.MustParseAddress("9LYiUYkuQz"), "account.012")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2czFgSFBXm7MQ3q96B2jRDDZPSDj53RFdmJQ7cnD6pe"), common.MustParsePublicHash("3UrXYdEcSkHFWK3L2JF2jff1HUjmsRmLFxB66S3cB6H"), common.MustParseAddress("AMtgyW4gTU"), "validator.013")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2czFgSFBXm7MQ3q96B2jRDDZPSDj53RFdmJQ7cnD6pe"), common.MustParseAddress("9JMEc4vzLW"), "account.013")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("qxbCTjmFGefofXjp9pkoQ2vG8sEzxP9zjciV1qxquB"), common.MustParsePublicHash("iAG7fJ5wwe8qN3qJybVWGrCQck8p6QFWcVo1Y8mVUP"), common.MustParseAddress("AKhD72EmNz"), "validator.014")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("qxbCTjmFGefofXjp9pkoQ2vG8sEzxP9zjciV1qxquB"), common.MustParseAddress("9G9kjb75G2"), "account.014")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2FZe1sie7FrSPGBGru7U8VJwfnaN53hpwgx22JmHj7A"), common.MustParsePublicHash("4gK9jii37FKBTar2E8hTu8eGdXcnwe2aAQ9ifNSyxBn"), common.MustParseAddress("AHVjEYQrJW"), "validator.015")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2FZe1sie7FrSPGBGru7U8VJwfnaN53hpwgx22JmHj7A"), common.MustParseAddress("9DxGs7HABY"), "account.015")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4gAL5SU5fdw99qYoCjkiEaGr3fJSVNhpVVWsfniXhFB"), common.MustParsePublicHash("naDDknodVR4Ko17EpECqSxonYFSKrSzXQF6wntvneF"), common.MustParseAddress("AFJFN4awE2"), "validator.016")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4gAL5SU5fdw99qYoCjkiEaGr3fJSVNhpVVWsfniXhFB"), common.MustParseAddress("9BknzdTF74"), "account.016")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("r1uVGZQnkUjAc7r8qzWqvjiHWFXdXSTzr9kZFA7PPB"), common.MustParsePublicHash("22YhsLMaMJRQfUhPS1KD9eoC1N1YQG7UFSLyWgeT7AG"), common.MustParseAddress("AWgcTRNMNi"), "validator.017")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("r1uVGZQnkUjAc7r8qzWqvjiHWFXdXSTzr9kZFA7PPB"), common.MustParseAddress("9kj13pqzUv"), "account.017")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3jX9CFrD69cv2p5niDTvGNmvSTRGvPzHSvss4dFpZmD"), common.MustParsePublicHash("2wGiiwsxateqd31CE99nhxFgWLhZEMpQinN92PQwjCe"), common.MustParseAddress("AUV8awYSJE"), "validator.018")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3jX9CFrD69cv2p5niDTvGNmvSTRGvPzHSvss4dFpZmD"), common.MustParseAddress("9iXXBM25QS"), "account.018")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2kpouKJxW4CzLkj3HxAuiJx5FV8TPWX2HZzN4FYmbTx"), common.MustParsePublicHash("45foUNmAAtycY9kXqCx8BS5UaCdsGKTsZcSxFo1SggB"), common.MustParseAddress("ASHeiTiXDk"), "validator.019")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2kpouKJxW4CzLkj3HxAuiJx5FV8TPWX2HZzN4FYmbTx"), common.MustParseAddress("9gL3JsCAKx"), "account.019")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4U95pTkLYUZ7cRk4DqrdUuHaMcpF9sUJXowFfkQ3VGQ"), common.MustParsePublicHash("SyoycJPMat1V6iXQQCjqnxmUgGotHun6ZmhSvRyrNj"), common.MustParseAddress("AQ6Aqytc9G"), "validator.020")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4U95pTkLYUZ7cRk4DqrdUuHaMcpF9sUJXowFfkQ3VGQ"), common.MustParseAddress("9e8ZSPNFFU"), "account.020")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4G1LnWK2RF4jFqLL1RUZJ4f7RHNgfRpkKzJwW7y5aqQ"), common.MustParsePublicHash("47YSJpuVtR9XtD4gHdtu64A3KvMwnjTMxsRkVB3ajEf"), common.MustParseAddress("AfUXwLg2Hx"), "validator.021")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4G1LnWK2RF4jFqLL1RUZJ4f7RHNgfRpkKzJwW7y5aqQ"), common.MustParseAddress("9bw5ZuYLAz"), "account.021")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2MYzudjxSySvn1GRmSDsw3UD53USdkbZhGPKrfL2kKj"), common.MustParsePublicHash("pqLQs4wrrT2MrPA4vbdugTq7xBPfid3PCFs2VYTKo7"), common.MustParseAddress("AdH44rr7DU"), "validator.022")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2MYzudjxSySvn1GRmSDsw3UD53USdkbZhGPKrfL2kKj"), common.MustParseAddress("9ZjbhRiR6W"), "account.022")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("32SvSJ1bDgot6ufRCqrGyyJBeCKHpU21rws5U1NToAc"), common.MustParsePublicHash("21cf5M8oYSTPdauaCMKkfwjqs1Jh8pkjupqMDSJ5MiD"), common.MustParseAddress("Ab5aCP2C8z"), "validator.023")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("32SvSJ1bDgot6ufRCqrGyyJBeCKHpU21rws5U1NToAc"), common.MustParseAddress("9XY7pwtW22"), "account.023")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2xgAmuDbup7ymKF6ZgJbYmKNVbUf7vxjHsq2QX9QUh1"), common.MustParsePublicHash("4eWSriyrM564mFaeXDtxJ7dgKTstmrgojJZ9oZ5JvZb"), common.MustParseAddress("AYt6KuCH4W"), "validator.024")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2xgAmuDbup7ymKF6ZgJbYmKNVbUf7vxjHsq2QX9QUh1"), common.MustParseAddress("9VLdxU4awY"), "account.024")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4hAi2mH6SmyC9jtitu9aWD6cR2KXKNfrccWiRp38THp"), common.MustParsePublicHash("3mjTH13gkdcUPUtLf3Ja7Z8S2xUFkUBnXJvcit1zoDU"), common.MustParseAddress("9nvUvJfcf"), "validator.025")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4hAi2mH6SmyC9jtitu9aWD6cR2KXKNfrccWiRp38THp"), common.MustParseAddress("A4Jr1fTLKQ"), "account.025")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4KvPNZeknC2tj6izqBskJ3TBVMypm6EWvYDYkmktjXu"), common.MustParsePublicHash("48GkVybHUBjohXzSCYNZ27LLAwVCzpv7jupVhMS86zL"), common.MustParseAddress("BzQMQ8aqy"), "validator.026")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4KvPNZeknC2tj6izqBskJ3TBVMypm6EWvYDYkmktjXu"), common.MustParseAddress("A27N9BdREv"), "account.026")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("31tTiBcugRiQ88ioHiunuA4kDvdd3V35wd4tC7w8yLz"), common.MustParsePublicHash("3CRHXWU6v1eAukDKdDPPcd8x2FWWJw1zZrygqPpMCiA"), common.MustParseAddress("EBtDsxW5H"), "validator.027")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("31tTiBcugRiQ88ioHiunuA4kDvdd3V35wd4tC7w8yLz"), common.MustParseAddress("9yutGhoWAS"), "account.027")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3VMKxX6VhhwrQaFDZq8nrdMn7EVHx9nQfA4Lpysmqbm"), common.MustParsePublicHash("LZosZ1tAHGYyxM5j5ViiZB3mvVrAYWW5T3fEeVfiUz"), common.MustParseAddress("GPN6MnRJb"), "validator.028")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3VMKxX6VhhwrQaFDZq8nrdMn7EVHx9nQfA4Lpysmqbm"), common.MustParseAddress("9wiQQDyb5x"), "account.028")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4pmPkZD9uHmSiHMmYTryPLkKbNCxyCYmCjXJkUdN12r"), common.MustParsePublicHash("fsu9kE3fPy37ZYJCohadBddoi5hpJs3gMaLEnaBsGT"), common.MustParseAddress("11111Jj"), "validator.029")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4pmPkZD9uHmSiHMmYTryPLkKbNCxyCYmCjXJkUdN12r"), common.MustParseAddress("9uWvXk9g1U"), "account.029")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("22WvzSTFd8sz7dRXkL8Bs3tnE2w4qTqAcXCLDtEnLhp"), common.MustParsePublicHash("2k5Rr4bbShNr1vAYsj2zgvVWEGEAZZ9G4JKQHN8izRZ"), common.MustParseAddress("3CUsUpvY3"), "validator.030")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("22WvzSTFd8sz7dRXkL8Bs3tnE2w4qTqAcXCLDtEnLhp"), common.MustParseAddress("9sKSfGKkvz"), "account.030")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2xvzKgvigdpmirHiVixsgjzUCnhU1bRnjsWrpqhPUZ4"), common.MustParsePublicHash("2Gp1n378cJ69vAnVA6uHyHwbBECVXFRSWEZvxJwhQH6"), common.MustParseAddress("5PxjxeqmM"), "validator.031")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2xvzKgvigdpmirHiVixsgjzUCnhU1bRnjsWrpqhPUZ4"), common.MustParseAddress("9q7xnnVqrW"), "account.031")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3iGGh2rYuaLo4TTWpiPwQSmgHFf6H67TQ1MHfDStVBv"), common.MustParsePublicHash("3PrdrEGWNqYYFxyC6UdMA2HCJhn3cddS1Ssa9CzWeDE"), common.MustParseAddress("7bScSUkzf"), "validator.032")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3iGGh2rYuaLo4TTWpiPwQSmgHFf6H67TQ1MHfDStVBv"), common.MustParseAddress("9nvUvJfvn2"), "account.032")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2wbHqCdH1DErQhUjK7bqDnpZryezZ5Y3mmj8d1aftFH"), common.MustParsePublicHash("EEhQqxkRtyCL4xcuSLzwzRgrkseXkq2knPB3jFpK3u"), common.MustParseAddress("TNmSkv1T9"), "validator.033")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2wbHqCdH1DErQhUjK7bqDnpZryezZ5Y3mmj8d1aftFH"), common.MustParseAddress("AMtgyW4g9t"), "account.033")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4TqkJFNNwVUmfdafS7Joxhb6cyzhRQSrA1wuTGYF49g"), common.MustParsePublicHash("2q5f9kiT2uThe34NuRcoe2AJJZStA2Yzqf7jb9Egttu"), common.MustParseAddress("VaFKEjvgT"), "validator.034")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4TqkJFNNwVUmfdafS7Joxhb6cyzhRQSrA1wuTGYF49g"), common.MustParseAddress("AKhD72Em5Q"), "account.034")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4ga92komN28myfh4ookb93pon7jJbqMHu3h4XZz6byX"), common.MustParsePublicHash("2MPf3ZBZzqbZrwdBwu1QqRgWpNUR3cuKGzo4W8FTiXH"), common.MustParseAddress("XmjBiZqum"), "validator.035")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4ga92komN28myfh4ookb93pon7jJbqMHu3h4XZz6byX"), common.MustParseAddress("AHVjEYQqzv"), "account.035")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2wPWm4sfnf7eaSLLZNYzV3grAhDPdwcDTBDLWb2MLqy"), common.MustParsePublicHash("2Mpfj7wTPaMjfvMZQQkkPQzhzo7DBJg1krnBjmZTkyN"), common.MustParseAddress("ZyD4CPm95"), "validator.036")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2wPWm4sfnf7eaSLLZNYzV3grAhDPdwcDTBDLWb2MLqy"), common.MustParseAddress("AFJFN4avvS"), "account.036")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3F3iYu4zvf4wb2yqsWVvbmDLtZANpjHNPRaSrGtEKbP"), common.MustParsePublicHash("3vpLqFXTCZiMqc9GDFwNXKuRQzu5r1CQwM9PGV1HMjG"), common.MustParseAddress("JaqxqcM9D"), "validator.037")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3F3iYu4zvf4wb2yqsWVvbmDLtZANpjHNPRaSrGtEKbP"), common.MustParseAddress("AD6mVam1qx"), "account.037")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("45tAMsbM3nrRHd4ivkHoocdFdHsFeUtV5R5ZagqDajk"), common.MustParsePublicHash("3Hx6k3rMP6VsWSWtSFtm78f68Dh4QBZ6wxvBqq7QErk"), common.MustParseAddress("LnKqKSGNX"), "validator.038")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("45tAMsbM3nrRHd4ivkHoocdFdHsFeUtV5R5ZagqDajk"), common.MustParseAddress("AAuHd6w6mU"), "account.038")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3fERwqZjJ2EuuRNKzkJekbuzG7h8XfJ2MtxTPwXpVr1"), common.MustParsePublicHash("2dF1UjqQbgKkUS4ToFLFP6QCnd8UxGZv2iuYNJ3R2e3"), common.MustParseAddress("NyohoGBbq"), "validator.039")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3fERwqZjJ2EuuRNKzkJekbuzG7h8XfJ2MtxTPwXpVr1"), common.MustParseAddress("A8hokd7Bgz"), "account.039")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2ixNC73voksuaxGaj7bqQJemiueLv7aYzdXfFX4sUEJ"), common.MustParsePublicHash("2SZFyqHx3nDsJheCRCazrT7kbBmrbzRusfmaKEN6k1T"), common.MustParseAddress("RBHaH66q9"), "validator.040")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2ixNC73voksuaxGaj7bqQJemiueLv7aYzdXfFX4sUEJ"), common.MustParseAddress("A6WKt9HGcW"), "account.040")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("9fnxZZq2tewH1dBZ8xh2p1aEpBvWRqcX7jyCEtydK1"), common.MustParsePublicHash("iCKLTcW5oQ4RQW3paGKVJ8xKV3i2UoFmrwr6bLcZpM"), common.MustParseAddress("kxcQbXMHd"), "validator.041")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("9fnxZZq2tewH1dBZ8xh2p1aEpBvWRqcX7jyCEtydK1"), common.MustParseAddress("AfUXwLg1zN"), "account.041")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2evxEUidaVmnMBVR49ge2xzNZ3JacCeSYFwSRb2Hd5X"), common.MustParsePublicHash("E69UuGkfdJxDosytGozZA6bVvSqMUj9ifSbC7Sjy2C"), common.MustParseAddress("oA6H5MGWw"), "validator.042")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2evxEUidaVmnMBVR49ge2xzNZ3JacCeSYFwSRb2Hd5X"), common.MustParseAddress("AdH44rr6ut"), "account.042")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3KTGv9JnrVwyXXvY7TGaKvmqKzw7MBHKYUK9mMKw4zk"), common.MustParsePublicHash("3JNB13QzoqXrR8yDUNHGgYsoAg52gz21CwFjFPhiWco"), common.MustParseAddress("qMa9ZBBkF"), "validator.043")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3KTGv9JnrVwyXXvY7TGaKvmqKzw7MBHKYUK9mMKw4zk"), common.MustParseAddress("Ab5aCP2BqQ"), "account.043")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("mLoc7vgSs6Z331iTrF95qQndUA2jXKGm5dPxxUJ6a8"), common.MustParsePublicHash("2KSc9wdWBJnK2HjcWjZbXzwQkHdH9PnrUdxiumZ1DZt"), common.MustParseAddress("sZ42316yZ"), "validator.044")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("mLoc7vgSs6Z331iTrF95qQndUA2jXKGm5dPxxUJ6a8"), common.MustParseAddress("AYt6KuCGkv"), "account.044")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3u2VzK6zX4QFxGsfV2ZcjzZWRyob5cyaDMraJckeMY9"), common.MustParsePublicHash("3iE2t6znYGg3tCUWTWx6CMckffZet1xNcyhV3G3U5bk"), common.MustParseAddress("cAgvgDgyh"), "validator.045")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3u2VzK6zX4QFxGsfV2ZcjzZWRyob5cyaDMraJckeMY9"), common.MustParseAddress("AWgcTRNMgS"), "account.045")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2sMTB2jGHKLMCz3Tc6twWWr9sEhdH7HHX5v7RiYRKE1"), common.MustParsePublicHash("3rBytTiAtGUQxLsZBzRUHH7GBU5HDWyqAfzFHLyBhrT"), common.MustParseAddress("eNAoA3cD1"), "validator.046")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2sMTB2jGHKLMCz3Tc6twWWr9sEhdH7HHX5v7RiYRKE1"), common.MustParseAddress("AUV8awYSbx"), "account.046")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2yidGkHRyRk34mvh3cybMqjW35mvr8TZFoDzaVL5Na1"), common.MustParsePublicHash("T5Zd4fV3pvP2FR5CGNhsvGAMVgHp57TbT2XZfgk2u3"), common.MustParseAddress("gZefdsXSK"), "validator.047")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2yidGkHRyRk34mvh3cybMqjW35mvr8TZFoDzaVL5Na1"), common.MustParseAddress("ASHeiTiXXU"), "account.047")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("28viaHDe7v8XwZ3B8ebxeNcFnYGzh2TDuJeM63K2Xku"), common.MustParsePublicHash("q7U2X3crFi6JChKifmGPivxfxHtCSthWtqh6sorkQ9"), common.MustParseAddress("im8Y7hSfd"), "validator.048")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("28viaHDe7v8XwZ3B8ebxeNcFnYGzh2TDuJeM63K2Xku"), common.MustParseAddress("AQ6AqytcSz"), "account.048")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2CFQcjSFd9xwiPPzGUx9RcFyhm8eq185oXg8pZhsgpq"), common.MustParsePublicHash("qZBd2SQSSMQejta98AxxMJxZ9mPbYMEM55kCUF7aGF"), common.MustParseAddress("24YTNS8h87"), "validator.049")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2CFQcjSFd9xwiPPzGUx9RcFyhm8eq185oXg8pZhsgpq"), common.MustParseAddress("JaqxqcLEK"), "account.049")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3F8ekcVneE4kVKWiKMw7x97D4ujUbybopxG11NxAJrN"), common.MustParsePublicHash("Vjp9U36nZryZo4zfJCYejCtxgoT9XVoqcySZdPrgm"), common.MustParseAddress("26jwEuxcMR"), "validator.050")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3F8ekcVneE4kVKWiKMw7x97D4ujUbybopxG11NxAJrN"), common.MustParseAddress("LnKqKSFTd"), "account.050")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4fKZ1Whq9SdiDeQGg1KyX9WUV9Vu89XL3ejDVmQtfWD"), common.MustParsePublicHash("3FJcxC8S2oA3xJ6iF2GT1NTFR8mcUWxZyf45KYJioAE"), common.MustParseAddress("28wR7PnXaj"), "validator.051")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4fKZ1Whq9SdiDeQGg1KyX9WUV9Vu89XL3ejDVmQtfWD"), common.MustParseAddress("NyohoGAgw"), "account.051")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3Hi4FcvNusXVWzy4uDHnBf6YfRrXMyYooioWVsc1gDw"), common.MustParsePublicHash("8YbkTnbXiynP3tn77rRxX1QxT9dzzcgVTyHYoLQfHD"), common.MustParseAddress("2B8tyscSp3"), "validator.052")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3Hi4FcvNusXVWzy4uDHnBf6YfRrXMyYooioWVsc1gDw"), common.MustParseAddress("RBHaH65vF"), "account.052")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("392Daga1ai7w8iVgGRfgTBKUBWaMkSa2jf5FLTS9Xrj"), common.MustParsePublicHash("t6GcMeWNDVikmwX21YfibjRaye5wvLwUTebzxoQVYi"), common.MustParseAddress("ukXtWq2pB"), "validator.053")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("392Daga1ai7w8iVgGRfgTBKUBWaMkSa2jf5FLTS9Xrj"), common.MustParseAddress("TNmSkv19Z"), "account.053")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("271LNtRGSzGSGespbyp8cRdmqTBdZVkwW7QD3SSQrTG"), common.MustParsePublicHash("mKQ1EKE171RFaH2s6P7sgiYoD5Ds6rfFG3QAJSWMfT"), common.MustParseAddress("wx1kzex3V"), "validator.054")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("271LNtRGSzGSGespbyp8cRdmqTBdZVkwW7QD3SSQrTG"), common.MustParseAddress("VaFKEjvNs"), "account.054")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("U7a1EMg7CXGLhL3d3LSHPbhSzAKXg9WFcpCJ6eprQ4"), common.MustParsePublicHash("kfBb7nwrN2mbBstxKRn2274a9B2bECU9BTbZ5ET961"), common.MustParseAddress("z9VdUUsGo"), "validator.055")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("U7a1EMg7CXGLhL3d3LSHPbhSzAKXg9WFcpCJ6eprQ4"), common.MustParseAddress("XmjBiZqcB"), "account.055")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3BfxWTJgPbHvPbpikbfnsbmpaXBsgEoYZJxAAAmuyjS"), common.MustParsePublicHash("yheJaqAqu1ZsVZxyjCw1AcNNy4ZbeA7DdZagxQcsAQ"), common.MustParseAddress("22LyVxJnW7"), "validator.056")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3BfxWTJgPbHvPbpikbfnsbmpaXBsgEoYZJxAAAmuyjS"), common.MustParseAddress("ZyD4CPkqV"), "account.056")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2H9jL766WVbMfkxT6bpSsGhk5Skmnx9ERr5VQ5xtPmX"), common.MustParsePublicHash("2egXd9jpJvXpKgckXqiZt66nvVttxVDPxchN7gGA3N9"), common.MustParseAddress("2N8JLGk2xb"), "validator.057")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2H9jL766WVbMfkxT6bpSsGhk5Skmnx9ERr5VQ5xtPmX"), common.MustParseAddress("11111cT"), "account.057")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3hDdxbF8TaTHmVUdg6Nbaz4gXCZtzfFZptiVqrMto9F"), common.MustParsePublicHash("2RmL4gffHrCpLZgwYjwf1zxDHZmDngUyCatzxU9GfAt"), common.MustParseAddress("2QKnCkZxBu"), "validator.058")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3hDdxbF8TaTHmVUdg6Nbaz4gXCZtzfFZptiVqrMto9F"), common.MustParseAddress("3CUsUpvqm"), "account.058")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2BXA3rzTup53kqzBKSGZRNfvyFqvFU4qbSBShGCNfor"), common.MustParsePublicHash("Eo7zwEKbBseKKPdYPXkWBqUvQtoeW8m2egMosUkgof"), common.MustParseAddress("2SXG5EPsRD"), "validator.059")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2BXA3rzTup53kqzBKSGZRNfvyFqvFU4qbSBShGCNfor"), common.MustParseAddress("5Pxjxer55"), "account.059")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2kFB9uoShXeyvB5Xa6tLAE5b57zkjzJVG5gHezAbo2u"), common.MustParsePublicHash("2om1qmbVJFJF63JNcRTgdecddY5nkugptDumVsh9EpW"), common.MustParseAddress("2UijwiDneX"), "validator.060")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2kFB9uoShXeyvB5Xa6tLAE5b57zkjzJVG5gHezAbo2u"), common.MustParseAddress("7bScSUmJP"), "account.060")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("g4oC8kZ1CfEheVtFbv84u1TbzUKcNHiqxVt6eLzsvF"), common.MustParsePublicHash("356j5ZvHWw2YUJwYsG2GFRbGcTGfgxyJTUnY1vZcCv7"), common.MustParseAddress("2DLNrMSNef"), "validator.061")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("g4oC8kZ1CfEheVtFbv84u1TbzUKcNHiqxVt6eLzsvF"), common.MustParseAddress("9nvUvJgXh"), "account.061")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4TTeBERaYtcvqfaZ4Fxr1J8AZADSEBKaWJxn7Yj33Cz"), common.MustParsePublicHash("4eNcadoEzB9tcJKbXyLaTUcBNU7uPyGkTZdf5zSkQAi"), common.MustParseAddress("2FXriqGHsy"), "validator.062")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4TTeBERaYtcvqfaZ4Fxr1J8AZADSEBKaWJxn7Yj33Cz"), common.MustParseAddress("BzQMQ8bm1"), "account.062")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4s7YAvduBT1LmUERGwYy7m195M9m31rmGZ3G3SsdxHF"), common.MustParsePublicHash("2eyf1YFV8QdsgNgwTUDGL6rA1j4Sw2Q72NRo5zFajJT"), common.MustParseAddress("2HjLbK6D7H"), "validator.063")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4s7YAvduBT1LmUERGwYy7m195M9m31rmGZ3G3SsdxHF"), common.MustParseAddress("EBtDsxWzK"), "account.063")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2E3pA9uAd93QuJNo4QccU1fUd49aDidsqZRq4LC1x5E"), common.MustParsePublicHash("2qvqjJmKHNRRxZGiELpDZzt9XZhvit4FL38B7d7MBEX"), common.MustParseAddress("2KvpTnv8Lb"), "validator.064")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2E3pA9uAd93QuJNo4QccU1fUd49aDidsqZRq4LC1x5E"), common.MustParseAddress("GPN6MnSDd"), "account.064")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3hhnfrf43N2yJC98bMhUC5EFXB7vr4WrMTBUExRDaz1"), common.MustParsePublicHash("2cJu1FqHsPDnNnsQgEMiH7fbutriew6MZdfzxKbYfBU"), common.MustParseAddress("2fi9J7MNo5"), "validator.065")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3hhnfrf43N2yJC98bMhUC5EFXB7vr4WrMTBUExRDaz1"), common.MustParseAddress("ukXtWq1uH"), "account.065")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4SNvTUN2iTmyTomqqYCYAzUxCV3rHhr3V6K1XtrxGU"), common.MustParsePublicHash("4jYYh8WFS9GVyF6z3VQvw11Fw4pQrbeeoU38eBGtdc7"), common.MustParseAddress("2hudAbBJ2P"), "validator.066")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4SNvTUN2iTmyTomqqYCYAzUxCV3rHhr3V6K1XtrxGU"), common.MustParseAddress("wx1kzew8b"), "account.066")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("41qy5i39SXpwW3FrKZf26HR78D33P5spxp2MF4oEcN1"), common.MustParsePublicHash("3nEovhCFaiaFgCorzz5MYUkvxUk7WjAbZzSFYa2gnQ2"), common.MustParseAddress("2k77351DFh"), "validator.067")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("41qy5i39SXpwW3FrKZf26HR78D33P5spxp2MF4oEcN1"), common.MustParseAddress("z9VdUUrMu"), "account.067")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("44eBNBURT83xYgogqb8d8BgWAM1nnb61xCZ3DEU2Ufb"), common.MustParsePublicHash("33SRNUT337stvERRt7tFmXYRffod6TwRUiovknPMmpZ"), common.MustParseAddress("2nJauYq8V1"), "validator.068")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("44eBNBURT83xYgogqb8d8BgWAM1nnb61xCZ3DEU2Ufb"), common.MustParseAddress("22LyVxJmbD"), "account.068")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3UuuFFZr4Larb2fUrgrnn6M17VNnbzaMd8B9ZYpJHsk"), common.MustParsePublicHash("3soiGaRf8yquS3dnFNCz8LGBp1gZ3eMiaoQDyMBUArW"), common.MustParseAddress("2WvDpC3iV9"), "validator.069")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3UuuFFZr4Larb2fUrgrnn6M17VNnbzaMd8B9ZYpJHsk"), common.MustParseAddress("24YTNS8gpX"), "account.069")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("g4dX4CvtvFCvWkp5ie5PB7Est3B7D7Bd6DTwq3Z6cD"), common.MustParsePublicHash("FXyJUFjzZDUjNvK1zbmDCoPGEFPjUJ2SoYm2qJ6t87"), common.MustParseAddress("2Z7hgfsdiT"), "validator.070")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("g4dX4CvtvFCvWkp5ie5PB7Est3B7D7Bd6DTwq3Z6cD"), common.MustParseAddress("26jwEuxc3q"), "account.070")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4ccAhQe4EcVbQqVAKhbYqwz3oHCdXxdguC8qhcvZFQQ"), common.MustParsePublicHash("3eh4kpEDTySo3dqatcnnSXgTyAMx82wz9RsGnGpdK77"), common.MustParseAddress("2bKBZ9hYwm"), "validator.071")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4ccAhQe4EcVbQqVAKhbYqwz3oHCdXxdguC8qhcvZFQQ"), common.MustParseAddress("28wR7PnXH9"), "account.071")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2Hvf3AZNzXCW7EDkHPbf91537zoVXP26RWRS7LmNs2n"), common.MustParsePublicHash("3GiMQX2QRYvnYoCDVv92Qah6u47bDjVoLUghVDZws8L"), common.MustParseAddress("2dWfRdXUB5"), "validator.072")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2Hvf3AZNzXCW7EDkHPbf91537zoVXP26RWRS7LmNs2n"), common.MustParseAddress("2B8tyscSWT"), "account.072")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("42kzHAxvAg14RDFCTUvnRpf66jVkw9zyzDT7q5hFuLT"), common.MustParsePublicHash("33sZ4era1pE35uZf6H8mwCk1ZogUBG1zKge4mhAcWNR"), common.MustParseAddress("2yHzFwxidZ"), "validator.073")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("42kzHAxvAg14RDFCTUvnRpf66jVkw9zyzDT7q5hFuLT"), common.MustParseAddress("cAgvgDhHR"), "account.073")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2r8oiEhJ9TGV6ud6RsCkufwoerxsJJUqiMWyu99UcC9"), common.MustParsePublicHash("4kj7Ci4zWfEFKESPZxiqXSFJS6MWZ7FKQvCuX6t9drY"), common.MustParseAddress("31VU8Rndrs"), "validator.074")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2r8oiEhJ9TGV6ud6RsCkufwoerxsJJUqiMWyu99UcC9"), common.MustParseAddress("eNAoA3cWj"), "account.074")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2VsMUhHCjNtVhJhxZ3uPapu21QxsBfYbxTJCMsE5tJZ"), common.MustParsePublicHash("2Lmy3UEgJL25qUzdHV1eb7ziizCM6moH8GDNbD4SUZr"), common.MustParseAddress("33gwzucZ6B"), "validator.075")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2VsMUhHCjNtVhJhxZ3uPapu21QxsBfYbxTJCMsE5tJZ"), common.MustParseAddress("gZefdsXk3"), "account.075")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("XJiu52v5WUjHrXdWSGAGYo4wBfiphGQJKgD7kUw3av"), common.MustParsePublicHash("4sjgvrnxxcfPhCN1N1fFMNt6MuRnGLgHBHdDNP9FQBw"), common.MustParseAddress("35tRsPSUKV"), "validator.076")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("XJiu52v5WUjHrXdWSGAGYo4wBfiphGQJKgD7kUw3av"), common.MustParseAddress("im8Y7hSyM"), "account.076")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4JdQ5HgaBkb65FLMwhWoZmWjoMa86V162pP9xYF71Am"), common.MustParsePublicHash("b3VqPQnFLsSFx28aWsCdqsS9ncjgschsz25tAw1THQ"), common.MustParseAddress("2pW4n2f4Kd"), "validator.077")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4JdQ5HgaBkb65FLMwhWoZmWjoMa86V162pP9xYF71Am"), common.MustParseAddress("kxcQbXNCf"), "account.077")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("fkS3tT8PKJLvpSeDtXM1mZnc6SSY9uZcN3aKXSkc2g"), common.MustParsePublicHash("4o5bTx8CUdYGMPWxqVdWo1dz9JLn7FnHw5nXGdg7yKQ"), common.MustParseAddress("2rhYeWUyYw"), "validator.078")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("fkS3tT8PKJLvpSeDtXM1mZnc6SSY9uZcN3aKXSkc2g"), common.MustParseAddress("oA6H5MHRy"), "account.078")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4B6PV5xPZtSN3HC5MMfwjBmhVr8hum6Tny2ow3qNNrz"), common.MustParsePublicHash("2aRWbeWybkZ2GZ3cwvztkw446qperMR4ZyiqRcjGKwS"), common.MustParseAddress("2tu2WzJtnF"), "validator.079")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4B6PV5xPZtSN3HC5MMfwjBmhVr8hum6Tny2ow3qNNrz"), common.MustParseAddress("qMa9ZBCfH"), "account.079")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4Dw6eis3PKzUaXNFuM6ppVGcYqXd2nTWanJREtazKNX"), common.MustParsePublicHash("4ojVnzK4zHtJgMbGrjNAyX8fcSogqm97y28nWiTtzmf"), common.MustParseAddress("2w6WPU8p1Z"), "validator.080")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4Dw6eis3PKzUaXNFuM6ppVGcYqXd2nTWanJREtazKNX"), common.MustParseAddress("sZ42317tb"), "account.080")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("4RaebXSgS7QQkvGKUvGhK1j9sVN5YcnKLHWXqUeZw61"), common.MustParsePublicHash("31wG2xryMCFhgkuTAf4vZxJLXDpQrpDwRiCovtLbKnA"), common.MustParseAddress("3GsqDna4U3"), "validator.081")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4RaebXSgS7QQkvGKUvGhK1j9sVN5YcnKLHWXqUeZw61"), common.MustParseAddress("2WvDpC3haF"), "account.081")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2QoDDPJDsTWBDk1AMR3gmtHLkP1qYy7pGuox8DUoeE3"), common.MustParsePublicHash("DxqL2vTr3HV6Eq54peVqnk1f3Ah2SiNocbJTDn4FY7"), common.MustParseAddress("3K5K6GPyhM"), "validator.082")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2QoDDPJDsTWBDk1AMR3gmtHLkP1qYy7pGuox8DUoeE3"), common.MustParseAddress("2Z7hgfscoZ"), "account.082")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2GuR54YbWVToL9zYrh5G1sCGJxbBkuKi7mWKXeCqU5P"), common.MustParsePublicHash("2KJFTmPgfiE4RHBCDdphV1u5DLqxTJ3cWAasszykVeD"), common.MustParseAddress("3MGnxkDtvf"), "validator.083")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2GuR54YbWVToL9zYrh5G1sCGJxbBkuKi7mWKXeCqU5P"), common.MustParseAddress("2bKBZ9hY2s"), "account.083")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2C8SEtfREYLLLiNWdE4dsUoqmLNWQiQEaUskcP9ofr5"), common.MustParsePublicHash("4UwmRgBF9oSyxnzmv9LADzTGJwoizypqyBHedoZ6LAp"), common.MustParseAddress("3PUGqE3p9y"), "validator.084")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2C8SEtfREYLLLiNWdE4dsUoqmLNWQiQEaUskcP9ofr5"), common.MustParseAddress("2dWfRdXTGB"), "account.084")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2zU3jyaaezcqibFJfR1GuFnsdYXEAkMG7Th1WwogyXj"), common.MustParsePublicHash("3RrfCZFk9jXoCDYQAuNCnuzUmqxyvxgq8xJzEgw35Ys"), common.MustParseAddress("385ujsGQA7"), "validator.085")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2zU3jyaaezcqibFJfR1GuFnsdYXEAkMG7Th1WwogyXj"), common.MustParseAddress("2fi9J7MNVV"), "account.085")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("pf2wpD3qSnKXtHsQmpcpPCVk2hJxz5MdtjZKz9TYJo"), common.MustParsePublicHash("4fBBBGrfbVMWELFh1H9bi7CDWDuicSFoEaM8yhRayKi"), common.MustParseAddress("3AHPcM6KPR"), "validator.086")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("pf2wpD3qSnKXtHsQmpcpPCVk2hJxz5MdtjZKz9TYJo"), common.MustParseAddress("2hudAbBHio"), "account.086")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("QRmPJwewLjUAMT2TYWkGPg2PthXxZ8S159YZFhiYtG"), common.MustParsePublicHash("4UiPRLe4rQB6SGfhnGgrNPCN7Pu2ZBUCfX8nGbnd3B2"), common.MustParseAddress("3CUsUpvEcj"), "validator.087")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("QRmPJwewLjUAMT2TYWkGPg2PthXxZ8S159YZFhiYtG"), common.MustParseAddress("2k77351Cx7"), "account.087")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3Bbd1PZVuQad1947AE1wxtB3iYd6C3fVCciik8zmkoj"), common.MustParsePublicHash("4hK69hkmd9RQ2tRHRVWjv5quBH8o3sg3G5CdLdanfFR"), common.MustParseAddress("3EgMMJk9r3"), "validator.088")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3Bbd1PZVuQad1947AE1wxtB3iYd6C3fVCciik8zmkoj"), common.MustParseAddress("2nJauYq8BR"), "account.088")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3JCVRaF1hRRFVisAuQfnYWXexnhFgHf5y2RKvaVm89s"), common.MustParsePublicHash("2o99JL7oLCQzmU2KaJULyga2Wxny8xGQVqYVWX9EdLr"), common.MustParseAddress("3aTgBdBQJX"), "validator.089")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3JCVRaF1hRRFVisAuQfnYWXexnhFgHf5y2RKvaVm89s"), common.MustParseAddress("2DLNrMSNxP"), "account.089")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("9Xe777dYusR8ZnueqJ7h75KL1TMQkqPLNZLct8GWeX"), common.MustParsePublicHash("36xBhKyguNH1dSu3rVHMNyCe6SbB5rGaqefVwq6emiX"), common.MustParseAddress("3cfA471KXq"), "validator.090")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("9Xe777dYusR8ZnueqJ7h75KL1TMQkqPLNZLct8GWeX"), common.MustParseAddress("2FXriqGJBh"), "account.090")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2LFYQy4pBCYxh9vD9ibqLY6ogkDf1dagKF39AjQQbEo"), common.MustParsePublicHash("yBjJ8feGtmXLoUnURLSaHLciKWnUN9qqKYjcu2CCmB"), common.MustParseAddress("3erdvaqEm9"), "validator.091")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2LFYQy4pBCYxh9vD9ibqLY6ogkDf1dagKF39AjQQbEo"), common.MustParseAddress("2HjLbK6DR1"), "account.091")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2cT7ntatW5ajbuAh873fj31nfVvWm6BodvgRK4tG6Av"), common.MustParsePublicHash("2WnQpfYk19Y2TjWAWMUj93zDo4si6bkZWLkFaeCmjVM"), common.MustParseAddress("3h47o4f9zT"), "validator.092")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2cT7ntatW5ajbuAh873fj31nfVvWm6BodvgRK4tG6Av"), common.MustParseAddress("2KvpTnv8eK"), "account.092")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("CDgb5d2dBPBA65iZkXybcLATYYZ1ymu6h3DMfM2YZp"), common.MustParsePublicHash("4Te9jiG8XBmQcpuWE6S5pRdzoHjiLDDX8RKynPYTFuv"), common.MustParseAddress("3Rfkhhsjzb"), "validator.093")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("CDgb5d2dBPBA65iZkXybcLATYYZ1ymu6h3DMfM2YZp"), common.MustParseAddress("2N8JLGk3sd"), "account.093")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("t3KZybSmC9NWeb7ccKY43xhFEAKQZ2vfDx3muSxojv"), common.MustParsePublicHash("4qWr495cubvySjnRKbPjeF8xnXwLdX1G9QVH4Ek3i16"), common.MustParseAddress("3TsEaBhfDu"), "validator.094")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("t3KZybSmC9NWeb7ccKY43xhFEAKQZ2vfDx3muSxojv"), common.MustParseAddress("2QKnCkZy6w"), "account.094")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3iVNbuvFgxuvpwbWKSMru6V2sdv2VpEve1cbikrWQts"), common.MustParsePublicHash("2LJKBZzTaDr9QRFQ9vBnAoshXptGDnTUo5s4WtXpsXU"), common.MustParseAddress("3W4iSfXaTD"), "validator.095")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3iVNbuvFgxuvpwbWKSMru6V2sdv2VpEve1cbikrWQts"), common.MustParseAddress("2SXG5EPtLF"), "account.095")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("2nzo4d7NY5tHGbCMyHaThEojq2QxSURXE5rMBNHov8b"), common.MustParsePublicHash("fKCVMoMA3VeuR63AhhR49tegRDQQcziuGFhfgKH5aU"), common.MustParseAddress("3YGCK9MVgX"), "validator.096")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2nzo4d7NY5tHGbCMyHaThEojq2QxSURXE5rMBNHov8b"), common.MustParseAddress("2UijwiDoZZ"), "account.096")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("Aj4tjQFtmPdr8Vz7A2dFRjxt1H3oyf4jYXGddtr1wP"), common.MustParsePublicHash("4b9Hk25qbuzUGDQH17fmuCqAtMdVNChJYojMtxSgfZs"), common.MustParseAddress("3t3X9Tnk91"), "validator.097")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("Aj4tjQFtmPdr8Vz7A2dFRjxt1H3oyf4jYXGddtr1wP"), common.MustParseAddress("385ujsGPFD"), "account.097")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3TCUeNpfBbdJR89wsq3NY9UPE3J3BGXHRcgppyn53PY"), common.MustParsePublicHash("tLX9YMiUw7ah9a5AcpscQVAXKguaYFayUvFwpAZUiC"), common.MustParseAddress("3vF11wcfNK"), "validator.098")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3TCUeNpfBbdJR89wsq3NY9UPE3J3BGXHRcgppyn53PY"), common.MustParseAddress("3AHPcM6JUX"), "account.098")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("81zEFdVNtyH1R9hwWdaPLESyh8cJ2hRNTK293r1ANu"), common.MustParsePublicHash("24NT2cxAoNM6xqcLFN4DCiayBu9h6xqB2rVPcA6jDX6"), common.MustParseAddress("3xSUtRSabd"), "validator.099")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("81zEFdVNtyH1R9hwWdaPLESyh8cJ2hRNTK293r1ANu"), common.MustParseAddress("3CUsUpvDhq"), "account.099")
		addHyperFormulator(sp, ctw, hyperPolicy, common.MustParsePublicHash("3fxKBUyyw8z5GecXJ44F8VzkbSRpchARCkM3WGqzKZe"), common.MustParsePublicHash("WPd8cx4Yt3XazEeTEqwDcZKaUBbLGrPjvcMarJwfiL"), common.MustParseAddress("3zdxkuGVpw"), "validator.100")
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3fxKBUyyw8z5GecXJ44F8VzkbSRpchARCkM3WGqzKZe"), common.MustParseAddress("3EgMMJk8w9"), "account.100")
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

func addSigmaFormulator(sp *vault.Vault, ctw *types.ContextWrapper, alphaPolicy *formulator.AlphaPolicy, sigmaPolicy *formulator.SigmaPolicy, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.SigmaFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount)),
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}

func addHyperFormulator(sp *vault.Vault, ctw *types.ContextWrapper, hyperPolicy *formulator.HyperPolicy, KeyHash common.PublicHash, GenHash common.PublicHash, addr common.Address, name string) {
	acc := &formulator.FormulatorAccount{
		Address_:       addr,
		Name_:          name,
		FormulatorType: formulator.HyperFormulatorType,
		KeyHash:        KeyHash,
		GenHash:        GenHash,
		Amount:         hyperPolicy.HyperCreationAmount,
		UpdatedHeight:  0,
		StakingAmount:  amount.NewCoinAmount(0, 0),
		Policy: &formulator.ValidatorPolicy{
			CommissionRatio1000: 60,
			MinimumStaking:      amount.NewCoinAmount(100, 0),
			PayOutInterval:      3,
		},
	}
	if err := ctw.CreateAccount(acc); err != nil {
		panic(err)
	}
}
