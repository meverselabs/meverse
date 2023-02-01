package chain

import (
	"log"

	"github.com/meverselabs/meverse/common"
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
	if err := cn.UpdateExecuteBlockOnContext(b, ctx, SigMap); err != nil {
		return err
	}
	return cn.updateBlockWithContext(b, ctx)
}

func (cn *Chain) updateBlockWithContext(b *types.Block, ctx *types.Context) error {
	if b.Header.ContextHash != ctx.Hash() {
		log.Println("CONNECT", ctx.Hash(), b.Header.ContextHash, ctx.Dump())
		panic("")
		// return errors.WithStack(ErrInvalidContextHash)
	}

	if ctx.StackSize() > 1 {
		return errors.WithStack(types.ErrDirtyContext)
	}

	if err := cn.store.UpdateContext(b, ctx); err != nil {
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

func (cn *Chain) UpdateExecuteBlockOnContext(b *types.Block, ctx *types.Context, sm map[hash.Hash256]common.Address) error {
	// Execute Transctions
	currentSlot := types.ToTimeSlot(b.Header.Timestamp)
	for _, tx := range b.Body.Transactions {
		slot := types.ToTimeSlot(tx.Timestamp)
		if slot < currentSlot-1 {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		} else if slot > currentSlot {
			return errors.WithStack(types.ErrInvalidTransactionTimeSlot)
		}

		sn := ctx.Snapshot()
		TxHash := tx.Hash(b.Header.Height)

		if err := ctx.UseTimeSlot(slot, string(TxHash[:])); err != nil {
			ctx.Revert(sn)
			return err
		}
		TXID := types.TransactionID(b.Header.Height, uint16(len(b.Body.Transactions)))
		if tx.To == common.ZeroAddr {
			if _, err := cn.ExecuteTransaction(ctx, tx, TXID); err != nil {
				ctx.Revert(sn)
				return err
			}
		} else {
			if err := ExecuteContractTx(ctx, tx, tx.From, TXID); err != nil {
				ctx.Revert(sn)
				return err
			}
		}
		ctx.Commit(sn)
	}
	if ctx.StackSize() > 1 {
		return errors.WithStack(types.ErrDirtyContext)
	}
	if b.Header.Height%prefix.RewardIntervalBlocks == 0 {
		if _, err := ctx.ProcessReward(ctx, b); err != nil {
			return err
		}
	}
	if ctx.StackSize() > 1 {
		return errors.WithStack(types.ErrDirtyContext)
	}
	return nil
}
