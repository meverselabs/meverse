package util

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

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
func ViewAmount(ctx *types.Context, contAddr common.Address, methodName string) (*amount.Amount, error) {
	cc, err := GetCC(ctx, contAddr, ZeroAddress)
	if err != nil {
		return nil, err
	}
	is, err := cc.Exec(cc, contAddr, methodName, []interface{}{})
	return is[0].(*amount.Amount), err
}
func ViewAddress(ctx *types.Context, contAddr common.Address, methodName string) (common.Address, error) {
	cc, err := GetCC(ctx, contAddr, ZeroAddress)
	if err != nil {
		return ZeroAddress, err
	}
	is, err := cc.Exec(cc, contAddr, methodName, []interface{}{})
	return is[0].(common.Address), err
}
