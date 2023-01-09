package token

import (
	"errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

var (
	tagTokenName        = byte(0x01)
	tagTokenSymbol      = byte(0x02)
	tagTokenMinter      = byte(0x03)
	tagTokenTotalSupply = byte(0x04)
	tagTokenGateway     = byte(0x05)
	tagTokenAmount      = byte(0x10)
	tagCollectedFee     = byte(0x11)
	tagTokenApprove     = byte(0x12)
	tagRouterAddress    = byte(0x13)
	tagRouterPaths      = byte(0x14)
	tagPause            = byte(0x15)
	tagVersion          = byte(0x16)
	tagDelegateInfo     = byte(0x17)
	tagTokenManager     = byte(0x18)
)

func MakeAllowanceTokenKey(sender common.Address) []byte {
	return makeTokenKey(sender, tagTokenApprove)
}
func makeTokenKey(sender common.Address, key byte) []byte {
	bs := make([]byte, 1+common.AddressLength)
	bs[0] = key
	copy(bs[1:], sender[:])
	return bs
}

type delegateInfo struct {
	spender           common.Address
	feeBanker         common.Address
	approveFee        *amount.Amount
	approveLowerLimit *amount.Amount
	transferFee       *amount.Amount
}

func getDelegateInfo(cc *types.ContractContext) (*delegateInfo, error) {
	bs := cc.ContractData([]byte{tagDelegateInfo})
	if len(bs) == 0 {
		return nil, errors.New("is not setup delegator")
	}
	is, err := bin.TypeReadAll(bs, 5)
	if err != nil {
		return nil, err
	}
	var ok bool
	di := &delegateInfo{}
	if di.spender, ok = is[0].(common.Address); !ok {
		return nil, errors.New("spender is not address")
	}
	if di.feeBanker, ok = is[1].(common.Address); !ok {
		return nil, errors.New("feeBanker is not address")
	}
	if di.approveFee, ok = is[2].(*amount.Amount); !ok {
		return nil, errors.New("approveFee is not amount")
	}
	if di.approveLowerLimit, ok = is[3].(*amount.Amount); !ok {
		return nil, errors.New("limit is not amount")
	}
	if di.transferFee, ok = is[4].(*amount.Amount); !ok {
		return nil, errors.New("transferFee is not amount")
	}
	return di, nil
}

func _setDelegateInfo(cc *types.ContractContext, spender common.Address, feeBanker common.Address, approveFee, approveLowerLimit, transferFee *amount.Amount) {
	bs := bin.TypeWriteAll(spender, feeBanker, approveFee, approveLowerLimit, transferFee)
	cc.SetContractData([]byte{tagDelegateInfo}, bs)
}
