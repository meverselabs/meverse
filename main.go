package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

/*
TODO: node 구성
TODO: observer 구성
*/

func main() {
	ChainID := big.NewInt(1)
	Version := uint16(0x0001)
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

	adminKey, err := key.NewMemoryKeyFromBytes(ChainID, []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		panic(err)
	}

	userKey, err := key.NewMemoryKeyFromBytes(ChainID, []byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		panic(err)
	}

	obkeys := []key.Key{}
	ObserverKeys := []common.PublicKey{}
	for i := 0; i < 5; i++ {
		pk, err := key.NewMemoryKeyFromBytes(ChainID, []byte{1, 1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			panic(err)
		}
		obkeys = append(obkeys, pk)
		ObserverKeys = append(ObserverKeys, pk.PublicKey())
	}
	frkeys := []key.Key{}
	frKeyMap := map[common.Address]key.Key{}
	for i := 0; i < 10; i++ {
		pk, err := key.NewMemoryKeyFromBytes(ChainID, []byte{1, 1, 1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			panic(err)
		}
		frkeys = append(frkeys, pk)
		frKeyMap[pk.PublicKey().Address()] = pk
	}

	cdb, err := piledb.Open("_data/chain", hash.Hash256{}, 0, 0)
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := chain.NewStore("_data/context", cdb, ChainID, Version)
	if err != nil {
		panic(err)
	}
	if st.Height() > st.InitHeight() {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cn := chain.NewChain(ObserverKeys, st, "main")
	genesis := types.NewEmptyContext()
	var tokenAddress common.Address
	var formulatorAddress common.Address
	if true {
		if err := genesis.SetAdmin(adminKey.PublicKey().Address(), true); err != nil {
			panic(err)
		}
		for _, v := range frkeys {
			if err := genesis.SetGenerator(v.PublicKey().Address(), true); err != nil {
				panic(err)
			}
		}
	}
	if true {
		arg := &token.TokenContractConstruction{
			Name:   "Test",
			Symbol: "TEST",
			InitialSupplyMap: map[common.Address]*amount.Amount{
				userKey.PublicKey().Address(): amount.NewAmount(2000000000, 0),
			},
		}
		bs, _, err := bin.WriterToBytes(arg)
		if err != nil {
			panic(err)
		}
		v, err := genesis.DeployContract(adminKey.PublicKey().Address(), ClassMap["Token"], bs)
		if err != nil {
			panic(err)
		}
		cont := v.(*token.TokenContract)
		cc := genesis.ContractContext(cont, adminKey.PublicKey().Address())
		fmt.Println(userKey.PublicKey().Address(), cont.BalanceOf(cc, userKey.PublicKey().Address()))
		fmt.Println(adminKey.PublicKey().Address(), cont.BalanceOf(cc, adminKey.PublicKey().Address()))
		tokenAddress = cont.Address()
	}
	formulators := []common.Address{}
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
				MiningFeeAddress:      adminKey.PublicKey().Address(),
				MiningFee1000:         300,
			},
		}
		bs, _, err := bin.WriterToBytes(arg)
		if err != nil {
			panic(err)
		}
		v, err := genesis.DeployContract(adminKey.PublicKey().Address(), ClassMap["Formulator"], bs)
		if err != nil {
			panic(err)
		}
		cont := v.(*formulator.FormulatorContract)
		cc := genesis.ContractContext(cont, userKey.PublicKey().Address())
		if false {
			for i := 0; i < 16; i++ {
				addr, err := cont.CreateAlpha(cc)
				if err != nil {
					panic(err)
				}
				formulators = append(formulators, addr)
			}
		}
		formulatorAddress = cont.Address()
	}
	if true {
		v, err := genesis.Contract(tokenAddress)
		if err != nil {
			panic(err)
		}
		cont := v.(*token.TokenContract)
		cc := genesis.ContractContext(cont, adminKey.PublicKey().Address())
		if err := cont.SetMinter(cc, formulatorAddress, true); err != nil {
			panic(err)
		}
	}
	fmt.Println("TokenAddress", tokenAddress)
	fmt.Println("FormulatorAddress", formulatorAddress)
	if err := cn.Init(genesis.Top()); err != nil {
		panic(err)
	}

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if err := cn.ConnectBlock(b, nil); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}); err != nil {
		if errors.Cause(err) == chain.ErrStoreClosed {
			return
		}
		panic(err)
	}

	GenCountMap := map[common.Address]uint32{}
	for q := 0; q < 17280; q++ {
		TimeoutCount := uint32(0)
		ctx := cn.NewContext()
		Generator, err := st.TopGenerator(TimeoutCount)
		if err != nil {
			panic(err)
		}
		GenCnt := int(prefix.MaxBlocksPerGenerator)

		var LastHash hash.Hash256
		for p := 0; p < GenCnt; p++ {
			if p > 0 {
				ctx = ctx.NextContext(LastHash, uint64(time.Now().UnixNano()))
			}
			bc := chain.NewBlockCreator(cn, ctx, Generator, TimeoutCount, uint64(time.Now().UnixNano()), 0)
			if true {
				bs := bin.TypeWriteAll(adminKey.PublicKey().Address(), amount.NewAmount(2, 0))
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: uint64(time.Now().UnixNano()),
					To:        tokenAddress,
					Method:    "Transfer",
					Args:      bs,
				}
				sig, err := userKey.Sign(tx.HashSig())
				if err != nil {
					panic(err)
				}
				if err := bc.AddTx(tx, sig); err != nil {
					panic(err)
				}
			}
			if false {
				for i := 0; i < 4; i++ {

					bs := bin.TypeWriteAll(formulators[i*4 : (i+1)*4])
					tx := &types.Transaction{
						ChainID:   ctx.ChainID(),
						Timestamp: uint64(time.Now().UnixNano()),
						To:        formulatorAddress,
						Method:    "CreateSigma",
						Args:      bs,
					}
					sig, err := userKey.Sign(tx.HashSig())
					if err != nil {
						panic(err)
					}
					if err := bc.AddTx(tx, sig); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("SigmaCreated")
					}
				}
				for i := 0; i < 2; i++ {
					bs := bin.TypeWriteAll([]common.Address{formulators[i*8], formulators[i*8+4]})
					tx := &types.Transaction{
						ChainID:   ctx.ChainID(),
						Timestamp: uint64(time.Now().UnixNano()),
						To:        formulatorAddress,
						Method:    "CreateOmega",
						Args:      bs,
					}
					sig, err := userKey.Sign(tx.HashSig())
					if err != nil {
						panic(err)
					}
					if err := bc.AddTx(tx, sig); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("OmegaCreated")
					}
				}
			}
			if false && q == 0 && p == 0 {
				tx := &types.Transaction{
					ChainID:   ctx.ChainID(),
					Timestamp: uint64(time.Now().UnixNano()),
					To:        formulatorAddress,
					Method:    "Stake",
					Args:      bin.TypeWriteAll(frkeys[0].PublicKey().Address(), amount.NewAmount(200000, 0)),
				}
				sig, err := userKey.Sign(tx.HashSig())
				if err != nil {
					panic(err)
				}
				if err := bc.AddTx(tx, sig); err != nil {
					panic(err)
				}
			}
			b, err := bc.Finalize(0)
			if err != nil {
				panic(err)
			}
			HeaderHash := bin.MustWriterToHash(&b.Header)

			LastHash = HeaderHash

			pk := frKeyMap[Generator]
			GenSig, err := pk.Sign(HeaderHash)
			if err != nil {
				panic(err)
			}
			b.Body.BlockSignatures = append(b.Body.BlockSignatures, GenSig)

			blockSign := &types.BlockSign{
				HeaderHash:         HeaderHash,
				GeneratorSignature: GenSig,
			}

			BlockSignHash := bin.MustWriterToHash(blockSign)

			idxes := rand.Perm(len(obkeys))
			for i := 0; i < len(obkeys)/2+1; i++ {
				pk := obkeys[idxes[i]]
				ObSig, err := pk.Sign(BlockSignHash)
				if err != nil {
					panic(err)
				}
				b.Body.BlockSignatures = append(b.Body.BlockSignatures, ObSig)
			}

			if err := cn.ConnectBlock(b, nil); err != nil {
				panic(err)
			}
			if false {
				ctx := cn.NewContext()
				v, err := ctx.Contract(tokenAddress)
				if err != nil {
					panic(err)
				}
				cont := v.(*token.TokenContract)
				cc := ctx.ContractLoader(cont.Address())
				fmt.Println(userKey.PublicKey().Address(), cont.BalanceOf(cc, userKey.PublicKey().Address()).String())
				fmt.Println(adminKey.PublicKey().Address(), cont.BalanceOf(cc, adminKey.PublicKey().Address()).String())
			}
			//fmt.Println(b.Header.Height, b.Header.Generator)

			if b.Header.Height%prefix.RewardIntervalBlocks == 0 {
				ctx := cn.NewContext()
				v, err := ctx.Contract(tokenAddress)
				if err != nil {
					panic(err)
				}
				cont := v.(*token.TokenContract)
				cc := ctx.ContractLoader(cont.Address())
				fmt.Println(userKey.PublicKey().Address(), cont.BalanceOf(cc, userKey.PublicKey().Address()).String())
				fmt.Println(adminKey.PublicKey().Address(), cont.BalanceOf(cc, adminKey.PublicKey().Address()).String())

				CountMap := map[common.Address]uint32{}
				for addr, GenCount := range GenCountMap {
					CountMap[addr] = GenCount
				}
				CountSum := uint32(0)
				for _, GenCount := range GenCountMap {
					CountSum += GenCount
				}
				CountPerFormulator := CountSum / uint32(len(GenCountMap))
				if true {
					v, err := ctx.Contract(formulatorAddress)
					if err != nil {
						panic(err)
					}
					cont := v.(*formulator.FormulatorContract)
					cc := ctx.ContractLoader(cont.Address())
					FormulatorMap, err := cont.FormulatorMap(cc)
					if err != nil {
						panic(err)
					}
					for _, fr := range FormulatorMap {
						CountMap[fr.Owner] += CountPerFormulator
					}
				}

				for addr, Count := range CountMap {
					if ctx.IsGenerator(addr) {
						fmt.Println("HYPER", addr, cont.BalanceOf(cc, addr).String(), Count)
					} else {
						fmt.Println("FORMULATOR", addr, cont.BalanceOf(cc, addr).String(), Count)
					}
				}
				fmt.Println("Generated")

				GenCountMap = map[common.Address]uint32{}
			}
			GenCountMap[b.Header.Generator]++
		}
	}
}
