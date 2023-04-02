package chain

import (
	"log"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// Init initializes the chain
func (cn *Chain) UpdateInit(genesisContextData *types.ContextData) error {
	cn.Lock()
	defer cn.Unlock()

	GenesisHash := hash.Hashes(hash.Hash(cn.store.ChainID().Bytes()), genesisContextData.Hash())
	Height := cn.store.Height()
	if Height > 0 {
		if h, err := cn.store.Hash(0); err != nil {
			return err
		} else {
			if GenesisHash != h {
				return errors.WithStack(piledb.ErrInvalidGenesisHash)
			}
		}
	} else {
		if err := cn.store.UpdateStoreGenesis(GenesisHash, genesisContextData); err != nil {
			return err
		}
	}
	if err := cn.store.UpdatePrepare(); err != nil {
		return err
	}

	// OnLoadChain
	ctx := types.NewContext(cn.store)
	for _, s := range cn.services {
		if err := s.OnLoadChain(ctx); err != nil {
			return err
		}
	}

	//log.Println("Chain loaded", cn.store.Height(), ctx.PrevHash().String())

	cn.isInit = true
	return nil
}

// ConnectBlock try to connect block to the chain
func (cn *Chain) UpdateBlock(b *types.Block, SigMap map[hash.Hash256]common.Address) error {
	cn.closeLock.RLock()
	defer cn.closeLock.RUnlock()
	if cn.isClose {
		return errors.WithStack(ErrChainClosed)
	}

	cn.Lock()
	defer cn.Unlock()

	if err := cn.validateHeader(&b.Header); err != nil {
		return err
	}
	ctx := types.NewContext(cn.store)
	if receipts, err := cn.UpdateExecuteBlockOnContext(b, ctx, SigMap); err != nil {
		return err
	} else {
		return cn.updateBlockWithContext(b, ctx, receipts)
	}
}

func (cn *Chain) updateBlockWithContext(b *types.Block, ctx *types.Context, receipts types.Receipts) error {
	if b.Header.ContextHash != ctx.Hash() {
		log.Println("CONNECT", ctx.Hash(), b.Header.ContextHash, ctx.Dump())
		panic("")
		// return errors.WithStack(ErrInvalidContextHash)
	}

	if cn.store.Version(b.Header.Height) > 1 {
		if b.Header.ReceiptHash != bin.MustWriterToHash(&receipts) {
			log.Println("CONNECT", bin.MustWriterToHash(&receipts), b.Header.ReceiptHash, receipts)
			panic("")
			return errors.WithStack(ErrInvalidReceiptHash)
		}
	}

	if ctx.StackSize() > 1 {
		return errors.WithStack(types.ErrDirtyContext)
	}

	if err := cn.store.UpdateContext(b, ctx, receipts); err != nil {
		return err
	}
	var ca []*common.SyncChan
	cn.waitLock.Lock()
	for _, c := range cn.waitChan {
		ca = append(ca, c)
	}
	cn.waitLock.Unlock()
	for _, c := range ca {
		c.Send(b.Header.Height)
	}

	for _, s := range cn.services {
		s.OnBlockConnected(b.Clone(), ctx)
	}
	return nil
}

func (cn *Chain) UpdateExecuteBlockOnContext(b *types.Block, ctx *types.Context, sm map[hash.Hash256]common.Address) (types.Receipts, error) {
	TxSigners, TxHashes, err := cn.validateTransactionSignatures(b, sm)
	if err != nil {
		return nil, err
	}

	types.CheckABI(b, cn.NewContext())

	// Execute Transctions
	currentSlot := types.ToTimeSlot(b.Header.Timestamp)
	receipts := types.Receipts{}
	for i, tx := range b.Body.Transactions {
		slot := types.ToTimeSlot(tx.Timestamp)
		if slot < currentSlot-1 {
			return nil, errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		} else if slot > currentSlot {
			return nil, errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}

		sn := ctx.Snapshot()
		if err := ctx.UseTimeSlot(slot, string(TxHashes[i][:])); err != nil {
			ctx.Revert(sn)
			return nil, err
		}
		TXID := types.TransactionID(b.Header.Height, uint16(len(b.Body.Transactions)))
		if tx.VmType != types.Evm {
			if tx.To == common.ZeroAddr {
				if !ctx.IsAdmin(TxSigners[i]) {
					ctx.Revert(sn)
					return nil, errors.WithStack(ErrInvalidAdminAddress)
				}
				if _, err := cn.ExecuteTransaction(ctx, tx, TXID); err != nil {
					ctx.Revert(sn)
					return nil, err
				}
			} else {
				if err := ExecuteContractTx(ctx, tx, TxSigners[i], TXID); err != nil {
					ctx.Revert(sn)
					return nil, err
				}
			}
			receipt := new(etypes.Receipt)
			receipts = append(receipts, receipt)
		} else {
			if _, receipt, err := cn.ApplyEvmTransaction(ctx, tx, uint16(i), TxSigners[i]); err != nil {
				ctx.Revert(sn)
				return nil, err
			} else {
				receipts = append(receipts, receipt)
			}
		}
		ctx.Commit(sn)
	}
	if ctx.StackSize() > 1 {
		return nil, errors.WithStack(types.ErrDirtyContext)
	}
	if b.Header.Height%prefix.RewardIntervalBlocks == 0 {
		if _, err := ctx.ProcessReward(ctx, b); err != nil {
			return nil, err
		}
	}
	if ctx.StackSize() > 1 {
		return nil, errors.WithStack(types.ErrDirtyContext)
	}
	return receipts, nil
}
