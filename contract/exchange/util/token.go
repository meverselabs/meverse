package util

import (
	"math/big"
	"strconv"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/types"
)

func DeployTokens(ctx *types.Context, classId uint64, size uint8, deployer common.Address) []common.Address {
	_coins := make([]common.Address, size, size)
	for k := uint8(0); k < size; k++ {
		tokenConstrunction := &token.TokenContractConstruction{
			Name:   "Token" + strconv.Itoa(int(k)),
			Symbol: "TOKEN" + strconv.Itoa(int(k)),
		}
		bs, _, _ := bin.WriterToBytes(tokenConstrunction)
		v, _ := ctx.DeployContract(deployer, classId, bs)
		tokenAContract := v.(*token.TokenContract)
		_coins[k] = tokenAContract.Address()
	}
	return _coins
}

// token.TotalSupply()
func TokenTotalSupply(cc *types.ContractContext, token common.Address) (*big.Int, error) {
	is, err := cc.Exec(cc, token, "TotalSupply", []interface{}{})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

// token.BalanceOf(from)
func TokenBalanceOf(cc *types.ContractContext, token, from common.Address) (*big.Int, error) {
	is, err := cc.Exec(cc, token, "BalanceOf", []interface{}{from})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

// token.BalanceOf(from)
func TokenAllowance(cc *types.ContractContext, token, owner, spender common.Address) (*big.Int, error) {
	is, err := cc.Exec(cc, token, "Allowance", []interface{}{owner, spender})
	if err != nil {
		return nil, err
	}
	return is[0].(*amount.Amount).Int, nil
}

// token.Transfer(to,Amount)
func SafeTransfer(cc *types.ContractContext, token, to common.Address, am *big.Int) error {
	_, err := cc.Exec(cc, token, "Transfer", []interface{}{to, ToAmount(am)})
	return err
}

// token.TansferFrom(from, to, Amount)
func SafeTransferFrom(cc *types.ContractContext, token, from, to common.Address, am *big.Int) error {
	_, err := cc.Exec(cc, token, "TransferFrom", []interface{}{from, to, ToAmount(am)})
	return err
}

// token.Apporve(to,Amount)
func TokenApprove(cc *types.ContractContext, token, to common.Address, am *big.Int) error {
	_, err := cc.Exec(cc, token, "Approve", []interface{}{to, ToAmount(am)})
	return err
}
