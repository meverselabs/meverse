package main // import "github.com/fletaio/fleta"

import (
	"bytes"
	"log"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
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

	ObserverKeyMap := types.NewPublicHashBoolMap()
	ObserverKeyMap.Put(common.NewPublicHash(common.PublicKey{0, 1, 2, 3, 4}), true)
	ObserverKeyMap.Put(common.NewPublicHash(common.PublicKey{1, 2, 3, 4, 5}), true)
	ObserverKeyMap.Put(common.NewPublicHash(common.PublicKey{2, 3, 4, 5, 6}), true)
	ObserverKeyMap.Put(common.NewPublicHash(common.PublicKey{3, 4, 5, 6, 7}), true)
	ObserverKeyMap.Put(common.NewPublicHash(common.PublicKey{4, 5, 6, 7, 8}), true)
	MaxBlocksPerFormulator := uint32(10)
	cs := pof.NewConsensus(ObserverKeyMap, MaxBlocksPerFormulator)
	app := &DApp{}
	cn := chain.NewChain(cs, app, st)
	if err := cn.Init(); err != nil {
		return err
	}

	TimeoutCount := uint32(0)
	Formulator := common.Address{0, 1, 2}
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(TimeoutCount); err != nil {
		return err
	}
	if err := enc.Encode(Formulator); err != nil {
		return err
	}
	bc := chain.NewBlockCreator(cn, buffer.Bytes())
	b, err := bc.Finalize()
	if err != nil {
		return err
	}
	Signatures := []common.Signature{
		common.Signature{0},
		common.Signature{1},
		common.Signature{2},
		common.Signature{3},
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
	cn *chain.Chain
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
		Address_:        common.NewAddress(common.NewCoordinate(ctp.TargetHeight(), 0), 0),
		Name_:           "fleta.001",
		FormulationType: pof.AlphaFormulatorType,
		KeyHash:         common.PublicHash{0, 1, 2},
		Amount:          amount.NewCoinAmount(0, 0),
	}
	if err := ctp.CreateAccount(acc); err != nil {
		return err
	}
	return nil
}
