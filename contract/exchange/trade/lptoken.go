package trade

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

type LPToken struct {
}

//////////////////////////////////////////////////
// LPToken : private reader function
//////////////////////////////////////////////////
func (self *LPToken) name(cc types.ContractLoader) string {
	return string(cc.ContractData([]byte{tagTokenName}))
}
func (self *LPToken) symbol(cc types.ContractLoader) string {
	return string(cc.ContractData([]byte{tagTokenSymbol}))
}
func (self *LPToken) decimals(cc types.ContractLoader) *big.Int {
	return big.NewInt(amount.FractionalCount)
}
func (self *LPToken) totalSupply(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagTokenTotalSupply})
	return big.NewInt(0).SetBytes(bs)
}

// Returns the amount of tokens owned by `account`.
func (self *LPToken) balanceOf(cc types.ContractLoader, _owner common.Address) *big.Int {
	bs := cc.AccountData(_owner, []byte{tagTokenAmount})
	return big.NewInt(0).SetBytes(bs)
}

// Returns the remaining number of tokens that `spender` will be
// allowed to spend on behalf of `owner` through {transferFrom}. This is
// big.NewInt(0) by default.
func (self *LPToken) allowance(cc types.ContractLoader, owner, spender common.Address) *big.Int {
	bs := cc.AccountData(owner, makeTokenKey(spender, tagTokenApprove))
	return big.NewInt(0).SetBytes(bs)
}

//////////////////////////////////////////////////
// LPToken Contract : private writer function
//////////////////////////////////////////////////
func (self *LPToken) _setName(cc *types.ContractContext, name string) {
	cc.SetContractData([]byte{tagTokenName}, []byte(name))
}
func (self *LPToken) _setSymbol(cc *types.ContractContext, symbol string) {
	cc.SetContractData([]byte{tagTokenSymbol}, []byte(symbol))
}
func (self *LPToken) _mint(cc *types.ContractContext, to common.Address, amount *big.Int) error {
	if amount.Cmp(Zero) < 0 {
		return errors.New("LPToken: MINT_NEGATIVE_AMOUNT")
	}
	balance := Add(self.balanceOf(cc, to), amount)
	total := Add(self.totalSupply(cc), amount)

	cc.SetAccountData(to, []byte{tagTokenAmount}, balance.Bytes())
	cc.SetContractData([]byte{tagTokenTotalSupply}, total.Bytes())
	return nil
}
func (self *LPToken) _burn(cc *types.ContractContext, from common.Address, amount *big.Int) error {
	if amount.Cmp(Zero) < 0 {
		return errors.New("LPToken: BURN_NEGATIVE_AMOUNT")
	}
	balance := self.balanceOf(cc, from)
	if balance.Cmp(amount) < 0 {
		return errors.New("LPToken: BURN_EXCEED_BALANCE")
	}
	balance = Sub(balance, amount)
	total := Sub(self.totalSupply(cc), amount)

	cc.SetAccountData(from, []byte{tagTokenAmount}, balance.Bytes())
	cc.SetContractData([]byte{tagTokenTotalSupply}, total.Bytes())
	return nil
}

