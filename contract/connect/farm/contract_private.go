package farm

import (
	"bytes"
	"errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////
func (cont *FarmContract) isOwner(cc *types.ContractContext, owner common.Address) bool {
	return cont.Owner(cc) == owner
}

func (cont *FarmContract) _poolInfo(cc *types.ContractContext, _pid uint64) (*PoolInfo, error) {
	bs := cc.ContractData(makePoolInfoKey(_pid))

	data := &PoolInfo{}
	if _, err := data.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, err
	}
	return data, nil
}

func (cont *FarmContract) _userInfo(cc *types.ContractContext, _pid uint64, _user common.Address) (*UserInfo, error) {
	a, b, err := cont.UserInfo(cc, _pid, _user)
	if err != nil {
		return nil, err
	}
	user := &UserInfo{a, b}
	return user, nil
}

func (cont *FarmContract) _sharesTotal(cc *types.ContractContext, pool *PoolInfo) (*amount.Amount, error) {
	ins, err := cc.Exec(cc, pool.Strat, "SharesTotal", []interface{}{})
	if err != nil {
		return nil, err
	}
	sharesTotal, ok := ins[0].(*amount.Amount)
	if !ok {
		return nil, errors.New("invalid strat SharesTotal")
	}
	return sharesTotal, nil
}

func (cont *FarmContract) _wantLockedTotal(cc *types.ContractContext, pool *PoolInfo) (*amount.Amount, error) {
	ins, err := cc.Exec(cc, pool.Strat, "WantLockedTotal", []interface{}{})
	if err != nil {
		return nil, err
	}
	wantLockedTotal, ok := ins[0].(*amount.Amount)
	if !ok {
		return nil, errors.New("invalid strat WantLockedTotal")
	}
	return wantLockedTotal, nil
}

func (cont *FarmContract) setTotalAllocPoint(cc *types.ContractContext, totalAllocPoint uint32) {
	cc.SetContractData([]byte{tagTotalAllocPoint}, bin.Uint32Bytes(totalAllocPoint))
}

func (cont *FarmContract) setPoolInfo(cc *types.ContractContext, pid uint64, pool *PoolInfo) error {
	bf := new(bytes.Buffer)
	_, err := pool.WriteTo(bf)
	if err != nil {
		return err
	}
	cc.SetContractData(makePoolInfoKey(pid), bf.Bytes())
	return nil
}

func (cont *FarmContract) setUserInfo(cc *types.ContractContext, pid uint64, user common.Address, userInfo *UserInfo) {
	bf := new(bytes.Buffer)
	userInfo.WriteTo(bf)
	cc.SetContractData(makeUserInfoKey(pid, user), bf.Bytes())
}

func (cont *FarmContract) addPoolLength(cc *types.ContractContext) uint64 {
	pl := cont.PoolLength(cc)
	pl++
	cc.SetContractData([]byte{tagPoolLength}, bin.Uint64Bytes(pl))
	return pl
}

func (cont *FarmContract) safeFarmTokenTransfer(cc *types.ContractContext, to common.Address, amt *amount.Amount) error {
	sendAmt := amt
	// uint256 CherryBal = IERC20(CherryAddr).balanceOf(address(this));
	farmToken := cont.FarmToken(cc)
	balanceOf, err := cont.getFarmAmountValue(cc, farmToken, "BalanceOf", cont.addr)
	if err != nil {
		return err
	} else if balanceOf.Cmp(amt.Int) < 0 {
		sendAmt = balanceOf
	}

	if _, err := cc.Exec(cc, farmToken, "Transfer", []interface{}{to, sendAmt}); err != nil {
		return err
	}
	return nil
}

func (cont *FarmContract) safeIncreaseAllowance(cc *types.ContractContext, token common.Address, spender common.Address, inc *amount.Amount) error {
	if allowance, err := cont.getFarmAmountValue(cc, token, "Allowance", cont.addr, spender); err != nil {
		return err
	} else {
		allowance = allowance.Add(inc)
		if _, err := cc.Exec(cc, token, "Approve", []interface{}{spender, allowance}); err != nil {
			return err
		}
	}
	return nil
}

func (cont *FarmContract) getFarmAmountValue(cc *types.ContractContext, token common.Address, method string, params ...interface{}) (*amount.Amount, error) {
	if amt, err := cont.callContAmountValue(cc, token, method, params...); err != nil {
		return nil, err
	} else {
		return amt, nil
	}
}

func (cont FarmContract) callContAmountValue(cc *types.ContractContext, conAddr common.Address, method string, params ...interface{}) (*amount.Amount, error) {
	if ins, err := cc.Exec(cc, conAddr, method, params); err != nil {
		return nil, err
	} else if len(ins) == 0 {
		return nil, errors.New("invalid " + conAddr.String() + " " + method)
	} else if val, ok := ins[0].(*amount.Amount); !ok {
		return nil, errors.New("invalid " + conAddr.String() + " " + method + " amount")
	} else {
		return val, nil
	}
}
