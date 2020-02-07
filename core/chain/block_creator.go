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
func NewBlockCreator(cn *Chain, ctx *types.Context, Generator common.Address, ConsensusData []byte, Timestamp uint64) *BlockCreator {
	bc := &BlockCreator{
		cn:       cn,
		ctx:      ctx,
		txHashes: []hash.Hash256{ctx.LastHash()},
		b: &types.Block{
			Header: types.Header{
				ChainID:       ctx.ChainID(),
				Version:       ctx.Version(),
				Height:        ctx.TargetHeight(),
				PrevHash:      ctx.LastHash(),
				Generator:     Generator,
				ConsensusData: ConsensusData,
				Timestamp:     Timestamp,
			},
			TransactionTypes:      []uint16{},
			Transactions:          []types.Transaction{},
			TransactionSignatures: [][]common.Signature{},
			Signatures:            []common.Signature{},
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
		} else if bc.ctx.StackSize() > 1 {
			return ErrDirtyContext
		}
	}
	if err := bc.cn.app.BeforeExecuteTransactions(types.NewContextWrapper(255, bc.ctx)); err != nil {
		return err
	} else if bc.ctx.StackSize() > 1 {
		return ErrDirtyContext
	}
	return nil
}

// AddTx validates, executes and adds transactions
func (bc *BlockCreator) AddTx(Generator common.Address, tx types.Transaction, sigs []common.Signature) error {
	t, err := bc.txFactory.TypeOf(tx)
	if err != nil {
		return err
	}
	TxHash := HashTransactionByType(bc.cn.Provider().ChainID(), t, tx)
	signers := []common.PublicHash{}
	for _, sig := range sigs {
		if pubkey, err := common.RecoverPubkey(TxHash, sig); err != nil {
			return err
		} else {
			signers = append(signers, common.NewPublicHash(pubkey))
		}
	}
	return bc.UnsafeAddTx(Generator, t, TxHash, tx, sigs, signers)
}

// UnsafeAddTx adds transactions without signer validation if signers is not empty
func (bc *BlockCreator) UnsafeAddTx(Generator common.Address, t uint16, TxHash hash.Hash256, tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	pid := uint8(t >> 8)
	p, err := bc.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, bc.ctx)

	currentSlot := types.ToTimeSlot(bc.b.Header.Timestamp)
	slot := types.ToTimeSlot(tx.Timestamp())
	if slot < currentSlot-1 {
		return types.ErrInvalidTransactionTimeSlot
	} else if slot > currentSlot {
		return types.ErrInvalidTransactionTimeSlot
	}

	sn := ctw.Snapshot()
	if err := bc.ctx.UseTimeSlot(slot, string(TxHash[:])); err != nil {
		return err
	}
	if err := tx.Validate(p, ctw, signers); err != nil {
		ctw.Revert(sn)
		return err
	}
	if err := tx.Execute(p, ctw, uint16(len(bc.b.Transactions))); err != nil {
		ctw.Revert(sn)
		return err
	}
	if Has, err := ctw.HasAccount(Generator); err != nil {
		ctw.Revert(sn)
		if err == types.ErrDeletedAccount {
			return ErrCannotDeleteGeneratorAccount
		} else {
			return err
		}
	} else if !Has {
		ctw.Revert(sn)
		return ErrCannotDeleteGeneratorAccount
	}
	ctw.Commit(sn)

	bc.b.TransactionTypes = append(bc.b.TransactionTypes, t)
	bc.b.Transactions = append(bc.b.Transactions, tx)
	bc.b.TransactionSignatures = append(bc.b.TransactionSignatures, sigs)
	bc.txHashes = append(bc.txHashes, TxHash)
	return nil
}

// Finalize generates block that has transactions adds by AddTx
func (bc *BlockCreator) Finalize() (*types.Block, error) {
	IDMap := map[int]uint8{}
	for id, idx := range bc.cn.processIndexMap {
		IDMap[idx] = id
	}

	LevelRootHash, err := BuildLevelRoot(bc.txHashes)
	if err != nil {
		return nil, err
	}
	bc.b.Header.LevelRootHash = LevelRootHash

	if bc.ctx.StackSize() > 1 {
		return nil, ErrDirtyContext
	}

	// AfterExecuteTransactions
	for i, p := range bc.cn.processes {
		if err := p.AfterExecuteTransactions(bc.b, types.NewContextWrapper(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		} else if bc.ctx.StackSize() > 1 {
			return nil, ErrDirtyContext
		}
	}
	if err := bc.cn.app.AfterExecuteTransactions(bc.b, types.NewContextWrapper(255, bc.ctx)); err != nil {
		return nil, err
	} else if bc.ctx.StackSize() > 1 {
		return nil, ErrDirtyContext
	}

	bc.b.Header.ContextHash = bc.ctx.Hash()

	return bc.b, nil
}
