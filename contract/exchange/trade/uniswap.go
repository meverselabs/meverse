package trade

import (
	"bytes"
	"math/big"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

type UniSwap struct {
	LPToken
	Exchange
}

func (self *UniSwap) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &UniSwapConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	self._setName(cc, data.Name)
	self._setSymbol(cc, data.Symbol)

	cc.SetContractData([]byte{tagExType}, []byte{byte(UNI)})
	cc.SetContractData([]byte{tagFactory}, data.Factory[:])

	self._setNTokens(cc, 2)
	bs := []byte{}
	bs = append(bs, data.Token0.Bytes()...)
	bs = append(bs, data.Token1.Bytes()...)
	cc.SetContractData([]byte{tagExTokens}, bs)
	cc.SetContractData([]byte{tagUniToken0}, data.Token0[:])
	cc.SetContractData([]byte{tagUniToken1}, data.Token1[:])
	if err := self._setPayToken(cc, data.PayToken); err != nil {
		return err
	}

	if err := self._setOwner(cc, data.Owner); err != nil {
		return err
	}
	self._setWinner(cc, data.Winner)
	self._setFee(cc, data.Fee)
	self._setAdminFee(cc, data.AdminFee)
	self._setWinnerFee(cc, data.WinnerFee)
	self._setWhiteList(cc, data.WhiteList)
	self._setGroupId(cc, data.GroupId)
	return nil
}

//////////////////////////////////////////////////
// UniSwap Contract : getter function
//////////////////////////////////////////////////
func (self *UniSwap) token0(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagUniToken0})
	return common.BytesToAddress(bs)
}
func (self *UniSwap) token1(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagUniToken1})
	return common.BytesToAddress(bs)
}
func (self *UniSwap) reserve0(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniReserve0})
	return big.NewInt(0).SetBytes(bs)
}
func (self *UniSwap) reserve1(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniReserve1})
	return big.NewInt(0).SetBytes(bs)
}
func (self *UniSwap) price0CumulativeLast(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniPrice0CumulativeLast})
	return big.NewInt(0).SetBytes(bs)
}
func (self *UniSwap) price1CumulativeLast(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniPrice1CumulativeLast})
	return big.NewInt(0).SetBytes(bs)

}
func (self *UniSwap) kLast(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniKLast})
	return big.NewInt(0).SetBytes(bs)
}
func (self *UniSwap) reserves(cc types.ContractLoader) (*big.Int, *big.Int, uint64) {
	return self.reserve0(cc), self.reserve1(cc), self.blockTimestampLast(cc)
}
func (self *UniSwap) mintedAdminBalance(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagUniAdminBalance})
	return big.NewInt(0).SetBytes(bs)
}

// 현재 balance + 향후 mint amount
func (self *UniSwap) adminBalance(cc types.ContractLoader) *big.Int {
	balance := self.mintedAdminBalance(cc)
	_, _, _, liquidity := self.getMintAdminFee(cc, self.reserve0(cc), self.reserve1(cc))

	return Add(balance, liquidity)

}

