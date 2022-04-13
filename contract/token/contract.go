package token

import (
	"bytes"
	"log"
	"math/big"

	"github.com/pkg/errors"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/fletaio/fleta_v2/core/types"
)

type TokenContract struct {
	addr   common.Address
	master common.Address
}

func (cont *TokenContract) Address() common.Address {
	return cont.addr
}

func (cont *TokenContract) Master() common.Address {
	return cont.master
}

func (cont *TokenContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *TokenContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &TokenContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagTokenName}, []byte(data.Name))
	cc.SetContractData([]byte{tagTokenSymbol}, []byte(data.Symbol))
	for k, v := range data.InitialSupplyMap {
		cont.addBalance(cc, k, v)
	}

	return nil
}

func (cont *TokenContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *TokenContract) addBalance(cc *types.ContractContext, addr common.Address, am *amount.Amount) error {
	if !am.IsPlus() {
		return errors.Errorf("invalid transfer amount %v", am.String())
	}
	bal := cont.BalanceOf(cc, addr)

	bal = bal.Add(am)

	cc.SetAccountData(addr, []byte{tagTokenAmount}, bal.Bytes())

	bs := cc.ContractData([]byte{tagTokenTotalSupply})
	total := amount.NewAmountFromBytes(bs).Add(am)
	cc.SetContractData([]byte{tagTokenTotalSupply}, total.Bytes())

	return nil
}

func (cont *TokenContract) subBalance(cc *types.ContractContext, addr common.Address, am *amount.Amount) error {
	bal := cont.BalanceOf(cc, addr)
	if bal.Less(am) {
		return errors.Errorf("invalid transfer amount %v less then %v", am.String(), bal)
	}
	bal = bal.Sub(am)
	if bal.IsZero() {
		cc.SetAccountData(addr, []byte{tagTokenAmount}, nil)
	} else {
		cc.SetAccountData(addr, []byte{tagTokenAmount}, bal.Bytes())
	}

	bs := cc.ContractData([]byte{tagTokenTotalSupply})
	total := amount.NewAmountFromBytes(bs).Sub(am)
	cc.SetContractData([]byte{tagTokenTotalSupply}, total.Bytes())

	return nil
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *TokenContract) ChargeFee(cc *types.ContractContext, fee *amount.Amount) error {
	if cont.addr != *cc.MainToken() {
		return errors.Errorf("is not fee token: %v", cont.addr.String())
	}
	if !fee.IsPlus() {
		return errors.Errorf("invalid ChargeFee amount %v", fee.String())
	}
	if err := cont.Burn(cc, fee); err != nil {
		return err
	}

	bal := cont.CollectedFee(cc)
	cc.SetContractData([]byte{tagCollectedFee}, bal.Add(fee).Bytes())
	return nil
}

func (cont *TokenContract) SubCollectedFee(cc *types.ContractContext, am *amount.Amount) error {
	if cont.addr != *cc.MainToken() {
		return errors.Errorf("is not fee token: %v", cont.addr.String())
	}

	isMinter := cont.IsMinter(cc, cc.From())

	if cc.From() != cont.Master() && !isMinter {
		return errors.New("not token minter")
	}

	sum := cont.CollectedFee(cc)
	if sum.Less(am) {
		return errors.Errorf("invalid SubCollectedFee amount %v but remine %v", am.String(), sum.String())
	}
	sum = sum.Sub(am)
	if sum.IsZero() {
		cc.SetContractData([]byte{tagCollectedFee}, nil)
	} else {
		cc.SetContractData([]byte{tagCollectedFee}, sum.Bytes())
	}
	return nil
}

func (cont *TokenContract) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	if err := cont.subBalance(cc, cc.From(), Amount); err != nil {
		return err
	}
	return cont.addBalance(cc, To, Amount)
}

func (cont *TokenContract) Burn(cc *types.ContractContext, am *amount.Amount) error {
	return cont.subBalance(cc, cc.From(), am)
}

func (cont *TokenContract) Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	isMinter := cont.IsMinter(cc, cc.From())
	if cc.From() != cont.Master() && !isMinter {
		return errors.New("not token minter")
	}
	return cont.addBalance(cc, To, Amount)
}

func (cont *TokenContract) MintBatch(cc *types.ContractContext, Tos []common.Address, Amounts []*amount.Amount) error {
	isMinter := cont.IsMinter(cc, cc.From())
	if cc.From() != cont.Master() && !isMinter {
		return errors.New("not token minter")
	}
	if len(Tos) != len(Amounts) {
		return errors.New("not match To and Amount")
	}
	for i, To := range Tos {
		if err := cont.addBalance(cc, To, Amounts[i]); err != nil {
			return err
		}
	}
	return nil
}

