package main // import "github.com/fletaio/fleta"

import (
	"bytes"
	"encoding/hex"
	"log"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/types"

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

	ObserverKeyMap := types.NewPublicHashBoolMap()
	for _, pubhash := range ObserverKeys {
		ObserverKeyMap.Put(pubhash, true)
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
	cs := pof.NewConsensus(ObserverKeyMap, MaxBlocksPerFormulator)
	app := &DApp{
		frkey: frkeys[0],
	}
	cn := chain.NewChain(cs, app, st)
	if err := cn.Init(); err != nil {
		return err
	}

	TimeoutCount := uint32(0)
	Formulator := common.NewAddress(common.NewCoordinate(0, 0), 0)
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(TimeoutCount); err != nil {
		return err
	}
	if err := enc.Encode(Formulator); err != nil {
		return err
	}
	bc := chain.NewBlockCreator(cn, buffer.Bytes())
	if err := bc.Init(); err != nil {
		return err
	}
	b, err := bc.Finalize()
	if err != nil {
		return err
	}
	bh := encoding.Hash(b.Header)
	sig0, _ := frkeys[0].Sign(bh)
	bs := &types.BlockSign{
		HeaderHash:         bh,
		GeneratorSignature: sig0,
	}
	bsh := encoding.Hash(bs)
	sig1, _ := obkeys[0].Sign(bsh)
	sig2, _ := obkeys[1].Sign(bsh)
	sig3, _ := obkeys[2].Sign(bsh)

	Signatures := []common.Signature{
		sig0,
		sig1,
		sig2,
		sig3,
	}
	b.Signatures = Signatures
	if err := cn.ConnectBlock(b); err != nil {
		return err
	}
	if true {
		b, err := cn.Provider().Block(cn.Provider().Height())
		if err != nil {
			return err
		}
		log.Println(cn.Provider().Height(), b, encoding.Hash(b.Header))
	}
	return nil
}

// DApp is app
type DApp struct {
	sync.Mutex
	*chain.ProcessBase
	cn    *chain.Chain
	frkey key.Key
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
func (app *DApp) Init(reg *chain.Register, cn *chain.Chain) error {
	app.cn = cn
	return nil
}

// InitGenesis initializes genesis data
func (app *DApp) InitGenesis(ctp *chain.ContextProcess) error {
	app.Lock()
	defer app.Unlock()
	acc := &pof.FormulationAccount{
		Address_:        common.NewAddress(common.NewCoordinate(0, 0), 0),
		Name_:           "fleta.001",
		FormulationType: pof.AlphaFormulatorType,
		KeyHash:         common.NewPublicHash(app.frkey.PublicKey()),
		Amount:          amount.NewCoinAmount(0, 0),
	}
	if err := ctp.CreateAccount(acc); err != nil {
		return err
	}
	return nil
}
