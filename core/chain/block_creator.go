package chain

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/bin"
	"github.com/fletaio/fleta_v2/common/hash"
	"github.com/fletaio/fleta_v2/core/prefix"
	"github.com/fletaio/fleta_v2/core/types"
)

// BlockCreator helps to create block
type BlockCreator struct {
	cn       *Chain
	ctx      *types.Context
	txHashes []hash.Hash256
	b        *types.Block
}

// NewBlockCreator returns a BlockCreator
func NewBlockCreator(cn *Chain, ctx *types.Context, Generator common.Address, TimeoutCount uint32, Timestamp uint64, gasLv uint16) *BlockCreator {
	bc := &BlockCreator{
		cn:       cn,
		ctx:      ctx,
		txHashes: []hash.Hash256{ctx.PrevHash()},
		b: &types.Block{
			Header: types.Header{
				ChainID:      ctx.ChainID(),
				Version:      ctx.Version(),
				Height:       ctx.TargetHeight(),
				PrevHash:     ctx.PrevHash(),
				TimeoutCount: TimeoutCount,
				Timestamp:    Timestamp,
				Generator:    Generator,
				Gas:          gasLv,
			},
			Body: types.Body{
				Transactions:          []*types.Transaction{},
				TransactionSignatures: []common.Signature{},
				BlockSignatures:       []common.Signature{},
			},
		},
	}
	return bc
}

// AddTx validates, executes and adds transactions
func (bc *BlockCreator) AddTx(tx *types.Transaction, sig common.Signature) error {
	TxHash := tx.Hash()
	pubkey, err := common.RecoverPubkey(tx.ChainID, TxHash, sig)
	if err != nil {
		return err
	}
	tx.From = pubkey.Address()
	return bc.UnsafeAddTx(TxHash, tx, sig, tx.From)
}

// UnsafeAddTx adds transactions without signer validation if signers is not empty
func (bc *BlockCreator) UnsafeAddTx(TxHash hash.Hash256, tx *types.Transaction, sig common.Signature, signer common.Address) (err error) {
	currentSlot := types.ToTimeSlot(bc.b.Header.Timestamp)
	slot := types.ToTimeSlot(tx.Timestamp)
	if slot < currentSlot-1 {
		return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
	} else if slot > currentSlot {
		return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
	}

	sn := bc.ctx.Snapshot()
	var en *types.Event
	defer func() {
		if err != nil {
			bc.ctx.Revert(sn)
			return
		}
		bc.ctx.Commit(sn)

		bc.b.Body.Transactions = append(bc.b.Body.Transactions, tx)
		if en != nil {
			bc.b.Body.Events = append(bc.b.Body.Events, en)
		}
		bc.b.Body.TransactionSignatures = append(bc.b.Body.TransactionSignatures, sig)
		bc.txHashes = append(bc.txHashes, TxHash)
	}()
	if err := bc.ctx.UseTimeSlot(slot, string(TxHash[:])); err != nil {
		return err
	}
	index := uint16(len(bc.b.Body.Transactions))
	TXID := types.TransactionID(bc.b.Header.Height, index)
	if tx.To == common.ZeroAddr {
		if err := bc.cn.ExecuteTransaction(bc.ctx, tx, TXID); err != nil {
			return err
		}
	} else {
		if en, err = ExecuteContractTx(bc.ctx, tx, signer); err != nil {
			return err
		}
		if en != nil {
			en.Index = index
		}
	}
	return nil
}

func ChargeFee(ctx *types.Context, useSize uint64, signer common.Address) error {
	addr := *ctx.MainToken()
	if mcont, err := ctx.Contract(addr); err != nil {
		return err
	} else if mt, ok := mcont.(types.ChargeFee); !ok {
		return errors.New("Maintoken not charge fee")
	} else {
		fee := ctx.BasicFee()
		if useSize > 0 {
			fee = fee.MulC(int64(useSize))
		}
		if err := mt.ChargeFee(ctx.ContractContext(mcont, signer), fee); err != nil {
			return err
		}
	}
	return nil
}

func ExecuteContractTx(ctx *types.Context, tx *types.Transaction, signer common.Address) (*types.Event, error) {
	s := ctx.GetPCSize()

	if tx.UseSeq {
		seq := ctx.AddrSeq(signer)
		if seq != tx.Seq {
			return nil, errors.Errorf("invalid signer sequence siger seq %v, got %v", seq, tx.Seq)
		}
		ctx.AddAddrSeq(signer)
	}

	data, err := types.TxArg(ctx, tx)
	if err != nil {
		return nil, err
	}
	var to common.Address = tx.To
	if !ctx.IsContract(tx.To) {
		data = append([]interface{}{tx.To}, data...)
		to = *ctx.MainToken()
	}
	cont, err := ctx.Contract(to)
	if err != nil {
		return nil, err
	}
	cc := ctx.ContractContext(cont, signer)
	intr := types.NewInteractor(ctx, cont, cc)
	cc.Exec = intr.Exec
	is, err := intr.Exec(cc, to, tx.Method, data)
	intr.Distroy()
	if err != nil {
		return nil, err
	}
	result := []interface{}{}
	for _, i := range is {
		if i != nil {
			if reflect.TypeOf(i).Kind() == reflect.Slice {
				s := reflect.ValueOf(i)
				if s.Len() > 0 {
					result = append(result, i)
				}
			} else {
				result = append(result, i)
			}
		}
	}
	var en *types.Event
	if len(result) > 0 {
		en = types.NewEvent(bin.TypeWriteAll(result...))
	}
	useSize := ctx.GetPCSize() - s
	// log.Println("fee", useSize)

	return en, ChargeFee(ctx, useSize, signer)
}

// Finalize generates block that has transactions adds by AddTx
func (bc *BlockCreator) Finalize(gasLv uint16) (*types.Block, error) {
	if bc.b.Header.Height%prefix.RewardIntervalBlocks == 0 {
		if rewardMap, err := bc.ctx.ProcessReward(bc.ctx, bc.b); err != nil {
			return nil, err
		} else if bs, err := types.MarshalAddressBytesMap(rewardMap); err != nil {
			return nil, err
		} else {
			if bc.b.Body.Events == nil {
				bc.b.Body.Events = []*types.Event{}
			}
			bc.b.Body.Events = append(bc.b.Body.Events, &types.Event{
				Type:   types.EventTagReward,
				Result: bs,
			})
		}
	}
	if bc.ctx.StackSize() > 1 {
		return nil, errors.WithStack(types.ErrDirtyContext)
	}

	LevelRootHash, err := BuildLevelRoot(bc.txHashes)
	if err != nil {
		return nil, err
	}
	bc.b.Header.LevelRootHash = LevelRootHash

	bc.b.Header.ContextHash = bc.ctx.Hash()

	// log.Println("BLOCK hash", bc.b.Header.ContextHash)
	// log.Println("BLOCK", bc.ctx.Dump())
	return bc.b, nil
}