//////////////////////////////////////////////////
// UniSwap Contract : setter function
//////////////////////////////////////////////////
func (self *UniSwap) setName(cc *types.ContractContext, name string) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setName(cc, name)
	return nil
}
func (self *UniSwap) setSymbol(cc *types.ContractContext, symbol string) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setSymbol(cc, symbol)
	return nil
}
func (self *UniSwap) setReserve0(cc *types.ContractContext, reserve0 *big.Int) {
	cc.SetContractData([]byte{tagUniReserve0}, reserve0.Bytes())
}
func (self *UniSwap) setReserve1(cc *types.ContractContext, reserve1 *big.Int) {
	cc.SetContractData([]byte{tagUniReserve1}, reserve1.Bytes())
}
func (self *UniSwap) setPrice0CumulativeLast(cc *types.ContractContext, price0CumulativeLast *big.Int) {
	cc.SetContractData([]byte{tagUniPrice0CumulativeLast}, price0CumulativeLast.Bytes())
}
func (self *UniSwap) setPrice1CumulativeLast(cc *types.ContractContext, price1CumulativeLast *big.Int) {
	cc.SetContractData([]byte{tagUniPrice1CumulativeLast}, price1CumulativeLast.Bytes())
}
func (self *UniSwap) setKLast(cc *types.ContractContext, kLast *big.Int) {
	cc.SetContractData([]byte{tagUniKLast}, kLast.Bytes())
}
func (self *UniSwap) setAdminBalance(cc *types.ContractContext, liquidityFee *big.Int) {
	cc.SetContractData([]byte{tagUniAdminBalance}, liquidityFee.Bytes())
}
func (self *UniSwap) addAdminBalance(cc *types.ContractContext, liquidityFee *big.Int) {
	cc.SetContractData([]byte{tagUniAdminBalance}, Add(self.mintedAdminBalance(cc), liquidityFee).Bytes())
}

//////////////////////////////////////////////////
// UniSwap Contract : private function
//////////////////////////////////////////////////
func (self *UniSwap) applyTransferOwnerWinner(cc *types.ContractContext) error {
	_, newOwner, err := self._applyTransferOwnerWinner(cc) // only Owner
	if err != nil {
		return err
	}
	balance := self.mintedAdminBalance(cc)
	if err := SafeTransfer(cc, self.addr, newOwner, balance); err != nil { // cc.From() = oldOwner
		return err
	}

	return nil
}
func (self *UniSwap) _update(cc *types.ContractContext, balance0, balance1, _reserve0, _reserve1 *big.Int) error {
	blockTimestamp := cc.LastTimestamp() / uint64(time.Second)
	timeElapsed := blockTimestamp - self.blockTimestampLast(cc) // overflow is not occur unit64
	if timeElapsed > 0 && _reserve0.Cmp(Zero) != 0 && _reserve1.Cmp(Zero) != 0 {
		price0CumulativeLast := Add(
			self.price0CumulativeLast(cc),
			Mul(MulDiv(_reserve1, big.NewInt(amount.FractionalMax), _reserve0), big.NewInt(int64(timeElapsed))))
		price1CumulativeLast := Add(
			self.price1CumulativeLast(cc),
			Mul(MulDiv(_reserve0, big.NewInt(amount.FractionalMax), _reserve1), big.NewInt(int64(timeElapsed))))

		self.setPrice0CumulativeLast(cc, price0CumulativeLast)
		self.setPrice1CumulativeLast(cc, price1CumulativeLast)
	}
	self.setReserve0(cc, balance0)
	self.setReserve1(cc, balance1)
	self._setBlockTimestampLast(cc, blockTimestamp)
	return nil
}

