package app

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/gateway"
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
	if p, err := app.pm.ProcessByName("fleta.admin"); err != nil {
		return err
	} else if ap, is := p.(*admin.Admin); !is {
		return types.ErrNotExistProcess
	} else {
		if err := ap.InitAdmin(ctw, app.addrMap); err != nil {
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
		totalDeligated := amount.NewCoinAmount(42222154, 72034396700000000)
		gatewaySupply := totalSupply.Sub(alphaCreated).Sub(sigmaCreated).Sub(hyperCreated).Sub(totalDeligated)
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4d5aH1brXeatoDYh6UauLemK1gYiHbEDyTqBj8FSKu6"), common.MustParseAddress("3CUsUpv9v"), "fleta.gateway", gatewaySupply)
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3VH7Bhv3Q2vnhQEcmbDoeXVvYu1VPfZUwz5SYs2iWKH"), common.MustParseAddress("5PxjxeqJq"), "fleta.formulator", amount.NewCoinAmount(0, 0))

		addHyperFormulator(sp, ctw, hyperPolicy, 80, common.MustParsePublicHash("fv5X9PVeujGRCGNg9AkSFG8ZPFVXzCfMxrk61RxYv4"), common.MustParsePublicHash("4pciwh34bUy1tcHkjZA7B3zp4YzqpqY8jMxU8RzgffB"), common.MustParseAddress("385ujsGNZt"), "HashTower")
		addHyperFormulator(sp, ctw, hyperPolicy, 80, common.MustParsePublicHash("UvixpAjKFckZZxu4gKoZvFGFTgC1CCXPztnTkS5kop"), common.MustParsePublicHash("2bhwWjkDVmxMxXxKFpk2Hij2NerxuykN8B4a2NZD9EP"), common.MustParseAddress("9nvUvJibL"), "Cosmostation")
		addHyperFormulator(sp, ctw, hyperPolicy, 80, common.MustParsePublicHash("Sd3xmbKWTsAwRq4W2irF1LFGJBE6L8B4WM64YtxseV"), common.MustParsePublicHash("39woEcwAcX4wSJPyK3KC7nWxHMVY1mzF1knmF7DP2sU"), common.MustParseAddress("7bScSUoST"), "Bitsonic")
		addHyperFormulator(sp, ctw, hyperPolicy, 80, common.MustParsePublicHash("38NorpBtMfe84EHcSJmwxH7WczEnUupidFUT1Sg74qa"), common.MustParsePublicHash("2m6kaF39t8nX9DL1u4jdWhFTMAH8Pq2Nb3umnKCtwni"), common.MustParseAddress("GPN6MnU3y"), "LikeLion")
		addHyperFormulator(sp, ctw, hyperPolicy, 68, common.MustParsePublicHash("4EUPe8ccDeu2Dpwgu2h96EfogT65tHzSAodBpkPG9jb"), common.MustParsePublicHash("2ijQEDfn8pfLKQoDQBwKnfiCfz4NFSQHY59f4c9PRq1"), common.MustParseAddress("3EgMMJk82X"), "FOROUR")
		addHyperFormulator(sp, ctw, hyperPolicy, 70, common.MustParsePublicHash("npBJ6QbJETVMiBNqMmnj7NgwvTRDnuW2J8ndDXYE7u"), common.MustParsePublicHash("43urLQLsGr7uGhhQLLKtfqC3ML9mntL5gkwv57CFnhg"), common.MustParseAddress("3AHPcM6Him"), "WBL")

		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3arHErBuSjjrLLqd1ti3E9Q76gu4kaXWLyqP7cCgDwQ"), common.MustParsePublicHash("2Bs7nRupga5ciCh7ahqGjiMQ1ywuYDcEE5yj1A9arHJ"), common.MustParseAddress("3t3X9Tnkkj"), "Ksw789")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4YPSs8T3kKN3ZCryZD67jHGGh6mHeAkV41hkVeNKWfL"), common.MustParsePublicHash("uaX3qd7MN7zNEW9D5gB4xsDfisNZkpyxfuZ8pxsWbE"), common.MustParseAddress("AKhD72EYoS"), "Cs")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4YRDPrEWhWemYofYx111j4eWNTB9oZ88ATyWL1MutQK"), common.MustParsePublicHash("4bCYXkTPfAhyezf4nEbRHSJjuJrBdRQFGd5YKKRHaGW"), common.MustParseAddress("8ab74xEPjL"), "fm1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("36CG6AxrFeD47zQQwkFUCxHjnwUCi3FNsG2poREAnyF"), common.MustParsePublicHash("zCu9noBR4Fp66kfatUyXpwmuXyQ3dfbqP9iV1qUTCx"), common.MustParseAddress("8YPdCUQUaT"), "fm2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4YcqhLTCcZFywo7HKP2yJG7sxvCphTDxyYDHzHbte1g"), common.MustParsePublicHash("QNZTG5P7at2fWvChXxUvZPUneCnBm1bUsTdJH4L1VF"), common.MustParseAddress("8WC9KzaZRa"), "fm3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("WJey1cCfkTpoFZcPPqZqwvp6mXBkoprzfEox3PokJ6"), common.MustParsePublicHash("32LoRsC2XbQ8PBrCUoVbRYmzo4AaAdzTjySWLP35Y4y"), common.MustParseAddress("8TzfTWkeGh"), "fm4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4D7xt25S2BeQmGgsEiFVWgfxGsMrCNPXkPPCHxPrWFp"), common.MustParsePublicHash("4DKoiRHCYT8u7vtHG8WiHXAT9YxcHg9hWj8hLGPRRUX"), common.MustParseAddress("3YGCK9MWNW"), "beauflo")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("29fojdBb167CDbfz2aCDiQhfvFtARwfwjkCcCN5nfEj"), common.MustParsePublicHash("2dKZDGcqXGs2yfpjUzPn7bUZdwLyLEgprPKZo4i3tK8"), common.MustParseAddress("3aTgBdBRXR"), "beauflo1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2QJ6f8dhwX3si18L6MchdfCKH76CwoQ1nm3mSWrMmw6"), common.MustParsePublicHash("2YK6wUxXLZUZ1JUQ8tstMPS6rWWapNLfSTJxbvftjbt"), common.MustParseAddress("5KZnDhBaSX"), "node1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2deqKGEygfTcq9BzFTcUwnsXcXBwkx8DLj3sUGay4at"), common.MustParsePublicHash("Bm5sTNJFYSqGndHYVTJ1fbqL29ZH97WJRbwNqcLCML"), common.MustParseAddress("9NkCM2ae2g"), "node2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("n9jaaXQFPiHCj61Api3b25xwPZcc1BupMiGL3DpJtk"), common.MustParsePublicHash("2ph26NGmaSnVtVA5MEkqaTqzMRMhF8X5pHAfMRekTug"), common.MustParseAddress("4VDD591RCo"), "Seo1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("21TRUjNYwAx5SdkfUhbTuySxCrjxp29dvs9ajWMKDWj"), common.MustParsePublicHash("AeJYTugUCRZsijjPf8dqbYpMxcprJGMKq4Uc4rBXkM"), common.MustParseAddress("9wiQQDyPG6"), "CROM")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3tzfe4TYBi8v4H7eARZBkdTzcjSWEsmxz9N22xjhb2B"), common.MustParsePublicHash("3mfbKb1wVUc8yors3YCybSbLH7x5aezJBGTNwYTZtXM"), common.MustParseAddress("9DxGs7GyGS"), "CHUNHO1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2wutigEjncpuJDkTbtsMNQS9azV7s8P3RuyQnYZk22U"), common.MustParsePublicHash("FVNJyy4s6EUoXD1XnHJUcweqMqxHB2UEWSZLvpmtmz"), common.MustParseAddress("5MmG6B1Vfq"), "AhnYH")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4tmZQcJF8qokftAVKevT3GHVo3tek9r9y346PALH3Zy"), common.MustParsePublicHash("4fqQzceEjECQ7aYAhBB25GgxxdfsT14XTRrSKLonHuP"), common.MustParseAddress("42qSdP6RHx"), "kyuonghwa1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3WUVjhFQVGasbKCPeH3e4VdEqFtnq8rBfSrG7TtxPfV"), common.MustParsePublicHash("4UaU6SKEHECSssFh86GQUrVFMgnH4UQvF5ELgGciccz"), common.MustParseAddress("6qV12tEEd1"), "KIMJUNYEON1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3UrAFm28ji59oF3ZGf7tinRxLfda9SNzQpGLcxsT52a"), common.MustParsePublicHash("4KAdq92ZumcenQ4RC816FuyjcKvZng3XFAVaY7SA2Q3"), common.MustParseAddress("6sgUuN49mv"), "KIMJUNYEON2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("495ZFcSDmuTd8X7fqDo9zkwfMfDfa8EFWMQfkRKPwuW"), common.MustParsePublicHash("3knDQ5mpbaBV27hrWn6g17i9tmQGw5AFx2Rn3LYGdwH"), common.MustParseAddress("76sN8F1ehk"), "Seunguk1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2kYLzsBV6qLCsSFJNZPDeWbZ4dUiEufPVCyb3xgiTAv"), common.MustParsePublicHash("CXPsXsVwS28Zo6yakWW2gN6eV2BU3uq11MFWQJirXo"), common.MustParseAddress("3mT5Y2K1DF"), "Seunguk2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("r58n3K4tYcwMD2AQE82B5x6WHuyGjxTvbfuzdeyrEn"), common.MustParsePublicHash("3rH9J3QTUdFL5hu79EK4SEfnewNrLvZWMPPCh8ExBjs"), common.MustParseAddress("5xvx1rEA6G"), "HOPEAA")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("46z4pGBpR4FpXDKkdr9ecbwo8PunecWbG4HC21mKRkf"), common.MustParsePublicHash("3nAWXX251Bkqf6xpwy9aER1Sdnx6csBNxyEiNCcofaK"), common.MustParseAddress("kxcQbXS7C"), "HOPEBB")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2xNm5cAtujEKhbGQnP4s4EmVs7MNjLVk4Z7BHgQeyn8"), common.MustParsePublicHash("4SocX3Sp6rHhpJ1HQh5SoUNym49mNToQydGxbgZ6eVW"), common.MustParseAddress("im8Y7hWxK"), "HOPECC")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4tfbgYpvSzh9k8tPUim2nN74rYDc4ajq2YhX5DuMNYi"), common.MustParsePublicHash("3cTR2Tswve4889jptAU8PFopH9ygs8u75QeNJd7QVwK"), common.MustParseAddress("3qr3Gyxqbp"), "tree")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4eXe3w4G24NYadoebHxWPBjQCMtaDbzNtWSrhSrC78V"), common.MustParsePublicHash("2FbLMhZkFXNwRucQ7NMXur1z1Shm92Y6WqEf2nrSPDY"), common.MustParseAddress("5PxjxewRK"), "tree1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4PwfEv1HUa8spKDgBhiTg6iXfitkddut2NmnMdFetXp"), common.MustParsePublicHash("eTHgqwstfsDM2CnyAVwHPp6wUu38QnD8ZCnH35zEC6"), common.MustParseAddress("2Z7hgfsgLD"), "Rooteater")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3w1FJ5oBDtPHyviSDFNK9wR8fKQLysNp5iWU9ZQxmgB"), common.MustParsePublicHash("4a79juvuteekAtRdbZVboZXsWHM1dBsgbdDa1WR2Y1Z"), common.MustParseAddress("6XuA53cuPy"), "Hyewon_Formulator")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4McJHYNQT3VvYBdERadb5WVSDw7ERmdedt8aBKxTUdj"), common.MustParsePublicHash("4dSHK87CHJP1AEEKc7FAu3VwxsFZzgTG1Es1sDLnBHD"), common.MustParseAddress("4ZcAp6fFSG"), "joystation2000")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("uWyp4i4J5q7o1kJ4d1fT4NxbfoJfH4iL37cy9YekJx"), common.MustParsePublicHash("3fqAMuMvnoWHH8uVpEBUAx12XrNTPwdz3C4GcH2gdhs"), common.MustParseAddress("4XQgwcqLHP"), "joystation")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2zymvk6jnvUtnx2vMUV5U4b3jfTgCWLxhysFpoRGaLm"), common.MustParsePublicHash("4csrJ549wEPMVZGrrqY8hfsiDgpf1erzEVnSvBD1Wq9"), common.MustParseAddress("XmjBiZwGH"), "joystation1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2VZe4UNseeQt6PkFeas3tXgUTw4ecxj2qwSTh9KsxVa"), common.MustParsePublicHash("29hniQeWmwLqZg3isSQf8bQMAG3m79nYnJWSbXazDAe"), common.MustParseAddress("ZyD4CPrRC"), "joystation2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2u29ySqdAjYY3cbHwsRsSB5kbtTEaPmRH6S64xvcoP2"), common.MustParsePublicHash("4rFFFgxCdBfUDU79BPhUnCRuELdGDbyqdXucegqivR4"), common.MustParseAddress("2pW4n2f6Uo"), "joystation3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2P8S8H7Uuk53EADc9VTT63DLkL4HXJuHv9PwvbuW6QT"), common.MustParsePublicHash("4nM9GHbj3LfL1qYD4oHixs7VApBqK7qrFkmjynUfLCT"), common.MustParseAddress("2rhYeWV1di"), "joystation4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4L6vDUiEzX9W9jWBMVLqNaRNh6WA17UTszFyHZJrT4k"), common.MustParsePublicHash("3tVXDaiQkh8upGyoXmKMDUeXcv68sWJ8wtiAULES3HF"), common.MustParseAddress("4pzXuTSfba"), "Sanghyun1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4KG4EjGsZTY9ZSW9oH4gEhiQz776MFrf3BCKhnCKH5e"), common.MustParsePublicHash("EC3tcyxf47YhuzwZsSMoJNvyrFvcU3atDTkMpenGu6"), common.MustParseAddress("54BR8LQAX4"), "Sanghyun2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("28FDpytQKMT9todKihCL5CZahRFme22mLuKubfGpAdJ"), common.MustParsePublicHash("2Swq7moGL4mbxJMmPGqME8t61sMtHx9pBN5VkUn1gPn"), common.MustParseAddress("56NtzpE5fy"), "Sanghyun3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("39CBH3jPbNdrDmDHgH44Ay14qx5bvNsYxpqfs1Y9J6m"), common.MustParsePublicHash("4gZg5pXHWZU8GFcWPe9f2ZekVdTJYwJLhRo4wraSxsz"), common.MustParseAddress("58aNsJ3zpt"), "Sanghyun4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("24rPWfkV46X1qPZEG9EkfAZ4ScVF77v3uimVp19w9c"), common.MustParsePublicHash("Lx9f1ac9ztfdRf5FtoowmouKB3GkM6z6CyxGfBGpeN"), common.MustParseAddress("5Amrjmsuyo"), "Sanghyun5")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("43mY2Z7rhDFZh9Xy7B6WXgu2m85qCSiVr1ssGVDhVbx"), common.MustParsePublicHash("KF5Rnk5rGyBgvmMdxAoMPzJyAsJmGsHosonfKYLvwG"), common.MustParseAddress("7UrAq3GpEY"), "Kdk1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2ZTaNHp7UeumcBF7xm8iiiPQJxS8sVzRdC5r68TTnk8"), common.MustParsePublicHash("4MjHNkbvQcvPK8hdFuktRBrxf7BmrQETHY9XAhr4ftk"), common.MustParseAddress("8ez4outDxf"), "OOCA1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2tjaYJZDPKdf7PJxxy8qMCAtxFENzHpCDsV636RRtze"), common.MustParsePublicHash("4dDcHNmHUPxKNkAz4AMsGZR6wMjjghJtHCFrmE4FkGY"), common.MustParseAddress("5MmG6B1Vjp"), "j1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4dDbFb35K2N5NimavmkPtekxVEAdXfFc6P3oQLDybGR"), common.MustParsePublicHash("4EuXxHPD6BWVyBsJCQer44AD2F9ysTLjfeywfhbzm2r"), common.MustParseAddress("5PxjxeqQtj"), "j2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2c2Rmy2bGFFFNLsTJzrtK5fwHqQouW3S278DPv4nbia"), common.MustParsePublicHash("4VBdKDu9Z6YxfDNeS5DESjS3v1s93x9mxWtTM2cPPyc"), common.MustParseAddress("5SADq8fL3e"), "j3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4ZqW22K86KKwf2pVRheHNtDNKDSR32tjsd2AZMnR1AR"), common.MustParsePublicHash("2Z4t5GUDeGaPZpAvzt7ac8nmf6dEwzanTUJVcbAhJ3R"), common.MustParseAddress("5UMhhcVFCZ"), "j4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("fyBhDkVYhUL82D55xK3FA6V3GByjDj7b3DaoZE3UMa"), common.MustParsePublicHash("46AK8hSo2CrJqQpcDL2ZXx43S9gYm3gCZ8L2UfSmmEv"), common.MustParseAddress("A27N9BdDZr"), "kkb621")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("BgrGAdr4Aov18h8PZNqqnECXUhqUxrSbEhDXfhAWn"), common.MustParsePublicHash("4sCEUnP2HZVDymsjrf34p6nnVdNU457P3FMfVA62zav"), common.MustParseAddress("8WC9KzaZM9"), "LSH")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2TpTwJpnNfzPLHdx7Ma4P4XH52Xqva1rqKV7kHoNpF9"), common.MustParsePublicHash("2fW1qBciEKvBBkHowYiQrYZmktrivRaEUSTUCvzmVFj"), common.MustParseAddress("4sC1mwGabR"), "THSG")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3TGJz9Pu9FWg5saE3vTKj8p7XsdpFQauPvDUJPFMki9"), common.MustParsePublicHash("Bj2ZanDtwpajRiwSaZF55F8k2DNDLhksQR6sWgjGLT"), common.MustParseAddress("5fM741cppU"), "THSJ")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("38HTRuYX4ucnoDaiTx6hufwSRSHkqQ7FBFeEU6JowjQ"), common.MustParsePublicHash("4nutYbcvMcLfHoWjJMRv4iV5uN1W68Wv9cr4XLB443J"), common.MustParseAddress("54BR8LQASF"), "forxiga1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4rzhGtacD31dy4YDeKKDssAgz6XXUfxWc1h9YGhMCPJ"), common.MustParsePublicHash("nf6DFeuHDVmiqQbGvtS318CUeH9FKd22MamCGUSVEJ"), common.MustParseAddress("51ywFraFHN"), "forxiga2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3pV9AKeEyM9mvBdadRx6EqWmPz542ZJKRexEfjJFiFq"), common.MustParsePublicHash("nU6SeLXHwinhcE4KcrVLDr6cwKN54tNMKvFc7Q9JUL"), common.MustParseAddress("8cnawS4JtD"), "omegacode01")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("48HrSKUNaJ6Ca5FrHSrXbt6NaVz3MUgtSNBdprh87a6"), common.MustParsePublicHash("41BXoPRdbc8vUUiQ6yz82QCi4k1kebFwWjybRDHbxDy"), common.MustParseAddress("9BknzdT4Bx"), "omegacode02")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("xNNVTfYr9gCZsx6RAwLyF8Rp5erhRd7mwb249q4kHv"), common.MustParsePublicHash("4Treoy323MiM2n75yTWX2jr21AJ2SDD7khi4Sf9Skrt"), common.MustParseAddress("4pzXuTSfWy"), "kellan-1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("28Hgj7Ns81iQFuxVFvNELerqFwmjusGD4WRsdST3kGQ"), common.MustParsePublicHash("4kfRc9TtnvLuNxSj3JgT8u6FvLSMRn3KA84qYoDtgEo"), common.MustParseAddress("8H1G77d4Rs"), "Lhs6775")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2NSJTtyGxVcTvyn6sGvV4xeVbh6963cUbiDMooBdKhg"), common.MustParsePublicHash("3PjwQBSJaYnZkgSnM8YrJmb3bc1jBZfwG2vRcRAKpy7"), common.MustParseAddress("8KCjybSyan"), "Lhs6776")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("xuRsQCcxb4muDwhyJ4BxFsWk9XkommhfYiKBFhHP6V"), common.MustParsePublicHash("3DqCveXAJzEJNyku8LYCXkEhS2KPoFdHWiCTWruLpwE"), common.MustParseAddress("AMtgyW4TxK"), "Warez")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3jqtm3sqVq9F9tiMu8LqQQNvbrb3oKiZogKRP688KFL"), common.MustParsePublicHash("ZURVwwegmmfLtankvmG24dvgoNkr91Lt5g4KUGRGXp"), common.MustParseAddress("5CyLcFhpyN"), "bluebird")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4Vz3LntovicYpH3mW4bzerVMtqGk6fhZLm5Xs6GVKsL"), common.MustParsePublicHash("tYu5dHYeWCjTxwKw8M6rMyDZ3P81s6AVvo3soBfSoq"), common.MustParseAddress("618RtL45KY"), "ptw1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2fJpcjBqjFomxYfjRWdCyU4vFdn7ujfVpGvcpCkPzyL"), common.MustParsePublicHash("2oJr3dZm9AoWqAUe7AozrU9Ut1WShqBKWZen1iU2aNc"), common.MustParseAddress("5xvx1rEAAf"), "ptw2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4R1CGqRme3CsmRrMHyJaJkENux5tToQispatFCFTzGr"), common.MustParsePublicHash("2H64rFR7WYuzUamogN69sC4jGMp7Xccwix6ob56Wp7s"), common.MustParseAddress("2B8tyscWns"), "G7")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3fSv2W8SafrD2vVRzxndASZT8ue3Vb8GZXHLPtJhif6"), common.MustParsePublicHash("3cJiYggSQmpUninVpwg9uuSgo3mgkMS4VZ5djCX9Mxq"), common.MustParseAddress("385ujsGRdx"), "Ck&JH")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2AfjVpFpnLtmpBTUt3QP8pz44SJZqUowqF1XFB1TStv"), common.MustParsePublicHash("2JvQpyHRaFzhY5zPyWY7pxRqA1n4Zh9PAEN3dy2a85a"), common.MustParseAddress("35tRsPSWV5"), "CK&JH1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3pNSjTs5ZJvgMYDg3o1uwgi4PhdgKuSDBxWmJgLt45o"), common.MustParsePublicHash("3LPH7aNkLdLU3gQnrtuKpxtHGct14JGJEHsQokKTASh"), common.MustParseAddress("4QpFLBMapk"), "S.B.Chung_1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3rNdcSazP6W69rkwLPwiMAM83TBgVD8koDoqEQhiyDo"), common.MustParsePublicHash("3FEy4BkeW6Cd3BaRdp67jnEhyP83RGnpQV8sJeTzNx7"), common.MustParseAddress("8AQpVg9K9e"), "seoae1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3ngzS3WUzNt7TLHVkPoFUoh1XNAFXeDdZ9DsVw2r9Fu"), common.MustParsePublicHash("3bMw2XucmzgL1bVMgNakMoiTdbGt6o9JPaeVqeLdmtq"), common.MustParseAddress("8CcJN9yEJZ"), "seoae2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2vHdzGesDjK4L3dok1HsfBp42UcReMvniZeWuZV5Zui"), common.MustParsePublicHash("C5Wz4WJHD7nca6yCuGYZQcPAZa9MdA3YHWzXMkFxac"), common.MustParseAddress("oA6H5MMBg"), "shin1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4NeKZt9ZK9xj6wPj9CYQ3L273UH5DTNgfBfb15xYPjU"), common.MustParsePublicHash("4jBRwKzADEcTxmUmkFBvZZN1VxEYJdEN4HeSZ3P8Gax"), common.MustParseAddress("gZefdsbj1"), "shin2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3GKixvkPwUpSBj4NWufctuDTgi1ieKKbFzZ1thvUkC8"), common.MustParsePublicHash("2DvR13hJp43Y4PU4DUkECmNcQyfgPmbzjdTzNCsK4vn"), common.MustParseAddress("im8Y7hWsv"), "shin3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3bB6Zqd7FnGwwj69kEK45Zwr9ydtRUTo6FQ3GfX1jqo"), common.MustParsePublicHash("mCceDCqUWbdMB7PwpVWknWz8jrjWYGFDqnmoEcCGvY"), common.MustParseAddress("AFJFN4aiRF"), "f1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("43iyemSdLWocsdYrvMtYYHX31FtQHerGRq18YwTjEMH"), common.MustParsePublicHash("3rTLp6vvrrWDcuKrrzPmtuPfi4irJRoThvYnHJwFwT1"), common.MustParseAddress("AHVjEYQdaA"), "f2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4G88u2RfQf3mfGDFmZ1rKhxobRAtx6miAMaEk5xzdrP"), common.MustParsePublicHash("71RjaK1saz9y9WgiVJWtBB5tAjVf9Wnp7uvwY4ggWU"), common.MustParseAddress("AAuHd6vt7V"), "f3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("48XdmWLg2js9QykJtvbtDvFJDosmXCTPuLd98zCpWsT"), common.MustParsePublicHash("34fQJ51xviEMtXWGWEDvgtUkX7ETMxHYGxcRN3YBfpA"), common.MustParseAddress("AD6mVakoGQ"), "f4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3sDov2Dv7Y99F21crzprbSTs5Ue2bAuha17Q2rWL2ZA"), common.MustParsePublicHash("4jujEzH6tRPRRXrdpD23aN8MuPNVM7Mjo7S3LpGvuu9"), common.MustParseAddress("47EQNLkFff"), "Moon1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4QArarWtwjAkaCPcbYiLDr8WMWLEqTYYEXMetYqQiSm"), common.MustParsePublicHash("3bemcQxdk6uoA1mSW61sAD3UX1tjURZNsxSHUg4W8Cu"), common.MustParseAddress("49RtEpaApa"), "Moon2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4HynSpG1ZAyVky66ZPrJ3YEktoDz4aj4iHKMs5FKqr1"), common.MustParsePublicHash("QVxc2Sdq3f8g9kmj3dApruyKnUptCvCDvsrkWTMSZX"), common.MustParseAddress("9uWvXk9U7D"), "jin1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4dW7SKEsNU1pSQZJ9cNVFu7Yvk2eNvrDSD7CSFSkz2T"), common.MustParsePublicHash("2QM1z5H7EuFeUvVShyh2TP9QpEdU2PeB9ZmmGTtF16d"), common.MustParseAddress("6TWCL5y5Ae"), "jin2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3pwviwPSEmgH4H5iP4BMLKm7N49e9r7cH83GuRmdocy"), common.MustParsePublicHash("4G3L2ji58HeDRnrw7vQq4i5uvkmzja1bi6Say7XPxTz"), common.MustParseAddress("65XPdHhudJ"), "NamA")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("PCWmSTULDK8Ph8DoVHQZXg5MDd8qjXambMwVWwP8Qy"), common.MustParsePublicHash("2BpujCBgURt7B9ciiiFTTgUjA4YKgrDRy2QjNozJJwu"), common.MustParseAddress("7NFjDbo4rJ"), "NamB")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3coA9KeqVELfCiDEtoXXyQLbNNarpcuSRBEmfewZf23"), common.MustParsePublicHash("4ontghDjuYrwA6c6ZkcVyyc1jqUP3mVgBzmPKpNCPbu"), common.MustParseAddress("7L4FM7y9hR"), "NamC")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4qjAfodjxyMgZ1h8d1FE3aQb6z97egJdh1MSsd45UWH"), common.MustParsePublicHash("4RzrYHZi9CqwQAdzCSKhyf6a7ZoV8FRWMio8k7fp9va"), common.MustParseAddress("6qV12tEEhS"), "ho1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3UiC1e1P9VPffMQmcqfs186LfGCXMeSYkVCeCEjHLAJ"), common.MustParsePublicHash("gnXE4DYJ1wfkdNUpnYbJ7QDdfUQrxBYyYahha3xmfx"), common.MustParseAddress("6itZRSkVEo"), "ho2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4KEyo6fTKG2vugKrGckNKigNztejHN1Bt8Nh4sFnyB9"), common.MustParsePublicHash("376Rxx7Exj6BhfyACuAjcMogKJy9cXdkYNXJqoDS9ut"), common.MustParseAddress("7DTojgVQAM"), "Sohntae")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("34CJNrVAdqjjaJaPCw8PeyVgP8gNfkk3W5LRQUsZtbu"), common.MustParsePublicHash("gSrCZEED4sGKKsR3i9ZTg4FqjyoeB6maju4QX5yRbR"), common.MustParseAddress("2QKnCka1ih"), "Parkss01")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("5Rg6Ge8aAhiQRANBUxBHmt9YZrJZVitPjoHx7uzA3x"), common.MustParsePublicHash("4Fx7hjvGMbekVsKaB3QddJnKcZKQ7k2ufdYMdrJ8fEy"), common.MustParseAddress("4wayWtvQuB"), "hongpa")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("wquMaoJBGpEad7XBUWAfx2gyW5KVwF43K9T59f8uFi"), common.MustParsePublicHash("U1gKcpGtE7zgkXrYvinjW22CdZMJwQWmjhdZXvqCA2"), common.MustParseAddress("4ynTPNkL46"), "hongpa1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("9yQWDQHYkDxz39LqAeViGXv5Se56psd9oZHwGS4nzY"), common.MustParsePublicHash("45raFZTssPb1QDhGFieTps2gvgEmh2jcqggdDPGH1Lw"), common.MustParseAddress("4uPVeR6VkL"), "hongpa2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2TncFhhNMUg3qdtWi3QiULBCYzSXWP3BjkpqkFbUnEj"), common.MustParsePublicHash("MiCqWNWE1CmAbyXUEUyrKdMScj6bdXzTdiXmsroWR8"), common.MustParseAddress("4no42yckHf"), "hongpa3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4XgWWkAUmNYse77ZmuHzNGuM2VfmQ7dGhR3HMYZ9ros"), common.MustParsePublicHash("bfz56wXDjUsjorMCeZeS75CjPDbBKrQ2w8UMnKYRq5"), common.MustParseAddress("4pzXuTSfSa"), "hongpa4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2ZyFm6Q9mJJTGmVW7SWtycm7NUrxkzEk3pnLEYt3uZu"), common.MustParsePublicHash("42nWF8pmuDyEg1GEYCootsAUropSBS8YGGbihMMYQat"), common.MustParseAddress("4iQ6J1xuyu"), "hongpa5")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("RFndB5K7AYB2iUL16DjgSEByPeUDqs5w8figk3akxB"), common.MustParsePublicHash("AhEHbbm5eYRW7ZHXR2QhELGLEJiGjVQSYoBoFSjxM6"), common.MustParseAddress("4kbaAVnq8p"), "hongpa6")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2nVy1Pzoc5tEN3oLvFo4qJy3y5dMdKGtKSqKKNxoGZ7"), common.MustParsePublicHash("4G1cfmwDzvzWyMYSC67p6mtzYPsvHhXTrszfB4QA6Za"), common.MustParseAddress("4e18Z4K5g9"), "hongpa7")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2CYd1NTRTKZ1XQABXVMejnqdTCZ1Zgyt9hCSUCa86FX"), common.MustParsePublicHash("25gMGhdag7iX1MT25QazKmrR7rPhJWMBTCpS5zmaGcq"), common.MustParseAddress("7L4FM7y9d2"), "tkwon042a")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("VoxCb1j8S1cbht8jAXTzs7RY33AVzDgBDUB44ymWTf"), common.MustParsePublicHash("4TtvSDdHv8VmL7PsWw6hPGCc5VWaCxfWKegUArh6zdy"), common.MustParseAddress("861rkiVUhA"), "tkwon042b")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3FUUD72FpvQBhmzFRoTtMMtsPFR5U4CS9jWw3Xd33KB"), common.MustParsePublicHash("PmdnBSc1eVsLnNByqbH58KZDMtidBxuXtV717cRYZw"), common.MustParseAddress("7yRR9H1jEV"), "tkwon042c")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("iTauni5G1oTkQ3LrzA4zKV1KYsjsWPdwhn5Jj1kntU"), common.MustParsePublicHash("edv9AcK9o6WyHQD1LyX66yYm4PexSCjwzBDBt3Rof"), common.MustParseAddress("51ywFraFN9"), "alex16")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2N3B78xXPcYK9h7kywdmRGBkTy4moi9BtFJBYhwoBND"), common.MustParsePublicHash("4rAbQ6TqW1C4hXG5PCQY7A6rWb8gXRokrD3BMYLnEkL"), common.MustParseAddress("2UijwiDqx4"), "1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2CkVg2fCSokt5Y8Lc7RHauZQ7xp2bzGo8Ttbx7dHiVu"), common.MustParsePublicHash("4NnPZUBJSdKDDhzTedyCqbYbnmMx2eftrvak8LMcKqw"), common.MustParseAddress("2N8JLGk6VP"), "2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3YNMknhKxQ53uu8BA5zuVTWXdnz8cjNvX1bip9oB9po"), common.MustParsePublicHash("3CDWrJ47ugUVhv8go8MoJKYkMVbSkokB5yTkDebHAnN"), common.MustParseAddress("2QKnCka1eJ"), "3")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("mMcZYmkYxzyKdnJiNSnpFnaJPZYehABDTPaeJvbpes"), common.MustParsePublicHash("zWj95zk1z4EyMm6tuT1FX9adZgMii8mLVtCL788hG8"), common.MustParseAddress("2HjLbK6GBd"), "4")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("5HbGA1nZfctRqNJtsy2VYqh17pVpgvxxFYmcSJFjS5"), common.MustParsePublicHash("STqjuvCrHkXh4FmqMNrQntft6FHSi53ctkqPpivErS"), common.MustParseAddress("2KvpTnvBLY"), "5")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("2G3T3vDkevEWwgZ2QR7oBSELiZsCpnD8KHpzYEt5QEj"), common.MustParsePublicHash("3QVHhy4sSzrh41W7C1tFudJ1LdbLKoispHKgZhwjuZw"), common.MustParseAddress("95AMPByJev"), "kim1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3mVLm4NY117nKYkZZSfD3rsq5VnUULqWnbKJ9yMaKAw"), common.MustParsePublicHash("46Q3jEzv6AbB2kG4oNpJFRaZXHv25H55P94S3w6mGHd"), common.MustParseAddress("8xZumkVZCF"), "kim2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3KFdDz4WBcb5p9JetuVqTXMtT6R7UnRbsh8DsDEbYjq"), common.MustParsePublicHash("smbWSzSkjBPSEbdKSb78mVaYt4SeDANQaF2hXoC3ow"), common.MustParseAddress("5vjU9NQEwM"), "ckh1")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4GZM8GmY7DoPBRtuAGpk3yKmXdLb9Nwpcw2p9kYxDhK"), common.MustParsePublicHash("2t3haUJzDMXqYGPJRUHie86hSyLHUYNpZyziiVdHG2Z"), common.MustParseAddress("2SXG5EPvo9"), "Yongdae_Kim")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4GmYu27jXXmXNWX6FeD4WybGDRSsppjQByHJQtXiGyg"), common.MustParsePublicHash("311QvrZUYFaKisd97E9TtktkEWxxdfBpcVKmkv5uKXd"), common.MustParseAddress("9bw5ZuY8on"), "Yongdae_Kim2")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("3Gt4xUVad1WjKZJvsk4mmswGLEEdhXkFqrJ9kdcrBEC"), common.MustParsePublicHash("4tqrj77F3y9kKudMhs3AAYQZA75Z38FZ13ppF6jTjGy"), common.MustParseAddress("Ab5aCP1xsb"), "kimxx")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("46LcPcCTPknU8uJvZYN4Y4Wa2Pcdmdg9qcGLe17VhCK"), common.MustParsePublicHash("27hk4nXJQFU3HgDPceg15hS66nZ69tGh1z1FBydSLtE"), common.MustParseAddress("8TzfTWkeCJ"), "Kyg")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("23pMjeYJHb1YfYajYX6QzkLLUpPWmTRHqpV9pAaMdBe"), common.MustParsePublicHash("2hKuKWGadZeyLm5ZVbiVKstxabvkqz9bSibSWrovzhf"), common.MustParseAddress("2pW4n2f6LV"), "zutenbe")
		addSigmaFormulator(sp, ctw, sigmaPolicy, alphaPolicy, common.MustParsePublicHash("4nPVfEU8bBp7FhJH4mPBo18u44pTaHXsWy4R1yrP1qd"), common.MustParsePublicHash("22FVTgdEzZGhwWsFLAM8oLxDUWJLcprJ1mJmzYZiiat"), common.MustParseAddress("2rhYeWV1VQ"), "zutenbe1")

		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3AGiR9vxBe9CPdChYa3ZbXpZ9DAoY8poof487kQFhY1"), common.MustParsePublicHash("2C6BsSU5toDNDLfpdzLVyajDJwBjudyDqnKscFSLNMn"), common.MustParseAddress("8RoBb2vj7p"), "fm5")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("uqcG6N5MQxR7BK94Ncq1Hbr7pMReaH3pAxZh1dJ3xD"), common.MustParsePublicHash("45kGmL6QSqgFdnVvPsGdV4xudxuWjKAqePCa7ynQYyN"), common.MustParseAddress("8PbhiZ6oxw"), "fm6")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("ELqktgfybkVMo3dJkuZDJasopCexbcHYpUugFG54Qa"), common.MustParsePublicHash("2mQRjAajzp29dQLqnAQ7W1wbwuR8NiJyeptHghfVHi"), common.MustParseAddress("4XQgwcqLMi"), "Seo2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4nkM2h4YxH5gVJNpotkkLfuMUHun2vhqhS17pequrDU"), common.MustParsePublicHash("2YK8gdJdByZJLHox17VW43FxGQhWvZyq535PXBroQ19"), common.MustParseAddress("69vMNFMk1m"), "AhnYH2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("oAnPu3d6X9obLBh1a2BdtDf4pV7mUzGpaY9aagd6DP"), common.MustParsePublicHash("2fQ8zTj9rqNycs4uN65ouJuU4vS6ovMe9KyndUbfsXT"), common.MustParseAddress("6C7qEjBfAg"), "AhnYH3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("suPskquzrvLWUnqZJMBEuTLzsyqr4qzYpDhuZQHGBp"), common.MustParsePublicHash("3idSR9ooK1N7HnBsDn7Rj26C1NU8G441ZZKCaUjemCq"), common.MustParseAddress("58aNsJ3zk1"), "kyounghwa11")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4iLYM11S2FvTHRP9zBLZz7dLuxV9WsATQNXYvCj8sZB"), common.MustParsePublicHash("2VL92hZFoERHKTkQn3ArndVFtQhXxKzFAhSrACVEYxy"), common.MustParseAddress("7bScSUraE"), "tree2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3DWcQezqF8k9b8E6wBSGKngBh8ez3H2i3Cg8SPr8K2q"), common.MustParsePublicHash("2PAgrdrYvhvYCSejfHy17LHVuqQmFZNxuGJDZ9upcMx"), common.MustParseAddress("9nvUvJmj9"), "tree3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("WMH7Y3gNseSE4UWQfdQYRLRGtymU2TAUw73huVDpRb"), common.MustParsePublicHash("Zt6GdJsNoxQKgPqwhrPaie1RyQtdUZbRzvshr4STxR"), common.MustParseAddress("BzQMQ8gt4"), "tree4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2gpStGSG4TQV64REcNky5ZESRx3gT2Ztijpvmn8axrF"), common.MustParsePublicHash("21EYQMzzjaeaxWpD6dCCG4tAcAKu9S6V3VnhhTU9m2Z"), common.MustParseAddress("7ZF8ZzveYJ"), "Jaein_Formulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2u3HfbAbyJsTVYB3d9Mk4jJjvagCzQE6H6zbKLBWQr8"), common.MustParsePublicHash("rbqYYydgTG7fQasti7pWj4udw94cVCTRBFXD8MTs5j"), common.MustParseAddress("7SegxZSu5d"), "Eunu_Formulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("82bkE1yUoPgdjBEWv3fFjHKMMCCEsWvkVcGz7vfztu"), common.MustParsePublicHash("EDH8ctQQbspb4WvxDzP2pkss65DetV73VH1SdGyByh"), common.MustParseAddress("7HrmUe9EU7"), "Seojun_Formulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("oBGKoZVvqx4oWkJSAZLfBNLRouX4GwDhHs65pWgDHd"), common.MustParsePublicHash("3tEJVrN7cFrj5oCSaD3HAdzEaB74wfAUrbF5juY5dYM"), common.MustParseAddress("5CyLcFhq8i"), "Sanghyun7")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("69XwnpALiLhFktaw92HT6uhaJTPw8J9jVYCK7Q38k3"), common.MustParsePublicHash("wT7PUVeAvpXQyNNtpAPhgdSLr1GfxbuJiNiBdTEkpj"), common.MustParseAddress("cAgvgDmZZ"), "Sanghyun6")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4ChCScnGz7tWwFTyq7VaQR8UKAmbuVkhKpejUtKVSLZ"), common.MustParsePublicHash("3jdN1MHLfXEGHzRGggnVqUsqBoJ1Jzf55robxddPnhd"), common.MustParseAddress("eNAoA3giU"), "Sanghyun8")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3hdbg6jzmgVwR9tUnY2WdD64NHfRbT6tvha4Gma1t3m"), common.MustParsePublicHash("4FN57NwhknpH5MiqwiWd6h2HcbWFMTfrjsj3ckjZ7Fm"), common.MustParseAddress("9yutGhoJQy"), "gazua")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3z5Tzky3yF2mYroYFDzjXfFrad76azJRfRFqbAgRZw"), common.MustParsePublicHash("2Hqzfn4p9jG8qUSZP3F1cw1gQW2Lbbe8LY4aU8p4QKY"), common.MustParseAddress("9q7xnnVdoT"), "gapnida")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2PHwbsKHy4wbToWX5D6AaAfzZbB13gYDrW5W5aMFYyw"), common.MustParsePublicHash("43bqba9DgZaGy8AM7JByz2eMUzHojCjdAty14ynVSic"), common.MustParseAddress("6VhgCZnzF8"), "LSH1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2WfRVgTBXc9bfYUUd5835qzhFxPJ1CABm2g3cKNFPC"), common.MustParsePublicHash("4DPFh6xbPhdKRsxf5ucgaBNFWbBk4N91QuVXy5ceW8Y"), common.MustParseAddress("99ZK89d935"), "omegacode03")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("pncYWyU3c5QHZCERTAZBCSxzGSwQFuq4uGhEhyWh3Q"), common.MustParsePublicHash("49Uew5MXK4WEg42hFmXqG5oV2EcWBzurSjhA3pCa8rX"), common.MustParseAddress("6RJiTc9A6A"), "kellan-2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("uLnsGCz81GhJiHLwcP2RoLczdm7rhbRMYHZ3xqpyke"), common.MustParsePublicHash("4dFBWDnXSTyJH72oWxFtGt2MBBb5mjHSu9T6Q4S7vMg"), common.MustParseAddress("7fqaBSQQ9m"), "kellan-3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("g9756mVaHABM2ymRU5Wj8mVJMXs3LzqnDim6YHGzzS"), common.MustParsePublicHash("361yyexBFiNTdVGr5JXYfCyRGFFfJARtiaKNfeyL44v"), common.MustParseAddress("7kEXvQ4ETb"), "kellan-4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2UgWKwWpQUAjtu81A9SfdbzZFk8tfWt2xZL92zdm8Pj"), common.MustParsePublicHash("ixLujo9UVgqjGU76cqduxGrAfRTkhWiDuevqXLhFEN"), common.MustParseAddress("AHVjEYQdeZ"), "Warez-2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3oeX3LvAAwJV7zudb9bWpbDznd92pXexz4RBB9wefbA"), common.MustParsePublicHash("3zo3zMypsbcJCEBatUziRt3V5FwhzCBAS7dfvmECLhp"), common.MustParseAddress("5vjU9NQF1n"), "ptw3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4dSLSq1wBrWQPxz8gKoT2t6FM4kGxJGUNzct98YpeCM"), common.MustParsePublicHash("2bniQeiFWHUWJGggvfqzLu6tXSVBi83fiJPqoyKnx9Z"), common.MustParseAddress("5tXzGtaKru"), "ptw4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2aHcbYM5nZt41h36H5mAMwHCz1pnrVFg3v7asVSk3Zm"), common.MustParsePublicHash("ATe4NzB2dj7wAQuyHJLr1RcFBLZL2BRvBUGVhPisoB"), common.MustParseAddress("A6WKt9H3oj"), "f5")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2uk53XxCGTUMVwY8fkRGTCGZRAyANGAZxpZM573p975"), common.MustParsePublicHash("44gwjXmVWJgKLuCUTbZRL87n4GATajs8NC6xoLaCRWN"), common.MustParseAddress("A8hokd6xxe"), "f6")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2iateVDG3WYAtVsnQUHizTeKiihr6vnHJ2fFU8YPm87"), common.MustParsePublicHash("31paG7cWgZbiXzcpescmtke1RLPoLMEkqxaLGCzqJME"), common.MustParseAddress("7HrmUe9EYY"), "NamD")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("D3M3Ut1UYix1Mwt7PvTatkriH84t14pdhcmWDytv6E"), common.MustParsePublicHash("4QffDFrQW9ggjZeFDPAeuH7ieFkQJXW7p4z4Zwk3Mnz"), common.MustParseAddress("7FfHcAKKPf"), "NamE")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("DB1DBmXby299tVKSwmCURSHsGkFMda6rdmS4to4zV8"), common.MustParsePublicHash("3kVq2TjqTyHwKM2MnNGEErxbCkxb94nGa8fJqEGKDB7"), common.MustParseAddress("7DTojgVQEn"), "NamF")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("31yavHbZ2mcRWBDx9bXvRT7ACGPZnzPtKJaEKWYS85A"), common.MustParsePublicHash("2XeeaZ9rTTKotg1ov6Snhxu4aSiQEjpgYCi1eokPTD1"), common.MustParseAddress("8CcJN9yE9m"), "Sohntae1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4f7Rjjj4Ycr6ihFPMeGd6cycy1YE67Lx5896YeHDAj9"), common.MustParsePublicHash("2ocPsQezjPGKFn6EYqMX5ahqUcH8vQRP15BuAx2xY9E"), common.MustParseAddress("8EonEdo9Jg"), "Sohntae2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("GpSg37pqs5EbsvapNC3cVgqpcsXujojEwV4rhFR23t"), common.MustParsePublicHash("4nxwvzsE9wf4u2LXrBg6ujZFbmhfymbicQR9RJo8cny"), common.MustParseAddress("2N8JLGk6Zp"), "Parkss02")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4freLrmNiFrbFhLbemdixeD2pH7oRvpJSRWwxtUtHtS"), common.MustParsePublicHash("A5qrF9tK6xAVuKUBgDuFLusyiizhzR8H7FH11VyzhM"), common.MustParseAddress("VaFKEk27N"), "alex14")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("VhyUbcAspCsgsHFudPFC7XypDF6Eerfb18GhJQtGfS"), common.MustParsePublicHash("3W5nAwuWAKUocGufhLStCk7DAp83E9U5ytL9ZCtK7GD"), common.MustParseAddress("5jk4nyGfCd"), "kima")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2eLcvHCWW7CFVK1RBi22oDZ2LgFFmC5gPdazwwhXePR"), common.MustParsePublicHash("3GKSnesDZ6K25f72jz5BhPBXS1giWK9MCuXoiyxasyS"), common.MustParseAddress("AdH44rqt2U"), "kimb")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("kRLtQiPvXCERyJwXGvwdyZEVVJbV4ARJzqyTMTTAPL"), common.MustParsePublicHash("3bxkZpJJiFZ3pPxQVq2yVuTEUa5W2gUhqCosPeLaF7q"), common.MustParseAddress("9sKSfGKYsu"), "fletaairdrop")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3w8xcjRhHCPAKRY6HMnBsy5cNrGC6v8HKCWr92Uc91x"), common.MustParsePublicHash("2VXBzNXAFc5HYponTjxAD4DEBqfxB5njqadfib4mem6"), common.MustParseAddress("7kEXvQ4EPe"), "Sms1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("22Mijt9whcD79GR3TdUnvnP6cytdREHq3vEJM5oZAdU"), common.MustParsePublicHash("2NCnuEmUji1sjf1qQhC6PSAmvmkQgkRU3LQQU2nJn8p"), common.MustParseAddress("7i343vEKEm"), "Ams2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3BV4XVYE4kNLiP3AE3VC5RvupXZFUZvxd2NTVfQBjJj"), common.MustParsePublicHash("pJ5QoE4q9H2EtpYKXa3ikz6CMiwSk9yXVVGcWj1h2D"), common.MustParseAddress("7fqaBSQQ5t"), "Sms3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("pQvn3Qorik4wvfkZF4AhxkdxE3QAHDucXfWXZuRWZa"), common.MustParsePublicHash("fpGUmhxXrnAcnhimkC5Gtbvyvry31SxiRVTqpH6Aq5"), common.MustParseAddress("ASHeiTiJBg"), "barge21")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3SxjY2Sx6yGC48yUkDDbbzDQ6VBLXqy7JkpkfwPSsGC"), common.MustParsePublicHash("4UnTvXHRpDzo56xj2prZx8x7fCf9zupEkBgvLsMtaPQ"), common.MustParseAddress("AKhD72EYj1"), "barge22")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3umaJkPK1p4vHjCwEks2VTFFTeiunLCCszNxwo43Np6"), common.MustParsePublicHash("2ceviL8yCT8RdbtSaSGdZYV4CXRBpnXAnhrEDE6pgqt"), common.MustParseAddress("AWgcTRN8Zq"), "JJAJJA")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("32T6xLbQJny6mvJVN1cnRyUtNsd1uDDwNNGCDDFu6g5"), common.MustParsePublicHash("bD6B185sRV1fPbHVX7u1K9pFYYnKNHti3YJa5mYiEy"), common.MustParseAddress("AUV8awYDQx"), "JJAJJA2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4nTT1zgaSc8ybGjMSThFWSPb3eRz3cj2yEzHUy1hvxr"), common.MustParsePublicHash("2tacGrMnEgEq7th9Ce1GewEq3kfFdEMdHsE3wRN6x1b"), common.MustParseAddress("9XY7pwtJaR"), "black0158")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2WyXkPxqGMmQkrYc7e4LFnrk88UphDwTGCuj21kNfg2"), common.MustParsePublicHash("4tR6UHJN7CUja7WRJuKmpxm5wPbocmu8K8vPXHLzmUu"), common.MustParseAddress("XmjBiZwBN"), "cho71")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3HeVHLbqabCdLHY56ZryyaXBG9W88A2aMLTJKAuBsFQ"), common.MustParsePublicHash("3agxPkKaexT3otJ5hYjt1kTQhBp7aRtZYAiEfXRU4jW"), common.MustParseAddress("VaFKEk22V"), "cho72")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2UPwzfxYJNWKqU3dGehRpapa2Bz1DRggYUnNhj4dTJp"), common.MustParsePublicHash("2KXKjwpzYDJhN2k2a2R76EQr6jfp1n5wjTA6RH2X3uS"), common.MustParseAddress("8YPdCUQUW4"), "cone")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4SiqVJfS9x7FRhyby6UqyUzeoGSRRJuXMuDJHULpMkp"), common.MustParsePublicHash("4H2o51pNPVXDFMdeSetqkfiP1Vp2WaHMXrceoFKZXrr"), common.MustParseAddress("8RoBb2vj3P"), "cone2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3QsFTKiA6wEUj2RA3JrhsckYCykoaEqumzEdFmZJoP9"), common.MustParsePublicHash("3gXmZ1rjv43xakaZz2eNfr5morCurwXeoungY3pH4rT"), common.MustParseAddress("8omzHqBtec"), "FLET")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("41JdbtL1MAsWNF2ZuJFXv1cecUuqwgfrLi4Q2RaWwEN"), common.MustParsePublicHash("43chxQaYXXAR7sUreSUDmpkmrajYrL9QaW4poJfcFgm"), common.MustParseAddress("3AHPcM6Lnq"), "choirj")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3LGS3DAXUTw1Gb2tmmwtXZJhWvh2Z8jcbWRNZL2BccJ"), common.MustParsePublicHash("2FLXynUj4N4mRV69A7e3kbcDn4ARYQLmaGo34UL4JSK"), common.MustParseAddress("5Amrjmsutt"), "LEE")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4MerKWinbtqNQb4NdmDgEoMRv9xDwQmkisghZ173MtD"), common.MustParsePublicHash("4e5HoabvTmDUJaRm8R7x7XydBVD5CNL8m1ARa7QHYcA"), common.MustParseAddress("8YPdCUQUf8"), "Hanui")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2xaQRjHGoruaYNHwjShxs1C7yoS24U4nnCpKGdgxg9m"), common.MustParsePublicHash("2qzUH5LTfohYo5ZYyiGZkurY1BtRDdTxeixKasydeCW"), common.MustParseAddress("2nJauYqBLS"), "kang1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("MUgXbLT7W425TLzF1EjuyqCsdB6U8nHMpAbpKBEHjy"), common.MustParsePublicHash("2pnj6Q5Ar9afrXKnScPUALJNBkzomM8ntifvWvnQEJZ"), common.MustParseAddress("8tAx2nqj2o"), "kang2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("kyEUPmHSu2phFZfpAbJvbHiRtccckf6cCirrQa5wn6"), common.MustParsePublicHash("3aoJYhWrUk5WVe4UsKXwzfebP9GcG4jZ38x3AviaVhc"), common.MustParseAddress("42qSdP6RMq"), "hit0ri")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3nN61TkVntzd3nZbCL7NEV7MYvrjKii2EBTEBxLuiZd"), common.MustParsePublicHash("3w4LHAibSJHM1GdnksmQraZtYZUqbVyZATkf5AoxnPa"), common.MustParseAddress("5UMhhcVF43"), "Se-Ah")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4efCFLGKrjsMDPoRtCbLXhHtvJZkYhmvAWjdTcy6KLM"), common.MustParsePublicHash("4tH55xAWor2ETWBRy84MEdTbHhzsryj7wubYuncMk4F"), common.MustParseAddress("8vNRuGfeBi"), "Sun-Hee")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("33eQFdQ72GFAitMYgHFHUjAatTiCjUJUS4NQiG2E1LC"), common.MustParsePublicHash("2neWYucSRyZoykUVUc1bWBaBkpBuBAJcms58KhcJdmm"), common.MustParseAddress("JaqxqcSG6"), "FLETA_YW")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("C6ghrKtzKApCmUdPgbG84Agw6QcXtNbwxJ26MgwM5w"), common.MustParsePublicHash("3EnuvF8jPyZrRjxk2PmLKC1NVNEG8QWzjQkUiEAk1uU"), common.MustParseAddress("GPN6MnX7D"), "cywoo84")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3iRnFxtnLZ6hFr28WBxpe389PoVmceruExrFVEW7jFe"), common.MustParsePublicHash("3nFyiiMWe5ynJW61FH7MMQAsPYZCaytm5GJv6fQ2Cud"), common.MustParseAddress("EBtDsxbxL"), "cywoo1129")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("mAEvSHj5Nxo6MmTuR6Ft6C7DJo8V2E4MuUNAxdcwK4"), common.MustParsePublicHash("S3TExQo3UJ9gJV46RB9KDECww3FqnT2Vk4JV43cRqK"), common.MustParseAddress("9G9kjb6tZw"), "woowengweng")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3WXKTv9Q3RG2LCy2rqkgGqa1d8nJqQ6mmHRjUFyi1Y"), common.MustParsePublicHash("xKbjsVYgak3CDaSLMAfA2vxvvpBxyVWodq9uA63mgQ"), common.MustParseAddress("6a6dwXSpYt"), "dlrhsdbsvmffpxk")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("yfei2zuzhcquc6ZChMzqRi1QeHtSQeTugCFjPsC8nE"), common.MustParsePublicHash("3x6SPfMySaykjSLL8AUKxhwEZueX8yJcLxHB1TaAgHr"), common.MustParseAddress("9BknzdT47X"), "dmdkwon")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("21sBX8ukWPKwPo1YMK7MQFfDa1u1VFv1z5uqu5WXcPG"), common.MustParsePublicHash("2avCTN2jvKr6cvwzMY2o9qWwqeskmsTW4dsXePc74Nc"), common.MustParseAddress("63KukoszUR"), "honamy1948")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2ftgbVCFxjpw5NukQhrYchTMKQdFRaNjnTbHjPNmfJC"), common.MustParsePublicHash("2Crftm6EMguecqRiA2nYLS77MN3WYAVLQCJZKghYe7z"), common.MustParseAddress("4e18Z4K5q1"), "drilar22")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("rofXku6cG5JKXsrYTiXkXQYyMUQjN8sRnvzV36VMH3"), common.MustParsePublicHash("Rqrb5oJ8pN5TyiYQRhzFXNKJKU4CBrdXAARe3rNQJa"), common.MustParseAddress("9gL3JsBy7Y"), "JS1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("AHBsNPXgs5BZN6HXeNRQXohmwYpB8B2Arg2ckC5y8L"), common.MustParsePublicHash("3hVuRrcwADLy4ZoBPho8Aaz3FWdVRSKCbmMBtHX59Pp"), common.MustParseAddress("3W4iSfXbDb"), "seungwookfleta")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4BpzM5GwfDJXB6gSkREuM67paJZSbv2ZwRa4BZHTA6m"), common.MustParsePublicHash("4QBdK3qEiFvNoHS3FdTY4oKZ99dnMNnpiquTde3vVnG"), common.MustParseAddress("3jFbfYV695"), "seungwookfleta1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4WQNnUZYQJBu51qswhJ9SWPLcbdwqdv6aF7phR9FjVH"), common.MustParsePublicHash("4kH5VLQwGSSFwTt9aCWufsHqn4QHM98xD6QfUtTJZcC"), common.MustParseAddress("6TWCL5y56D"), "two_stars")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3PLZjCs6pEDVkJ3d5JQmELhfDjo6qNfMUEExvJNYb3V"), common.MustParsePublicHash("3H2egtZC1hd6mGZLExyFa4tpqMGBSziNEpuNwDw7skg"), common.MustParseAddress("5mwYfT6aH5"), "key")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3eVZCayrQYdreoWketnc8p8u2BDX8UGToiSaAhsXeoo"), common.MustParsePublicHash("45EmHSGtB4Lt3zc4nTeinPAtuzKy9DxSxSTkQ1UuEho"), common.MustParseAddress("5WZBa6KACx"), "key2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4YddbMvwsebswBJAG8BNLxRozjDAe6tWCU4rRvX6Eb6"), common.MustParsePublicHash("1RJNRkqzUEumcp6fVfXLfNmoCzeyLosY8pRPVchjus"), common.MustParseAddress("5PxjxeqQkH"), "key3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("Q7WQmDKR7N4MNxCNMwq25XhyRMANSz1HbdvS1pk4zZ"), common.MustParsePublicHash("2KRGike9SDh8Bf439coVgGsStPdMV6CNQtQWaB194gM"), common.MustParseAddress("2QKnCka1oJ"), "EOM01")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("9Xsh1YoHmHWC3MWoUambfLDYC6k674JJ5D9BdAPWj8"), common.MustParsePublicHash("3umfk9y53L6xM1w7ywZUSFztA1pQNLiHeRWvLng3AU8"), common.MustParseAddress("wx1kzf1sb"), "haley-4our")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3YMr9kbhMjXp3TQJ1rVeYB9V13tZR3Lazi8ti72o3ov"), common.MustParsePublicHash("4VfTwvoxktbmikguVnRnj1GTE76fSJD2ZQ2SUKEkpaf"), common.MustParseAddress("63KukoszQ2"), "GerrardFormulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("YhTeEx1sCVC6kMA6ToTsm4mRWUoojrqWYpmjFPX2ES"), common.MustParsePublicHash("rJN7e333rJPGL5BnLgcJdPwuGobQBK9C2EbWTDTQvz"), common.MustParseAddress("6x5SeKhz5g"), "HyunjiFormulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2AzLa45P74NLeV51ZNr2P4KPwSPsnkqQrkuxcbksbdC"), common.MustParsePublicHash("3LmZQrdLR4HvccaogvBzhb9phDZh7AeVonfwcTDyMK1"), common.MustParseAddress("3oeZQW8vN8"), "vanhhhhha")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("7PBtfPjyVwsiAjfZACVEwWG3eLaMgrDbUjwqXJMxWk"), common.MustParsePublicHash("4uGhquGT3QpS77hjPyeLdgBGCj6zXNGyVyjnCsUTNs3"), common.MustParseAddress("z9VdUUw2U"), "hidaegu")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3VcEsxBQb85jDnmwAbF26CYtcHwCCs3xH8e76dnLgvR"), common.MustParsePublicHash("42AUcV2aJL2zRRnCiwPjeFpyJcaeMkLMjvas1vfwMPq"), common.MustParseAddress("8qyUAK1oj6"), "Hangki_Choi")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("332vLWXv7Hurhc9cDTuKppgHEHmndSWdTd9mGcHffA"), common.MustParsePublicHash("4ryTKVCHxeUv4VRVGZP8p8UTqMtxw7uHbqVZA48W7z6"), common.MustParseAddress("8maWRMMyRL"), "Choi_Hangki")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("49yrXngKscYdN5HHPoPT7cKWoXJuNa7XGAhAVs15qxD"), common.MustParsePublicHash("4iyzYJhEWKFZoBdvGv1aScmu216seaAtnffxKmdbK2f"), common.MustParseAddress("4G2KrG3vDE"), "kongkk")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4U61EU3ibf4FCrMza87JvrF1JmDLENPGBjhytYUpoqD"), common.MustParsePublicHash("41yjDLvH4M8d8Yn2bnwZRSgfgRaecHkB72CGbdCQzye"), common.MustParseAddress("5KZnDhBaWx"), "alphatosigma")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4bK2XwwqX3azrfZ3TXvJ1sS9LWSp2dohaX2dHHHRtsw"), common.MustParsePublicHash("gQe99ADqgHwtJcqaragD8UYmr1QRTsWZ3uBk1oPVuS"), common.MustParseAddress("7FfHcAKKKG"), "HYUNG_JOON_KIM")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2U3JkoLcftEaVaqQCzV1zxUhX6FYjYSaYPMh8iJscDR"), common.MustParsePublicHash("HMLUh6TnWYZFnCzU4CVmU1sxkKmcBeRse4CWYzoyx6"), common.MustParseAddress("7pdVfMi4cy"), "HHYUNG_JOON2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3TdmR7dAHXoHR54UdB2ezu9uFTEGrLKb9GbC7iVvYuv"), common.MustParsePublicHash("3NvsdB7sr9oKuuLCPA5KGmXFF6RAcJFAZXGSQHbwwZn"), common.MustParseAddress("7i343vEKAN"), "HYUNG_JOON3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("HmudNdmtEpb5yody4B4TxN9PCVfJ8pTMcWmJJ7To1U"), common.MustParsePublicHash("xrTaiVTXRaf5dinSVSsBQvtqrpmf9Gfbwru5ZwiKaD"), common.MustParseAddress("8maWRMMyVj"), "ilhesu")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("44W2qjoybfJrD8PhwVtPBeRXzcb9W1zzyqWiDRzvp4K"), common.MustParsePublicHash("4W5UHwxpHZCfUnUEAs3oqsyFsmk966BddbKT7y2Jx2B"), common.MustParseAddress("9BknzdT4G7"), "Leejae")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4RHGgBXVmSFUNrpLJEZoj5h96fHBQBwrAxPnsgm1viC"), common.MustParsePublicHash("3suxLxHhH5zJzWS1JX2RRzMmNfi6satnzqmL4tMiJjF"), common.MustParseAddress("5UMhhcVF8U"), "Hys")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3eLj4Jn8qJgmHWg2efJyvQZiX8KkLoc9EdwGhavNn72"), common.MustParsePublicHash("2yBCFZ1ovQ8vcN4w1m5tqEV9PVqvCxUGLqxRtNqXB5H"), common.MustParseAddress("97MqFfoDxH"), "Nuvola")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("hbD5zFREur2snCWHP4jEy5JmQ3yh5ZzRwAzKqm2n1t"), common.MustParsePublicHash("VjQM13jndYAhfBV9ShdSNan9VrtKhzf5wV48sJZhLB"), common.MustParseAddress("99ZK89d97C"), "Nebbia")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2NwFKq9HC1gi9hMK75WPUs5aRHmif1NUM1xpZCXCrwc"), common.MustParsePublicHash("2JRaW6jWrejPYjNz82tdQhqqQTeHGWmiimqUBFGonim"), common.MustParseAddress("8tAx2nqixv"), "JKL123")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("7ynSBHvumLM7UGWk1MCPx5VapiMpfjmbz4d1jnHfdM"), common.MustParsePublicHash("2MndX7nLpPUa8imMWKWyCNrHF173aP6GCYQXVrdNbaT"), common.MustParseAddress("6oHXAQQKYZ"), "formulator1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2eLRfXWdb7baPcVG5BtdJxfZnBUmaVimv3zNUe9brvG"), common.MustParsePublicHash("9rUcwiKxu3UvvsywTVVfkBaNgCTW8VjDTbnumYsSNp"), common.MustParseAddress("92xsWi9PeT"), "formulator2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("23CCDeby1RA36JNXXH8Tp8VTCFmVr1CufKMs3ygUrF7"), common.MustParsePublicHash("3BamrkAJb7LGJx7X9hd7kGjXdW5toGCmZU9e5GgH63G"), common.MustParseAddress("5d9dBXnujz"), "jksystem")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3ZVYDBJBVCVUUWAUaPMhTQVXdkibQ15oa9XDaMqir2v"), common.MustParsePublicHash("2GQdx4p1KLmddQ9gf1stiiASJGxbjxQfwsZoEaE35Gx"), common.MustParseAddress("7wDwGoBpA3"), "Ruie")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2hb1VtmZ8NbJzfQ8QFu9JpQSnvvzheREkyu3Fid2WoC"), common.MustParsePublicHash("2duj9QqoZBAZCy362T55yEAVF5tWKeM6K5Lz1XmqfLz"), common.MustParseAddress("35tRsPSWZC"), "Sunny")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("23wtg8fiSnUJEYyXUXA8jQnGHSnUgcQL29r1nCQGDDB"), common.MustParsePublicHash("2b1ETvRLKs2rMYA7iEYhcxjg8XDHo7juvym9B5P4J4M"), common.MustParseAddress("3PUGqE3qi7"), "kimaqua1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3oa2eSjkqghnd9H8rEfk7nhUeVu4x1PPhRZru7Jew56"), common.MustParsePublicHash("2g4nu5mhrSzd7SsPmRBMochYCLKfRX1nUZF9psXHuno"), common.MustParseAddress("3K5K6GQ1QM"), "kimaqua2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4Wm8RFMgoAkbv6jAMq6idrsgcA12qXrXhyaSd89VwqZ"), common.MustParsePublicHash("4ictfocRxiWYJ4f5PFW4CDc3LWCpYLTNaiqZ7gKgd27"), common.MustParseAddress("ukXtWq6ii"), "richkmh")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3CpX28mXhVBiNnePaUdXcLQuDvV6bhu9PPNFdsFpnsE"), common.MustParsePublicHash("3PCAYgKXiya3Hnw6A6uFCkqAnQxvgYQ784UCZraCfno"), common.MustParseAddress("qMa9ZBGQx"), "richkmh1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4Wr5N67e2PRvJvf3sH93cwFDm5V5te9e5buuS6zo9Ez"), common.MustParsePublicHash("4YxPEFKGyaDE5rZHgMZchu9bgWTShDx1YXQHwhpvKb8"), common.MustParseAddress("7u2TQKMu1A"), "Kms2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3Ja6FMLNz6Nk5rQYhV25j57y6dHCF82CyTmAVkGtmFN"), common.MustParsePublicHash("LXV97QEtAbsLpXBJ671je8foQAnW8dcstPjFNRKwdF"), common.MustParseAddress("7rpyXqXyrH"), "Kms3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4nsNCbBXuEJjBvSxtowdVqKAy5Dx2RydGUvxkBcq8WA"), common.MustParsePublicHash("48dTykEaPR6E2uuX3U6Y299exp8bqs92mZWEDeR3iAU"), common.MustParseAddress("7pdVfMi4hQ"), "Kms4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3NdA9qhDVidhmJGrGQDgNcCPncFzK6mJhFPS3YvxoH8"), common.MustParsePublicHash("3ezuWmqX4ujafZ4BAQhNTF1fnPPXcEZNM9TcFpHKHX7"), common.MustParseAddress("8ab74xEPp3"), "Suho")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3tV2UruytXYRP7bniorPNy1qpmy3oVWrk4JDvnc1Kua"), common.MustParsePublicHash("45Vx6amN7HyTH6AHHFiRFFwvovnCd1W5RPKdUsKouSv"), common.MustParseAddress("2DLNrMSRxJ"), "koreajk")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("44bbefDQ6WMGUAfaLyGb3wQkwjuCCog4XRywjf2hFf4"), common.MustParsePublicHash("mznjQkubtMXYNdwcpptM4r5QoEA6GFBUx9BXBYmGZB"), common.MustParseAddress("4no42yckN6"), "ArtFormulator01")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3u6DffCedEm2HrVSDs6VGT2gNjDAvt7qYwPME39MjvE"), common.MustParsePublicHash("bDKQqgaEeCdwKWeTfA7foC3r1KWNaFx8mb1zEKsaxi"), common.MustParseAddress("9LYiUYkism"), "Choik")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4AorYiTxLxpCUhnXqUx2YFpRUa2NnHBWoYFs6zpu4Gx"), common.MustParsePublicHash("3zZxneKdhYbdNpXJweqLcZfG7RKFbxysq3prCrZUTrg"), common.MustParseAddress("oA6H5MMG5"), "Kimsj")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("33NSzxx4z5XDWcXQJJbResXZNQNr1gS65STn9hErfSE"), common.MustParsePublicHash("4GTUpdUKS8E9XFDQb6xHXKy3gitrd3UK6qP4LtxPiHj"), common.MustParseAddress("2w6WPU8qoA"), "FLETA")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3xEnBt2k9KAashJctrhaNApvp9cDKmhBcnyprK8nJdb"), common.MustParsePublicHash("rnAoY55Vd1AGagAMUZmzgvtNR12D7MrZpK5S8srJC6"), common.MustParseAddress("AFJFN4aiVg"), "Luckystar")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("gyrwaGGdRx4aK8MavgtojjVhhG2JqNNobBdZka4BHa"), common.MustParsePublicHash("mafMWZFa6yzzzSFEfbuAELGe2tyhWgsgjGc7GdPe5"), common.MustParseAddress("6qV12tEEmf"), "STL_Fomul")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("TFgBQcUcLtheRA3Pk9eaVRBDgtLdRuwcRgJakuobpf"), common.MustParsePublicHash("vm8Kx9fdxj5BSPyR8YXeeQi2ThbzK2qbHZSb5cxF4e"), common.MustParseAddress("2UijwiDr2T"), "lajblajb")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("ssQAR3a9QvmSJSxuTtVe9s9h6iXDNjFE6EHjhLtyTi"), common.MustParsePublicHash("4puB9ww9oXJko5g6Z4qetBBRaxYuVfXL5w2Jkd68reZ"), common.MustParseAddress("9ZjbhRiDes"), "mydreams1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("bEEucGsYd1MLgZxjJvBkE447Xbs4PR3kcrkc3z6sm4"), common.MustParsePublicHash("4rnEFXZrJB68T1AqrVGJpQSHimbcm81jsc5CAACYUXs"), common.MustParseAddress("6m63HvaQPg"), "mydreams2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3DRJNcTDXDMpokpAe94dX5USwvNSUrvuuuU5aRw93DB"), common.MustParsePublicHash("8EwkCneMRqowhRS4fgSGk6BvAgjFVVL9ApXLLh8iZ4"), common.MustParseAddress("7UrAq3GpJw"), "mydreams3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3GeFDP4XetJ7ynoDvwng6MwekoHfBbCY8shtTjG5ExV"), common.MustParsePublicHash("48db1a3kCkQoyKrKFfPLWJ27BxRZ7Z6jBPpRkAd1No"), common.MustParseAddress("AQ6AqytP2m"), "arirang")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3eq44Ud1GxfndTR9LDkdvuUeEWnAgBiQ86tqFBKFZC3"), common.MustParsePublicHash("fsnQSkyb7SYoUcRwU8sScMdPZ3q76FqV6h7ZU2yEUz"), common.MustParseAddress("3GsqDna6FU"), "leefwang")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3yjbHrgPmQShx1mb4449b5Kuwk4oDADoAbHdeoGRCSq"), common.MustParsePublicHash("3xFJZRuGby6f2YgvibuBP3oN4CzGBoXTvgepXoLFzwj"), common.MustParseAddress("4sC1mwGafr"), "MINHO")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2NwXhbFKzpZ9T3zmeywvjBQ6uivaE8bEsKFmeMVKd2W"), common.MustParsePublicHash("3ZM531GpT4aRq1s7dZmrVXhrwv1iJFcd8BksCtFgv5E"), common.MustParseAddress("9yutGhoJLa"), "Shin_Junghoon")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("42y9B98mhmc8xd5e8EJwxwj3gAVEJUrwMnYDnAZDdMb"), common.MustParsePublicHash("3bo57NbPvSCGEdSKcWTdrL4Z5Hkwx5uCGwewi9syvuB"), common.MustParseAddress("9uWvXk9U2p"), "sjh2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2g6XWYvpSo6MV9tjUGFz86rpNjFigyHuFbUUBsrFP8p"), common.MustParsePublicHash("89ap6J6Fi7PjZAW6Y9P7t5Tc81kAh4TajGVodnoAHi"), common.MustParseAddress("9nvUvJfia9"), "sjh3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4gvZprUvqpVYv6LYEFerKDK8hq8JCeEFJorS4HZkkth"), common.MustParsePublicHash("38zNJD18ntJcHjfG1anJWiMUUJoKxzn522vgV9B4uLZ"), common.MustParseAddress("6x5SeKhzEQ"), "avellaneda")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3spa3vRy6LLbdDhEsmSGx34xcU5mxWgnNAR9ugKn1dH"), common.MustParsePublicHash("35MHgw3QQcBXTXiL3GSgyVpZxYShAnGW6BZewTUcydA"), common.MustParseAddress("3TsEaBhg4g"), "glory_of_fomulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("EQZe1JkQBA4Ft6Rftuhw4fNMzvYPwtvDTppqCrXE3x"), common.MustParsePublicHash("4cJgpBrgAuXA2oY2fQ2rN7LeBAmW9Ny9qpxQefneRtn"), common.MustParseAddress("3EgMMJkB6b"), "ock1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4B5Dt1F5DqdXm4CYEB5vWmiqmTH2M7Dh2WBjmQ9ESEe"), common.MustParsePublicHash("2bocCdJBf7ywg332XiiPQzDTBdTWFNcgmSRELe2ATaA"), common.MustParseAddress("3CUsUpvFwi"), "ock2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3JhKe6jPkrBkSETFSe7k5ihtPdv26LphxyEtpcRB7Ar"), common.MustParsePublicHash("3B15D6wmyptWaNiCqBnfJJVzxfzto3Q81Q6JQgafiiM"), common.MustParseAddress("2hudAbBM2c"), "345000")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("wNvAotPEP3SmyUHYn3xBmN3nWzZmEemyYmYrsjUKF7"), common.MustParsePublicHash("3HwFAbHekavCJUbeAK5qA9Rw5ZvvbCqXXbEM6KRyKT2"), common.MustParseAddress("2SXG5EPvsa"), "pacetoface")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4JSMo2H4hfqPP6UZeGBfiZrSDX3Fnu9X6H5YxJg2xrz"), common.MustParsePublicHash("3Qf1sQt7uDA5Bwj82FbHawrFY1pxy3GRgzxtnENpTV7"), common.MustParseAddress("2WvDpC3mBL"), "tigerbalm1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3TcZrq3jL6HqqaiLfx2vnAhvcWRir8tthi2RVgnWLZY"), common.MustParsePublicHash("4cGraxgxVMMz2K3KLWUi7fhT9qyU5BbJhGm6QLDR4cR"), common.MustParseAddress("2Z7hgfsgQx"), "kimboksu")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("rhRnBPYcPmSG1SWF4nbYJyqCSZN13jwt9LXnragr4T"), common.MustParsePublicHash("4o9prMyyFRCKZkgzjdkaycZtSt9v24ReEa8XVenagmL"), common.MustParseAddress("67isVmXprr"), "pre-ico1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("64JmeTFQxgzgL6xnLS5o1K6GS6qRPyDTnktbQLvvUS"), common.MustParsePublicHash("AXetkwEnjAHtBePt5UzMzAsAQBuMRcRzPUEKFETEQN"), common.MustParseAddress("9e8ZSPN3xd"), "SupaDupa")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3Jf4b5QkSCS8FfiiBYMLa2mf5dB5JXnt95C7PbZ1Fwd"), common.MustParsePublicHash("HJL3qGs4qTkkDRNciCwznjLepwAsCKzLvjNRvgAW1b"), common.MustParseAddress("5hYavVSk3k"), "JIAH")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4A7EEXVjyVDZK8hBCYG8eV5WJWARmDHGjdpSaXYGSTY"), common.MustParsePublicHash("ZHZeNvHCi9ZvvN6XxaEyKjEBo9wZLDx3Rk5Q3CqDcX"), common.MustParseAddress("5fM741cpts"), "HJA")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3SL7u5H1TdiqExiejSou1kTwkoJSSvDTnKoFVZ7fefw"), common.MustParsePublicHash("3F3ArMmFdvWb4Nu75eph9ySPpyqgdS2pqDQbtaHw996"), common.MustParseAddress("BzQMQ8goT"), "1G")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2fBTovpr7jB5bx11pEWu34PSbqHr3ZpvXmdezaH6xkv"), common.MustParsePublicHash("2Lje24vqtjnSeMVrZo6TuqQPh8pmq8ExPbfs9PhYqBe"), common.MustParseAddress("9nvUvJmea"), "2G")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("f2vdEEzfedW5st4mHXP7gcAYBmNMpWAYA3ikEervuJ"), common.MustParsePublicHash("SZcdZeJHZVNZDfBDmSUbfR8wnWrQFm8TGbCGdhKbnU"), common.MustParseAddress("7bScSUrVh"), "3G")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("RBh27zSGTdoFHiysJ8PuFidC7FZDA59JunZ19KXHoy"), common.MustParsePublicHash("4UyUk71jWsMfULNrGw5GVmsw6R4efowjZ2cMmGJMZRG"), common.MustParseAddress("5PxjxewLp"), "4G")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4qh8JXjaEt5J1zT26FGx8qowsPz7viofE8K6ERZF3GK"), common.MustParsePublicHash("HZ2aXQCXUv5VA8Su24wvGNjp7g9SovkE8SfE3YWmt2"), common.MustParseAddress("3CUsUq2Bw"), "G5")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("hZkfxgPfWhoBGRcEbjUcYCiiqKMsvMhKq2mpKMksNp"), common.MustParsePublicHash("31Te9d4PPtLZ9RpPcnZc6He9EZG9Pd1BodS15Q11uyv"), common.MustParseAddress("11111734"), "G6")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2FZCsWUjYxTLCc5e6QKaM9jtoFfxvZE1QfQD1pAh6SK"), common.MustParsePublicHash("34F9oWyq2BSbWRpCUBnWcqceHmoV2otubTq63mwJCm5"), common.MustParseAddress("49RtEpaAkb"), "s39404104@")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("waqgaV7YQ4j75WDYSSSbt8VUyvXU7DQzCAdnGgFPEN"), common.MustParsePublicHash("47DmqdciCL4x7F4sPozLK4sznGMMaexJ8f8JpJFr9XL"), common.MustParseAddress("47EQNLkFbi"), "s39404104")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3mSGDpF1RV2KNEF6p9qtzJtptYoawmCwFUTtJVjsPiA"), common.MustParsePublicHash("2bK7PMf3KXLHVJqyumbLF8Hbho2EFYyZxCr5iq181RY"), common.MustParseAddress("3CUsUq2GQ"), "Tigers1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3PkZg8inpnzmQBSXiUY34a4SVEu2RuP6EYFjTRxut7Y"), common.MustParsePublicHash("C2G5Q45Dx6CtoPTVnUZdAEVqWvyuDbZQ4YReUYxbNB"), common.MustParseAddress("NyohoGGed"), "Tigers2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3VLFFdqiLK16LWGptFHGh6KHSH9QTky2QJzGSXwazKu"), common.MustParsePublicHash("2eXmiBxD5S7WxTPN2AhVU8Wg3ogVkzfYqkgRBnrn1AR"), common.MustParseAddress("4wayWtvQyc"), "170r1bn")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4NQApXbUFLtfvwHKV7GVFUYBLWDVBd1zbZggEHkRJXp"), common.MustParsePublicHash("Cu1rAXpzj5BLtTDBWHHqWfcD31Ysrve7qXBahZKr1i"), common.MustParseAddress("4iQ6J1xv4L"), "Forour_Eddy")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2oWbNgJcwpiBZNF8GQogKV32PifzWDwtKbwR3hDqb5j"), common.MustParsePublicHash("LPJy6yVb95NJ2ktdLPnvYGEQ3TgoXWuTePgDn9XHh4"), common.MustParseAddress("5PxjxeqQpi"), "djshim1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("ohjtn63Rrgj5io8KpeEre9io4smG5635ZzqZbwx9sX"), common.MustParsePublicHash("2jJJzY6Hrh7qKGna2GEopaiPDwhYoWNi3M9FnYLFb6"), common.MustParseAddress("AMtgyW4Tsv"), "JH")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4Ha2hydfGM6NtuMfYa4bByeTYskiiZm1MxA5Nn3gFZu"), common.MustParsePublicHash("25T4fFMTKmLHwAskPqNGN369HYnPn7QxqTpG9uqfrQi"), common.MustParseAddress("3W4iSfXb96"), "BULESKY")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("utJJ7iJtFTpGk4nYREPSy9HLifPZdVsj3DBg1LRCJ4"), common.MustParsePublicHash("43L41YQL9B188vu2fpm6oPy8HuqKESjhH5gRK9MZyTP"), common.MustParseAddress("7BGKsCfV1W"), "Leehan")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("ztxZecXWFM5gUqjwyn3EUUdJywr3r551UisXtH6GSe"), common.MustParsePublicHash("3LCHXdawq1sJvCegngv3qkuSW7KXRsYnLzmzj6CBACh"), common.MustParseAddress("7wDwGoBp5e"), "Leehann")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2JbXKjUis5kPVkQoyiiFoa6y2Jkp7MHSGqFq8QX6d2F"), common.MustParsePublicHash("3vVaKDF7T5bAe6soNa4u9jQZ6Ejc5frELfs6ZVs6x6"), common.MustParseAddress("4T1jCfBVyd"), "wonhakyeon")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("E5vUvSF5pkPMTPU7pGMrQ82DtpFqgq6K9LoUBVfXkd"), common.MustParsePublicHash("3SCyq69xdNnzYKejDcqMdCHdskZL5fQ1EdaZyK25TaC"), common.MustParseAddress("AWgcTRN8VS"), "SONCBFLETA1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("DYLu56qFGZ3aFjvUAK7gxXiCF1wqkj5639VbZ2xyFm"), common.MustParsePublicHash("21NoPnkjPwGWDXuZaPmtedcPkFHUCvUTTXmgDREPK1W"), common.MustParseAddress("4DpqynE14M"), "SONCBFLETA2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3P8Ua7yHh5TKi67EC5w79nVXMdZHwXRM6tL6epjYHck"), common.MustParsePublicHash("2iYJyp7Yn8MWmNdXY5AH3FKdkHfzXkKCRAEgVGfF7rR"), common.MustParseAddress("8RoBb2vjCP"), "KGY0732")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4KWRCvk9JCGD9eWjdm3hYV41PoiDrnLcsGVaRYK4Ssw"), common.MustParsePublicHash("KW4V2JaazFedeym1ZtnqnnxnMb4edjRZUn4iML2ZiM"), common.MustParseAddress("8TzfTWkeMJ"), "KGY07321")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("zsJyD6MzLJSyfD3J9cmdwFcXNRcKJxhrWoAgCvdjKz"), common.MustParsePublicHash("3nA3itphSxobNQtVTd4PerEs8kAromjN6CJ8WpvqtDT"), common.MustParseAddress("8WC9KzaZWD"), "KGY07322")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("41fUFn3hBG3ksj16PU2XuAepcYNpbtq916CF5TYBVDg"), common.MustParsePublicHash("CLwNNWhwcjtzNrWz2knXczvd6V8mUnbS6gwQWEM4uK"), common.MustParseAddress("7NFjDbo4ms"), "Shwjsjxb")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4G6pnyUzJzHiwAH8kLSdknJ6VLxqHY44uZ6LSWqJDG"), common.MustParsePublicHash("dqgWZGum1ariXiyRWjMSZjji1bVBn3d6MksTgG367"), common.MustParseAddress("8MQDr5Gtjd"), "SHWA")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3DHgEoop4hPgFZLL6aa7mSYks2dfc97GLFXsfjsTSrp"), common.MustParsePublicHash("4EGZAUwpRzD7Epvi7J95n6j5dizxsTxYxFKjwCrnbZZ"), common.MustParseAddress("8PbhiZ6otY"), "SHWA2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("4o8a4P7VvL6umtd7tWg2dd88m9kTc4iP7PV9G8ScmNo"), common.MustParsePublicHash("3WdwN6VpnDMiFBPXGr17JPTeEbRtCFGC1niZBHjxrYj"), common.MustParseAddress("2Z7hgfsgFp"), "sung")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3uPEwkFgsDTz3ahLkLRoTUeW82j4ne92jshmnaDJEmU"), common.MustParsePublicHash("2TMVPzbD9iHv3m8KivibXXzHcQ17Fd8KxVopSydeX2R"), common.MustParseAddress("81cu1kqeXz"), "sung2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("33FWneYXFtMvewrwFeH2cYzEaA61gZnhZ2BfRDgahLX"), common.MustParsePublicHash("2MQwBaS5UJqopUuTCBVNitYw4iSaMgudRdVypUb7BkZ"), common.MustParseAddress("9kj13pqoRJ"), "Donghyun1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("2QgPWK74bLLSr66dDb4ajwyxxgXm7dUxBPFn7iH72fK"), common.MustParsePublicHash("69EyFtC2kQc6whA1AgQdjRFJ9uRxnacArj1BMFJzkg"), common.MustParseAddress("2pW4n2f6Qv"), "doodoo")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("3sTwebTueuWMTZWKiiNruCc4LHqTNiWqcExUNhzQrQv"), common.MustParsePublicHash("3grZLC5B9SJy6g1m1b2fyVpPRVyuiCD2z1Jrc1sPQ5Y"), common.MustParseAddress("8qyUAK1ooV"), "doodoo1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 5184000, common.MustParsePublicHash("49h21RPKJ8GpWFfFKHcuV8TcrsHFzZhejxZAC3pwBRS"), common.MustParsePublicHash("3Piff52ULgP6PDRYQ34aqBKXaQwNWgL9thssqgtD18W"), common.MustParseAddress("22LyVxJrBM"), "ctrl_Z")
		addAlphaFormulator(sp, ctw, alphaPolicy, 2821822, common.MustParsePublicHash("bp8zJpGY4YpFEMnF3KmjSmLkuYRN6zzVdoftjh5Pdb"), common.MustParsePublicHash("4Rm8hXwy9BHjWyfSZEM9YzTawfs9U8NAcDWzVq4nnUa"), common.MustParseAddress("A6WKt9H3x3"), "Djdjd")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1982426, common.MustParsePublicHash("34YC2pFrDUp4c9sw7hD31Sd4fqWNjXrdiDKNwP3pMNk"), common.MustParsePublicHash("4thJ47PSxe3PiJBWsbTc5rsaT6FJC9gfnvxnz5YtKc9"), common.MustParseAddress("AHVjEYQdic"), "JJAJJA3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1982426, common.MustParsePublicHash("3DxeWmg543yEs8zhWFN43KsSpPp7h9kWyXj6xo8jZQ2"), common.MustParsePublicHash("3rrdz3VyU1wjusKyZvth2efAHU5TfEaeZUNjgULUBZH"), common.MustParseAddress("AYt6KuC3o1"), "JJAJJA4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1853778, common.MustParsePublicHash("2nNnHahvFd57j39GeiYgSRKDf5gqrRkBPRzfKWECspK"), common.MustParsePublicHash("3G2SPcVZSR3TN2qn7PgwUE3VDgmpcCdqvrxb2rmZZzF"), common.MustParseAddress("Ab5aCP1xwv"), "FelipeProfitFormulator")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1774400, common.MustParsePublicHash("43sGaZuxzZSiBHLGmSiBe49gAztwpJS3QvJMvmTev43"), common.MustParsePublicHash("4BtA2ctGZ8iA8zUgVskpR29RrLAFs4LCxJwt1qCDBfC"), common.MustParseAddress("8hBYgPi9Gn"), "grandlyrule1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1774400, common.MustParsePublicHash("36R5Hf7pdZUMjMvV5RDtEoWZzTzZyZfpxGS2Lb3uR7h"), common.MustParsePublicHash("3yUJ2JVaiSUN9BXA8NV1Wcq51wctYxQc4o3StnqbqKf"), common.MustParseAddress("AdH44rqt6q"), "grandlyrule2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1663342, common.MustParsePublicHash("9TfGMzKwQ6Xuwn3QRbC3RTpJVbWxc7tKfatiVzAp6W"), common.MustParsePublicHash("2E67aM7Mz1ZUSqZKnijpJwXSotmbf2stDvatsGjRPDM"), common.MustParseAddress("8qyUAK1otS"), "flwjd")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463978, common.MustParsePublicHash("4RC6rMfMuBMMRsSyFzfs4MMPM86cSoZiMgZJ44wJ8Az"), common.MustParsePublicHash("4oNJC79h96SqGBhao2vVTFcXciWKfUZDFXBZGxFfB7A"), common.MustParseAddress("9VLdxU4PVy"), "PSJ1")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463602, common.MustParsePublicHash("iF9mMgWryZtNEcDh7RfiRVXkViL1Yr9NzZEBYMQ8Rg"), common.MustParsePublicHash("4TQd8B9dbmTuDcT9rnz7nCuUFQDWsZMTnEVdhbgpr7T"), common.MustParseAddress("9XY7pwtJet"), "PSJ2")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463376, common.MustParsePublicHash("2K4Yn6jiyk8GDn7oBVTWoMJECCsL6MnARUkn3mKQPYn"), common.MustParsePublicHash("38kDThHydeJF4mbtVfJJjbLaP9RvSF5x12F7YrGY8xx"), common.MustParseAddress("9ZjbhRiDoo"), "PSJ3")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463376, common.MustParsePublicHash("PmKAu7C7jkQy55GhSdbPJvkWecd8Sv9sRCb75ro95N"), common.MustParsePublicHash("21gvNapW6FTHwPvchbNSJFF9HSC6yofZNaw8EJUwHR7"), common.MustParseAddress("9bw5ZuY8xi"), "PSJ4")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463376, common.MustParsePublicHash("DDCUBpHj8DDybUa1ch6HnANKdTpWg8YhyhWDGai8an"), common.MustParsePublicHash("4mSFytMMVxAAdn8CD9kiJdXXKuWEy9X4ULHopJk9LQL"), common.MustParseAddress("9e8ZSPN47d"), "PSJ5")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463376, common.MustParsePublicHash("4kiS8f3oce9poSsKyzRaS6iMSUc6MM9igvYFt1CPHmj"), common.MustParsePublicHash("2d8ee9yArR8TrxCWyvViwfDC9RSt5b218n1CsiRaBXu"), common.MustParseAddress("9gL3JsByGY"), "PSJ6")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463310, common.MustParsePublicHash("322x6w92hHpyVog26DinYWFJDEysTE4umM8WpYfwVqU"), common.MustParsePublicHash("4fk8UuZC41XG9PG3u2kspCwaigHeZJqf2XW2cpzqCPA"), common.MustParseAddress("9iXXBM1tRT"), "PSJ7")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463310, common.MustParsePublicHash("4JBkELspPFBGuJKLog7oks6Es93FLEjwckENHT9h8ae"), common.MustParsePublicHash("2WhyerifyDW1nBoZNKMrvk3SrDhVmiiKLr5AefGb4B6"), common.MustParseAddress("9kj13pqoaN"), "PSJ8")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463290, common.MustParsePublicHash("3oreYRqvh2UaVbbxCZUTRapT7oebCrommswG79e11ii"), common.MustParsePublicHash("2ia8fpF4dSmNZ9V6V7kWUsuiyfP4GAQ83nPMnnrTTq9"), common.MustParseAddress("9nvUvJfijH"), "PSJ9")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1463290, common.MustParsePublicHash("2nHRcN5rymhQHQzCyhG9ESaZSEBzGXPLhdkkjZ5R8EJ"), common.MustParsePublicHash("hVp5J48Ukv79GF4pfFmmmQxCifmMhf33eToxoGC2gM"), common.MustParseAddress("9q7xnnVdtC"), "PSJ10")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1458332, common.MustParsePublicHash("2pEdXC2dpUKxie74WvqTWgQiQ2ZbRAb5du9Fxh5hppA"), common.MustParsePublicHash("4XZtAtUXKsMUfyDBkHiD9D6KqYbWxWvwwQuw2pe2zRj"), common.MustParseAddress("9sKSfGKZ37"), "PSJ11")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1458228, common.MustParsePublicHash("37K9x9GxVxF7ybiMEcFCZFU5pyHyodjwmTDUHr5wfCh"), common.MustParsePublicHash("2oCig1VdmRJP8sJ7AxtuRvUMfrDxvXkhGTw3msWwQXW"), common.MustParseAddress("9uWvXk9UC2"), "PSJ12")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1423432, common.MustParsePublicHash("4pRPkQQj6hikgwrQ8s8Vm1x6H1WEh2fGx355pq6KEFF"), common.MustParsePublicHash("23gQrr6L2Z5umyoCXQ2WAQBkePjmvjKwTqrxH5SdnYn"), common.MustParseAddress("9wiQQDyPLw"), "PSJ13")
		addAlphaFormulator(sp, ctw, alphaPolicy, 1114330, common.MustParsePublicHash("4CpKUrnHh4K22uD9nskY84bhTFuA2DNyvmhSdg4yafc"), common.MustParsePublicHash("3F5pPt9bVD1V4d4rJpqrzLRBkrRKTbQjD2ztB5RK7pi"), common.MustParseAddress("6VhgCZnzPS"), "DLI-FLETA-01")
		addAlphaFormulator(sp, ctw, alphaPolicy, 113360, common.MustParsePublicHash("Q4RqLBPtz47PcmWq9GvwX5g4pNAJKiyXrmtvaRq6EA"), common.MustParsePublicHash("2WxprTRTi12tbV5hhMD98VFy4JXjSwvdkUawvkV8yLq"), common.MustParseAddress("6sgUuN49va"), "STL_Lee")
		addAlphaFormulator(sp, ctw, alphaPolicy, 17220, common.MustParsePublicHash("3igbqnfZsDNgpCZZXWQm1JkJYE2QLwMUwYFuyuCYoCb"), common.MustParsePublicHash("3VfL3VY5pYwubzAF2snoeWquZq1Q5c1bZ1qzSgptocx"), common.MustParseAddress("5vjU9NQF6H"), "ehowl2")

		addSingleAccount(sp, ctw, common.MustParsePublicHash("Vb3J1ZQtcoMtzjf4H9mdHwS3dfo5h1a5Ts51WK2fuW"), common.MustParseAddress("3W4iSfXe3Y"), "manjae@firstchain.co", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("AEHGi2sfiWg13Y6ww35LMRQ5yMp166B3tQg6meqteP"), common.MustParseAddress("3t3X9TnoaY"), "chappie@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3xo8p9Yt3L3w8GfTGYTMbqL9GtLXGDY9oc7qMRxSNBm"), common.MustParseAddress("3xSUtRSdtN"), "mrcgs@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("DQ6fVX3kZJ9R7F3uRospcweWDWkFbDp2t2bLiuvGCu"), common.MustParseAddress("3mT5Y2K47u"), "katekim@firstchain.co", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3M3K4dnnMzr4o3HDC5WBN2wxTBybpQnZe4tUGcVhrGB"), common.MustParseAddress("3qr3GyxtRj"), "joonsoocom@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2gjHhhrv2KwsJe64K3zbypBWUcjpXhD8NMf5gUtaycT"), common.MustParseAddress("3oeZQW8yGr"), "namu3111@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("z7HJu6uUZ2gngD5GvsQNKMH9h79VyKZBQsYabpiJbH"), common.MustParseAddress("31VU8Rnj43"), "cta1123@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2LchSfa7jV1znr7QPQBwQVYw8LtQ4qqFbRoTdLxAjiE"), common.MustParseAddress("2rhYeWV4SX"), "meaowe@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4WHA1q4J2FTMqFffLQ6LuVWsauRS44Z5Hso5qnGLyno"), common.MustParseAddress("2w6WPU8tkM"), "strecs@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2t4BYS3ZhcQVh3n8HMFkQDccsxKYW6R31tVdkoyeftW"), common.MustParseAddress("3CUsUpvJpn"), "aes0519@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3KapewtooEaEF2fXp33sAPjgyiZQgbwGTThDAHZdJ4A"), common.MustParseAddress("2UijwiDtuo"), "garangyo@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("TJoxWpHVde2CKKhFt3zjLCQai8kZSuxfpBWSYTXYvj"), common.MustParseAddress("2SXG5EPykv"), "zutenbe@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4rUkaYR3o78aTaSUhPZYpyzMaaVnZvhWzrcMWjDPwtW"), common.MustParseAddress("ukXtWq9cp"), "yelllim8@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2BtnJP3PYYim4DQ9hncvVPg5H5kXfpgwAJak1v5AGw"), common.MustParseAddress("22LyVxJu5X"), "tc8863@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3HZy7k9hkjXNQXeT6Gg7xnbHeevNyP6GSa7StyRopJd"), common.MustParseAddress("RBHaH6EdT"), "sanae5078@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3PoTcHr5S7z9Rpug4w8rB4frLCeNZ4Z1aknd8DaU7dM"), common.MustParseAddress("AKhD72EbgE"), "wjarud8100@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3YCF5XAZqZBdWW9AJ6HmR6hTiJitHVTzVhfd6YVE38v"), common.MustParseAddress("A8hokd71um"), "asteroid4756@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("468kdTddL8ZTmBoyGq4mJCHdzZJQnc5UGi14u6tzMDH"), common.MustParseAddress("AD6mVakrDb"), "crypto.hit0ri@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2moDV6vPECuNWzaQukiQx4wze2jKF7Ugn5JL1pm1MgU"), common.MustParseAddress("Ab5aCP21kb"), "omphalos87@nate.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("bcFQkHLZwSKUr1f9iKTikNS7yoBLcK6tZQgFhs9PtW"), common.MustParseAddress("9yutGhoMJX"), "a01033351844@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("46rc5tn3fodXNgRrTof9LW2sWf7yshmFNXmtpPAkJDj"), common.MustParseAddress("9wiQQDyS9e"), "sunghun30@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3ebQzRkAuuwx7aYV9RgGDztfCuYspc2kSgiPZWt7TCh"), common.MustParseAddress("8vNRuGfh1d"), "kdw@jd-networks.co.kr", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("376EH9dziYbnHhqtfXqcEkGWNj3vAStCVJgE6gAZpP8"), common.MustParseAddress("7wDwGoBs2q"), "chchim0425@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3tMMpMzuKhLPubR9ERVUVhEijxoh5cK3dB4D5EMhPvB"), common.MustParseAddress("7kEXvQ4HGS"), "amani2@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("hAsuz1ac6b56t1pMpmKrjBBRws85dPRpeaAggkqfXQ"), common.MustParseAddress("8CcJN9yH7G"), "fnintagi7@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("pyr8WJKaTYjcwnoZSL2LAZEYgf1hzZwr26mnzcTHRn"), common.MustParseAddress("74ftFmBnWY"), "l123j@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("28WMiECKLXQqFtojNi4giimCVgraRD9Qj3YtAh6juvM"), common.MustParseAddress("6m63HvaTHn"), "tel115@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("WPK1DDpLuAm2ZuWC4x7LjFqBan34psZzK4KYP2oPK"), common.MustParseAddress("65XPdHhxXk"), "rtysm5@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("g5HFPenHPG3rKMFrD59T8D38WLUXLByyGk7eM4u3Tg"), common.MustParseAddress("6JiGrAfTTE"), "parksj0623@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2cxaxqwbamwZmqTmzbWqhHzH6cNb7QEzXxZxJ6e3BwH"), common.MustParseAddress("6P7Eb8KHm4"), "Tazanmj@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("XimyNFoULsA94V7JimmvgLexwZZvA4hcNPuZnqAk6g"), common.MustParseAddress("5KZnDhBdU7"), "md318i@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4bW2StZk87pQ21jykqdxSwvXr28Hjrrv1tZbTg9GA51"), common.MustParseAddress("5MmG6B1Yd2"), "kyung2200@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("FPC2s17vt1X7Kgy4ZsbCFLbW3AwPLMo9N35WrM84j"), common.MustParseAddress("4no42ycoKo"), "kjyoocf@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("49w7dJMTrdTKmMxZfRKmqnktgRYx7MLwV1NGbKuWm22"), common.MustParseAddress("56NtzpE8Z7"), "2151530@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2BTE5EAFiGraNkkfDurgwu9MkSiVKCuLG7UvXW5z1va"), common.MustParseAddress("4wayWtvTwb"), "tallkid77@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("TFGsFegMqyY3cXNA9DApK5uiAVhZfqQoyfVi1TDge3"), common.MustParseAddress("4BdN7JQ8sj"), "kibos0125@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3dx9xCfGB9cY5v3WyjUXr8CC5EfKHMnYqM2hXznTiBF"), common.MustParseAddress("42qSdP6UGD"), "leeyj622012@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("AzSUfwNU2Q4Vk7TFhzRGm5PW8QuXBvjrsLWH5qgcnX"), common.MustParseAddress("3aTgBdBURf"), "jungho555@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3W4ZNamDyWTxTL763eKg6irHaan7zJjnWFJ76bUCRqy"), common.MustParseAddress("3jFbfYV93T"), "beadsmap@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("Mro7153cAeUzEGBEhtcG2jHXduLQYxKZ3GcDznuMMY"), common.MustParseAddress("2pW4n2f9N5"), "kjhun3kr@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2Vk5t2T2Uifc2zEkxLpEtpCKVJzsbNhwoUDDAFLUixw"), common.MustParseAddress("2rhYeWV4Wz"), "judasowelu@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("QmGefrRbN9M8jiqQTFBeYNdzGCnXKkLxAuRvBA9T2C"), common.MustParseAddress("3GsqDna9Cu"), "ekgmlsmsdy@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3XCKmG2zw461CcGCVGoAe5YCeKPUuNeHM341EFxMM8U"), common.MustParseAddress("3CUsUpvJu9"), "hkyjk67@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3qTVxpYVKUsw8qB5YQ8xCZCmqbKLWJ8ScqTeNmZEWN2"), common.MustParseAddress("385ujsGUbP"), "naranja75@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("7NkWzqS4gZ5jiWFYQPReCyd5xmTPXECd3CojCRmbuY"), common.MustParseAddress("2QKnCka4gS"), "tnt2bl@msn.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("26ktKKJHVReyG7RF5mWsYuKdgeJytek7Dc6UAQvA9z3"), common.MustParseAddress("2nJauYqEDW"), "soncb@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2A5M2gxHC4N64B3sLDZqAd4q9CFqHtMDudvN1MQ2TVv"), common.MustParseAddress("2hudAbBPuk"), "chamjea@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4ExW9bgGvkeTc12vinvqchue1a3rbYMsoFAHepPDzUJ"), common.MustParseAddress("2bKBZ9heT5"), "yhjo1202@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("22tSLhQRDLy9f2Sz7yYMwTkmsT5Hvj31RKy43UupZsH"), common.MustParseAddress("2WvDpC3p9K"), "leech4312@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("34KNZPcxd3vzSFHihisZWQBTS3YbkXJ9vmHvjHsNh9d"), common.MustParseAddress("kxcQbXV5T"), "loverjini1004@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2PRNgfWqx5ygxxcZ9mQy4mDjAPWng5LDmXcCRYPvJir"), common.MustParseAddress("oA6H5MQEN"), "ace8987@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2m4Y399haDbQugazTswU4mpciD19gE8KVrvbyfwu8JN"), common.MustParseAddress("cAgvgDpTw"), "sm2913@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4QwNGnCw1c6mvM1UBP5R3PcnSFVSFGzrgGbbB7L9HTn"), common.MustParseAddress("RBHaH6Ehr"), "sky7332@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2g2H7bC25qNDYzRMw88i1zj2bRWDTqfLba38cGzj6vD"), common.MustParseAddress("9yutGhoMNz"), "nana432@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3SAJE8MFy2yUGyzxfLQrptJkHQ2HESCuA1izRnYKtDy"), common.MustParseAddress("9q7xnnVgmU"), "1014152sunhee@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("coMkYCb8SxzBtenB3Fw2o5qu4CfNSXbnKw7o7MAQPb"), common.MustParseAddress("99ZK89dC1N"), "itvadu@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4Z6xivbrb5EA5f6Qwqoxy2qz6dakjZzkmfA1jsi3FrV"), common.MustParseAddress("92xsWi9SYh"), "fleta1004@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2SwAhNzRPUvLCYCko7HQ2yXUW53ifMYC8oyuzMwNWXG"), common.MustParseAddress("8zmPeEKXPr"), "funhanstory@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("cHSCgTrUJFjRGCsgMEHqZtZPZgK6dhv6VNefChrg52"), common.MustParseAddress("9NkCM2agvv"), "kyoung5481@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("oTaTVVcoGrxhrnRBHTi1fk3hj359NHadNEj21Tf7ZY"), common.MustParseAddress("8RoBb2vn6d"), "abchhy1237@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("6gf2k4xhtmpf6eVyJGJM3DnuCytFtrrFuqcvsHUM8p"), common.MustParseAddress("8MQDr5Gwns"), "mm9778@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2bvMfUP3k1Af1Fk7JC4j6h4CLjpfLm1MFWTpzxmVs19"), common.MustParseAddress("8H1G77d7V7"), "powerer@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3dT6ZCgJEaCT5FRSNe4UkJnTime33dhNvBSpLJv2xCN"), common.MustParseAddress("8qyUAK1rnc"), "wwlndy@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2yR7exTLNdJ783zrBqg78uy6DriMKKJfTr48Lk4Y22m"), common.MustParseAddress("7rpyXqY2oU"), "dlrhsdbs@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4EbFT47dyecDA2LLuxBF7LXh2fraKpPKuTjFhHUSXW4"), common.MustParseAddress("7BGKsCfY3e"), "godada1@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2Gz17txGbiuCk7MVWaNaGTgqjmMk7T7X6gbKicMMJGP"), common.MustParseAddress("74ftFmBnay"), "pier6321@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("21wTCyTUXzJLrfPAPtsWGenNDywQDLPxa5nRZsm8gEy"), common.MustParseAddress("72UQPHMsSQ"), "ohsm1806@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2C2KZ3BzUVEPPQLJv7t1EnQBsuHn84TuVb77YSnWfjE"), common.MustParseAddress("6m63HvaTND"), "av6996@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2SNDAqERwDJfifWo7WezFTSQ82jiXToS7emivk24dqL"), common.MustParseAddress("6oHXAQQNX8"), "jayuin2001@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3BSknDC89KLCFS4kDyPBw63ZHNuhXuDKsmm3omqUxhT"), common.MustParseAddress("618RtL48JM"), "ereswony@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3MjsB56849TpS8F2Y3VzJ93J6a1nm3h9g4CHArkC9Ef"), common.MustParseAddress("6C7qEjBi54"), "jisegong111@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3dDi4uhQU4amvref5fwt9QBfNc6t2JfyNN1TbSDkeTu"), common.MustParseAddress("5SADq8fP1B"), "kimaqua@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2XL2BWmwGibkX4VahbEB3Ji2YCXEB5Bv9bMtq5BAerE"), common.MustParseAddress("5jk4nyGiEV"), "cta1123@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2GWDq2BWd9GFkZX2dFFvRQphWn3gD8ZUFUYAmccqe47"), common.MustParseAddress("5fM741csvj"), "kjho100@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("26jXYJMKYUV7ygJHnLPeWVbMeBifvFtVS3FgHeha3tB"), common.MustParseAddress("51ywFraJKn"), "vgbyhs@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("16bUWLMRyjxMzVeWBeh2s7rsYqwj2H7E2xe7Hop78L"), common.MustParseAddress("4BdN7JQ8xA"), "bsh6557@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("oLcTLffXJsi4RVBua2v3nuZNqtddL2MKBx6HiqwKqB"), common.MustParseAddress("42qSdP6ULe"), "solbarama@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2jkSkLHHA68Jd9L4HVT1K3Jbj35ERTKiyGVxYUgkPrS"), common.MustParseAddress("4XQgwcqPLM"), "vividsu252@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("24fyRTCXjjdomCWNANXQ6H2GSHm13jsCu9T2tZqDY9q"), common.MustParseAddress("3aTgBdBUW6"), "a01064792995@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4BagyWFwUxtTVRMQ434L8Xe4BQs2BpqTyZE2aj3bebu"), common.MustParseAddress("3W4iSfXeCL"), "podo7866@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("1oqKE9y9aV5mDFhNrKJgFBYtQ78C1xH6cMVpaHAEGE"), common.MustParseAddress("3jFbfYV97t"), "sda.sunyou@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("43KNdyzaMKVVUnThFomHMphZSFenUHD59AFReCMSNvS"), common.MustParseAddress("2w6WPU8tu9"), "daumgle@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("aEcFZ36dEnXVxTtCM3VVx6AQosXdM2jfvXZDyyeqLS"), common.MustParseAddress("2rhYeWV4bP"), "signalinguu@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2LwNLxYAPGAuvKnKJZs1fBe6L3jEEoA8HS7B5vwp5sF"), common.MustParseAddress("2pW4n2f9SW"), "insunga@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("28b1e6biGSenosbAK8j4JTFD8Gqc3YmLQfd9x71TgSg"), common.MustParseAddress("3MGnxkDyb6"), "kcc3311@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3f79zX4KRNVfjR5QM73yP1i7agkXwNF8uzde7Vovzbm"), common.MustParseAddress("3EgMMJkE8T"), "hanscall1000@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4TeLq5YmQMSFz2Ns5XW754zu2QfASuL7Cs2s1hZHvet"), common.MustParseAddress("3CUsUpvJya"), "xoxcenter@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("swiJGxq1fB6AsSHrXsqHa1cSEaK6iz66ZZMdTZMyYZ"), common.MustParseAddress("3AHPcM6Pph"), "koinpc@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("36xxWRtvdLrh87RqK3UJAGZG6aAGCpxs6HgvCTF6U1f"), common.MustParseAddress("385ujsGUfp"), "navercoco@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("6yGeE4mb2VTDLyPMUt1ZD2AwkXCkrxf6gj9noHuV2x"), common.MustParseAddress("2hudAbBPz9"), "ukpakaemmanuel@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3NxyMTzHVqxsekRvgmM6txkn7wmoMwTgKLFh3hL33Bg"), common.MustParseAddress("oA6H5MQJm"), "hss1012@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("oaEFzRjzcNeUKBfwvAie9HNWPyCfFq7vq44u1Ltcok"), common.MustParseAddress("wx1kzf4vZ"), "a01090574706@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3Pz78nVPEynvgRdhTArHtg8D6sJVKcX9LBvvRwcDdGp"), common.MustParseAddress("ukXtWq9mg"), "Culikim6804@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2JjKVtmWMaiop7xEDFpUVaShbJt4gZgN2drAxNdgFxh"), common.MustParseAddress("XmjBiZzEt"), "hidaegu@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4LPP2uQ6QoFc2Jz8YrwFgLLecJjQTAb2B8diKB3s9u3"), common.MustParseAddress("RBHaH6EnF"), "mero912000@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3NZ2furwZvc8VD76mMiEqgqrBaYHvDyRJaj5kvZA36w"), common.MustParseAddress("LnKqKSQUV"), "big1219@hotmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3uyKDbJnmtruBNPESD1ibiYeZ9itNHjPbupLJB56E6V"), common.MustParseAddress("AFJFN4amXG"), "kes5364@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("Ku8nFSh9yCpanLfae2rfroPRa1Zm8x3L8pz29oKzVC"), common.MustParseAddress("AAuHd6vwDW"), "king8676@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2ELpxLmCjHYxP8t2BV91BwS7R7qjWNNN679bRPjL1d4"), common.MustParseAddress("A8hokd724d"), "himjay@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4VXv5DKNkbFtt5w4gqkneakfE5XeANRLrFtFnh1sUQy"), common.MustParseAddress("AdH44rqw4L"), "lcykool6@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("Nh3YFkhUR2SAt6hU4NR5LDSBeHAJ3NrRU7f6Ry94g5"), common.MustParseAddress("Ab5aCP21uT"), "Ksjykr@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2FsBtsPkTTb9tAn1g49Zk1oZtFSfp8qRfDjZcRGdjJ6"), common.MustParseAddress("AYt6KuC6ka"), "hunterterran@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("44X9nHhifrEFkopSN7upntFdwL8Ck7rYZuh3ivymz7a"), common.MustParseAddress("AWgcTRNBbh"), "dydwls17@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("24K6vxDk692FoCQtJAdynZMriH1MexkTb924YWc3kbE"), common.MustParseAddress("ASHeiTiMHw"), "kwak8513@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3Kpq4JR1CWkrdd1hstauUgRvUd6QgpVrv2hmq5QrhAe"), common.MustParseAddress("9VLdxU4STg"), "dkim0120@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("24LiVZiztsHvoF77D3sTrybcmQdZLPiwoXcNxsuPT7L"), common.MustParseAddress("A27N9BdGcG"), "stin1010@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("PSbZ77APZWQQfGBPqRJCa4ez6DmsugnPRLgeJR4b4D"), common.MustParseAddress("92xsWi9Sd8"), "eth00700@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4CuofGVxQxBvY8qReqJb4ukmD6rSRyupUc1owBRXmd3"), common.MustParseAddress("9LYiUYkmrS"), "jkp7969@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4LUDnstgbCfnfRmue8aQiuMsHVsbdmsYReZsdPDpwWL"), common.MustParseAddress("9BknzdT7Ev"), "manipedi.oscar@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("h6fXcPzg31trmsXTcYWUTupn9PakUacBUBNrBGFEVM"), common.MustParseAddress("8YPdCUQXdh"), "the_ri@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2kvGb7jwDuyMX5teHKJ7eqF7jQ2DwHDsG1P3X6H7eJE"), common.MustParseAddress("8WC9KzacUp"), "tmdql777@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2tb3eYCsmX4mq7XuN5jhWx1ZN899hN3EdCi7Kkx53jg"), common.MustParseAddress("8RoBb2vnB4"), "carry10@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2j4abHp7Qhg8LAee4PizRq6L6K9TgZ5zzJzjWwZXyJe"), common.MustParseAddress("8MQDr5GwsJ"), "beauflo@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("mSKjiTJPYLXA6sKqhfgsoXNDwyd3VecEcLti6XY6GR"), common.MustParseAddress("8H1G77d7ZY"), "lsywind3@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("mePJH8QvQGYbnutH7SdDmZs3bazFNCuQhZh46wCwn2"), common.MustParseAddress("8qyUAK1rs1"), "nulmisohj@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("SHvtuYyk1bkx7XnKNNSREZGDAm5KXxtAqdWX8Bid1N"), common.MustParseAddress("8omzHqBwi8"), "moobin98@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2bmNLokXi5WxnEkX3a3SP1amknnzGqRNqzBN82R373p"), common.MustParseAddress("7L4FM7yCjZ"), "kangjinhwan1@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4fwmc5thkFrSNAUMUSNmKdY777u5EJnMFE1aBz83jeE"), common.MustParseAddress("7HrmUe9Hag"), "talleban228@nate.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3bR2Vu2E4aRSsb24aBvnEQhnmnLo6HLE2RbhqYCkV64"), common.MustParseAddress("7DTojgVTGv"), "qifls3@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4A6ASNXA2ZteQDsiTWwoegDqhT6mctYkUNYFVTQA2mA"), common.MustParseAddress("76sN8F1hpH"), "ekzmghtm666@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("XPf9CXqSHfHwLQwjGXiQ5miRkCKmHMAeHuF2Wk2cwb"), common.MustParseAddress("7ZF8Zzvhf7"), "handsome890@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("5dQU3To3nzzUJG5q8AMbLEMCYkdrph4B8TQ9n9mm8A"), common.MustParseAddress("7UrAq3GsMM"), "dr.changjw@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4Yom8YwdNd2J1PJzaHDN9VXXHP6Z5L8fzd53JzBgdJe"), common.MustParseAddress("7SegxZSxCU"), "sdyeun0821@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("X9iuDg5cteBEyZmb8m74acndXEDV52wv1uWdLM33yF"), common.MustParseAddress("6itZRSkYHV"), "mainncoin@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("7dPqeTZHUxTgfAbtGGbZCKwLNgqL6LzNtAjXSZzQqw"), common.MustParseAddress("6JiGrAfTc6"), "6multiverse@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2WksLbTxEJ7VZLj3yyCwQEJjJoY8o5PSjYGsEXd1WnG"), common.MustParseAddress("6GWnygqYTD"), "cheildc@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("391aqEWfVWhMG1DPQ39w6t3ybTbqFEVjy7dGEZ7eA8f"), common.MustParseAddress("4ZcAp6fJZb"), "webnes@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2oUjXJ1GVfYf5tQ59RBhmZBLjgPsmKWjSvpphchTq3q"), common.MustParseAddress("42qSdP6UR5"), "rokakys@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3MM7LJMj1PoUu8ykzqvt5BFYvYQtpERXhTRYzZMzrRe"), common.MustParseAddress("4JDoijstVU"), "ossion0715@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2eFp5iu8nXhpoH43kE3DrnGn1wCPEZ1LSe7LaiVdFfC"), common.MustParseAddress("3oeZQW8yVs"), "colordrama@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4SYfKG1CkgGCqiqUM1tR1LGhYodTdW1FL2EUyf9ps38"), common.MustParseAddress("3Rfkhhsoy1"), "billar369@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3xmWJEHzbvNgSBStAu5Fuxb4Bk6vjV5tgXRJSp8gFKp"), common.MustParseAddress("3erdvaqJtV"), "bigkingstar@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2VERxq9oiQuTcAmvZKJ2SaEXXAZ91AVzsXY7vohuKXp"), common.MustParseAddress("3AHPcM6Ptt"), "skyjin48108@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4R3Az7f8PZaZfp1yxdNYWBovBPgaUpyvB2NBXUjgj4P"), common.MustParseAddress("3EgMMJkECi"), "fushigilove@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3sCwT8CdJU2fqCkJf9XhvoLoVtggBJy7VNkbDZtpQxR"), common.MustParseAddress("3GsqDna9Md"), "jjambbung@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("gba2hioeHeoVaz7abebhy7dNzSB3Dr5ktjTQ33LKTZ"), common.MustParseAddress("3K5K6GQ4WY"), "hys2667416@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("roU9LLYEs16M1ZXvUZhpAb2bbXBFrLj5h3gTvjZAnJ"), common.MustParseAddress("3MGnxkDyfT"), "dlkkmita@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2S3ShbBq2LApJ1xiMiWeX9z3WV4MX1NqYtSw3azJhjb"), common.MustParseAddress("2rhYeWV4fr"), "donoee@paran.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4HpS2CaLNrKHnT5z1fLizmaDgmp6Weg912cL2eMvzqE"), common.MustParseAddress("2tu2WzJypm"), "ab81215@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("49WVhBWPTuy7BwbdvhXZDxUtTGBd4C6Mi4ufPMHqgXG"), common.MustParseAddress("2w6WPU8tyg"), "kgfa8811@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4CcUzckrzz1AtQ6inaLvC1gwLzsW3AfnuvssF4u8bGP"), common.MustParseAddress("2yHzFwxp8b"), "96carrot@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("jayrt98WPZGoSSQH5DpRD4iJkCv6gJyPrT1VsG7sLz"), common.MustParseAddress("33gwzuceSR"), "pjp8@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3T8X6FT2aVU1NkGPdondmob7ntbdDyynZGy1kZ3fcXr"), common.MustParseAddress("2hudAbBQ4U"), "yoikoko@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3uScbhzWP1KCqBFbR9rSWwKxbHVsiZxTscBjg4ZZgQi"), common.MustParseAddress("2DLNrMSV4s"), "vvip0413@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4eyt8i8fZ7GtQDnTe3Mq4JWCuU51ccVEVvEPnLy1gQH"), common.MustParseAddress("2QKnCka4qS"), "candko@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("gW7mLCJbEuLRmR8b8h9B4dmegRvGSzVfygGpcMY5a5"), common.MustParseAddress("oA6H5MQPN"), "jsg042482@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("264DmCn4X9ByLS1V2baqNC49fT1fy7VH87sYG6EXMaV"), common.MustParseAddress("qMa9ZBKYH"), "gyeongdal.lee@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3qTjafJC5HivfMbWqiZHAgiGK433cYK1y627CP6jgUF"), common.MustParseAddress("VaFKEk5AL"), "gimme04k@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3FuCxZZWYKcvKDQJr5kUZdte4Y2SdobyGqN2ZhYPfVa"), common.MustParseAddress("XmjBiZzKF"), "chancem35@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("48jp6wRxFYisatyDE3KhMuJPWXDTyt6XediHMZmV7QE"), common.MustParseAddress("9nvUvJpnP"), "sogang0902@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("89QT5UQvQKZLER8kD5nADjndzTw2F4AczVTBpthaKq"), common.MustParseAddress("A6WKt9H6zB"), "youngkchoe@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("hgxSMJo2cVQYczt4WgfLZNvFzPXSxtKyyxj8YsFsQo"), common.MustParseAddress("AMtgyW4X4a"), "godbowling@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("C2Skdt8XFWHW2n6duxvKPijFvBX7SppMj84Nm6JDkm"), common.MustParseAddress("8jP2YsY7Uf"), "jaeeuncore@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("7QRLcmeeWtpNgb1yYW8DjJPBHJtU4zgFGSchfsqc94"), common.MustParseAddress("7yRR9H1nQw"), "sda-boy@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("xLNwqQD6fFeNFEnaFnRiYc1mdegqVf3R1a7n292UQt"), common.MustParseAddress("81cu1kqhZr"), "jozianaida@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("48Nn4KwVnR5abXDbFk6dLkop8c35QSEAoBt46MSSsPY"), common.MustParseAddress("8AQpVg9NBW"), "kissforsky@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2i4nr3wcq4QAAR6VKk7avPN7t7t7rU7h7m2ZzDXaryz"), common.MustParseAddress("7rpyXqY2xU"), "yong_heui@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3JsVfKkoPPNCy6FMxXuJwE5kpuBG26nZ9uN1HE8HBmr"), common.MustParseAddress("69vMNFMo4j"), "kdqp304@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2fA9bQRd9D65L1Unf488rDoz5XVibQXL4NJHrwuNFc8"), common.MustParseAddress("6EKK7D1dNZ"), "wtiger922@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("QedXcdTLAd3QP6xhmj5KURQFcfefLccwcuSYW9XTTr"), common.MustParseAddress("6GWnygqYXU"), "flaubert33@empas.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("pCQTWUh4oukbgmpTnfp2gnmnWwwBEDc4fjonKD98Jj"), common.MustParseAddress("5xvx1rEDJS"), "dasomasam@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2UowZ7EVCQWjghUE1LkMJcUc6xnyPqW7o6VVggZ5DNt"), common.MustParseAddress("5fM741ct5P"), "munchilgim@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4QppHFBNEhimhCsMdUWm2JoiefFvGEjgJWKbdqNnP24"), common.MustParseAddress("5jk4nyGiPD"), "withinliferefresh@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3MdcyQ2NX4hAiVyiDMKfWZvoYtyk4grNVuZq7WCcH2u"), common.MustParseAddress("5HNJMDMiYX"), "h23334@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("35vgPbeiCg2KY3ppeD4vaiqU4zJsb8R8R9QMf4fkMHk"), common.MustParseAddress("5FApUjXoPe"), "hy-rnd@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3W4Xs9zVyj9pZLqkhUbF3FkeTisDF6esv6TqvCMdRDx"), common.MustParseAddress("5SADq8fPAB"), "likejunho@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3XodPvZFLP9ej6dGCK4GHmUKRwhtNccdHCUcyCGYNnN"), common.MustParseAddress("5PxjxeqU1J"), "skiong1@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("429BabCWAi6rMmMiPuiK94oRoyhTELA1cY6pKv3rFMy"), common.MustParseAddress("4ynTPNkPKV"), "sin21231@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("pRJkG11ae3qBFPqVwzgFusLCfDqs1AD9sBh8kyND21"), common.MustParseAddress("4no42ycoZE"), "gdae.park@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2HFchn5iuTRDGAjPkSTyNrmF5YuJ48bfZijXVbaNnGp"), common.MustParseAddress("4ZcAp6fJe2"), "tedjunoksunchan0409@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2eomEd3mBxgR9n5Jrvj7y9stMFAd8kRNtgSMth4UihF"), common.MustParseAddress("3mT5Y2K4RM"), "boogunk@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4pRt25sAcPabDd6ajqABAkrGYT9nA2rXpWz5DKuonCv"), common.MustParseAddress("3YGCK9MZW9"), "wal007@nate.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("1R269ZfPs9hVRoXn5BJfvv5Kf8S2DfCtsZRmd7tP1f"), common.MustParseAddress("3CUsUpvK8E"), "mteguhs1396@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3JQ3rAWL5jaLcMEe7vmjg7a574L2xRa1Uea6C9oZnfQ"), common.MustParseAddress("3PUGqE3ttm"), "jungrok3217@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3UCzpxEvkDkt5Ct3CaGuMVnAy22QsPLuT9SMRydUS4b"), common.MustParseAddress("2tu2WzJyuC"), "sales@pansolution.kr", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("i7BAPcsZhVPtiGxaEtQBRronTy3qF2vrHxf75XAErj"), common.MustParseAddress("31VU8RnjMu"), "hschung1345@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2H2PeyHDnwqz7ySoB4Fi2p4FrELFLvT46GK6E2DCU3X"), common.MustParseAddress("2Z7hgfsjXD"), "namkyu0@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4ZmUvqE2cbrVBaXjfffYmSSzMKt5tzqpq1UKJjh81qb"), common.MustParseAddress("2nJauYqESh"), "7343101@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4Z6ZfMsVJE9qah7MwptoXaQ76YjgZjW4Qt16ZWAhNsM"), common.MustParseAddress("2UijwiDuDf"), "goodway999@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2Mwva3yUhbQnc574RQqAG26W2QzWnUTYjrEb1JLNAu4"), common.MustParseAddress("kxcQbXVJt"), "ahg4999@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3gSb4WcBGR174qG2nQqZJRLUZ3xt9Yji8MkanASWamX"), common.MustParseAddress("sZ4231Emb"), "pkproman@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4CxkVrH5j9vrJLzjz85GLEJftHqaFKaZFboY4L3uwBk"), common.MustParseAddress("JaqxqcVUC"), "rick.jwkim@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3K3wnN6Uy25DLWVTQ5HqZJ8rAauK8vryP4tifSAuE4e"), common.MustParseAddress("VaFKEk5Ej"), "kms6542@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2ZbwSj7thH34qLjZXQqsN73Ps3tzBt8JJMSogDQ8rfU"), common.MustParseAddress("EBtDsxfAe"), "k3084572@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2A4uyXgLqEkia3D2eah9Sd2K8NadHtfoWgGFA6Nmn1K"), common.MustParseAddress("ASHeiTiMSX"), "mas_m3x@ymail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("bprVSWwuBudf83b5CMh9Fa4M6roFLXBipacYB556fu"), common.MustParseAddress("AQ6AqytSHe"), "rouichan@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("22r6zbKHY1MqwtSdXuHRDYKgK98RzWKMuXZiyBbUtBq"), common.MustParseAddress("AAuHd6vwNS"), "parkgb1018@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2fM32PJWmpMqvqwYy7aaivjDfyE4Durpj14WmnFtp2N"), common.MustParseAddress("AFJFN4amgG"), "wkorando@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2vXw9Xurynr8qFM8mFw7gBa4pukSanqRA6BQUifZAY8"), common.MustParseAddress("A4Jr1fTBuw"), "kyokushinway@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3DXKWKwbx63X6dtawVjf41M1SWcec2hiMJ6MYZexh3t"), common.MustParseAddress("9DxGs7H2YP"), "yjg7461@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2WGeHiRhvJW1NdqhFQ5yfZBpGyv7PmYtVUY17Pd5wa2"), common.MustParseAddress("8jP2YsY7Z6"), "ayubatope1@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("yA9H3DdD6LkMNqqVsc8tLgv6kC6AG5vu7JawVy5bPe"), common.MustParseAddress("8TzfTWkhUw"), "baguny42@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("aXeBuKHLHWkXZWoBFS6UK8kxdQtrvjdJbB6oEnEDGu"), common.MustParseAddress("7i343vENRD"), "jsfil@daum.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2hkKWmba2kHKKDWq8jq93LPBmhF1wG6tcXWiCv2uiou"), common.MustParseAddress("7kEXvQ4HaA"), "freedom0710@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3ipmKfodxwE8mWmTHbATFkDWTusQnNHLoN78iUzzwWu"), common.MustParseAddress("7rpyXqY32s"), "braveyongjoo@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3LmQKMu1i9J2Ti1pQNdtJdACv9ee9ytqFWqN28KCJmd"), common.MustParseAddress("7wDwGoBsLh"), "cuism2000@hanmail.net", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("mVeHvDjQirgEoikazbnDTmSj4TMHmU1fcgXNw5Fy1c"), common.MustParseAddress("7X3ehX6nex"), "purple65@nate.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4BD2yFbokg5nCCT87QdbGfgCFBQ18LH7Be51niUGu1h"), common.MustParseAddress("7bScSUkcxn"), "jjsky7@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("EFSefZtKS6D3T91jwpNjniJUSiqUTH5rYTHZ9kLmtd"), common.MustParseAddress("76sN8F1hy9"), "jeongjinsuk6822@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4PkbqrL63x9VCwX1KxdWUvLbc4JPr93vJNxePbmXsnV"), common.MustParseAddress("7FfHcAKNao"), "sonchawoo@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("44dVULGGnB1pgLxpJyuwM15WVeHMvbP3brA5oHNFMYC"), common.MustParseAddress("6m63HvaTbE"), "sheep318@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4WfaaJuhBZgSzxJAyHfztVzFuggjHv9xZsQu96Jr6JY"), common.MustParseAddress("6zGvWoXxWi"), "joe82s@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4UsYvpPmbVTEfyPNSSmfyNNU7ASzXLgHtR1cu6fzXb9"), common.MustParseAddress("6LukieVNuh"), "mywisdomforever@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3MuZ5ztujKJ3xHoijEoRZxqaWKLVD7t5bhr8PAAeMnz"), common.MustParseAddress("5d9dBXnxzs"), "gim0772@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("49Fh4G8P7em111qAKnkVbZZKNm7FP1hCQQ4aH2LAjzv"), common.MustParseAddress("5jk4nyGiTg"), "ccarisma11@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4dQXQu4GFwsoLHtkUy4HKNasrSDD1PUe6cCUacTM7Sr"), common.MustParseAddress("4gCcRY94Av"), "yt.kasegu100@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("4hZw7kLWecofTiADmsutz5NuXP1HQpPWHmb7goaKrBh"), common.MustParseAddress("4boegaVDsJ"), "ljs@dr.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3fMGzU4g84XXsbNgiU565WpWaMcRDF4ZD58WAPmQdQu"), common.MustParseAddress("3t3X9TnoxZ"), "alessandropasqualetti@outlook.it", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2HQY2eWYYxVug5GHPjL2npQwiu2bNzkXqtBKXtMmxav"), common.MustParseAddress("2pW4n2f9fo"), "a01053732416@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3WAkPeYsC9KTCKmJXJYVWNPkzEMGtRYpyiusjBusfQb"), common.MustParseAddress("2bKBZ9hekX"), "babyfungus@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2rYYsA4XZcT2GoejbbraQ2uLhZXw6KEvoY7593HsXwQ"), common.MustParseAddress("2dWfRdXZuS"), "leesm7410@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3WJE6Rq33keRTQTm5SzjqbWkH6HmU6QgCR7RQSRbq7a"), common.MustParseAddress("2WvDpC3pSm"), "smurpe@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("1Ss8yNSyph9bqhxsSENaGCFQMqAAjXME315Skb7xnc"), common.MustParseAddress("2FXriqGQNe"), "yakupovdn@gmail.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("3KXK2UZ2QuDjVqWETFNVvjXhkLCnyQqjT5s35zv73w1"), common.MustParseAddress("2UijwiDuJ4"), "dadachilla@naver.com", amount.NewCoinAmount(0, 0))
		addSingleAccount(sp, ctw, common.MustParsePublicHash("2a7CyuNJHKY2VbMs8cAJK2rR54CcGb3KY5qP1igs4r9"), common.MustParseAddress("2N8JLGk9qP"), "sung3845@naver.com", amount.NewCoinAmount(0, 0))
	}
	if p, err := app.pm.ProcessByName("fleta.formulator"); err != nil {
		return err
	} else if fp, is := p.(*formulator.Formulator); !is {
		return types.ErrNotExistProcess
	} else {
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("3W4iSfXe3Y"), amount.MustParseAmount("4500000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3t3X9TnoaY"), amount.MustParseAmount("50543.17072692"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3t3X9TnoaY"), amount.MustParseAmount("13400.72518335"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3t3X9TnoaY"), amount.MustParseAmount("11076"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3t3X9TnoaY"), amount.MustParseAmount("11154"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3t3X9TnoaY"), amount.MustParseAmount("107770"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3xSUtRSdtN"), amount.MustParseAmount("30000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3mT5Y2K47u"), amount.MustParseAmount("19327.10743689"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("3oeZQW8yGr"), amount.MustParseAmount("20000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("31VU8Rnj43"), amount.MustParseAmount("4500000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("2rhYeWV4SX"), amount.MustParseAmount("22758.7193845"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2w6WPU8tkM"), amount.MustParseAmount("31235"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2w6WPU8tkM"), amount.MustParseAmount("200000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("2UijwiDtuo"), amount.MustParseAmount("120000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2SXG5EPykv"), amount.MustParseAmount("5000"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("ukXtWq9cp"), amount.MustParseAmount("17950"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("22LyVxJu5X"), amount.MustParseAmount("4208.28433145"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("AKhD72EbgE"), amount.MustParseAmount("13284"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("A8hokd71um"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("AD6mVakrDb"), amount.MustParseAmount("3940"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("Ab5aCP21kb"), amount.MustParseAmount("47000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("Ab5aCP21kb"), amount.MustParseAmount("13653"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("9yutGhoMJX"), amount.MustParseAmount("90000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("9yutGhoMJX"), amount.MustParseAmount("10000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("9wiQQDyS9e"), amount.MustParseAmount("250000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("8vNRuGfh1d"), amount.MustParseAmount("51551"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("7wDwGoBs2q"), amount.MustParseAmount("32440"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("7kEXvQ4HGS"), amount.MustParseAmount("200000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("8CcJN9yH7G"), amount.MustParseAmount("26668"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("74ftFmBnWY"), amount.MustParseAmount("9271.53558515"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("74ftFmBnWY"), amount.MustParseAmount("491281"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("6JiGrAfTTE"), amount.MustParseAmount("1346"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("6P7Eb8KHm4"), amount.MustParseAmount("3600"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("6P7Eb8KHm4"), amount.MustParseAmount("500"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("5KZnDhBdU7"), amount.MustParseAmount("89950"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("5MmG6B1Yd2"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("5MmG6B1Yd2"), amount.MustParseAmount("16900"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("4no42ycoKo"), amount.MustParseAmount("355871"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("56NtzpE8Z7"), amount.MustParseAmount("539950"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("4wayWtvTwb"), amount.MustParseAmount("143023.82"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("42qSdP6UGD"), amount.MustParseAmount("102580"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("3aTgBdBURf"), amount.MustParseAmount("2550"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3jFbfYV93T"), amount.MustParseAmount("48146.5"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("2pW4n2f9N5"), amount.MustParseAmount("12000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2rhYeWV4Wz"), amount.MustParseAmount("144511.6"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("3GsqDna9Cu"), amount.MustParseAmount("40000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("385ujsGUbP"), amount.MustParseAmount("428130"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("2hudAbBPuk"), amount.MustParseAmount("80002"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("2bKBZ9heT5"), amount.MustParseAmount("80000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2WvDpC3p9K"), amount.MustParseAmount("2000001"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("kxcQbXV5T"), amount.MustParseAmount("200000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("oA6H5MQEN"), amount.MustParseAmount("48660"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("cAgvgDpTw"), amount.MustParseAmount("18000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("RBHaH6Ehr"), amount.MustParseAmount("140000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("9q7xnnVgmU"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("9q7xnnVgmU"), amount.MustParseAmount("18788.66319774"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("9q7xnnVgmU"), amount.MustParseAmount("99900"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("8zmPeEKXPr"), amount.MustParseAmount("625"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("8zmPeEKXPr"), amount.MustParseAmount("95811.72"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("8zmPeEKXPr"), amount.MustParseAmount("200"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("9NkCM2agvv"), amount.MustParseAmount("66620"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("8RoBb2vn6d"), amount.MustParseAmount("72000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7rpyXqY2oU"), amount.MustParseAmount("159950"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("74ftFmBnay"), amount.MustParseAmount("104396"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("72UQPHMsSQ"), amount.MustParseAmount("178731"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("6m63HvaTND"), amount.MustParseAmount("28385"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("618RtL48JM"), amount.MustParseAmount("27090"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("618RtL48JM"), amount.MustParseAmount("64270"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("6C7qEjBi54"), amount.MustParseAmount("210"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("6C7qEjBi54"), amount.MustParseAmount("17790"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("5SADq8fP1B"), amount.MustParseAmount("100000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("5jk4nyGiEV"), amount.MustParseAmount("400000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("5fM741csvj"), amount.MustParseAmount("199950"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("51ywFraJKn"), amount.MustParseAmount("143950"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("4BdN7JQ8xA"), amount.MustParseAmount("127000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("4XQgwcqPLM"), amount.MustParseAmount("54000"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("3aTgBdBUW6"), amount.MustParseAmount("93338"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("2rhYeWV4bP"), amount.MustParseAmount("500000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2pW4n2f9SW"), amount.MustParseAmount("140572"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3MGnxkDyb6"), amount.MustParseAmount("100000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("3EgMMJkE8T"), amount.MustParseAmount("2000000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("3CUsUpvJya"), amount.MustParseAmount("2000000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("3AHPcM6Pph"), amount.MustParseAmount("1000000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("385ujsGUfp"), amount.MustParseAmount("2000000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("2hudAbBPz9"), amount.MustParseAmount("2760"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("oA6H5MQJm"), amount.MustParseAmount("3100"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("oA6H5MQJm"), amount.MustParseAmount("70000"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("wx1kzf4vZ"), amount.MustParseAmount("475"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("RBHaH6EnF"), amount.MustParseAmount("27033"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("LnKqKSQUV"), amount.MustParseAmount("100000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("AFJFN4amXG"), amount.MustParseAmount("16220"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("AAuHd6vwDW"), amount.MustParseAmount("1500000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("9VLdxU4STg"), amount.MustParseAmount("386032"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("A27N9BdGcG"), amount.MustParseAmount("19365"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("A27N9BdGcG"), amount.MustParseAmount("1844.2"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("92xsWi9Sd8"), amount.MustParseAmount("340616"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("9BknzdT7Ev"), amount.MustParseAmount("108132"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("8YPdCUQXdh"), amount.MustParseAmount("168098.02"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("8RoBb2vnB4"), amount.MustParseAmount("178418"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("8MQDr5GwsJ"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("7L4FM7yCjZ"), amount.MustParseAmount("34603"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("7HrmUe9Hag"), amount.MustParseAmount("35999"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7DTojgVTGv"), amount.MustParseAmount("100000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7DTojgVTGv"), amount.MustParseAmount("10000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("76sN8F1hpH"), amount.MustParseAmount("35950"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7ZF8Zzvhf7"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7ZF8Zzvhf7"), amount.MustParseAmount("149901"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("7UrAq3GsMM"), amount.MustParseAmount("36000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("6itZRSkYHV"), amount.MustParseAmount("3000000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("4ZcAp6fJZb"), amount.MustParseAmount("40000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3Rfkhhsoy1"), amount.MustParseAmount("1095"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("3Rfkhhsoy1"), amount.MustParseAmount("508.9375209"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("3erdvaqJtV"), amount.MustParseAmount("6000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("3AHPcM6Ptt"), amount.MustParseAmount("27033"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("3EgMMJkECi"), amount.MustParseAmount("65000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3MGnxkDyfT"), amount.MustParseAmount("50000"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2tu2WzJypm"), amount.MustParseAmount("20000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2w6WPU8tyg"), amount.MustParseAmount("140000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2hudAbBQ4U"), amount.MustParseAmount("121109"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2DLNrMSV4s"), amount.MustParseAmount("990"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2DLNrMSV4s"), amount.MustParseAmount("560000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2QKnCka4qS"), amount.MustParseAmount("34603"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("oA6H5MQPN"), amount.MustParseAmount("16264"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("qMa9ZBKYH"), amount.MustParseAmount("30000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("A6WKt9H6zB"), amount.MustParseAmount("162000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("AMtgyW4X4a"), amount.MustParseAmount("54066"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("8jP2YsY7Uf"), amount.MustParseAmount("120000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7yRR9H1nQw"), amount.MustParseAmount("100000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7yRR9H1nQw"), amount.MustParseAmount("33240"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("8AQpVg9NBW"), amount.MustParseAmount("54066"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("7rpyXqY2xU"), amount.MustParseAmount("160000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("69vMNFMo4j"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("69vMNFMo4j"), amount.MustParseAmount("200"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("6EKK7D1dNZ"), amount.MustParseAmount("75505"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("6GWnygqYXU"), amount.MustParseAmount("130000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("5fM741ct5P"), amount.MustParseAmount("26618"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("5HNJMDMiYX"), amount.MustParseAmount("133477.6666666667"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("5FApUjXoPe"), amount.MustParseAmount("750000"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("5SADq8fPAB"), amount.MustParseAmount("750000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("4ZcAp6fJe2"), amount.MustParseAmount("143950"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("3mT5Y2K4RM"), amount.MustParseAmount("79770"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("3PUGqE3ttm"), amount.MustParseAmount("800000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2tu2WzJyuC"), amount.MustParseAmount("162198"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("31VU8RnjMu"), amount.MustParseAmount("37260"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("31VU8RnjMu"), amount.MustParseAmount("66670"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("2Z7hgfsjXD"), amount.MustParseAmount("200"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2nJauYqESh"), amount.MustParseAmount("1450"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("2nJauYqESh"), amount.MustParseAmount("1300"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("sZ4231Emb"), amount.MustParseAmount("190"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("sZ4231Emb"), amount.MustParseAmount("830"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("sZ4231Emb"), amount.MustParseAmount("450"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("JaqxqcVUC"), amount.MustParseAmount("200000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("VaFKEk5Ej"), amount.MustParseAmount("30000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("VaFKEk5Ej"), amount.MustParseAmount("1000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("AAuHd6vwNS"), amount.MustParseAmount("200"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("AFJFN4amgG"), amount.MustParseAmount("178874"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("A4Jr1fTBuw"), amount.MustParseAmount("17302"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7rpyXqY32s"), amount.MustParseAmount("40000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7X3ehX6nex"), amount.MustParseAmount("1255"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7X3ehX6nex"), amount.MustParseAmount("520"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("7bScSUkcxn"), amount.MustParseAmount("5000000"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("7FfHcAKNao"), amount.MustParseAmount("300"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("6m63HvaTbE"), amount.MustParseAmount("540"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("5d9dBXnxzs"), amount.MustParseAmount("2084"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("5jk4nyGiTg"), amount.MustParseAmount("51759.40200083"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("4boegaVDsJ"), amount.MustParseAmount("2537"))
		addStaking(fp, ctw, common.MustParseAddress("385ujsGNZt"), common.MustParseAddress("3t3X9TnoxZ"), amount.MustParseAmount("400"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("2pW4n2f9fo"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("2pW4n2f9fo"), amount.MustParseAmount("285"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("2bKBZ9hekX"), amount.MustParseAmount("100"))
		addStaking(fp, ctw, common.MustParseAddress("7bScSUoST"), common.MustParseAddress("2dWfRdXZuS"), amount.MustParseAmount("477"))
		addStaking(fp, ctw, common.MustParseAddress("GPN6MnU3y"), common.MustParseAddress("2WvDpC3pSm"), amount.MustParseAmount("10000"))
		addStaking(fp, ctw, common.MustParseAddress("9nvUvJibL"), common.MustParseAddress("2FXriqGQNe"), amount.MustParseAmount("420"))
		addStaking(fp, ctw, common.MustParseAddress("3EgMMJk82X"), common.MustParseAddress("2UijwiDuJ4"), amount.MustParseAmount("140000"))
		addStaking(fp, ctw, common.MustParseAddress("3AHPcM6Him"), common.MustParseAddress("2N8JLGk9qP"), amount.MustParseAmount("70000"))
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
		FormulatorType: formulator.AlphaFormulatorType,
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
	if acc, err := ctw.Account(HyperAddress); err != nil {
		panic(err)
	} else if frAcc, is := acc.(*formulator.FormulatorAccount); !is {
		panic(formulator.ErrInvalidFormulatorAddress)
	} else if frAcc.FormulatorType != formulator.HyperFormulatorType {
		panic(formulator.ErrNotHyperFormulator)
	}
	if has, err := ctw.HasAccount(StakingAddress); err != nil {
		panic(err)
	} else if !has {
		panic(types.ErrNotExistAccount)
	}
	fp.AddStakingAmount(ctw, HyperAddress, StakingAddress, am)

	gAmount = gAmount.Add(am)
}
