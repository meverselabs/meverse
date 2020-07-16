package gateway

import (
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// HasTokenTXID returns the token txid has processed or not
func (p *Gateway) HasTokenTXID(loader types.Loader, TokenPlatform string, TokenTXID hash.Hash256) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toTokenTXIDKey(TokenPlatform, TokenTXID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Gateway) setTokenTXID(ctw *types.ContextWrapper, TokenPlatform string, TokenTXID hash.Hash256) {
	ctw.SetProcessData(toTokenTXIDKey(TokenPlatform, TokenTXID), []byte{1})
}

// HasOutCoinTXID returns the out txid has processed or not
func (p *Gateway) HasOutCoinTXID(loader types.Loader, CoinTXID string) bool {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toOutCoinTXIDKey(CoinTXID)); len(bs) > 0 {
		return true
	} else {
		return false
	}
}

func (p *Gateway) setOutCoinTXID(ctw *types.ContextWrapper, CoinTXID string) {
	ctw.SetProcessData(toOutCoinTXIDKey(CoinTXID), []byte{1})
}

// GetPolicy returns the gateway policy of the token platform
func (p *Gateway) GetPolicy(loader types.Loader, TokenPlatform string) (*Policy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.ProcessData(toPolicyKey(TokenPlatform))
	if len(bs) == 0 {
		return nil, ErrNotExistPolicy
	}
	policy := &Policy{}
	if err := encoding.Unmarshal(bs, &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (p *Gateway) setPolicy(ctw *types.ContextWrapper, TokenPlatform string, policy *Policy) error {
	if bs, err := encoding.Marshal(policy); err != nil {
		return err
	} else {
		ctw.SetProcessData(toPolicyKey(TokenPlatform), bs)
		return nil
	}
}
