package test

import (
	"bytes"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/tests/lib"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// formulatorOwnerMap type
type formulatorOwnerMap map[common.Address]bool

// add owner
func (m formulatorOwnerMap) add(owner common.Address) {
	m[owner] = true
}

// delete owner
func (m formulatorOwnerMap) delete(owner common.Address) {
	delete(m, owner)
}

// TestContract is a Go TestContractContract Contract
type testContractContract struct {
	GoContract
}

// SetGeneratorTx returns the SetGenerator method tx of the TestContractContract contract
func (c *testContractContract) SetGeneratorTx(senderKey key.Key) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "SetGenerator")
}

var _ = Describe("Formulator to Generator Test", func() {

	userKeys, err := GetSingers(ChainID)
	if err != nil {
		panic(err)
	}

	aliceKey, bobKey, charlieKey, eveKey := userKeys[0], userKeys[1], userKeys[2], userKeys[3]
	alice, bob, charlie, eve := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address(), eveKey.PublicKey().Address()

	var tb *TestBlockChain

	var initialAmount = amount.NewAmount(1000000000000000000, 0)

	assertFunc := func(frCont *FormulatorContract, frOwners formulatorOwnerMap) {
		// formulatorMap
		frMap := NewJsonClient(tb).ViewCall(frCont.Address, "FormulatorMap").([]interface{})[0].(map[common.Address]*formulator.Formulator)
		frOwnerSet := make(map[common.Address]bool)
		for _, v := range frMap {
			frOwnerSet[v.Owner] = true
		}
		Expect(len(frOwnerSet)).To(Equal(len(frOwners)))
		for k := range frOwnerSet {
			Expect(frOwners[k]).To(BeTrue())
		}

		// generator
		generators, err := tb.Store.Generators()
		Expect(err).To(Succeed())
		Expect(len(generators)).To(Equal(len(frOwners) + 1))

		expectedGenerators := []common.Address{}
		for k := range frOwners {
			expectedGenerators = append(expectedGenerators, k)
		}
		for k := range tb.FrKeyMap {
			expectedGenerators = append(expectedGenerators, k)
		}
		Expect(HaveSameElements(generators, expectedGenerators)).To(BeTrue())
	}

	createAlpha := func(ownerKey key.Key, frCont *FormulatorContract, frOwners formulatorOwnerMap) common.Address {
		var tokenID common.Address
		b := tb.MustAddBlock([]*TxWithSigner{frCont.CreateAlphaTx(ownerKey)})
		for _, event := range b.Body.Events {
			if event.Index == 0 && event.Type == ctypes.EventTagTxMsg {
				tokenID = common.BytesToAddress(event.Result)
			}
		}
		// add formularOwner
		frOwners.add(ownerKey.PublicKey().Address())
		return tokenID
	}

	createSigma := func(ownerKey key.Key, frCont *FormulatorContract, frOwners formulatorOwnerMap) common.Address {
		tokenIDs := []common.Address{}
		for i := uint8(0); i < DefaultFormulatorPolicy.SigmaCount; i++ {
			tokenIDs = append(tokenIDs, createAlpha(ownerKey, frCont, frOwners))
		}
		// after sigmablocks
		for i := uint32(0); i < DefaultFormulatorPolicy.SigmaBlocks; i++ {
			tb.AddBlock([]*TxWithSigner{})
		}

		tb.AddBlock([]*TxWithSigner{frCont.CreateSigmaTx(ownerKey, tokenIDs)})
		return tokenIDs[0]
	}

	createOmega := func(ownerKey key.Key, frCont *FormulatorContract, frOwners formulatorOwnerMap) common.Address {
		sigmaIDs := []common.Address{}
		for i := uint8(0); i < DefaultFormulatorPolicy.OmegaCount; i++ {
			sigmaIDs = append(sigmaIDs, createSigma(ownerKey, frCont, frOwners))
		}

		// after omegablocks
		for i := uint32(0); i < DefaultFormulatorPolicy.OmegaBlocks; i++ {
			tb.AddBlock([]*TxWithSigner{})
		}
		tb.AddBlock([]*TxWithSigner{frCont.CreateOmegaTx(ownerKey, sigmaIDs)})
		return sigmaIDs[0]
	}

	Describe("SetGenerator", func() {

		var contractAddress common.Address

		BeforeEach(func() {
			intialize := func(ctx *types.Context, classMap map[string]uint64) error {
				initSupplyMap := make(map[common.Address]*amount.Amount)
				initSupplyMap[alice] = initialAmount
				_, err := MevInitialize(ctx, classMap, alice, initSupplyMap)
				if err != nil {
					return err
				}
				return nil
			}

			tb = NewTestBlockChain(ChainDataPath, true, ChainID, Version, alice, intialize, DefaultInitContextInfo)

			// deploy contract
			data := &chain.DeployContractData{
				Owner:   alice,
				ClassID: tb.ClassMap["NotFormulatorContract"],
				Args:    []byte{},
			}

			bs2 := bytes.NewBuffer([]byte{})
			_, err := data.WriteTo(bs2)
			if err != nil {
				panic(err)
			}

			tx := &TxWithSigner{
				Tx: &types.Transaction{
					ChainID:   tb.ChainID,
					Timestamp: uint64(time.Now().UnixNano()),
					To:        common.Address{},
					Method:    "Contract.Deploy",
					Args:      bs2.Bytes(),
				},
				Signer: aliceKey,
			}

			b := tb.MustAddBlock([]*TxWithSigner{tx})

			for _, event := range b.Body.Events {
				if event.Index == 0 && event.Type == ctypes.EventTagTxMsg {
					contractAddress = common.BytesToAddress(event.Result)
				}
			}

		})
		AfterEach(func() {
			tb.Close()
		})

		It("only formulator allowed", func() {
			tc := &testContractContract{GoContract: GoContract{Address: &contractAddress, Provider: tb.Provider}}
			_, err := tb.AddBlock([]*TxWithSigner{tc.SetGeneratorTx(aliceKey)})
			Expect(err).To(MatchError(types.ErrOnlyFormulatorAllowed))
		})
	})

	Describe("SyncGenerator", func() {

		BeforeEach(func() {})
		AfterEach(func() {
			tb.Close()
		})

		testSyncGenerator := func(senderKey key.Key, initAlphas, initSigmas, initOmegas []common.Address) error {

			var formulatorAddress *common.Address

			intialize := func(ctx *types.Context, classMap map[string]uint64) error {
				initSupplyMap := make(map[common.Address]*amount.Amount)
				for _, v := range userKeys {
					initSupplyMap[v.PublicKey().Address()] = initialAmount
				}
				mevAddress, err := MevInitialize(ctx, classMap, alice, initSupplyMap)
				if err != nil {
					return err
				}
				formulatorAddress, err = FormulatorInitialize(ctx, classMap, *mevAddress, alice, initAlphas, initSigmas, initOmegas)
				if err != nil {
					return err
				}
				return nil
			}

			tb = NewTestBlockChain(ChainDataPath, true, ChainID, Version, alice, intialize, DefaultInitContextInfo)
			frCont := BindFormulatorContract(formulatorAddress, tb.Provider)
			_, err := tb.AddBlock([]*TxWithSigner{frCont.SyncGeneratorTx(senderKey)})
			if err != nil {
				return err
			}

			frOwners := make(map[common.Address]bool)
			for _, owners := range [][]common.Address{initAlphas, initSigmas, initOmegas} {
				for _, v := range owners {
					frOwners[v] = true
				}
			}

			assertFunc(frCont, frOwners)

			return nil

		}

		It("alpha", func() {
			initAlphas := []common.Address{bob}
			testSyncGenerator(aliceKey, initAlphas, nil, nil)
		})
		It("sigma", func() {
			initSigmas := []common.Address{bob, charlie}
			testSyncGenerator(aliceKey, nil, initSigmas, nil)
		})
		It("omega", func() {
			initOmegas := []common.Address{alice, bob, charlie}
			testSyncGenerator(aliceKey, nil, nil, initOmegas)
		})

		It("only admin", func() {
			initOmegas := []common.Address{alice, bob, charlie}
			err := testSyncGenerator(bobKey, nil, nil, initOmegas)
			Expect(err).To(MatchError("is not master"))
		})

	})

	Describe("After Sync", func() {

		var frCont *FormulatorContract
		var frOwners formulatorOwnerMap
		initAlphas := []common.Address{alice}
		initSigmas := []common.Address{bob}
		initOmegas := []common.Address{bob}

		BeforeEach(func() {

			var mevAddress, formulatorAddress *common.Address
			frOwners = formulatorOwnerMap{}

			intialize := func(ctx *types.Context, classMap map[string]uint64) error {
				initSupplyMap := make(map[common.Address]*amount.Amount)
				for _, v := range userKeys {
					initSupplyMap[v.PublicKey().Address()] = initialAmount
				}
				var err error
				mevAddress, err = MevInitialize(ctx, classMap, alice, initSupplyMap)
				if err != nil {
					return err
				}

				formulatorAddress, err = FormulatorInitialize(ctx, classMap, *mevAddress, alice, initAlphas, initSigmas, initOmegas)
				if err != nil {
					return err
				}
				return nil
			}

			tb = NewTestBlockChain(ChainDataPath, true, ChainID, Version, alice, intialize, DefaultInitContextInfo)

			mev := BindTokenContract(mevAddress, tb.Provider)
			// alice, bob, charlie eve approve
			for i := 0; i < len(userKeys); i++ {
				tb.AddBlock([]*TxWithSigner{mev.ApproveTx(userKeys[i], *formulatorAddress, MaxUint256)})
			}

			frCont = BindFormulatorContract(formulatorAddress, tb.Provider)
			tb.AddBlock([]*TxWithSigner{frCont.SyncGeneratorTx(aliceKey)})

			// frOwners initialize
			for _, owners := range [][]common.Address{initAlphas, initSigmas, initOmegas} {
				for _, v := range owners {
					frOwners.add(v)
				}
			}

		})
		AfterEach(func() {
			tb.Close()
		})

		Describe("Create", func() {
			DescribeTable("",
				func(f func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					f(charlieKey, frCont, frOwners)
					assertFunc(frCont, frOwners)
				},
				Entry("alpha", createAlpha),
				Entry("sigma", createSigma),
				Entry("omega", createOmega),
			)
		})

		Describe("Revoke", func() {
			DescribeTable("sigle",
				func(f func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID := f(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(charlieKey, tokenID)})
					frOwners.delete(charlie)
					assertFunc(frCont, frOwners)
				},
				Entry("alpha", createAlpha),
				Entry("sigma", createSigma),
				Entry("omega", createOmega),
			)

			DescribeTable("twos revoke first",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID := f1(charlieKey, frCont, frOwners)
					f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(charlieKey, tokenID)})
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(3))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)
			DescribeTable("twos revoke first",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					f1(charlieKey, frCont, frOwners)
					tokenID := f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(charlieKey, tokenID)})
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(3))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)
			DescribeTable("twos revoke all",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID1 := f1(charlieKey, frCont, frOwners)
					tokenID2 := f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(charlieKey, tokenID1)})
					tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(charlieKey, tokenID2)})
					frOwners.delete(charlie)
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(2))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)
		})

		Describe("TansferFrom", func() {
			DescribeTable("sigle",
				func(f func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID := f(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(charlieKey, charlie, eve, tokenID)})
					frOwners.delete(charlie)
					frOwners.add(eve)
					assertFunc(frCont, frOwners)
				},
				Entry("alpha", createAlpha),
				Entry("sigma", createSigma),
				Entry("omega", createOmega),
			)

			DescribeTable("twos transfer first",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID := f1(charlieKey, frCont, frOwners)
					f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(charlieKey, charlie, eve, tokenID)})
					frOwners.add(eve)
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(4))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)

			DescribeTable("twos transfer second",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID := f1(charlieKey, frCont, frOwners)
					f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(charlieKey, charlie, eve, tokenID)})
					frOwners.add(eve)
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(4))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)

			DescribeTable("twos transfer all",
				func(f1, f2 func(key.Key, *FormulatorContract, formulatorOwnerMap) common.Address) {
					tokenID1 := f1(charlieKey, frCont, frOwners)
					tokenID2 := f2(charlieKey, frCont, frOwners)
					tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(charlieKey, charlie, eve, tokenID1)})
					tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(charlieKey, charlie, eve, tokenID2)})
					frOwners.delete(charlie)
					frOwners.add(eve)
					assertFunc(frCont, frOwners)
					Expect(len(frOwners)).To(Equal(3))
				},
				Entry("alpha, alpha", createAlpha, createAlpha),
				Entry("alpha, sigma", createAlpha, createSigma),
				Entry("alpha, omega", createAlpha, createOmega),
				Entry("sigma, alpha", createSigma, createAlpha),
				Entry("sigma, sigma", createSigma, createSigma),
				Entry("sigma, omega", createSigma, createOmega),
				Entry("omega, alpha", createOmega, createAlpha),
				Entry("omega, sigma", createOmega, createSigma),
				Entry("omega, omega", createOmega, createOmega),
			)
		})

		// BuyFormulator 기능 삭제 :  20230626
		// Describe("BuyFormulator", func() {
		// 	DescribeTable("single",
		// 		func(f func(key.Key) common.Address) {
		// 			tokenID := f(charlieKey)
		// 			tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(charlieKey, tokenID, amount.NewAmount(1, 0))})
		// 			tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(eveKey, tokenID)})
		// 			frOwners.delete(charlie)
		// 			frOwners.add(eve)
		// 			assertFunc(frCont, frOwners)
		// 		},
		// 		Entry("alpha", createAlpha),
		// 		Entry("sigma", createSigma),
		// 		Entry("omega", createOmega),
		// 	)

		// 	DescribeTable("twos buy first",
		// 		func(f1, f2 func(key.Key) common.Address) {
		// 			tokenID := f1(charlieKey)
		// 			f2(charlieKey)
		// 			tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(charlieKey, tokenID, amount.NewAmount(1, 0))})
		// 			tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(eveKey, tokenID)})
		// 			frOwners.add(eve)
		// 			assertFunc(frCont, frOwners)
		// 			Expect(len(frOwners)).To(Equal(4))
		// 		},
		// 		Entry("alpha, alpha", createAlpha, createAlpha),
		// 		Entry("alpha, sigma", createAlpha, createSigma),
		// 		Entry("alpha, omega", createAlpha, createOmega),
		// 		Entry("sigma, alpha", createSigma, createAlpha),
		// 		Entry("sigma, sigma", createSigma, createSigma),
		// 		Entry("sigma, omega", createSigma, createOmega),
		// 		Entry("omega, alpha", createOmega, createAlpha),
		// 		Entry("omega, sigma", createOmega, createSigma),
		// 		Entry("omega, omega", createOmega, createOmega),
		// 	)

		// 	DescribeTable("twos buy second",
		// 		func(f1, f2 func(key.Key) common.Address) {
		// 			f1(charlieKey)
		// 			tokenID := f2(charlieKey)
		// 			tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(charlieKey, tokenID, amount.NewAmount(1, 0))})
		// 			tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(eveKey, tokenID)})
		// 			frOwners.add(eve)
		// 			assertFunc(frCont, frOwners)
		// 			Expect(len(frOwners)).To(Equal(4))
		// 		},
		// 		Entry("alpha, alpha", createAlpha, createAlpha),
		// 		Entry("alpha, sigma", createAlpha, createSigma),
		// 		Entry("alpha, omega", createAlpha, createOmega),
		// 		Entry("sigma, alpha", createSigma, createAlpha),
		// 		Entry("sigma, sigma", createSigma, createSigma),
		// 		Entry("sigma, omega", createSigma, createOmega),
		// 		Entry("omega, alpha", createOmega, createAlpha),
		// 		Entry("omega, sigma", createOmega, createSigma),
		// 		Entry("omega, omega", createOmega, createOmega),
		// 	)

		// 	DescribeTable("twos buy all",
		// 		func(f1, f2 func(key.Key) common.Address) {
		// 			tokenID1 := f1(charlieKey)
		// 			tokenID2 := f2(charlieKey)
		// 			tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(charlieKey, tokenID1, amount.NewAmount(1, 0))})
		// 			tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(eveKey, tokenID1)})
		// 			tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(charlieKey, tokenID2, amount.NewAmount(2, 0))})
		// 			tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(eveKey, tokenID2)})
		// 			frOwners.add(eve)
		// 			frOwners.delete(charlie)
		// 			assertFunc(frCont, frOwners)
		// 			Expect(len(frOwners)).To(Equal(3))
		// 		},
		// 		Entry("alpha, alpha", createAlpha, createAlpha),
		// 		Entry("alpha, sigma", createAlpha, createSigma),
		// 		Entry("alpha, omega", createAlpha, createOmega),
		// 		Entry("sigma, alpha", createSigma, createAlpha),
		// 		Entry("sigma, sigma", createSigma, createSigma),
		// 		Entry("sigma, omega", createSigma, createOmega),
		// 		Entry("omega, alpha", createOmega, createAlpha),
		// 		Entry("omega, sigma", createOmega, createSigma),
		// 		Entry("omega, omega", createOmega, createOmega),
		// 	)

		// })

		Describe("Random", func() {
			const maxIters = 1000

			const users = 4
			//actions := []string{"createAlpha", "createSigma", "createOmega", "revoke", "transfer", "buy"}
			actions := []string{"createAlpha", "createSigma", "createOmega", "revoke", "transfer"}
			tokenIDs := [users][]common.Address{}

			It("", func() {
				// initialize tokenIDs
				frMap := NewJsonClient(tb).ViewCall(frCont.Address, "FormulatorMap").([]interface{})[0].(map[common.Address]*formulator.Formulator)
				for k, v := range frMap {
					for i := 0; i < users; i++ {
						if userKeys[i].PublicKey().Address() == v.Owner {
							tokenIDs[i] = append(tokenIDs[i], k)
						}
					}
				}

				Expect(len(tokenIDs[0])).To(Equal(1)) // alice 1
				Expect(len(tokenIDs[1])).To(Equal(2)) // bob 2
				Expect(len(tokenIDs[2])).To(Equal(0)) // charlie 0
				Expect(len(tokenIDs[3])).To(Equal(0)) // eve 0

				i := 0
				for i < maxIters {
					action := actions[rand.Intn(len(actions))]
					sIdx := rand.Intn(users) // sender index
					sKey := userKeys[sIdx]
					sAddress := sKey.PublicKey().Address()

					switch action {
					case "createAlpha":
						tokenID := createAlpha(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "createSigma":
						tokenID := createSigma(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "createOmega":
						tokenID := createOmega(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "revoke":
						if len(tokenIDs[sIdx]) == 0 {
							continue
						}

						tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(sKey, tokenIDs[sIdx][0])})
						tokenIDs[sIdx] = tokenIDs[sIdx][1:]
						if len(tokenIDs[sIdx]) == 0 {
							frOwners.delete(sAddress)
						}
					case "transfer":
						if len(tokenIDs[sIdx]) == 0 {
							continue
						}
						// not same
						var toIdx int
						for {
							toIdx = rand.Intn(users)
							if toIdx != sIdx {
								break
							}
						}
						tAddress := userKeys[toIdx].PublicKey().Address()
						tokenID := tokenIDs[sIdx][0]
						tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(sKey, sAddress, tAddress, tokenID)})
						tokenIDs[sIdx] = tokenIDs[sIdx][1:]
						if len(tokenIDs[sIdx]) == 0 {
							frOwners.delete(sAddress)
						}
						tokenIDs[toIdx] = append(tokenIDs[toIdx], tokenID)
						frOwners.add(tAddress)
					// case "buy":
					// 	size := len(tokenIDs[sIdx])
					// 	if size == 0 {
					// 		continue
					// 	}
					// 	// not same
					// 	var buyerIdx int
					// 	for {
					// 		buyerIdx = rand.Intn(users)
					// 		if buyerIdx != sIdx {
					// 			break
					// 		}
					// 	}
					// 	buyerKey := userKeys[buyerIdx]
					// 	buyerAddress := buyerKey.PublicKey().Address()

					// 	tokenID := tokenIDs[sIdx][0]
					// 	tb.AddBlock([]*TxWithSigner{frCont.RegisterSalesTx(sKey, tokenID, amount.NewAmount(1, 0))})
					// 	tb.AddBlock([]*TxWithSigner{frCont.BuyFormulatorTx(buyerKey, tokenID)})
					// 	tokenIDs[sIdx] = tokenIDs[sIdx][1:]
					// 	if len(tokenIDs[sIdx]) == 0 {
					// 		frOwners.delete(sAddress)
					// 	}
					// 	tokenIDs[buyerIdx] = append(tokenIDs[buyerIdx], tokenID)
					// 	frOwners.add(buyerAddress)
					// 	Expect(len(tokenIDs[sIdx])).To(Equal(size - 1))
					default:
						panic("non-existent action : " + action)
					}

					assertFunc(frCont, frOwners)
					i++

					// log.Printf("%v : %v : %v", i, sIdx, action)
				}
			})
		})
	})

	// to show no change before SyncGenerator
	Describe("Before Sync", func() {

		var frCont *FormulatorContract
		var frOwners formulatorOwnerMap
		initAlphas := []common.Address{alice}
		initSigmas := []common.Address{bob}
		initOmegas := []common.Address{bob}

		BeforeEach(func() {

			var mevAddress, formulatorAddress *common.Address
			frOwners = formulatorOwnerMap{}

			intialize := func(ctx *types.Context, classMap map[string]uint64) error {
				initSupplyMap := make(map[common.Address]*amount.Amount)
				for _, v := range userKeys {
					initSupplyMap[v.PublicKey().Address()] = initialAmount
				}
				var err error
				mevAddress, err = MevInitialize(ctx, classMap, alice, initSupplyMap)
				if err != nil {
					return err
				}

				formulatorAddress, err = FormulatorInitialize(ctx, classMap, *mevAddress, alice, initAlphas, initSigmas, initOmegas)
				if err != nil {
					return err
				}
				return nil
			}

			tb = NewTestBlockChain(ChainDataPath, true, ChainID, Version, alice, intialize, DefaultInitContextInfo)

			mev := BindTokenContract(mevAddress, tb.Provider)
			// alice, bob, charlie eve approve
			for i := 0; i < len(userKeys); i++ {
				tb.AddBlock([]*TxWithSigner{mev.ApproveTx(userKeys[i], *formulatorAddress, MaxUint256)})
			}

			frCont = BindFormulatorContract(formulatorAddress, tb.Provider)

			// frOwners initialize
			for _, owners := range [][]common.Address{initAlphas, initSigmas, initOmegas} {
				for _, v := range owners {
					frOwners.add(v)
				}
			}
		})
		AfterEach(func() {
			tb.Close()
		})

		Describe("Random", func() {
			const maxIters = 1000

			const users = 4
			//actions := []string{"createAlpha", "createSigma", "createOmega", "revoke", "transfer", "buy"}
			actions := []string{"createAlpha", "createSigma", "createOmega", "revoke", "transfer"}
			tokenIDs := [users][]common.Address{}

			It("no change in generators", func() {
				// initialize tokenIDs
				frMap := NewJsonClient(tb).ViewCall(frCont.Address, "FormulatorMap").([]interface{})[0].(map[common.Address]*formulator.Formulator)
				for k, v := range frMap {
					for i := 0; i < users; i++ {
						if userKeys[i].PublicKey().Address() == v.Owner {
							tokenIDs[i] = append(tokenIDs[i], k)
						}
					}
				}

				Expect(len(tokenIDs[0])).To(Equal(1)) // alice 1
				Expect(len(tokenIDs[1])).To(Equal(2)) // bob 2
				Expect(len(tokenIDs[2])).To(Equal(0)) // charlie 0
				Expect(len(tokenIDs[3])).To(Equal(0)) // eve 0

				i := 0
				for i < maxIters {
					action := actions[rand.Intn(len(actions))]
					sIdx := rand.Intn(users) // sender index
					sKey := userKeys[sIdx]
					sAddress := sKey.PublicKey().Address()

					switch action {
					case "createAlpha":
						tokenID := createAlpha(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "createSigma":
						tokenID := createSigma(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "createOmega":
						tokenID := createOmega(sKey, frCont, frOwners)
						tokenIDs[sIdx] = append(tokenIDs[sIdx], tokenID)
					case "revoke":
						if len(tokenIDs[sIdx]) == 0 {
							continue
						}

						tb.AddBlock([]*TxWithSigner{frCont.RevokeTx(sKey, tokenIDs[sIdx][0])})
						tokenIDs[sIdx] = tokenIDs[sIdx][1:]
						if len(tokenIDs[sIdx]) == 0 {
							frOwners.delete(sAddress)
						}
					case "transfer":
						if len(tokenIDs[sIdx]) == 0 {
							continue
						}
						// not same
						var toIdx int
						for {
							toIdx = rand.Intn(users)
							if toIdx != sIdx {
								break
							}
						}
						tAddress := userKeys[toIdx].PublicKey().Address()
						tokenID := tokenIDs[sIdx][0]
						tb.AddBlock([]*TxWithSigner{frCont.TransferFromTx(sKey, sAddress, tAddress, tokenID)})
						tokenIDs[sIdx] = tokenIDs[sIdx][1:]
						if len(tokenIDs[sIdx]) == 0 {
							frOwners.delete(sAddress)
						}
						tokenIDs[toIdx] = append(tokenIDs[toIdx], tokenID)
						frOwners.add(tAddress)
					default:
						panic("non-existent action : " + action)
					}

					// generator only 1
					generators, err := tb.Store.Generators()
					Expect(err).To(Succeed())
					Expect(len(generators)).To(Equal(1))

					i++
					// log.Printf("%v : %v : %v", i, sIdx, action)
				}
			})
		})
	})

})
