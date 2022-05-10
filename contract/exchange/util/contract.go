package util

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

func Owner(cc *types.ContractContext, cont common.Address) (common.Address, error) {
	is, err := cc.Exec(cc, cont, "Owner", []interface{}{})
	if err != nil {
		return ZeroAddress, err
	}
	return is[0].(common.Address), nil
}

func ExType(cc *types.ContractContext, cont common.Address) (uint8, error) {
	is, err := cc.Exec(cc, cont, "ExType", []interface{}{})
	if err != nil {
		return 255, err
	}
	return is[0].(uint8), nil
}

func TokenIndex(cc *types.ContractContext, cont, _token common.Address) (uint8, error) {
	is, err := cc.Exec(cc, cont, "TokenIndex", []interface{}{_token})
	if err != nil {
		return 255, err
	}
	return is[0].(uint8), nil
}

func GetMintAdminFee(cc *types.ContractContext, cont common.Address, _reserve0, _reserve1 *big.Int) (*big.Int, error) {
	is, err := cc.Exec(cc, cont, "GetMintAdminFee", []interface{}{ToAmount(_reserve0), ToAmount(_reserve1)})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

func MintedAdminBalance(cc *types.ContractContext, cont common.Address) (*big.Int, error) {
	is, err := cc.Exec(cc, cont, "MintedAdminBalance", []interface{}{})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

func GetDy(cc *types.ContractContext, cont common.Address, in, out uint8, dx *big.Int, from common.Address) (*big.Int, error) {
	is, err := cc.Exec(cc, cont, "GetDy", []interface{}{in, out, dx, from})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

func FeeAddress(cc *types.ContractContext, cont, from common.Address) (uint64, error) {
	is, err := cc.Exec(cc, cont, "FeeAddress", []interface{}{from})
	//is, err := cc.Exec(cc, cont, "Fee", []interface{}{})
	if err != nil {
		return 0, err
	}
	return is[0].(uint64), nil
}

/*
func Fee(cc *types.ContractContext, cont common.Address) (uint64, error) {
	is, err := cc.Exec(cc, cont, "Fee", []interface{}{})
	if err != nil {
		return 0, err
	}
	return is[0].(uint64), nil
}
*/
