package gateway

import (
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// HasERC20TXID returns the erc20 txid has processed or not
func (p *Gateway) HasERC20TXID(lw types.LoaderWrapper, TXID hash.Hash256) bool {
	lw = types.SwitchLoaderWrapper(p.pid, lw)

	if bs := lw.ProcessData(toERC20TXIDKey(TXID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Gateway) setERC20TXID(ctw *types.ContextWrapper, TXID hash.Hash256) {
	ctw.SetProcessData(toERC20TXIDKey(TXID), []byte{1})
}
