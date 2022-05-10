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

func (f *front) AddGroup(cc *types.ContractContext, delegate common.Address, method string, params []interface{}, checkResult string, result []byte) (hash.Hash256, error) {
	return f.cont.AddGroup(cc, delegate, method, params, checkResult, result)
}

func (f *front) UpdateGroupData(cc *types.ContractContext, groupId hash.Hash256, delegate common.Address, method string, params []interface{}, checkResult string, result []byte) error {
	return updateGroupData(cc, groupId, delegate, method, params, checkResult, result)
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (f *front) GroupData(cc *types.ContractContext, groupId hash.Hash256, user common.Address) ([]byte, error) {
	return f.cont.GroupData(cc, groupId, user)
}
