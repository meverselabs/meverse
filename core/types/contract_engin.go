package types

type IEngin interface {
	ContractInvoke(_cc interface{}, method string, params []interface{}) (interface{}, error)
	InitContract(_cc interface{}, contract []byte, InitArgs []interface{}) error
	UpdateContract(_cc interface{}, contract []byte) error
}

type InvokeableContract interface {
	ContractInvoke(cc *ContractContext, method string, params []interface{}) (interface{}, error)
}
