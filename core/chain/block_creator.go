package chain

import (
	"time"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/factory"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// TODO : BlockCreator에 txpool 연동해서 이미 서명 검사한 것은 signers 목록을 받아올 수 있도록 처리

// BlockCreator helps to create block
type BlockCreator struct {
	cn        *Chain
	ctx       *types.Context
	txHashes  []hash.Hash256
	b         *types.Block
	txFactory *factory.Factory
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
		if err := p.BeforeExecuteTransactions(types.NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return err
		}
	}
	if err := bc.cn.app.BeforeExecuteTransactions(types.NewContextProcess(255, bc.ctx)); err != nil {
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

	th := HashTransaction(t, tx)

	signers := []common.PublicHash{}
	for _, sig := range sigs {
		if pubkey, err := common.RecoverPubkey(th, sig); err != nil {
			return err
		} else {
			signers = append(signers, common.NewPublicHash(pubkey))
		}
	}
	ctp := types.NewContextProcess(uint8(t>>8), bc.ctx)
	if err := tx.Validate(ctp, signers); err != nil {
		return err
	}
	if err := tx.Execute(ctp, uint16(len(bc.b.Transactions))); err != nil {
		return err
	}
	bc.b.TransactionTypes = append(bc.b.TransactionTypes, t)
	bc.b.Transactions = append(bc.b.Transactions, tx)
	bc.b.TranactionSignatures = append(bc.b.TranactionSignatures, sigs)
	bc.txHashes = append(bc.txHashes, th)
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
		if err := p.AfterExecuteTransactions(bc.b, types.NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		}
	}
	if err := bc.cn.app.AfterExecuteTransactions(bc.b, types.NewContextProcess(255, bc.ctx)); err != nil {
		return nil, err
	}

	// ProcessReward
	for i, p := range bc.cn.processes {
		if err := p.ProcessReward(bc.b, types.NewContextProcess(IDMap[i], bc.ctx)); err != nil {
			return nil, err
		}
	}
	if err := bc.cn.app.ProcessReward(bc.b, types.NewContextProcess(255, bc.ctx)); err != nil {
		return nil, err
	}
	if err := bc.cn.consensus.ProcessReward(bc.b, types.NewContextProcess(0, bc.ctx)); err != nil {
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