// 아직 mint 되지 않은 mintFee 조회
func (self *UniSwap) getMintAdminFee(cc types.ContractLoader, _reserve0, _reserve1 *big.Int) (bool, common.Address, *big.Int, *big.Int) {
	_owner := self.owner(cc)
	adminFeeNominator := self.adminFee(cc)
	adminFeeOn := (_owner != ZeroAddress) && (adminFeeNominator != uint64(0))
	_kLast := self.kLast(cc)
	liquidity := big.NewInt(0)
	if adminFeeOn {
		if _kLast.Cmp(Zero) != 0 {
			rootK := Sqrt(Mul(_reserve0, _reserve1))
			rootKLast := Sqrt(_kLast)
			if rootK.Cmp(rootKLast) > 0 {
				numerator := Mul(self.totalSupply(cc), Sub(rootK, rootKLast))
				denominator := Add(Sub(MulDivCC(rootK, FEE_DENOMINATOR, int64(adminFeeNominator)), rootK), rootKLast)
				liquidity = Div(numerator, denominator)
			}
		}
	}
	return adminFeeOn, _owner, _kLast, liquidity
}
func (self *UniSwap) _mintAdminFee(cc *types.ContractContext, _reserve0, _reserve1 *big.Int) bool {
	adminFeeOn, owner, kLast, liquidityFee := self.getMintAdminFee(cc, _reserve0, _reserve1)
	if adminFeeOn {
		if liquidityFee.Cmp(Zero) > 0 {
			self.addAdminBalance(cc, liquidityFee)
			self._mint(cc, owner, liquidityFee)
		}
	} else if kLast.Cmp(Zero) != 0 {
		self.setKLast(cc, Zero)
	}
	return adminFeeOn
}
func (self *UniSwap) mint(cc *types.ContractContext, to common.Address) (*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return nil, errors.New("Exchange: KILLED") // is killed
	}

	_reserve0, _reserve1, _ := self.reserves(cc)
	balance0, err := TokenBalanceOf(cc, self.token0(cc), self.addr)
	if err != nil {
		return nil, err
	}
	balance1, err := TokenBalanceOf(cc, self.token1(cc), self.addr)
	if err != nil {
		return nil, err
	}

	feeOn := self._mintAdminFee(cc, _reserve0, _reserve1)
	_totalSupply := self.totalSupply(cc)

	amount0 := Sub(balance0, _reserve0)
	if amount0.Cmp(Zero) < 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_0_INPUT")
	}
	amount1 := Sub(balance1, _reserve1)
	if amount1.Cmp(Zero) < 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_1_INPUT")
	}
	var liquidity *big.Int
	if _totalSupply.Cmp(Zero) == 0 {
		liquidity = SubC(Sqrt(Mul(amount0, amount1)), MINIMUM_LIQUIDITY)
		self._mint(cc, ZeroAddress, big.NewInt(MINIMUM_LIQUIDITY))
	} else {
		liquidity = Min(MulDiv(amount0, _totalSupply, _reserve0), MulDiv(amount1, _totalSupply, _reserve1))
	}
	if !(liquidity.Cmp(Zero) > 0) {
		return nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY_MINTED")
	}
	self._mint(cc, to, liquidity)

	if err := self._update(cc, balance0, balance1, _reserve0, _reserve1); err != nil {
		return nil, err
	}

	if feeOn {
		self.setKLast(cc, Mul(self.reserve0(cc), self.reserve1(cc)))
	}

	return liquidity, nil
}

// 실제 balance와 reserve가 벌어짐
// 실제 balance조회 삭제 - Transfer가 나중에
func (self *UniSwap) _burnBeforeTransfer(cc *types.ContractContext, _token0, _token1 common.Address, feeOn bool) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	_reserve0, _reserve1 := self.reserve0(cc), self.reserve1(cc)
	balance0, err := TokenBalanceOf(cc, _token0, self.addr)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	balance1, err := TokenBalanceOf(cc, _token1, self.addr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	liquidity := self.balanceOf(cc, self.addr)
	_totalSupply := self.totalSupply(cc)
	amount0 := MulDiv(liquidity, balance0, _totalSupply)
	amount1 := MulDiv(liquidity, balance1, _totalSupply)
	if !(amount0.Cmp(Zero) > 0 && amount1.Cmp(Zero) > 0) {
		return nil, nil, nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY_BURNED")
	}

	if err := self._burn(cc, self.addr, liquidity); err != nil {
		return nil, nil, nil, nil, err
	}

	balance0.Set(Sub(balance0, amount0))
	balance1.Set(Sub(balance1, amount1))
	if err := self._update(cc, balance0, balance1, _reserve0, _reserve1); err != nil {
		return nil, nil, nil, nil, err
	}

	if feeOn {
		self.setKLast(cc, Mul(balance0, balance1))
	}

	return amount0, amount1, balance0, balance1, nil
}

