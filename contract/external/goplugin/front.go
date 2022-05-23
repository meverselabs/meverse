package goplugin

import "github.com/meverselabs/meverse/core/types"

func (cont *PluginContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *PluginContract
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (f *front) ContractInvoke(cc *types.ContractContext, method string, params []interface{}) ([]interface{}, error) {
	return f.cont.ContractInvoke(cc, method, params)
}
