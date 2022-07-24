package txsearch

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
)

var TAG = "TXSEARCH"

func plog(str ...interface{}) {
	ss := []interface{}{TAG}
	fmt.Println(append(ss, str...)...)
}

func NewTxSearch(Path string, api *apiserver.APIServer, st *chain.Store, cn *chain.Chain, initHeight uint32) *TxSearch {
	db, err := leveldb.OpenFile(Path, nil)
	if err != nil {
		panic(err)
	}

	t := &TxSearch{
		db:  db,
		st:  st,
		cn:  cn,
		api: api,
	}

	if err := t.initFromStore(st, initHeight); err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	t.SetupApi()

	return t
}

// Name returns the name of the service
func (t *TxSearch) Name() string {
	return "fleta.txsearch"
}

// OnLoadChain called when the chain loaded
func (t *TxSearch) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (t *TxSearch) OnBlockConnected(b *types.Block, loader types.Loader) {
	t.ReadBlock(b)
}

// OnLoadChain called when the chain loaded
func (t *TxSearch) OnTransactionInPoolExpired(txs []*types.Transaction) {
}

// OnTransactionFail called when the tx fail
func (t *TxSearch) OnTransactionFail(height uint32, txs []*types.Transaction, err []error) {
	for i, tx := range txs {
		t.db.Put(toTxFailKey(tx.Hash(height)), bin.TypeWriteAll(height, err[i].Error()), nil)
	}
}

func (t *TxSearch) initFromStore(st *chain.Store, initHeight uint32) error {
	if !t.isInitDB() {
		t.db.Put([]byte{tagInitHeight}, bin.Uint32Bytes(st.InitHeight()), nil)
		t.setHeight(1)
		t.setInitDB()
	}
	if t.Height() < initHeight {
		t.setHeight(initHeight)
	}
	plog(t.Height(), st.Height())
	for t.Height() < st.Height() {
		b, err := st.Block(t.Height() + 1)
		if err != nil {
			return err
		}
		err = t.ReadBlock(b)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TxSearch) Height() uint32 {
	bs, err := t.db.Get([]byte{tagHeight}, nil)
	if err != nil {
		plog("Cannot getHeight")
		return 0
	}
	return bin.Uint32(bs)
}

func (t *TxSearch) setHeight(h uint32) error {
	err := t.db.Put([]byte{tagHeight}, bin.Uint32Bytes(h), nil)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return &ErrCannotSetHeight{err, h}
	}
	return nil
}

func (t *TxSearch) isInitDB() bool {
	bs, err := t.db.Get([]byte{tagInitDB}, nil)
	if err != nil {
		return false
	}
	return len(bs) > 0
}

func (t *TxSearch) setInitDB() error {
	if err := t.db.Put([]byte{tagInitDB}, []byte{1}, nil); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
