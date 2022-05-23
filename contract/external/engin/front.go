package engin

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *EnginContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *EnginContract
}

func (f *front) AddEngin(cc *types.ContractContext, Name string, Description string, EnginURL string) error {
	return f.cont.addEngin(cc, Name, Description, EnginURL)
}
func (f *front) LoadEngin(cc *types.ContractContext, Name string, Version uint32) (types.IEngin, error) {
	return f.cont.loadEngin(cc, Name, Version)
}
func (f *front) DeploryContract(cc *types.ContractContext, EnginName string, EnginVersion uint32, contract []byte, initArgs []interface{}, updateable bool) (addr common.Address, err error) {
	return f.cont.deploryContract(cc, EnginName, EnginVersion, contract, initArgs, updateable)
}
func (f *front) EnginDescription(cc *types.ContractContext, Name string, Version uint32) (string, error) {
	return f.cont.EnginDescription(cc, Name, Version)
}
func (f *front) EnginVersion(cc *types.ContractContext, name string) uint32 {
	return f.cont.EnginVersion(cc, name)
}
