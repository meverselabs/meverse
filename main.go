package main // import "github.com/fletaio/fleta"

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"

	"github.com/fletaio/fleta/pof"

	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/encoding"
)

func main() {
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
	for i := 0; i < 20000; i++ {
		tx := &Transaction{
			Timestamp_: 0,
			KeyHash:    common.PublicHash{},
			Amount:     amount.NewCoinAmount(uint64(i+1), 0),
		}
		sig, _ := frkeys[0].Sign(encoding.Hash(tx))
		sigs = append(sigs, []common.Signature{sig})
		txs = append(txs, tx)
	}

	if true {
		begin := time.Now().UnixNano()
		for i := 0; i < 10000; i++ {
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
	reg.RegisterTransaction(1, &Transaction{})
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
			SigmaEfficiency1000:   1000,
			OmegaEfficiency1000:   1000,
			HyperEfficiency1000:   1000,
			StakingEfficiency1000: 1000,
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
	if true {
		acc := &formulator.FormulatorAccount{
			Address_:       common.NewAddress(0, 1, 0),
			Name_:          "admin",
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

// Transaction is an interface that defines common transaction functions
type Transaction struct {
	Timestamp_ uint64
	KeyHash    common.PublicHash
	Amount     *amount.Amount
}

func (tx *Transaction) Timestamp() uint64 {
	return tx.Timestamp_
}

func (tx *Transaction) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

func (tx *Transaction) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	return nil
}

func (tx *Transaction) Execute(p types.Process, ctx *types.ContextWrapper, index uint16) error {
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
