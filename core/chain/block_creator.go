package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/factory"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// BlockCreator helps to create block
type BlockCreator struct {
	cn        *Chain
	ctx       *types.Context
	txHashes  []hash.Hash256
	b         *types.Block
	txFactory *factory.Factory
}

// NewBlockCreator returns a BlockCreator
func NewBlockCreator(cn *Chain, ctx *types.Context, Generator common.Address, ConsensusData []byte) *BlockCreator {
	bc := &BlockCreator{
		cn:       cn,
		ctx:      ctx,
		txHashes: []hash.Hash256{ctx.LastHash()},
		b: &types.Block{
			Header: types.Header{
				Version:       ctx.Version(),
				Height:        ctx.TargetHeight(),
				PrevHash:      ctx.LastHash(),
				Generator:     Generator,
				ConsensusData: ConsensusData,
			},
			Transactions:         []types.Transaction{},
			TranactionSignatures: [][]common.Signature{},
			Signatures:           []common.Signature{},
		},
		txFactory: encoding.Factory("transaction"),
	}
	return bc
}

// Init initializes the block creator
func (bc *BlockCreator) Init() error {
	IDMap := map[int]uint8{}
	for id, idx := range bc.cn.processIndexMap {
		IDMap[idx] = id
	}

	// BeforeExecuteTransactions
	for i, p := range bc.cn.processes {
		if err := p.BeforeExecuteTransactions(types.NewContextWrapper(IDMap[i], bc.ctx)); err != nil {
			return err
		}
	}
	if err := bc.cn.app.BeforeExecuteTransactions(types.NewContextWrapper(255, bc.ctx)); err != nil {
		return err
	}
	return nil
}

// AddTx validates, executes and adds transactions
func (bc *BlockCreator) AddTx(tx types.Transaction, sigs []common.Signature) error {
	t, err := bc.txFactory.TypeOf(tx)
	if err != nil {
		return err
	}
	TxHash := HashTransactionByType(t, tx)
	signers := []common.PublicHash{}
	for _, sig := range sigs {
		if pubkey, err := common.RecoverPubkey(TxHash, sig); err != nil {
			return err
		} else {
			signers = append(signers, common.NewPublicHash(pubkey))
		}
	}
	return bc.UnsafeAddTx(t, TxHash, tx, sigs, signers)
}

// UnsafeAddTx adds transactions without signer validation if signers is not empty
func (bc *BlockCreator) UnsafeAddTx(t uint16, TxHsah hash.Hash256, tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	ctw := types.NewContextWrapper(uint8(t>>8), bc.ctx)
	pid := uint8(t >> 8)
	p, err := bc.cn.Process(pid)
	if err != nil {
		return err
	}
	sn := ctw.Snapshot()
	if err := tx.Validate(p, ctw, signers); err != nil {
		return err
	}
	if err := tx.Execute(p, ctw, uint16(len(bc.b.Transactions))); err != nil {
		ctw.Revert(sn)
		return err
	}
	ctw.Commit(sn)

	bc.b.TransactionTypes = append(bc.b.TransactionTypes, t)
	bc.b.Transactions = append(bc.b.Transactions, tx)
	bc.b.TranactionSignatures = append(bc.b.TranactionSignatures, sigs)
	bc.txHashes = append(bc.txHashes, TxHsah)
	return nil
}

// Finalize generates block that has transactions adds by AddTx
func (bc *BlockCreator) Finalize(Timestamp uint64) (*types.Block, error) {
	IDMap := map[int]uint8{}
	for id, idx := range bc.cn.processIndexMap {
		IDMap[idx] = id
	}

	LevelRootHash, err := BuildLevelRoot(bc.txHashes)
	if err != nil {
		return nil, err
	}
	bc.b.Header.LevelRootHash = LevelRootHash

	// AfterExecuteTransactions
	for i, p := range bc.cn.processes {
		if err := p.AfterExecuteTransactions(bc.b, types.NewContextWrapper(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		}
	}
	if err := bc.cn.app.AfterExecuteTransactions(bc.b, types.NewContextWrapper(255, bc.ctx)); err != nil {
		return nil, err
	}

	if bc.ctx.StackSize() > 1 {
		return nil, ErrDirtyContext
	}

	bc.b.Header.Timestamp = Timestamp
	bc.b.Header.ContextHash = bc.ctx.Hash()
	return bc.b, nil
}
