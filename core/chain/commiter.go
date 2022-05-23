package chain

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

// Committer enables to commit block with pre-executed context
type Committer interface {
	Generators() ([]common.Address, error)
	ValidateHeader(bh *types.Header) error
	ExecuteBlockOnContext(b *types.Block, ctx *types.Context, SigMap map[hash.Hash256]common.Address) error
	ConnectBlockWithContext(b *types.Block, ctx *types.Context) error
	NewContext() *types.Context
}

type chainCommiter struct {
	cn *Chain
}

func NewChainCommiter(cn *Chain) *chainCommiter {
	ct := &chainCommiter{
		cn: cn,
	}
	return ct
}

func (ct *chainCommiter) Generators() ([]common.Address, error) {
	return ct.cn.store.Generators()
}

func (ct *chainCommiter) ValidateHeader(bh *types.Header) error {
	ct.cn.Lock()
	defer ct.cn.Unlock()

	return ct.cn.validateHeader(bh)
}

func (ct *chainCommiter) ExecuteBlockOnContext(b *types.Block, ctx *types.Context, sm map[hash.Hash256]common.Address) error {
	ct.cn.Lock()
	defer ct.cn.Unlock()

	return ct.cn.executeBlockOnContext(b, ctx, sm)
}

func (ct *chainCommiter) ConnectBlockWithContext(b *types.Block, ctx *types.Context) error {
	ct.cn.Lock()
	defer ct.cn.Unlock()

	return ct.cn.connectBlockWithContext(b, ctx)
}

func (ct *chainCommiter) NewContext() *types.Context {
	ct.cn.Lock()
	defer ct.cn.Unlock()

	return types.NewContext(ct.cn.store)
}
