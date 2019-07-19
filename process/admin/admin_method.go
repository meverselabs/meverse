package admin

import "github.com/fletaio/fleta/common"

// AdminAddress returns the admin address
func (p *Admin) AdminAddress() common.Address {
	return p.adminAddress.Clone()
}
