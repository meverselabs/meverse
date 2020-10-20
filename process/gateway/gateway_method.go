package gateway

import (
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// IsProcessedERC20TXID returns the erc20 txid has processed or not
func (p *Gateway) IsProcessedERC20TXID(loader types.Loader, Platform string, ERC20TXID hash.Hash256) (bool, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toPlatformKey(Platform)); len(bs) == 0 {
		return false, ErrNotSupportedPlatform
	}
	if bs := lw.ProcessData(toERC20TXIDKey(Platform, ERC20TXID)); len(bs) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// SetERC20TXIDProcessed sets the given erc20 txid as processed
func (p *Gateway) SetERC20TXIDProcessed(ctw *types.ContextWrapper, Platform string, ERC20TXID hash.Hash256) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs := ctw.ProcessData(toPlatformKey(Platform)); len(bs) == 0 {
		return ErrNotSupportedPlatform
	}
	ctw.SetProcessData(toERC20TXIDKey(Platform, ERC20TXID), []byte{1})
	return nil
}

// IsProcessedOutTXID returns the out txid has processed or not
func (p *Gateway) IsProcessedOutTXID(loader types.Loader, Platform string, CoinTXID string) (bool, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toPlatformKey(Platform)); len(bs) == 0 {
		return false, ErrNotSupportedPlatform
	}
	if bs := lw.ProcessData(toOutTXIDKey(Platform, CoinTXID)); len(bs) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// SetOutTXIDProcessed sets the given txid as a processed
func (p *Gateway) SetOutTXIDProcessed(ctw *types.ContextWrapper, Platform string, CoinTXID string) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs := ctw.ProcessData(toPlatformKey(Platform)); len(bs) == 0 {
		return ErrNotSupportedPlatform
	}
	ctw.SetProcessData(toOutTXIDKey(Platform, CoinTXID), []byte{1})
	return nil
}

// GetPolicy returns the gateway policy of the platform
func (p *Gateway) GetPolicy(loader types.Loader, Platform string) (*Policy, error) {
	lw := types.NewLoaderWrapper(p.pid, loader)

	bs := lw.ProcessData(toPolicyKey(Platform))
	if len(bs) == 0 {
		return nil, ErrNotSupportedPlatform
	}
	var policy *Policy
	if err := encoding.Unmarshal(bs, &policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// SetPolicy sets the policy of the platform
func (p *Gateway) SetPolicy(ctw *types.ContextWrapper, Platform string, policy *Policy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs := ctw.ProcessData(toPlatformKey(Platform)); len(bs) == 0 {
		return ErrNotSupportedPlatform
	}
	if bs, err := encoding.Marshal(policy); err != nil {
		return err
	} else {
		ctw.SetProcessData(toPolicyKey(Platform), bs)
	}
	return nil
}

// Platforms returns the supported platform list of the gateway
func (p *Gateway) Platforms(loader types.Loader) []string {
	lw := types.NewLoaderWrapper(p.pid, loader)

	list := []string{}
	cnt := p.getPlatformCount(lw)
	for i := uint32(0); i < cnt; i++ {
		Platform := p.getPlatformName(lw, i)
		list = append(list, Platform)
	}
	return list
}

// AddPlatform adds the platform and the policy to the gateway
func (p *Gateway) AddPlatform(ctw *types.ContextWrapper, Platform string, policy *Policy) error {
	ctw = types.SwitchContextWrapper(p.pid, ctw)

	if bs := ctw.ProcessData(toPlatformKey(Platform)); len(bs) > 0 {
		return ErrAlreadySupportedPlatform
	}
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

func (p *Gateway) getPlatformCount(lw types.LoaderWrapper) uint32 {
	bs := lw.ProcessData(tagPlatformCount)
	if len(bs) == 0 {
		return 0
	}
	return binutil.LittleEndian.Uint32(bs)
}

func (p *Gateway) getPlatformName(lw types.LoaderWrapper, index uint32) string {
	bs := lw.ProcessData(toPlatformIndexKey(index))
	return string(bs)
}