// @dev Approve the passed address to spend the specified amount of tokens on behalf of msg.sender.
// 	 Beware that changing an allowance with this method brings the risk that someone may use both the old
// 	 and the new allowance by unfortunate transaction ordering. One possible solution to mitigate this
// 	 race condition is to first reduce the spender's allowance to 0 and set the desired value afterwards:
// 	 https://github.com/ethereum/EIPs/issues/20#issuecomment-263524729
// @param spender The address which will spend the funds.
// @param amount The amount of tokens to be spent.
func (self *LPToken) _approve(cc *types.ContractContext, owner, spender common.Address, amount *big.Int) error {
	if owner == ZeroAddress {
		return errors.New("LPToken: APPROVE_FROM_ZEROADDRESS")
	}
	if spender == ZeroAddress {
		return errors.New("LPToken: APPROVE_TO_ZEROADDRESS")
	}
	if amount.Cmp(Zero) < 0 {
		return errors.New("LPToken: APPROVE_NEGATIVE_AMOUNT")
	}
	cc.SetAccountData(owner, makeTokenKey(spender, tagTokenApprove), amount.Bytes())
	return nil
}
func (self *LPToken) _transfer(cc *types.ContractContext, from, to common.Address, amount *big.Int) error {
	if from == ZeroAddress {
		return errors.New("LPToken: TRANSFER_FROM_ZEROADDRESS")
	}
	if to == ZeroAddress {
		return errors.New("LPToken: TRANSFER_TO_ZEROADDRESS")
	}
	if amount.Cmp(Zero) < 0 {
		return errors.New("LPToken: TRANSFER_NEGATIVE_AMOUNT")
	}
	fromBalance := self.balanceOf(cc, from)
	if fromBalance.Cmp(amount) < 0 {
		return errors.New("LPToken: TRANSFER_EXCEED_BALANCE")
	}
	fromBalance = Sub(fromBalance, amount)
	cc.SetAccountData(from, []byte{tagTokenAmount}, fromBalance.Bytes())
	cc.SetAccountData(to, []byte{tagTokenAmount}, Add(self.balanceOf(cc, to), amount).Bytes())
	return nil
}

// Sets `amount` as the allowance of `spender` over the caller's tokens.
// Returns a boolean value indicating whether the operation succeeded.
func (self *LPToken) approve(cc *types.ContractContext, spender common.Address, amount *big.Int) error {
	return self._approve(cc, cc.From(), spender, amount)
}

// @notice Increase the allowance granted to `_spender` by the caller
// @dev This is alternative to {approve} that can be used as a mitigation for
// 	the potential race condition
// @param _spender The address which will transfer the funds
// @param _added_value The amount of to increase the allowance
// @return bool success
func (self *LPToken) increaseAllowance(cc *types.ContractContext, spender common.Address, addAmount *big.Int) error {
	if addAmount.Cmp(Zero) < 0 {
		return errors.New("LPToken: INCREASEALLOWANCE_NEGATIVE_AMOUNT")
	}
	allowance := Add(self.allowance(cc, cc.From(), spender), addAmount)
	return self._approve(cc, cc.From(), spender, allowance)
}

// @notice Decrease the allowance granted to `_spender` by the caller
// @dev This is alternative to {approve} that can be used as a mitigation for
// 	the potential race condition
// @param _spender The address which will transfer the funds
// @param _subtracted_value The amount of to decrease the allowance
// @return bool success
func (self *LPToken) decreaseAllowance(cc *types.ContractContext, spender common.Address, subtractAmount *big.Int) error {
	if subtractAmount.Cmp(Zero) < 0 {
		return errors.New("LPToken: DECREASEALLOWANCE_NEGATIVE_AMOUNT")
	}
	allowance := Sub(self.allowance(cc, cc.From(), spender), subtractAmount)
	return self._approve(cc, cc.From(), spender, allowance)
}

// Moves `amount` tokens from the caller's account to `to`.
func (self *LPToken) transfer(cc *types.ContractContext, to common.Address, amount *big.Int) error {
	return self._transfer(cc, cc.From(), to, amount)
}

// Moves `amount` tokens from `from` to `to` using the
// allowance mechanism. `amount` is then deducted from the caller's
// allowance.
func (self *LPToken) transferFrom(cc *types.ContractContext, from, to common.Address, amount *big.Int) error {
	spender := cc.From()
	currentAllowance := self.allowance(cc, from, spender)
	if amount.Cmp(currentAllowance) > 0 {
		return errors.New("LPToken: TRANSFER_EXCEED_ALLOWANCE")
	}
	if currentAllowance.Cmp(MaxUint256.Int) != 0 {
		self._approve(cc, from, spender, Sub(currentAllowance, amount))
	}
	return self._transfer(cc, from, to, amount)
}
