package farm

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

type FarmContract struct {
	addr   common.Address
	master common.Address
}

func (cont *FarmContract) Name() string {
	return "FarmContract"
}

func (cont *FarmContract) Address() common.Address {
	return cont.addr
}

func (cont *FarmContract) Master() common.Address {
	return cont.master
}

func (cont *FarmContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *FarmContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &FarmContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagOwner}, data.Owner[:])
	cc.SetContractData([]byte{tagFarmToken}, data.FarmToken[:])

	cc.SetContractData([]byte{tagOwnerReward}, bin.Uint16Bytes(data.OwnerReward))

	cc.SetContractData([]byte{tagTokenMaxSupply}, data.TokenMaxSupply.Bytes())
	cc.SetContractData([]byte{tagTokenPerBlock}, data.TokenPerBlock.Bytes())
	cc.SetContractData([]byte{tagStartBlock}, bin.Uint32Bytes(data.StartBlock))
	return nil
}

func (cont *FarmContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////
func (cont *FarmContract) Owner(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagOwner})
	var owner common.Address
	copy(owner[:], bs)
	return owner
}

func (cont *FarmContract) OwnerReward(cc *types.ContractContext) uint16 {
	bs := cc.ContractData([]byte{tagOwnerReward})
	if len(bs) == 2 {
		return bin.Uint16(bs)
	}
	return 0
}

func (cont *FarmContract) FarmToken(cc *types.ContractContext) common.Address {
	var FarmToken common.Address
	bs := cc.ContractData([]byte{tagFarmToken})
	copy(FarmToken[:], bs)
	return FarmToken
}

func (cont *FarmContract) TokenMaxSupply(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagTokenMaxSupply})
	am := amount.NewAmountFromBytes(bs)
	return am
}

func (cont *FarmContract) TokenPerBlock(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagTokenPerBlock})
	am := amount.NewAmountFromBytes(bs)
	return am
}

func (cont *FarmContract) StartBlock(cc *types.ContractContext) uint32 {
	bs := cc.ContractData([]byte{tagStartBlock})
	if len(bs) == 4 {
		return bin.Uint32(bs)
	}
	return 0
}

func (cont *FarmContract) PoolLength(cc *types.ContractContext) uint64 {
	bs := cc.ContractData([]byte{tagPoolLength})
	if len(bs) == 8 {
		return bin.Uint64(bs)
	}
	return 0
}

func (cont *FarmContract) PoolInfo(cc *types.ContractContext, pid uint64) (common.Address, uint32, uint32, *amount.Amount, common.Address, error) {
	data, err := cont._poolInfo(cc, pid)
	if err != nil {
		return common.Address{}, 0, 0, nil, common.Address{}, err
	}

	return data.Want, data.AllocPoint, data.LastRewardBlock, data.AccTokenPerShare, data.Strat, nil
}

func (cont *FarmContract) UserInfo(cc *types.ContractContext, pid uint64, user common.Address) (*amount.Amount, *amount.Amount, error) {
	bs := cc.ContractData(makeUserInfoKey(pid, user))

	if len(bs) != 0 {
		data := &UserInfo{}
		if _, err := data.ReadFrom(bytes.NewReader(bs)); err != nil {
			return nil, nil, err
		}
		return data.Shares, data.RewardDebt, nil
	} else {
		return &amount.Amount{Int: big.NewInt(0)}, &amount.Amount{Int: big.NewInt(0)}, nil
	}
}

func (cont *FarmContract) TotalAllocPoint(cc *types.ContractContext) uint32 {
	bs := cc.ContractData([]byte{tagTotalAllocPoint})
	if len(bs) == 4 {
		return bin.Uint32(bs)
	}
	return 0
}

// Return reward multiplier over the given _from to _to block.
func (cont *FarmContract) GetMultiplier(cc *types.ContractContext, _from uint32, _to uint32) (uint32, error) {
	farmToken := cont.FarmToken(cc)
	ins, err := cc.Exec(cc, farmToken, "TotalSupply", []interface{}{})
	if err != nil {
		return 0, err
	}
	if tokenSupply, ok := ins[0].(*amount.Amount); ok {
		TokenMaxSupply := cont.TokenMaxSupply(cc)
		if tokenSupply.Cmp(TokenMaxSupply.Int) >= 0 {
			return 0, nil
		}
		return _to - _from, nil
	}
	return 0, errors.New("invalid token supply")
}

// View function to see pending Cherry on frontend.
func (cont *FarmContract) PendingReward(cc *types.ContractContext, _pid uint64, _user common.Address) (*amount.Amount, error) {
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return nil, err
	}
	user, err := cont._userInfo(cc, _pid, _user)
	if err != nil {
		return nil, err
	}

	var total *amount.Amount
	if !cont.isHoldShares2(cc) {
		sharesTotal, err := cont._sharesTotal(cc, pool)
		if err != nil {
			return nil, err
		}
		total = sharesTotal
	} else {
		wantLockedTotal, err := cont._wantLockedTotal(cc, pool)
		if err != nil {
			return nil, err
		}
		total = wantLockedTotal
	}
	accTokenPerShare := amount.NewAmountFromBytes(pool.AccTokenPerShare.Bytes())

	if cc.TargetHeight() > pool.LastRewardBlock && len(total.Bytes()) != 0 {
		// uint256 multiplier = getMultiplier(pool.lastRewardBlock, block.number);
		multiplier, err := cont.GetMultiplier(cc, pool.LastRewardBlock, cc.TargetHeight())
		if err != nil {
			return nil, err
		}
		totalAllocPoint := cont.TotalAllocPoint(cc)
		// uint256 CherryReward = multiplier.mul(CherryPerBlock).mul(pool.allocPoint).div(totalAllocPoint);
		CherryReward := amount.NewAmount(0, 0)
		if totalAllocPoint > 0 {
			tokenPerBlock := cont.TokenPerBlock(cc)
			CherryReward = tokenPerBlock.MulC(int64(multiplier)).MulC(int64(pool.AllocPoint))
			CherryReward = CherryReward.DivC(int64(totalAllocPoint))
		}

		// accCherryPerShare = accCherryPerShare.add(CherryReward.mul(1e12).div(sharesTotal));
		accTokenPerShare = accTokenPerShare.Add(CherryReward.Div(total))
	}
	if cont.isHoldShares3(cc) && _pid == 0 {
		return user.Shares.Mul(accTokenPerShare).Sub(user.RewardDebt).DivC(2), nil
	} else {
		return user.Shares.Mul(accTokenPerShare).Sub(user.RewardDebt), nil
	}
}

// View function to see staked Want tokens on frontend.
func (cont *FarmContract) StakedWantTokens(cc *types.ContractContext, _pid uint64, _user common.Address) (*amount.Amount, error) {
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return nil, err
	}
	user, err := cont._userInfo(cc, _pid, _user)
	if err != nil {
		return nil, err
	}

	if !cont.isHoldShares2(cc) {
		sharesTotal, err := cont._sharesTotal(cc, pool)
		if err != nil {
			return nil, err
		}
		wantLockedTotal, err := cont._wantLockedTotal(cc, pool)
		if err != nil {
			return nil, err
		}
		if len(sharesTotal.Bytes()) == 0 {
			return amount.NewAmount(0, 0), nil
		}
		return user.Shares.Mul(wantLockedTotal).Div(sharesTotal), nil
	} else {
		if _pid == 0 {
			return user.Shares.DivC(2), nil
		} else {
			return user.Shares, nil
		}
	}
}
