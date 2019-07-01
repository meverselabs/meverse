package chain

import "github.com/fletaio/fleta/core/types"

// Committer enables to commit block with pre-executed context
type Committer interface {
	ConnectBlockWithContext(b *types.Block, ctx *types.Context) error
}

type chainCommiter struct {
	cn *Chain
}

func newChainCommiter(cn *Chain) *chainCommiter {
	ct := &chainCommiter{
		cn: cn,
	}
	return ct
}

func (ct *chainCommiter) ConnectBlockWithContext(b *types.Block, ctx *types.Context) error {
	ct.cn.Lock()
	defer ct.cn.Unlock()

	return ct.cn.connectBlockWithContext(b, ctx)
}
