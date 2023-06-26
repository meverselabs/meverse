package notformulator

import (
	"github.com/meverselabs/meverse/core/types"
)

func (cont *NotFormulatorContract) Front() interface{} {
	return &NotFormulatorFront{
		cont: cont,
	}
}

type NotFormulatorFront struct {
	cont *NotFormulatorContract
}

func (f *NotFormulatorFront) SetGenerator(cc *types.ContractContext) error {
	return f.cont.setGenerator(cc)
}
