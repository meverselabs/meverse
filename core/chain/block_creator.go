package chain

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// BlockCreator helps to create block
type BlockCreator struct {
	ctx      *types.Context
	txHashes []hash.Hash256
	b        *types.Block
}

// NewBlockCreator returns a BlockCreator
func NewBlockCreator(cn *Chain, ConsensusData []byte) *BlockCreator {
	ctx := types.NewContext(cn.store)
	bc := &BlockCreator{
		ctx:      ctx,
		txHashes: []hash.Hash256{ctx.LastHash()},
		b: &types.Block{
			Header: types.Header{
				Version:       ctx.Version(),
				Height:        ctx.TargetHeight(),
				PrevHash:      ctx.LastHash(),
				ContextHash:   ctx.Hash(),
				ConsensusData: ConsensusData,
			},
			Transactions:         []types.Transaction{},
			TranactionSignatures: [][]common.Signature{},
			Signatures:           []common.Signature{},
		},
	}
	return bc
}

// AddTx validates, executes and adds transactions
func (bc *BlockCreator) AddTx(tx types.Transaction, sigs []common.Signature, signers []common.PublicHash) error {
	if err := tx.Validate(bc.ctx, signers); err != nil {
		return err
	}
	if err := tx.Execute(bc.ctx, uint16(len(bc.b.Transactions))); err != nil {
		return err
	}
	bc.b.Transactions = append(bc.b.Transactions, tx)
	bc.b.TranactionSignatures = append(bc.b.TranactionSignatures, sigs)
	return nil
}

// Finalize generates block that has transactions adds by AddTx
func (bc *BlockCreator) Finalize(Signatures []common.Signature) (*types.Block, error) {
	LevelRootHash, err := BuildLevelRoot(bc.txHashes)
	if err != nil {
		return nil, err
	}
	now := uint64(time.Now().UnixNano())
	if now <= bc.ctx.LastTimestamp() {
		now = bc.ctx.LastTimestamp() + 1
	}
	bc.b.Header.Timestamp = now
	bc.b.Header.LevelRootHash = LevelRootHash
	bc.b.Signatures = Signatures
	return bc.b, nil
}
