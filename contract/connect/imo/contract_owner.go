package imo

import (
	"errors"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// Public Owner Write Functions
//////////////////////////////////////////////////
func (cont *ImoContract) setEndBlock(cc *types.ContractContext, endBlock uint32) error {
	if !cont.IsOwner(cc) {
		return errors.New("only owner")
	}
	cc.SetContractData([]byte{tagEndBlock}, bin.Uint32Bytes(endBlock))
	return nil
}

func (cont *ImoContract) setOfferingAmount(cc *types.ContractContext, projectOffer *amount.Amount) error {
	if !cont.IsOwner(cc) {
		return errors.New("only owner")
	}
	startBlock := cont.StartBlock(cc)

	if cc.TargetHeight() >= startBlock {
		return errors.New("started imo")
	}
	cc.SetContractData([]byte{tagProjectOffering}, projectOffer.Bytes())
	return nil
}

func (cont *ImoContract) setRaisingAmount(cc *types.ContractContext, projectRaising *amount.Amount) error {
	if !cont.IsOwner(cc) {
		return errors.New("only owner")
	}
	startBlock := cont.StartBlock(cc)

	if cc.TargetHeight() >= startBlock {
		return errors.New("started imo")
	}

	cc.SetContractData([]byte{tagProjectRaising}, projectRaising.Bytes())
	return nil
}

func (cont *ImoContract) finalWithdraw(cc *types.ContractContext, payAmount *amount.Amount, offerAmount *amount.Amount) error {
	if !cont.IsOwner(cc) {
		return errors.New("only owner")
	}
	payToken := cont.PayToken(cc)
	projectToken := cont.ProjectToken(cc)
	var payAmt *amount.Amount
	var offerAmt *amount.Amount
	var ok bool
	//	require (_lpAmount < lpToken.balanceOf(address(this)), 'not enough token 0');
	if ins, err := cc.Exec(cc, payToken, "BalanceOf", []interface{}{cont.addr}); err != nil {
		return err
	} else if len(ins) == 0 {
		return errors.New("invalid " + payToken.String() + " BalanceOf return nothing")
	} else if payAmt, ok = ins[0].(*amount.Amount); !ok {
		return errors.New("invalid " + payToken.String() + " BalanceOf return !amount")
	} else if payAmt.Cmp(payAmount.Int) < 0 {
		return errors.New("not enough pay token")
	}
	// require (_offerAmount < offeringToken.balanceOf(address(this)), 'not enough token 1');
	if ins, err := cc.Exec(cc, projectToken, "BalanceOf", []interface{}{cont.addr}); err != nil {
		return err
	} else if len(ins) == 0 {
		return errors.New("invalid " + payToken.String() + " BalanceOf return nothing")
	} else if offerAmt, ok = ins[0].(*amount.Amount); !ok {
		return errors.New("invalid " + payToken.String() + " BalanceOf return !amount")
	} else if offerAmt.Cmp(offerAmount.Int) < 0 {
		return errors.New("not enough offer token")
	}

	// lpToken.safeTransfer(address(msg.sender), _lpAmount)
	if _, err := cc.Exec(cc, payToken, "Transfer", []interface{}{cc.From(), payAmount}); err != nil {
		return err
	}
	// offeringToken.safeTransfer(address(msg.sender), _offerAmount)
	if _, err := cc.Exec(cc, projectToken, "Transfer", []interface{}{cc.From(), offerAmount}); err != nil {
		return err
	}
	return nil
}

func (cont *ImoContract) usdcReclaim(cc *types.ContractContext) error {
	if !cont.IsOwner(cc) {
		return errors.New("only owner")
	}
	payToken := cont.PayToken(cc)
	var payAmt *amount.Amount
	var ok bool
	//	require (_lpAmount < lpToken.balanceOf(address(this)), 'not enough token 0');
	if ins, err := cc.Exec(cc, payToken, "BalanceOf", []interface{}{cont.addr}); err != nil {
		return err
	} else if len(ins) == 0 {
		return errors.New("invalid " + payToken.String() + " BalanceOf return nothing")
	} else if payAmt, ok = ins[0].(*amount.Amount); !ok {
		return errors.New("invalid " + payToken.String() + " BalanceOf return !amount")
	}
	// lpToken.safeTransfer(address(msg.sender), _lpAmount)
	if _, err := cc.Exec(cc, payToken, "Transfer", []interface{}{cc.From(), payAmt}); err != nil {
		return err
	}
	return nil
}
