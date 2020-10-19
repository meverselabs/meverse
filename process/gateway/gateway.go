package gateway

import (
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/encoding"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// Gateway manages balance of accounts of the chain
type Gateway struct {
	*types.ProcessBase
	pid uint8
}

// NewGateway returns a Gateway
func NewGateway(pid uint8) *Gateway {
	p := &Gateway{
		pid: pid,
	}
	return p
}

// ID returns the id of the process
func (p *Gateway) ID() uint8 {
	return p.pid
}

// Name returns the name of the process
func (p *Gateway) Name() string {
	return "fleta.gateway"
}

// RequireProcessNames returns process names that used in the transaction
func (p *Gateway) RequireProcessNames() []string {
	return []string{"fleta.admin"}
}

// InitPolicy called at OnInitGenesis of an application
func (p *Gateway) InitPolicy(ctx types.Context, Platform string, policy *Policy) error {
	ctw := ctx.ProcessContext(p.pid)

	if bs, err := encoding.Marshal(policy); err != nil {
		return err
	} else {
		ctw.SetProcessData(toPlatformKey(Platform), []byte{1})
		cnt := p.getPlatformCount(ctw)
		ctw.SetProcessData(toPlatformIndexKey(cnt), []byte(Platform))
		cnt++
		ctw.SetProcessData(tagPlatformCount, binutil.LittleEndian.Uint32ToBytes(cnt))
		ctw.SetProcessData(toPolicyKey(Platform), bs)
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (p *Gateway) OnLoadChain(ps []types.Process, cv types.ProcessContextView) error {
	ap := ps[0].(*admin.Admin)
	ap.AdminAddress(cv, p.Name())
	Platforms := p.Platforms(cv)
	for _, v := range Platforms {
		if bs := cv.ProcessData(toPolicyKey(v)); len(bs) == 0 {
			return ErrPolicyShouldBeSetupInApplication
		}
	}
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (p *Gateway) AfterExecuteTransactions(ps []types.Process, b *types.Block, ctw types.ProcessContext) error {
	return nil
}
