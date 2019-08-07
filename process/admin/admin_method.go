package admin

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
)

// AdminAddress returns the admin address
func (p *Admin) AdminAddress(loader types.Loader, name string) common.Address {
	lw := types.NewLoaderWrapper(p.pid, loader)

	if bs := lw.ProcessData(toAdminAddressKey(name)); len(bs) == 0 {
		panic(ErrNotExistAdminAddress)
	} else {
		var addr common.Address
		copy(addr[:], bs)
		return addr
	}
}
