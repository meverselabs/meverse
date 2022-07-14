package imo

import (
	"errors"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Write Functions
//////////////////////////////////////////////////
func (cont *ImoContract) Deposit(cc *types.ContractContext, _amount *amount.Amount) error {
	allow, err := cont.CheckWhiteList(cc, cc.From())
	if err != nil {
		return err
	}
	if !allow {
		return errors.New("not allow")
	}
	targetHeight := cc.TargetHeight()
	// require (block.number > startBlock && block.number < endBlock, 'not ifo time');
	if targetHeight < cont.StartBlock(cc) || targetHeight > cont.EndBlock(cc) {
		return errors.New("not imo time")
	}
	// require (_amount > 0, 'need _amount > 0');
	if !_amount.IsPlus() {
		return errors.New("need _amount > 0")
	}

	ui, err := cont._userInfo(cc, cc.From())
	if err != nil {
		return err
	}
	pl := cont.PayLimit(cc)
	if len(pl.Int.Bytes()) > 0 && _amount.Add(ui.Amt).Cmp(pl.Int) > 0 {
		return errors.New("over pay limit")
	}

	// lpToken.safeTransferFrom(address(msg.sender), address(this), _amount);
	payToken := cont.PayToken(cc)
	if _, err := cc.Exec(cc, payToken, "TransferFrom", []interface{}{cc.From(), cont.addr, _amount}); err != nil {
		return err
	}

	// if userInfo[msg.sender].amount == 0 {
	//   addressList.push(address(msg.sender))
	// }
	if ui.Amt.IsZero() {
		if err := cont.addAddress(cc, cc.From()); err != nil {
			return err
		}
	}
	// userInfo[msg.sender].amount = userInfo[msg.sender].amount.add(_amount)
	ui.Amt = ui.Amt.Add(_amount)
	cont.setUserInfo(cc, cc.From(), ui)
	// totalAmount = totalAmount.add(_amount)
	cont.addTotalAmount(cc, _amount)
	return nil
}

func (cont *ImoContract) Harvest(cc *types.ContractContext) error {
	targetHeight := cc.TargetHeight()
	// require (block.number > endBlock, 'not harvest time');
	if targetHeight < cont.EndBlock(cc) {
		return errors.New("not harvest time")
	}
	ui, err := cont._userInfo(cc, cc.From())
	if err != nil {
		return err
	}
	// require (userInfo[msg.sender].amount > 0, 'have you participated?');
	if !ui.Amt.IsPlus() {
		return errors.New("have you participated?")
	}
	// require (!userInfo[msg.sender].claimed, 'nothing to harvest');
	if ui.Claimed {
		return errors.New("nothing to harvest")
	}

	offeringTokenAmount, err := cont.GetOfferingAmount(cc, cc.From())
	if err != nil {
		return err
	}
	refundingTokenAmount, err := cont.GetRefundingAmount(cc, cc.From())
	if err != nil {
		return err
	}

	// offeringToken.safeTransfer(address(msg.sender), offeringTokenAmount);
	projectToken := cont.ProjectToken(cc)
	if _, err := cc.Exec(cc, projectToken, "Transfer", []interface{}{cc.From(), offeringTokenAmount}); err != nil {
		return err
	}
	if refundingTokenAmount.IsPlus() {
		// lpToken.safeTransfer(address(msg.sender), refundingTokenAmount);
		payToken := cont.PayToken(cc)
		if _, err := cc.Exec(cc, payToken, "Transfer", []interface{}{cc.From(), refundingTokenAmount}); err != nil {
			return err
		}
	}

	ui.Claimed = true
	cont.setUserInfo(cc, cc.From(), ui)
	return nil
}
