package chain

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// BlockCreator helps to create block
type BlockCreator struct {
	cn       *Chain
	ctx      *types.Context
	txHashes []hash.Hash256
	b        *types.Block
}

// NewBlockCreator returns a BlockCreator
func NewBlockCreator(cn *Chain, ConsensusData []byte) *BlockCreator {
	ctx := types.NewContext(cn.store)
	bc := &BlockCreator{
		cn:       cn,
		ctx:      ctx,
		txHashes: []hash.Hash256{ctx.LastHash()},
		b: &types.Block{
			Header: types.Header{
				Version:       ctx.Version(),
				Height:        ctx.TargetHeight(),
				PrevHash:      ctx.LastHash(),
				ConsensusData: ConsensusData,
			},
			Transactions:         []types.Transaction{},
			TranactionSignatures: [][]common.Signature{},
			Signatures:           []common.Signature{},
		},
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
		if err := p.BeforeExecuteTransactions(NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return err
		}
	}
	if err := bc.cn.app.BeforeExecuteTransactions(NewContextProcess(255, bc.ctx)); err != nil {
		return err
	}
	if err := bc.cn.consensus.BeforeExecuteTransactions(NewContextProcess(0, bc.ctx)); err != nil {
		return err
	}
	return nil
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

	// AfterExecuteTransactions
	for i, p := range bc.cn.processes {
		if err := p.AfterExecuteTransactions(bc.b, NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		}
	}
	if err := bc.cn.app.AfterExecuteTransactions(bc.b, NewContextProcess(255, bc.ctx)); err != nil {
		return nil, err
	}
	if err := bc.cn.consensus.AfterExecuteTransactions(bc.b, NewContextProcess(0, bc.ctx)); err != nil {
		return nil, err
	}

	// ProcessReward
	for i, p := range bc.cn.processes {
		if err := p.ProcessReward(bc.b, NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		}
	}
	if err := bc.cn.app.ProcessReward(bc.b, NewContextProcess(255, bc.ctx)); err != nil {
		return nil, err
	}
	if err := bc.cn.consensus.ProcessReward(bc.b, NewContextProcess(0, bc.ctx)); err != nil {
		return nil, err
	}

	now := uint64(time.Now().UnixNano())
	if now <= bc.ctx.LastTimestamp() {
		now = bc.ctx.LastTimestamp() + 1
	}
	bc.b.Header.Timestamp = now
	bc.b.Header.ContextHash = bc.ctx.Hash()
	return bc.b, nil
}