func (self *UniSwap) burn(cc *types.ContractContext, to common.Address) (*big.Int, *big.Int, error) {
	self.Lock()
	defer self.Unlock()

	// Owner가 pair를 직접 call하는 경우, liquidity는 전송된 상태, 잔액 > mintedAdminBalance
	if cc.From() == self.owner(cc) {
		ownerBalance, err := TokenBalanceOf(cc, self.addr, cc.From())
		if err != nil {
			return nil, nil, err
		}
		if ownerBalance.Cmp(self.mintedAdminBalance(cc)) < 0 {
			return nil, nil, errors.New("Exchange: OWNER_LIQUIDITY")
		}
	}

	_reserve0, _reserve1 := self.reserve0(cc), self.reserve1(cc)
	feeOn := self._mintAdminFee(cc, _reserve0, _reserve1)

	_token0, _token1 := self.token0(cc), self.token1(cc)
	amount0, amount1, _, _, err := self._burnBeforeTransfer(cc, _token0, _token1, feeOn)
	if err != nil {
		return nil, nil, err
	}
	if err := SafeTransfer(cc, _token0, to, amount0); err != nil {

		return nil, nil, err
	}
	if err := SafeTransfer(cc, _token1, to, amount1); err != nil {
		return nil, nil, err
	}
	return amount0, amount1, nil
}

// fee : whitelist Fee 반영
func (self *UniSwap) swap(cc *types.ContractContext, amount0Out, amount1Out *big.Int, to common.Address, data []byte, from common.Address) error {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return errors.New("Exchange: KILLED") // is killed
	}

	fee, err := self._feeAddress(cc, from)
	if err != nil {
		return err
	}

	if amount0Out.Cmp(Zero) < 0 || amount1Out.Cmp(Zero) < 0 {
		return errors.New("Exchange: INSUFFICIENT_OUTPUT_AMOUNT")
	}

	if !(amount0Out.Cmp(Zero) > 0 || amount1Out.Cmp(Zero) > 0) {
		return errors.New("Exchange: INSUFFICIENT_OUTPUT_AMOUNT")
	}
	_reserve0, _reserve1, _ := self.reserves(cc)
	if !(amount0Out.Cmp(_reserve0) < 0 && amount1Out.Cmp(_reserve1) < 0) {
		return errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
	}
	_token0 := self.token0(cc)
	_token1 := self.token1(cc)
	/*
		if !(to != _token0 && to != _token1) {
			return errors.New("Exchange: INVALID_TO")
		}
	*/

	// flash swap : start
	if amount0Out.Cmp(Zero) > 0 {
		if err := SafeTransfer(cc, _token0, to, amount0Out); err != nil {
			return err
		}
	}
	if amount1Out.Cmp(Zero) > 0 {
		if err := SafeTransfer(cc, _token1, to, amount1Out); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		if _, err := cc.Exec(cc, to, "FlashSwapCall", []interface{}{cc, cc.From(), ToAmount(amount0Out), ToAmount(amount1Out), data}); err != nil {
			return err
		}
	}
	balance0, err := TokenBalanceOf(cc, _token0, self.addr)
	if err != nil {
		return err
	}
	balance1, err := TokenBalanceOf(cc, _token1, self.addr)
	if err != nil {
		return err
	}
	// flash swap : end

	var amount0In, amount1In *big.Int
	if balance0.Cmp(Sub(_reserve0, amount0Out)) > 0 {
		amount0In = Sub(balance0, Sub(_reserve0, amount0Out))
	} else {
		amount0In = big.NewInt(0)
	}
	if balance1.Cmp(Sub(_reserve1, amount1Out)) > 0 {
		amount1In = Sub(balance1, Sub(_reserve1, amount1Out))
	} else {
		amount1In = big.NewInt(0)
	}

	if !(amount0In.Cmp(Zero) > 0 || amount1In.Cmp(Zero) > 0) {
		return errors.New("Exchange: INSUFFICIENT_INPUT_AMOUNT")
	}

	balance0Adjusted := Sub(MulC(balance0, FEE_DENOMINATOR), MulC(amount0In, int64(fee)))
	balance1Adjusted := Sub(MulC(balance1, FEE_DENOMINATOR), MulC(amount1In, int64(fee)))
	if Mul(balance0Adjusted, balance1Adjusted).Cmp(Mul(Mul(_reserve0, _reserve1), Mul(big.NewInt(FEE_DENOMINATOR), big.NewInt(FEE_DENOMINATOR)))) < 0 {
		return errors.New("Exchange: K")
	}

	if err := self._update(cc, balance0, balance1, _reserve0, _reserve1); err != nil {
		return err
	}
	return nil
}

