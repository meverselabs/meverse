package depositpool

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

type DepositPoolContract struct {
	addr   common.Address
	master common.Address
}

func (cont *DepositPoolContract) Name() string {
	return "DepositPool"
}

func (cont *DepositPoolContract) Address() common.Address {
	return cont.addr
}

func (cont *DepositPoolContract) Master() common.Address {
	return cont.master
}

func (cont *DepositPoolContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *DepositPoolContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &DepositPoolContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagOwner}, data.Owner[:])
	cc.SetContractData([]byte{tagDepositToken}, data.Token[:])
	cc.SetContractData([]byte{tagDepositAmount}, data.Amt.Bytes())
	cc.SetContractData([]byte{tagDepositLock}, []byte{0})
	cc.SetContractData([]byte{tagWithdrawLock}, []byte{1})
	return nil
}

func (cont *DepositPoolContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *DepositPoolContract) addDepositUser(cc *types.ContractContext, from common.Address) {
	acclen := cont.Holder(cc)
	acclen.Add(acclen, big.NewInt(1))
	cc.SetContractData([]byte{tagAccountLength}, acclen.Bytes())

	cc.SetContractData(makeAccountIndexKey(acclen), from[:])
	cc.SetContractData(makeHolderKey(from), []byte{1})
}

//////////////////////////////////////////////////
// Public Write only owner Functions
//////////////////////////////////////////////////

func (cont *DepositPoolContract) IsOwner(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagOwner})
	var owner common.Address
	copy(owner[:], bs)
	return cc.From() == owner
}

func (cont *DepositPoolContract) LockDeposit(cc *types.ContractContext) error {
	if !cont.IsOwner(cc) {
		return errors.New("not owner")
	}
	cc.SetContractData([]byte{tagDepositLock}, []byte{1})
	return nil
}

func (cont *DepositPoolContract) UnlockWithdraw(cc *types.ContractContext) error {
	if !cont.IsOwner(cc) {
		return errors.New("not owner")
	}
	cc.SetContractData([]byte{tagWithdrawLock}, []byte{0})
	return nil
}

func (cont *DepositPoolContract) ReclaimToken(cc *types.ContractContext, token common.Address, amt *amount.Amount) error {
	if !cont.IsOwner(cc) {
		return errors.New("not owner")
	}
	_, err := cc.Exec(cc, token, "Transfer", []interface{}{cc.From(), amt})
	return err
}

//////////////////////////////////////////////////
// Public Write Functions
//////////////////////////////////////////////////
func (cont *DepositPoolContract) Deposit(cc *types.ContractContext) error {
	if ok, err := cont.IsDepositLock(cc); err != nil {
		return err
	} else if ok {
		return errors.New("Locked")
	}
	tokenbs := cc.ContractData([]byte{tagDepositToken})
	token := common.BytesToAddress(tokenbs)

	if cont.IsHolder(cc, cc.From()) {
		return errors.New("aleady deposit")
	}

	cont.addDepositUser(cc, cc.From())

	amtbs := cc.ContractData([]byte{tagDepositAmount})
	amt := amount.NewAmountFromBytes(amtbs)
	_, err := cc.Exec(cc, token, "TransferFrom", []interface{}{cc.From(), cont.addr, amt})
	return err
}

func (cont *DepositPoolContract) Withdraw(cc *types.ContractContext) error {
	if ok, err := cont.IsWithdrawLock(cc); err != nil {
		return err
	} else if ok {
		return errors.New("Locked")
	}
	tokenbs := cc.ContractData([]byte{tagDepositToken})
	token := common.BytesToAddress(tokenbs)

	amtbs := cc.ContractData([]byte{tagDepositAmount})
	amt := amount.NewAmountFromBytes(amtbs)

	if cont.IsHolder(cc, cc.From()) {
		_, err := cc.Exec(cc, token, "Transfer", []interface{}{cc.From(), amt})
		return err
	} else {
		return errors.New("no deposit")
	}
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *DepositPoolContract) IsDepositLock(cc *types.ContractContext) (bool, error) {
	bs := cc.ContractData([]byte{tagDepositLock})
	if len(bs) == 0 {
		return false, errors.New("not found DepositLock")
	}
	return bs[0] == 1, nil
}

func (cont *DepositPoolContract) IsWithdrawLock(cc *types.ContractContext) (bool, error) {
	bs := cc.ContractData([]byte{tagWithdrawLock})
	if len(bs) == 0 {
		return false, errors.New("not found WithdrawLock")
	}
	return bs[0] == 1, nil
}

func (cont *DepositPoolContract) Holder(cc *types.ContractContext) *big.Int {
	acclen := big.NewInt(0)
	bs := cc.ContractData([]byte{tagAccountLength})
	if len(bs) == 0 {
		return acclen
	}
	acclen.SetBytes(bs)
	return acclen
}

func (cont *DepositPoolContract) Holders(cc *types.ContractContext) []common.Address {
	acclen := cont.Holder(cc)
	i := big.NewInt(1)
	data := []common.Address{}
	for acclen.Cmp(i) >= 0 {
		bs := cc.ContractData(makeAccountIndexKey(i))
		addr := common.BytesToAddress(bs)
		data = append(data, addr)
		i.Add(i, big.NewInt(1))
	}
	return data
}

func (cont *DepositPoolContract) IsHolder(cc *types.ContractContext, addr common.Address) bool {
	bs := cc.ContractData(makeHolderKey(addr))
	return len(bs) != 0
}
