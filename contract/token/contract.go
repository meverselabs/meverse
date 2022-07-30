package token

import (
	"bytes"
	"fmt"
	"log"
	"math/big"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
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
	if cont.isPause(cc) {
		return errors.New("paused")
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
	if !am.IsPlus() {
		return errors.Errorf("invalid transfer amount %v", am.String())
	}
	if cont.isPause(cc) {
		return errors.New("paused")
	}
	bal := cont.BalanceOf(cc, addr)
	if bal.Less(am) {
		return errors.Errorf("invalid transfer amount %v less then %v", am.String(), bal.String())
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
	if cc.From() == common.ZeroAddr {
		return errors.New("Token: TRANSFER_FROM_ZEROADDRESS")
	}

	if To == common.ZeroAddr {
		return errors.New("Token: TRANSFER_TO_ZEROADDRESS")
	}

	if Amount.IsMinus() {
		return errors.New("minus amount")
	}

	fromBalance := cont.BalanceOf(cc, cc.From())
	if fromBalance.Cmp(Amount.Int) < 0 {
		return fmt.Errorf("Token: TRANSFER_EXCEED_BALANCE %v %v %v %v", cc.From().String(), To.String(), fromBalance.String(), Amount.String())
	}

	if Amount.IsZero() {
		return nil
	}
	if err := cont.subBalance(cc, cc.From(), Amount); err != nil {
		return err
	}
	return cont.addBalance(cc, To, Amount)
}

func (cont *TokenContract) Burn(cc *types.ContractContext, am *amount.Amount) error {
	if am.IsMinus() {
		return errors.New("minus amount")
	}
	return cont.subBalance(cc, cc.From(), am)
}

func (cont *TokenContract) Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	isMinter := cont.IsMinter(cc, cc.From())
	if cc.From() != cont.Master() && !isMinter {
		return errors.New(cc.From().String() + ": not token minter")
	}
	if Amount.IsPlus() {
		return cont.addBalance(cc, To, Amount)
	}
	return nil
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

func (cont *TokenContract) Approve(cc *types.ContractContext, spender common.Address, Amount *amount.Amount) error {
	if cc.From() == common.ZeroAddr {
		return errors.New("Token: APPROVE_FROM_ZEROADDRESS")
	}

	if spender == common.ZeroAddr {
		return errors.New("Token: APPROVE_TO_ZEROADDRESS")
	}

	if Amount.IsMinus() {
		return errors.New("Token: APPROVE_NEGATIVE_AMOUNT")
	}

	cont._approve(cc, cc.From(), spender, Amount)
	return nil
}

func (cont *TokenContract) _approve(cc *types.ContractContext, owner common.Address, spender common.Address, Amount *amount.Amount) {
	cc.SetAccountData(owner, MakeAllowanceTokenKey(spender), Amount.Bytes())
}

func (cont *TokenContract) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	if Amount.IsZero() {
		return nil
	}
	balance := cont.BalanceOf(cc, From)
	if Amount.Cmp(balance.Int) > 0 {
		return errors.Errorf("the token holding quantity is insufficient balance: %v Amount: %v From: %v cont: %v", balance.String(), Amount.String(), From.String(), cont.addr.String())
	}

	allowedValue := cont.Allowance(cc, From, cc.From())
	if Amount.Cmp(allowedValue.Int) > 0 {
		return errors.Errorf("the token allowance is insufficient token: %v cc.From: %v form: %v to: %v Amount: %v allowedValue: %v", cont.addr.String(), cc.From().String(), From.String(), To.String(), Amount, allowedValue)
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

func (cont *TokenContract) SetRouter(cc *types.ContractContext, router common.Address, path []common.Address) error {
	if cc.From() != cont.Master() {
		log.Println("err not token master")
		return errors.New("not token master")
	}

	cc.SetContractData([]byte{tagRouterAddress}, router[:])
	bs := make([]byte, common.AddressLength*len(path))
	for i, addr := range path {
		copy(bs[i*common.AddressLength:], addr[:])
	}
	cc.SetContractData([]byte{tagRouterPaths}, bs)
	return nil
}

func (cont *TokenContract) SwapToMainToken(cc *types.ContractContext, amt *amount.Amount) (*amount.Amount, error) {
	var router common.Address
	{
		bs := cc.ContractData([]byte{tagRouterAddress})
		if len(bs) == 0 {
			return nil, errors.New("this token not supported swap to fee token")
		}
		copy(router[:], bs)
	}

	var path []common.Address
	{
		bs := cc.ContractData([]byte{tagRouterPaths})
		if len(bs) == 0 {
			return nil, errors.New("this token not supported swap to fee token")
		}
		path = make([]common.Address, len(bs)/common.AddressLength)
		for i := 0; i < len(bs); i += common.AddressLength {
			path[i/common.AddressLength] = common.BytesToAddress(bs[i : i+common.AddressLength])
		}
	}

	mt := *cc.MainToken()
	if cont.addr == mt {
		return nil, errors.New("this is fee token")
	}

	err := cont.Transfer(cc, cont.addr, amt)
	if err != nil {
		return nil, err
	}

	cont._approve(cc, cont.addr, router, amt)
	var swapAmt *amount.Amount
	if is, err := cc.Exec(cc, router, "SwapExactTokensForTokens", []interface{}{amt, amount.NewAmount(0, 0), path}); err != nil {
		return nil, err
	} else {
		if len(is) < 1 {
			return nil, errors.New("invalid swap result")
		} else if swapResult, ok := is[0].([]*amount.Amount); !ok {
			return nil, errors.New("invalid swap result depth 2")
		} else if len(swapResult) < 2 {
			return nil, errors.New("invalid swap result count")
		} else {
			swapAmt = swapResult[len(swapResult)-1]
		}
	}

	_, err = cc.Exec(cc, *cc.MainToken(), "Transfer", []interface{}{cc.From(), swapAmt})
	return swapAmt, err
}

func (cont *TokenContract) SetName(cc *types.ContractContext, name string) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}
	cc.SetContractData([]byte{tagTokenName}, []byte(name))
	return nil
}

func (cont *TokenContract) SetSymbol(cc *types.ContractContext, symbol string) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}
	cc.SetContractData([]byte{tagTokenSymbol}, []byte(symbol))
	return nil
}

func (cont *TokenContract) isPause(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagPause})
	if len(bs) == 1 && bs[0] == 1 {
		return true
	}
	return false
}

func (cont *TokenContract) Pause(cc *types.ContractContext) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}
	cc.SetContractData([]byte{tagPause}, []byte{1})
	return nil
}

func (cont *TokenContract) Unpause(cc *types.ContractContext) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}
	cc.SetContractData([]byte{tagPause}, nil)
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
