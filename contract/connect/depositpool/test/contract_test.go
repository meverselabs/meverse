package test

import (
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/contract/connect/depositpool"
	"github.com/meverselabs/meverse/extern/test/util"
)

func init() {
}

const (
	None uint8 = 1 + iota
	ErrDepositNoApprove
	Holder0
	Deposit
	ErrDepositDuplicate
	Holder1
	Holder9
	LockDeposit
	ErrDepositAfterLock
	AdminIsHolder
	PoolBalanceOf300
	ErrWithdrawBeforeUnlock
	UnlockWithdraw
	ErrWithdrawNoDeposit
	WithdrawDuplicate
	PoolBalanceOf2700
	PoolBalanceOf0
	AdminBalanceOf_2400
	ErrReclaimTokenNotOwner
	ErrReclaimTokenExceedBalance
	AfterReclaimTokenAdminBalanceOf
)

var tokenAddr common.Address
var poolAddr common.Address

func scenario(point uint8, t *testing.T) (inf interface{}, err error) {
	util.RegisterContractClass(&depositpool.DepositPoolContract{}, "JsContract")

	tc := util.NewTestContext()
	tokenAddr = tc.MakeToken("TestToken", "TESTTOKEN", "1000000")

	tokenContArgs := &depositpool.DepositPoolContractConstruction{
		Owner: util.Users[0],
		Token: tokenAddr,
		Amt:   amount.MustParseAmount("300"),
	}
	tokenContType := &depositpool.DepositPoolContract{}

	poolAddr = tc.DeployContract(tokenContType, tokenContArgs)

	inf, err = tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[0], amount.NewAmount(10, 0))
	if err != nil {
		t.Error("Transfer", err, inf)
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Deposit")
	if point == ErrDepositNoApprove {
		return
	}

	_, err = tc.MakeTx(util.AdminKey, tokenAddr, "Approve", poolAddr, amount.NewAmount(3000, 0))
	if err != nil {
		t.Error("Approve", err)
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Holder")
	if point == Holder0 {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Deposit")
	if point == Deposit {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Deposit")
	if point == ErrDepositDuplicate {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Holder")
	if point == Holder1 {
		return
	}

	for i, u := range util.Users {
		if i < 2 {
			continue
		}
		tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", u, amount.NewAmount(1, 0))
		tc.MakeTx(util.AdminKey, tokenAddr, "Transfer", u, amount.NewAmount(300, 0))
		tc.MakeTx(util.UserKeys[i], tokenAddr, "Approve", poolAddr, amount.NewAmount(300, 0))
		tc.MakeTx(util.UserKeys[i], poolAddr, "Deposit")
	}
	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Holder")
	if point == Holder9 {
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], poolAddr, "LockDeposit")
	if point == LockDeposit {
		return
	}

	tc.MakeTx(util.AdminKey, tc.MainToken, "Transfer", util.Users[1], amount.NewAmount(1, 0))
	tc.MakeTx(util.AdminKey, tokenAddr, "Transfer", util.Users[1], amount.NewAmount(300, 0))
	tc.MakeTx(util.UserKeys[1], tokenAddr, "Approve", poolAddr, amount.NewAmount(300, 0))
	inf, err = tc.MakeTx(util.UserKeys[1], poolAddr, "Deposit")
	if point == ErrDepositAfterLock {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "IsHolder", util.Admin)
	if point == AdminIsHolder {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, tokenAddr, "BalanceOf", poolAddr)
	if point == PoolBalanceOf300 {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Withdraw")
	if point == ErrWithdrawBeforeUnlock {
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], poolAddr, "UnlockWithdraw")
	if point == UnlockWithdraw {
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], poolAddr, "Withdraw")
	if point == ErrWithdrawNoDeposit {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "Withdraw")
	if point == WithdrawDuplicate {
		return
	}

	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", poolAddr)
	if point == PoolBalanceOf2700 {
		return
	}

	for i, k := range util.UserKeys {
		if i < 2 {
			continue
		}
		tc.MakeTx(k, poolAddr, "Withdraw")
	}
	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", poolAddr)
	if point == PoolBalanceOf0 {
		return
	}

	inf = tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", util.Admin)
	if point == AdminBalanceOf_2400 {
		return
	}

	inf, err = tc.MakeTx(util.AdminKey, tokenAddr, "Transfer", poolAddr, amount.NewAmount(10, 0))
	if err != nil {
		t.Error("Transfer", err, inf)
	}

	inf, err = tc.MakeTx(util.AdminKey, poolAddr, "ReclaimToken", tokenAddr, amount.MustParseAmount("20"))
	if point == ErrReclaimTokenNotOwner {
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], poolAddr, "ReclaimToken", tokenAddr, amount.MustParseAmount("20"))
	if point == ErrReclaimTokenExceedBalance {
		return
	}

	inf, err = tc.MakeTx(util.UserKeys[0], poolAddr, "ReclaimToken", tokenAddr, amount.MustParseAmount("10"))
	if err != nil {
		t.Error("Transfer", err, inf)
	}

	poolBal := tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", poolAddr)
	adminBal := tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", util.Admin)
	user0Bal := tc.MustSendTx(util.AdminKey, tokenAddr, "BalanceOf", util.Users[0])
	if point == AfterReclaimTokenAdminBalanceOf {
		poolAmt := poolBal[0].(*amount.Amount)
		adminAmt := adminBal[0].(*amount.Amount)
		user0Amt := user0Bal[0].(*amount.Amount)
		return []*amount.Amount{poolAmt, adminAmt, user0Amt}, nil
	}
	return
}

func TestSuccess(t *testing.T) {
	inf, err := scenario(None, t)
	if err != nil {
		t.Error("None", err, inf)
	}
}

func TestErrDepositNoApprove(t *testing.T) {
	inf, err := scenario(ErrDepositNoApprove, t)
	if err == nil {
		t.Error("ErrDepositNoApprove", err, inf)
	}
}

func TestHolder0(t *testing.T) {
	inf, err := scenario(Holder0, t)
	if err != nil {
		t.Error("Holder0", err, inf)
	}
}

func TestDeposit(t *testing.T) {
	inf, err := scenario(Deposit, t)
	if err != nil {
		t.Error("Deposit", err, inf)
	}
}
func TestErrDepositDuplicate(t *testing.T) {
	inf, err := scenario(ErrDepositDuplicate, t)
	if err == nil {
		t.Error("ErrDepositDuplicate", err, inf)
	}
}
func TestHolder1(t *testing.T) {
	inf, err := scenario(Holder1, t)
	if err != nil {
		t.Error("Holder1", err, inf)
	}
}
func TestHolder9(t *testing.T) {
	inf, err := scenario(Holder9, t)
	if err != nil {
		t.Error("Holder9", err, inf)
	}
}
func TestLockDeposit(t *testing.T) {
	inf, err := scenario(LockDeposit, t)
	if err != nil {
		t.Error("LockDeposit", err, inf)
	}
}
func TestErrDepositAfterLock(t *testing.T) {
	inf, err := scenario(ErrDepositAfterLock, t)
	if err == nil {
		t.Error("ErrDepositAfterLock", err, inf)
	}
}
func TestAdminIsHolder(t *testing.T) {
	inf, err := scenario(AdminIsHolder, t)
	if err != nil {
		t.Error("AdminIsHolder", err, inf)
	}
}
func TestPoolBalanceOf300(t *testing.T) {
	inf, err := scenario(PoolBalanceOf300, t)
	if err != nil {
		t.Error("PoolBalanceOf300", err, inf)
	}
}
func TestErrWithdrawBeforeUnlock(t *testing.T) {
	inf, err := scenario(ErrWithdrawBeforeUnlock, t)
	if err == nil {
		t.Error("ErrWithdrawBeforeUnlock", err, inf)
	}
}
func TestUnlockWithdraw(t *testing.T) {
	inf, err := scenario(UnlockWithdraw, t)
	if err != nil {
		t.Error("UnlockWithdraw", err, inf)
	}
}
func TestErrWithdrawNoDeposit(t *testing.T) {
	inf, err := scenario(ErrWithdrawNoDeposit, t)
	if err == nil {
		t.Error("Withdraw", err, inf)
	}
}
func TestWithdrawDuplicate(t *testing.T) {
	inf, err := scenario(WithdrawDuplicate, t)
	if err != nil {
		t.Error("WithdrawDuplicate", err, inf)
	}
}
func TestPoolBalanceOf2700(t *testing.T) {
	inf, err := scenario(PoolBalanceOf2700, t)
	if err != nil {
		t.Error("PoolBalanceOf2700", err, inf)
	}
}
func TestPoolBalanceOf0(t *testing.T) {
	inf, err := scenario(PoolBalanceOf0, t)
	if err != nil {
		t.Error("PoolBalanceOf0", err, inf)
	}
}
func TestAdminBalanceOf_2400(t *testing.T) {
	inf, err := scenario(AdminBalanceOf_2400, t)
	if err != nil {
		t.Error("AdminBalanceOf_2400", err, inf)
	}
}
func TestErrReclaimTokenNotOwner(t *testing.T) {
	inf, err := scenario(ErrReclaimTokenNotOwner, t)
	if err == nil {
		t.Error("ErrReclaimTokenNotOwner", err, inf)
	}
}
func TestErrReclaimTokenExceedBalance(t *testing.T) {
	inf, err := scenario(ErrReclaimTokenExceedBalance, t)
	if err == nil {
		t.Error("ErrReclaimTokenExceedBalance", err, inf)
	}
}
func TestAfterReclaimTokenAdminBalanceOf(t *testing.T) {
	inf, err := scenario(AfterReclaimTokenAdminBalanceOf, t)
	if err != nil {
		t.Error("AfterReclaimTokenAdminBalanceOf", err, inf)
	}
}
