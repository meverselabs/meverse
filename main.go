package main // import "github.com/fletaio/fleta"

import (
	"bytes"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
)

func main() {
	obround := pof.NewVoteRound(1, 10)
	voteTicker := time.NewTimer(100 * time.Millisecond)
	go func() {
		for {
			select {
			case <-voteTicker.C:
				if obround.RoundState == pof.RoundVoteState {
					//send round vote
				}
			}
		}
	}()

	func() error {
		msgCh := make(chan interface{})
		for m := range msgCh {
			switch msg := m.(type) {
			case *pof.RoundVoteMessage:
				log.Println(msg)
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState != pof.RoundVoteState {
					//reply round vote if recv not reply -> x
					return nil
				}
				//[same state]
				//update round vote map
				//[if get round vote more than 4]
				//send vote ack with min vote
				//change to RoundVoteAckState
			case *pof.RoundVoteAckMessage:
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < pof.RoundVoteAckState {
					//save to wait
					return nil
				} else if obround.RoundState != pof.RoundVoteAckState {
					//reply round vote ack if recv not reply -> x
				}
				//[same state]
				//update round vote ack map
				//[if get same min more than 3]
				//send round setup using proof(3 min vote pack)
				//send gen req to formulator if min vote voter
				//change to BlockWaitState
			case *pof.RoundSetupMessage:
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < pof.BlockWaitState {
					//[if valid]
					//change to BlockVoteState
					//send round setup
				} else if obround.RoundState == pof.BlockVoteState {
					//reply block gen if recv not reply -> x
				}
			case *pof.BlockGenMessage:
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < pof.BlockWaitState {
					//save to wait
				} else if obround.RoundState == pof.BlockVoteState {
					//reply block vote if recv not reply -> x
				}
				//[same state]
				//broadcast block gen if min.PubHash == me
				//[if valid block]
				//send block vote
				//change to BlockVoteState
			case *pof.BlockVoteMessage:
				//diff round -> send status if recv old -> x
				//[same round]
				if obround.RoundState < pof.BlockVoteState {
					//save to wait
				}
				//[same state]
				//update block vote map
				//[if get more than 3]
				//connect block
				//send block onsign to formulator
				//decrease RemainBlocks
				//[if RemainBlocks > 0]
				//change to BlockWaitState
				//[if RemainBlocks <= 0]
				//clear round
			}
		}
		return nil
	}()
	return

	if err := test(); err != nil {
		panic(err)
	}
}

func test() error {
	st, err := chain.NewStore("./_data", "FLEAT Mainnet", 0x0001, true)
	if err != nil {
		return err
	}

	obstrs := []string{
		"cd7cca6359869f4f58bb31aa11c2c4825d4621406f7b514058bc4dbe788c29be",
		"d8744df1e76a7b76f276656c48b68f1d40804f86518524d664b676674fccdd8a",
		"387b430fab25c03313a7e987385c81f4b027199304e2381561c9707847ec932d",
		"a99fa08114f41eb7e0a261cf11efdc60887c1d113ea6602aaf19eca5c3f5c720",
		"a9878ff3837700079fbf187c86ad22f1c123543a96cd11c53b70fedc3813c27b",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	ObserverKeys := make([]common.PublicHash, 0, len(obstrs))
	for _, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			pubhash := common.NewPublicHash(Key.PublicKey())
			ObserverKeys = append(ObserverKeys, pubhash)
		}
	}

	frstrs := []string{
		"67066852dd6586fa8b473452a66c43f3ce17bd4ec409f1fff036a617bb38f063",
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
	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := &DApp{
		frkey:        frkeys[0],
		adminAddress: common.NewAddress(0, 1, 0),
	}
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(vault.NewVault(1))
	cn.MustAddProcess(formulator.NewFormulator(2, app.adminAddress))
	if err := cn.Init(); err != nil {
		return err
	}

	TimeoutCount := uint32(0)
	Formulator := common.NewAddress(0, 2, 0)
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(TimeoutCount); err != nil {
		return err
	}
	bc := chain.NewBlockCreator(cn, Formulator, buffer.Bytes())
	if err := bc.Init(); err != nil {
		return err
	}
	txs := []types.Transaction{}
	sigs := [][]common.Signature{}
	Seq := cn.Seq(common.NewAddress(0, 1, 0))
	for i := 0; i < 1; i++ {
		Seq++
		tx := &vault.Transfer{
			Timestamp_: 0,
			Seq_:       Seq,
			From:       common.NewAddress(0, 1, 0),
			To:         common.NewAddress(0, 2, 0),
			Amount:     amount.NewCoinAmount(1, 0),
		}
		sig, _ := frkeys[0].Sign(chain.HashTransaction(tx))
		sigs = append(sigs, []common.Signature{sig})
		txs = append(txs, tx)
	}

	if true {
		begin := time.Now().UnixNano()
		for i := 0; i < len(txs); i++ {
			if err := bc.AddTx(txs[i], sigs[i]); err != nil {
				return err
			}
		}
		end := time.Now().UnixNano()
		log.Println((end - begin) / int64(time.Millisecond))
	}
	b, err := bc.Finalize()
	if err != nil {
		return err
	}

	bh := encoding.Hash(b.Header)
	sig0, _ := frkeys[0].Sign(bh)
	Signatures := []common.Signature{
		sig0,
	}
	b.Signatures = Signatures

	// TODO

	// Header
	// Context
	// Signature[0]

	bs := &types.BlockSign{
		HeaderHash:         bh,
		GeneratorSignature: sig0,
	}
	bsh := encoding.Hash(bs)
	sig1, _ := obkeys[0].Sign(bsh)
	sig2, _ := obkeys[1].Sign(bsh)
	sig3, _ := obkeys[2].Sign(bsh)

	b.Signatures = append(b.Signatures, sig1)
	b.Signatures = append(b.Signatures, sig2)
	b.Signatures = append(b.Signatures, sig3)

	if err := cn.ConnectBlock(b); err != nil {
		return err
	}

	if true {
		b, err := cn.Provider().Block(cn.Provider().Height())
		if err != nil {
			return err
		}
		log.Println(cn.Provider().Height(), len(b.Transactions), encoding.Hash(b.Header))
	}
	return nil
}

// DApp is app
type DApp struct {
	sync.Mutex
	*types.ProcessBase
	pm           types.ProcessManager
	cn           types.Provider
	frkey        key.Key
	adminAddress common.Address
}

// ID must returns 255
func (app *DApp) ID() uint8 {
	return 255
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
	return nil
}
