package mappfarm

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

func (cont *FarmContract) initPool(cc *types.ContractContext, want common.Address) error {
	lastRewardBlock := cont.StartBlock(cc)
	if cc.TargetHeight() > lastRewardBlock {
		lastRewardBlock = cc.TargetHeight()
	}

	cont.setPoolInfo(cc, &PoolInfo{
		// Want:             _want,
		// AllocPoint:       _allocPoint,
		LastRewardBlock:  lastRewardBlock,
		AccTokenPerShare: amount.NewAmount(0, 0),
		// Strat:            _strat,
	})
	// cont.addPoolLength(cc)
	return cont.pool.SetWant(cc, want)
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
