package whitelist

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *WhiteListContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *WhiteListContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (f *front) AddWhiteListGorup(cc *types.ContractContext, groupType string) hash.Hash256 {
	return f.cont.AddWhiteListGorup(cc, groupType)
}

func (f *front) AddWhiteList(cc *types.ContractContext, groupId []byte, addrs []common.Address) error {
	return f.cont.AddWhiteList(cc, groupId, addrs)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) WhiteListGroupType(cc *types.ContractContext, groupId hash.Hash256) string {
	return f.cont.WhiteListGroupType(cc, groupId)
}

func (f *front) WhiteList(cc *types.ContractContext, groupId hash.Hash256) (string, []common.Address, error) {
	return f.cont.WhiteList(cc, groupId)
}

func (f *front) IsAllow(cc *types.ContractContext, groupId hash.Hash256, user common.Address) bool {
	return f.cont.IsAllow(cc, groupId, user)
}

func (f *front) GroupData(cc *types.ContractContext, groupId hash.Hash256, user common.Address) []byte {
	return f.cont.GroupData(cc, groupId, user)
}