func (cont *TokenContract) SetMinter(cc *types.ContractContext, To common.Address, Is bool) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}

	isMinter := cont.IsMinter(cc, To)

	if Is {
		if isMinter {
			return errors.New("already token minter")
		}
		cc.SetAccountData(To, []byte{tagTokenMinter}, []byte{1})
	} else {
		if !isMinter {
			return errors.New("not token minter")
		}
		cc.SetAccountData(To, []byte{tagTokenMinter}, nil)
	}
	return nil
}

func (cont *TokenContract) Approve(cc *types.ContractContext, To common.Address, Amount *amount.Amount) {
	cc.SetAccountData(cc.From(), MakeAllowanceTokenKey(To), Amount.Bytes())
}

func (cont *TokenContract) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	balance := cont.BalanceOf(cc, From)
	if Amount.Cmp(balance.Int) > 0 {
		return errors.New("the token holding quantity is insufficient")
	}

	allowedValue := cont.Allowance(cc, From, cc.From())
	if Amount.Cmp(allowedValue.Int) > 0 {
		return errors.New("the token allowance is insufficient")
	}
	nAllow := allowedValue.Sub(Amount)
	cc.SetAccountData(From, MakeAllowanceTokenKey(To), nAllow.Bytes())

	if err := cont.subBalance(cc, From, Amount); err != nil {
		return err
	}
	if err := cont.addBalance(cc, To, Amount); err != nil {
		return err
	}
	return nil
}

func (cont *TokenContract) TokenInRevert(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	isGateway := cont.IsGateway(cc, cc.From())
	if cc.From() != cont.Master() && !isGateway {
		return errors.New("not token gateway")
	}

	if err := cont.subBalance(cc, to, Amount); err != nil {
		return err
	}
	if err := cont.addBalance(cc, cc.From(), Amount); err != nil {
		return err
	}
	return nil
}

func (cont *TokenContract) SetGateway(cc *types.ContractContext, Gateway common.Address, Is bool) error {
	if cc.From() != cont.Master() {
		log.Println("err not token master")
		return errors.New("not token master")
	}

	isGateway := cont.IsGateway(cc, Gateway)

	if Is {
		if isGateway {
			log.Println("err already token gateway")
			return errors.New("already token gateway")
		}
		cc.SetAccountData(Gateway, []byte{tagTokenGateway}, []byte{1})
	} else {
		if !isGateway {
			log.Println("err not token gateway")
			return errors.New("not token gateway")
		}
		cc.SetAccountData(Gateway, []byte{tagTokenGateway}, nil)
	}
	log.Println("SetGateway", Gateway, Is)
	return nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *TokenContract) Name(cc types.ContractLoader) string {
	return string(cc.ContractData([]byte{tagTokenName}))
}

func (cont *TokenContract) Symbol(cc types.ContractLoader) string {
	return string(cc.ContractData([]byte{tagTokenSymbol}))
}

// todo
func (cont *TokenContract) TotalSupply(cc types.ContractLoader) *amount.Amount {
	cdbs := cc.ContractData([]byte{tagCollectedFee})
	CollectedFee := amount.NewAmountFromBytes(cdbs)

	bs := cc.ContractData([]byte{tagTokenTotalSupply})
	total := amount.NewAmountFromBytes(bs).Add(CollectedFee)
	return total
}

func (cont *TokenContract) Decimals(cc types.ContractLoader) *big.Int {
	return big.NewInt(amount.FractionalCount)
}

func (cont *TokenContract) BalanceOf(cc types.ContractLoader, from common.Address) *amount.Amount {
	bs := cc.AccountData(from, []byte{tagTokenAmount})
	return amount.NewAmountFromBytes(bs)
}

func (cont *TokenContract) IsMinter(cc types.ContractLoader, addr common.Address) bool {
	bs := cc.AccountData(addr, []byte{tagTokenMinter})
	if len(bs) == 1 && bs[0] == 1 {
		return true
	}
	return false
}

func (cont *TokenContract) IsGateway(cc types.ContractLoader, addr common.Address) bool {
	bs := cc.AccountData(addr, []byte{tagTokenGateway})
	if len(bs) == 1 && bs[0] == 1 {
		return true
	}
	return false
}

func (cont *TokenContract) CollectedFee(cc types.ContractLoader) *amount.Amount {
	bs := cc.ContractData([]byte{tagCollectedFee})
	return amount.NewAmountFromBytes(bs)
}

func (cont *TokenContract) Allowance(cc types.ContractLoader, _owner common.Address, _spender common.Address) *amount.Amount {
	bs := cc.AccountData(_owner, MakeAllowanceTokenKey(_spender))
	return amount.NewAmountFromBytes(bs)
}