func (self *UniSwap) skim(cc *types.ContractContext, to common.Address) error {
	self.Lock()
	defer self.Unlock()

	if err := self.onlyOwner(cc); err != nil {
		return err
	}

	_token0 := self.token0(cc)
	_token1 := self.token1(cc)
	balance0, err := TokenBalanceOf(cc, _token0, self.addr)
	if err != nil {
		return err
	}
	balance1, err := TokenBalanceOf(cc, _token1, self.addr)
	if err != nil {
		return err
	}
	reserve0 := self.reserve0(cc)
	reserve1 := self.reserve1(cc)
	amount0 := Sub(balance0, reserve0)
	if amount0.Cmp(Zero) < 0 {
		return errors.New("Exchange: INSUFFICIENT_0_BALANCE")
	}
	amount1 := Sub(balance1, reserve1)
	if amount1.Cmp(Zero) < 0 {
		return errors.New("Exchange: INSUFFICIENT_1_BALANCE")
	}
	if amount0.Cmp(Zero) > 0 {
		if err := SafeTransfer(cc, _token0, to, amount0); err != nil {
			return err
		}
	}
	if amount1.Cmp(Zero) > 0 {
		if err := SafeTransfer(cc, _token1, to, amount1); err != nil {
			return err
		}
	}
	return nil
}

func (self *UniSwap) sync(cc *types.ContractContext) error {
	self.Lock()
	defer self.Unlock()

	_token0 := self.token0(cc)
	_token1 := self.token1(cc)
	balance0, err := TokenBalanceOf(cc, _token0, self.addr)
	if err != nil {
		return err
	}
	balance1, err := TokenBalanceOf(cc, _token1, self.addr)
	if err != nil {
		return err
	}
	reserve0 := self.reserve0(cc)
	reserve1 := self.reserve1(cc)

	return self._update(cc, balance0, balance1, reserve0, reserve1)
}

