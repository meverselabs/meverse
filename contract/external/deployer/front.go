package deployer

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *DeployerContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *DeployerContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (f *front) SetOwner(cc *types.ContractContext, NewOwner common.Address) error {
	return f.cont.SetOwner(cc, NewOwner)
}

func (f *front) Update(cc *types.ContractContext, EnginName string, EnginVersion uint32, contract []byte) error {
	return f.cont.Update(cc, EnginName, EnginVersion, contract)
}

func (f *front) InitContract(cc *types.ContractContext, contract []byte, params []interface{}) error {
	return f.cont.InitContract(cc, contract, params)
}

func (f *front) ContractInvoke(cc *types.ContractContext, method string, params []interface{}) (interface{}, error) {
	return f.cont.ContractInvoke(cc, method, params)
}

func (f *front) IsUpdateable(cc *types.ContractContext) bool {
	return isUpdateable(cc)
}
