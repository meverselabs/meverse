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
				ChainID:       ctx.ChainID(),
				Version:       ctx.Version(),
				Height:        ctx.TargetHeight(),
				PrevHash:      ctx.LastHash(),
				Generator:     Generator,
				ConsensusData: ConsensusData,
			},
			TransactionTypes:      []uint16{},
			Transactions:          []types.Transaction{},
			TransactionSignatures: [][]common.Signature{},
			Signatures:            []common.Signature{},
			TransactionResults:    []uint8{},
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
func (bc *BlockCreator) UnsafeAddTx(Generator common.Address, t uint16, TxHsah hash.Hash256, tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	pid := uint8(t >> 8)
	p, err := bc.cn.Process(pid)
	if err != nil {
		return err
	}
	ctw := types.NewContextWrapper(pid, bc.ctx)

	Result := uint8(0)

	sn := ctw.Snapshot()
	if err := tx.Validate(p, ctw, signers); err != nil {
		ctw.Revert(sn)
		return err
	}
	if at, is := tx.(AccountTransaction); is {
		if at.Seq() != ctw.Seq(at.From())+1 {
			ctw.Revert(sn)
			return err
		}
		ctw.AddSeq(at.From())
		if err := tx.Execute(p, ctw, uint16(len(bc.b.Transactions))); err != nil {
			Result = 0
		} else {
			Result = 1
		}
	} else {
		if err := tx.Execute(p, ctw, uint16(len(bc.b.Transactions))); err != nil {
			ctw.Revert(sn)
			return err
		}
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
	bc.b.TransactionResults = append(bc.b.TransactionResults, Result)
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

	bc.b.Header.Timestamp = Timestamp
	bc.b.Header.ContextHash = bc.ctx.Hash()

	return bc.b, nil
}
