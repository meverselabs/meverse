package gateway

import (
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// HasERC20TXID returns the erc20 txid has processed or not
func (p *Gateway) HasERC20TXID(lw types.LoaderWrapper, ERC20TXID hash.Hash256) bool {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.ProcessData(toERC20TXIDKey(ERC20TXID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Gateway) setERC20TXID(ctw *types.ContextWrapper, ERC20TXID hash.Hash256) {
	ctw.SetProcessData(toERC20TXIDKey(ERC20TXID), []byte{1})
}

// HasOutTXID returns the out txid has processed or not
func (p *Gateway) HasOutTXID(lw types.LoaderWrapper, CoinTXID string) bool {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.ProcessData(toOutTXIDKey(CoinTXID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Gateway) setOutTXID(ctw *types.ContextWrapper, CoinTXID string) {
	ctw.SetProcessData(toOutTXIDKey(CoinTXID), []byte{1})
}
