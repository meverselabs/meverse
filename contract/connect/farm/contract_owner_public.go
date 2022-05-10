package farm

import (
	"errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////
func (cont *FarmContract) setOwner(cc *types.ContractContext, To common.Address) error {
	if cc.From() != cont.master {
		return errors.New("ownable: caller is not the master")
	}

	cc.SetContractData([]byte{tagOwner}, To[:])
	return nil
}

func (cont *FarmContract) setOwnerReward(cc *types.ContractContext, OwnerReward uint16) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagOwnerReward}, bin.Uint16Bytes(OwnerReward))
	return nil
}

func (cont *FarmContract) setTokenMaxSupply(cc *types.ContractContext, TokenMaxSupply *amount.Amount) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagTokenMaxSupply}, TokenMaxSupply.Bytes())
	return nil
}

func (cont *FarmContract) setTokenPerBlock(cc *types.ContractContext, TokenPerBlock *amount.Amount) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagTokenPerBlock}, TokenPerBlock.Bytes())
	return nil
}

func (cont *FarmContract) setStartBlock(cc *types.ContractContext, StartBlock uint32) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagStartBlock}, bin.Uint32Bytes(StartBlock))
	return nil
}

func (cont *FarmContract) Add(cc *types.ContractContext, _allocPoint uint32, _want common.Address, _withUpdate bool, _strat common.Address) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}

	if _withUpdate {
		err := cont.MassUpdatePools(cc)
		if err != nil {
			return err
		}
	}
	lastRewardBlock := cont.StartBlock(cc)
	if cc.TargetHeight() > lastRewardBlock {
		lastRewardBlock = cc.TargetHeight()
	}

	cont.setTotalAllocPoint(cc, cont.TotalAllocPoint(cc)+_allocPoint)
	newPid := cont.PoolLength(cc)
	cont.setPoolInfo(cc, newPid, &PoolInfo{
		Want:             _want,
		AllocPoint:       _allocPoint,
		LastRewardBlock:  lastRewardBlock,
		AccTokenPerShare: amount.NewAmount(0, 0),
		Strat:            _strat,
	})
	cont.addPoolLength(cc)

	if cont.isHoldShares(cc) {
		bsHeight := cc.ContractData([]byte{tagHoldSharesHeight})
		height := bin.Uint32(bsHeight)
		if _, err := cc.Exec(cc, _strat, "SetHoldShares", []interface{}{height}); err != nil {
			return err
		}
	}

	return nil
}

// Update the given pool's Cherry allocation point. Can only be called by the owner.
func (cont *FarmContract) Set(cc *types.ContractContext, _pid uint64, _allocPoint uint32, _withUpdate bool) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}

	if _withUpdate {
		err := cont.MassUpdatePools(cc)
		if err != nil {
			return err
		}
	}
	pool, err := cont._poolInfo(cc, _pid)
	if err != nil {
		return err
	}

	totalAllocPoint := cont.TotalAllocPoint(cc) - pool.AllocPoint + _allocPoint
	cont.setTotalAllocPoint(cc, totalAllocPoint)

	pool.AllocPoint = _allocPoint
	cont.setPoolInfo(cc, _pid, pool)
	return nil
}

func (cont *FarmContract) InCaseTokensGetStuck(cc *types.ContractContext, _token common.Address, _amount *amount.Amount) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}

	farmToken := cont.FarmToken(cc)
	if farmToken == _token {
		return errors.New("!safe")
	}

	// IERC20(_token).safeTransfer(msg.sender, _amount)
	if _, err := cc.Exec(cc, _token, "Transfer", []interface{}{cc.From(), _amount}); err != nil {
		return err
	}
	return nil
}

func (cont *FarmContract) SetHoldShares(cc *types.ContractContext, height uint32) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagHoldShares}, []byte{1})
	cc.SetContractData([]byte{tagHoldSharesHeight}, bin.Uint32Bytes(height))

	length := uint64(cont.PoolLength(cc))
	var pid uint64 = 0
	for ; pid < length; pid++ {
		pool, err := cont._poolInfo(cc, pid)
		if err != nil {
			return err
		}
		if _, err := cc.Exec(cc, pool.Strat, "SetHoldShares", []interface{}{height}); err != nil {
			return err
		}
	}
	return nil
}

func (cont *FarmContract) UnsetHoldShares(cc *types.ContractContext) error {
	isOwner := cont.isOwner(cc, cc.From())
	if cc.From() != cont.master && !isOwner {
		return errors.New("ownable: caller is not the owner")
	}
	cc.SetContractData([]byte{tagHoldShares}, nil)

	length := uint64(cont.PoolLength(cc))
	var pid uint64 = 0
	for ; pid < length; pid++ {
		pool, err := cont._poolInfo(cc, pid)
		if err != nil {
			return err
		}
		if _, err := cc.Exec(cc, pool.Strat, "UnsetHoldShares", []interface{}{}); err != nil {
			return err
		}
	}

	return nil
}

func (cont *FarmContract) isHoldShares(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagHoldShares})
	if len(bs) == 0 {
		return false
	}
	bsHeight := cc.ContractData([]byte{tagHoldSharesHeight})
	height := bin.Uint32(bsHeight)
	return cc.TargetHeight() > height
}

func (cont *FarmContract) isHoldShares2(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagHoldShares})
	if len(bs) == 0 {
		return false
	}

	// 13031760
	// 13042760
	bsHeight := cc.ContractData([]byte{tagHoldSharesHeight})
	height := bin.Uint32(bsHeight) + 11000
	return cc.TargetHeight() > height
}

func (cont *FarmContract) isHoldShares3(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagHoldShares})
	if len(bs) == 0 {
		return false
	}

	// 13031760
	// 13378760
	bsHeight := cc.ContractData([]byte{tagHoldSharesHeight})
	height := bin.Uint32(bsHeight) + 347000
	return cc.TargetHeight() > height
}