// stableswap과 logic 동일
func (self *UniSwap) withdrawAdminFees2(cc *types.ContractContext) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if err := self.onlyOwner(cc); err != nil { // only owner
		return nil, nil, nil, nil, nil, err
	}

	_reserve0 := self.reserve0(cc)
	_reserve1 := self.reserve1(cc)
	feeOn := self._mintAdminFee(cc, _reserve0, _reserve1)
	adminBalance := self.mintedAdminBalance(cc)

	if adminBalance.Cmp(Zero) <= 0 {
		return adminBalance, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil
	}

	if err := SafeTransfer(cc, self.addr, self.addr, adminBalance); err != nil { // self.addr = pair
		return nil, nil, nil, nil, nil, err
	}

	_token0, _token1 := self.token0(cc), self.token1(cc)
	amount0, amount1, balance0, balance1, err := self._burnBeforeTransfer(cc, _token0, _token1, feeOn) // reserve 변경됨
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	self.setAdminBalance(cc, Zero)

	payToken := self.payToken(cc)
	if payToken != ZeroAddress {
		fee, err := self.feeAddress(cc, cc.From())
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if payToken == _token0 { // token1 in -> token0 out
			outAmt0, err := UniGetAmountOut(fee, amount1, balance1, balance0)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if !(outAmt0.Cmp(balance0) < 0) {
				return nil, nil, nil, nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
			}
			err = self._update(cc, Sub(balance0, outAmt0), Add(balance1, amount1), _reserve0, _reserve1)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			amount0 = Add(amount0, outAmt0)
			amount1 = big.NewInt(0)
		} else if payToken == _token1 { // token0 in -> token1 out
			outAmt1, err := UniGetAmountOut(fee, amount0, balance0, balance1)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if !(outAmt1.Cmp(_reserve1) < 0) {
				return nil, nil, nil, nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
			}
			if err := self._update(cc, Add(balance0, amount0), Sub(balance1, outAmt1), _reserve0, _reserve1); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			amount0 = big.NewInt(0)
			amount1 = Add(amount1, outAmt1)
		} else {
			return nil, nil, nil, nil, nil, errors.New("Exchange: PAYTOKEN_NOT_EXIST")
		}
	}

	tokens := self.tokens(cc)
	var adminFees []*big.Int
	if tokens[0] == _token0 {
		adminFees = []*big.Int{amount0, amount1}
	} else {
		adminFees = []*big.Int{amount1, amount0}
	}

	ownnerFees, winnerFees, err := self._divideFee(cc, adminFees)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return adminBalance, ownnerFees[0], ownnerFees[1], winnerFees[0], winnerFees[1], nil
}

// stableswap과 logic 동일
func (self *UniSwap) withdrawAdminFees(cc *types.ContractContext) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if err := self.onlyOwner(cc); err != nil { // only owner
		return nil, nil, nil, nil, nil, err
	}

	feeOn := self._mintAdminFee(cc, self.reserve0(cc), self.reserve1(cc))
	adminBalance := self.mintedAdminBalance(cc)

	if adminBalance.Cmp(Zero) <= 0 {
		return adminBalance, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil
	}

	if err := SafeTransfer(cc, self.addr, self.addr, adminBalance); err != nil { // self.addr = pair
		return nil, nil, nil, nil, nil, err
	}

	_token0, _token1 := self.token0(cc), self.token1(cc)
	amount0, amount1, _reserve0, _reserve1, err := self._burnBeforeTransfer(cc, _token0, _token1, feeOn) // reserve 변경됨
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	self.setAdminBalance(cc, Zero)

	payToken := self.payToken(cc)
	if payToken != ZeroAddress {
		fee, err := self.feeAddress(cc, cc.From())
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if payToken == _token0 {
			outAmt0, err := UniGetAmountOut(fee, amount1, _reserve1, _reserve0)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if !(outAmt0.Cmp(_reserve0) < 0) {
				return nil, nil, nil, nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
			}
			err = self._update(cc, Add(_reserve0, outAmt0), Add(_reserve1, amount1), _reserve0, _reserve1)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			amount0 = Add(amount0, outAmt0)
			amount1 = big.NewInt(0)
		} else if payToken == _token1 {
			outAmt1, err := UniGetAmountOut(fee, amount0, _reserve0, _reserve1)
			if err != nil {
				return nil, nil, nil, nil, nil, err
			}
			if !(outAmt1.Cmp(_reserve1) < 0) {
				return nil, nil, nil, nil, nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY")
			}
			if err := self._update(cc, Sub(_reserve0, amount0), Sub(_reserve1, outAmt1), _reserve0, _reserve1); err != nil {
				return nil, nil, nil, nil, nil, err
			}
			amount0 = big.NewInt(0)
			amount1 = Add(amount1, outAmt1)
		} else {
			return nil, nil, nil, nil, nil, errors.New("Exchange: PAYTOKEN_NOT_EXIST")
		}
	}

	ownnerFees, winnerFees, err := self._divideFee(cc, []*big.Int{amount0, amount1})
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return adminBalance, ownnerFees[0], ownnerFees[1], winnerFees[0], winnerFees[1], nil
}
