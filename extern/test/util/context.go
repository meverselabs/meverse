package util

import (
	"sync"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

type TestContext struct {
	Ctx       *types.Context
	Cn        *chain.Chain
	Idx       int
	MainToken common.Address
}

var idx int
var idxLock sync.Mutex

func NewTestContext() *TestContext {
	idxLock.Lock()
	tc := &TestContext{
		Idx: idx,
		Ctx: types.NewEmptyContext(),
	}
	idx++
	idxLock.Unlock()

	tc.MainToken = tc.InitMainToken(Admin, ClassMap)

	err := tc.InitChain(Admin)
	if err != nil {
		panic(err)
	}
	tc.Ctx = tc.Cn.NewContext()
	err = tc.Sleep(60, nil, nil)
	if err != nil {
		panic(err)
	}
	return tc
}

/////////// context ///////////
func GetCC(ctx *types.Context, contAddr common.Address, user common.Address) (*types.ContractContext, error) {

	cont, err := ctx.Contract(contAddr)
	if err != nil {
		return nil, err
	}
	cc := ctx.ContractContext(cont, user)
	intr := types.NewInteractor(ctx, cont, cc, "000000000000", false)
	cc.Exec = intr.Exec

	return cc, nil
}

func Exec(ctx *types.Context, user common.Address, contAddr common.Address, methodName string, args []interface{}) ([]interface{}, error) {
	cc, err := GetCC(ctx, contAddr, user)
	if err != nil {
		return nil, err
	}
	is, err := cc.Exec(cc, contAddr, methodName, args)
	return is, err
}
