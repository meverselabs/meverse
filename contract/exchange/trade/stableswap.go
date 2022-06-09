package trade

import (
	"bytes"
	"math/big"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

// 참조 curve-contrct/contracts/pool-templates/base/SwapTemplateBase.vy

type StableSwap struct {
	LPToken
	Exchange
}

func (self *StableSwap) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &StableSwapConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	self._setName(cc, data.Name)
	self._setSymbol(cc, data.Symbol)

	self._setExtype(cc, STABLE)
	cc.SetContractData([]byte{tagFactory}, data.Factory[:])

	self._setNTokens(cc, data.NTokens)
	bs := []byte{}
	for k := uint8(0); k < data.NTokens; k++ {
		if data.Tokens[k] == ZeroAddress {
			return errors.New("Exchange: ZERO_ADDRESS")
		}
		bs = append(bs, data.Tokens[k].Bytes()...)
	}
	cc.SetContractData([]byte{tagExTokens}, bs)
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
	self.setInitialA(cc, Mul(data.Amp, big.NewInt(A_PRECISION)))
	self.setFutureA(cc, Mul(data.Amp, big.NewInt(A_PRECISION)))

	bs = []byte{}
	for k := uint8(0); k < data.NTokens; k++ {
		bs = append(bs, bin.Uint64Bytes(data.PrecisionMul[k])...)
	}
	cc.SetContractData([]byte{tagStablePrecisionMul}, bs)

	for k := uint8(0); k < data.NTokens; k++ {
		key := []byte{tagStableRates, byte(k)}
		cc.SetContractData(key, data.Rates[k].Bytes())
	}
	return nil
}

//////////////////////////////////////////////////
// StableSwap Contract : getter function
//////////////////////////////////////////////////
func (self *StableSwap) tokenIndex(cc types.ContractLoader, _token common.Address) (uint8, error) {
	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	for k := uint8(0); k < N; k++ {
		if tokens[k] == _token {
			return k, nil
		}
	}
	return 255, errors.New("Exchange: TOKEN_INDEX")
}
func (self *StableSwap) rates(cc types.ContractLoader) []*big.Int {
	N, result := self.makeSlice(cc)
	for k := uint8(0); k < N; k++ {
		key := []byte{tagStableRates, byte(k)}
		bs := cc.ContractData(key)
		if k == 0 && len(bs) == 0 {
			return result
		}
		result[k].SetBytes(bs)
	}
	return result
}
func (self *StableSwap) precisionMul(cc types.ContractLoader) []uint64 {
	N := self.nTokens(cc)
	result := []uint64{}
	bs := cc.ContractData([]byte{tagStablePrecisionMul})
	for k := uint8(0); k < N; k++ {
		result = append(result, bin.Uint64(bs[k*8:(k+1)*8]))
	}
	return result
}
func (self *StableSwap) reserves(cc types.ContractLoader) []*big.Int {
	N, result := self.makeSlice(cc)
	for k := uint8(0); k < N; k++ {
		key := []byte{tagStableReserves, byte(k)}
		bs := cc.ContractData(key)
		if k == 0 && len(bs) == 0 { // 값이 존재하지 않으면
			return result
		}
		result[k].SetBytes(bs)
	}
	return result
}
func (self *StableSwap) initialA(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagStableInitialAmp})
	return big.NewInt(0).SetBytes(bs)
}
func (self *StableSwap) futureA(cc types.ContractLoader) *big.Int {
	bs := cc.ContractData([]byte{tagStableFutureAmp})
	return big.NewInt(0).SetBytes(bs)
}

func (self *StableSwap) initialATime(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagStableInitialAmpTime})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *StableSwap) futureATime(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagStableFutureAmpTime})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}

//////////////////////////////////////////////////
// StableSwap Contract : setter function
//////////////////////////////////////////////////
func (self *StableSwap) setName(cc *types.ContractContext, name string) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setName(cc, name)
	return nil
}
func (self *StableSwap) setSymbol(cc *types.ContractContext, symbol string) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setSymbol(cc, symbol)
	return nil
}
func (self *StableSwap) setReserves(cc *types.ContractContext, _reserves []*big.Int) {
	for k := 0; k < len(_reserves); k++ {
		key := []byte{tagStableReserves, byte(k)}
		cc.SetContractData(key, _reserves[k].Bytes())
	}

	blockTimestamp := cc.LastTimestamp() / uint64(time.Second)
	self._setBlockTimestampLast(cc, blockTimestamp)
}
func (self *StableSwap) setInitialA(cc *types.ContractContext, amp *big.Int) {
	cc.SetContractData([]byte{tagStableInitialAmp}, amp.Bytes())
}
func (self *StableSwap) setFutureA(cc *types.ContractContext, amp *big.Int) {
	cc.SetContractData([]byte{tagStableFutureAmp}, amp.Bytes())
}
func (self *StableSwap) setInitialATime(cc *types.ContractContext, _time uint64) {
	cc.SetContractData([]byte{tagStableInitialAmpTime}, bin.Uint64Bytes(_time))
}
func (self *StableSwap) setFutureATime(cc *types.ContractContext, _time uint64) {
	cc.SetContractData([]byte{tagStableFutureAmpTime}, bin.Uint64Bytes(_time))
}

//////////////////////////////////////////////////
// StableSwap Contract : private function
//////////////////////////////////////////////////
func (self *StableSwap) _A(cc types.ContractLoader) *big.Int {

	t1 := self.futureATime(cc)
	A1 := self.futureA(cc)

	timestamp := cc.LastTimestamp() / uint64(time.Second)
	if timestamp < t1 {
		A0 := self.initialA(cc)
		t0 := self.initialATime(cc)
		// Expressions in uint256 cannot have negative numbers, thus "if"
		if A1.Cmp(A0) > 0 {
			return Add(A0, DivC(MulC(Sub(A1, A0), int64(timestamp-t0)), int64(t1-t0)))
		} else {
			return Sub(A0, DivC(MulC(Sub(A0, A1), int64(timestamp-t0)), int64(t1-t0)))
		}
	} else { // when t1 == 0 or block.timestamp >= t1
		return A1
	}
}
func (self *StableSwap) a(cc types.ContractLoader) *big.Int {
	return DivC(self._A(cc), A_PRECISION)
}
func (self *StableSwap) aPrecise(cc types.ContractLoader) *big.Int {
	return self._A(cc)
}
func (self *StableSwap) _xp(cc types.ContractLoader) []*big.Int {
	N, result := self.cloneSlice(cc, self.rates(cc))
	reserves := self.reserves(cc)
	for k := uint8(0); k < N; k++ {
		result[k] = MulDivC(result[k], reserves[k], PRECISION)
	}
	return result
}
func (self *StableSwap) _xp_mem(cc types.ContractLoader, _reserves []*big.Int) []*big.Int {
	N, result := self.cloneSlice(cc, self.rates(cc))
	for k := uint8(0); k < N; k++ {
		result[k] = MulDivC(result[k], _reserves[k], PRECISION)
	}
	return result
}

// D invariant calculation in non-overflowing integer operations
// iteratively
// A * sum(x_i) * n**n + D = A * D * n**n + D**(n+1) / (n**n * prod(x_i))
// Converging solution:
// D[j+1] = (A * n**n * sum(x_i) - D[j]**(n+1) / (n**n prod(x_i))) / (A * n**n - 1)
func (self *StableSwap) _get_D(cc types.ContractLoader, _xp []*big.Int, _amp *big.Int) (*big.Int, error) {

	N := int64(self.nTokens(cc))

	S := Sum(_xp)
	Dprev := big.NewInt(0)

	if S.Cmp(Zero) == 0 {
		return big.NewInt(0), nil
	}

	D := Clone(S)
	Ann := MulC(_amp, N)

	for k := 0; k < 255; k++ {
		// If division by 0, this will be borked: only withdrawal will work. And that is good
		D_P := Clone(D)
		for j := int64(0); j < N; j++ {
			if _xp[j].Cmp(Zero) <= 0 {
				return nil, errors.New("Exchange: RESERVE_NOT_POSITIVE")
			}
			D_P = MulDiv(D_P, D, MulC(_xp[j], N))
		}
		Dprev.Set(D)
		//D = (Ann * S / A_PRECISION + D_P * N) * D / ((Ann - A_PRECISION) * D / A_PRECISION + (N + 1) * D_P)
		D = MulDiv(
			Add(MulDivC(Ann, S, A_PRECISION), MulC(D_P, N)),
			D,
			Add(MulDivC(SubC(Ann, A_PRECISION), D, A_PRECISION), MulC(D_P, N+1)))

		// Equality with the precision of 1
		if Abs(Sub(D, Dprev)).Cmp(big.NewInt(1)) <= 0 {
			return D, nil
		}
	}
	// convergence typically occurs in 4 rounds or less, this should be unreachable!
	// if it does happen the pool is borked and LPs can withdraw via `remove_liquidity`
	return nil, errors.New("Exchange: D")
}

func (self *StableSwap) _get_D_mem(cc types.ContractLoader, _reserves []*big.Int, _amp *big.Int) (*big.Int, error) {
	return self._get_D(cc, self._xp_mem(cc, _reserves), _amp)
}

//   @notice The current virtual price of the pool LP token
//   @dev Useful for calculating profits
//   @return LP token virtual price normalized to 1e18
func (self *StableSwap) getVirtualPrice(cc types.ContractLoader) (*big.Int, error) {
	D, err := self._get_D(cc, self._xp(cc), self._A(cc))
	// D is in the units similar to DAI (e.g. converted to precision 1e18)
	// When balanced, D = n * x_u - total virtual value of the portfolio
	if err != nil {
		return nil, err
	}
	token_supply := self.totalSupply(cc)
	return MulDiv(D, big.NewInt(PRECISION), token_supply), nil
}

//   @notice Calculate addition or reduction in token supply from a deposit or withdrawal
//   @dev This calculation accounts for slippage, but not fees.
//  	 Needed to prevent front-running, not for precise calculations!
//   @param _amounts Amount of each coin being deposited
//   @param _is_deposit set True for deposits, False for withdrawals
//   @return Expected amount of LP tokens received
//   수수료 포함 변경 2022.04 조광현
func (self *StableSwap) calcLPTokenAmount(cc *types.ContractContext, _amounts []*big.Int, _is_deposit bool) (*big.Int, uint64, error) {
	amp := self._A(cc)
	N, old_reserves := self.cloneSlice(cc, self.reserves(cc))

	for k := uint8(0); k < N; k++ {
		if _amounts[k].Cmp(Zero) < 0 {
			return nil, 0, errors.New("Exchange: INSUFFICIENT_INPUT")
		}
	}

	D0, err := self._get_D_mem(cc, old_reserves, amp)
	if err != nil {
		return nil, 0, err
	}
	new_reserves := make([]*big.Int, N, N)
	for k := uint8(0); k < N; k++ {
		if _is_deposit {
			new_reserves[k] = Add(old_reserves[k], _amounts[k])
		} else {
			new_reserves[k] = Sub(old_reserves[k], _amounts[k])
		}

	}
	D1, err := self._get_D_mem(cc, new_reserves, amp)
	if err != nil {
		return nil, 0, err
	}
	token_supply := self.totalSupply(cc)

	fees := make([]*big.Int, N, N)
	reserves := make([]*big.Int, N, N)
	var D2 *big.Int
	if token_supply.Cmp(Zero) > 0 {
		_fee, err := self.feeAddress(cc, cc.From())
		if err != nil {
			return nil, 0, err
		}
		fee := MulDiv(big.NewInt(int64(_fee)), big.NewInt(int64(N)), big.NewInt(int64(4*(N-1))))
		admin_fee := big.NewInt(int64(self.adminFee(cc)))
		for k := uint8(0); k < N; k++ {
			idead_balance := MulDiv(D1, old_reserves[k], D0)
			new_balance := new_reserves[k]
			difference := Abs(Sub(idead_balance, new_balance))

			fees[k] = MulDivC(fee, difference, FEE_DENOMINATOR)
			reserves[k] = Sub(new_balance, MulDivC(fees[k], admin_fee, FEE_DENOMINATOR))
			new_reserves[k] = Sub(new_reserves[k], fees[k])
		}
		D2, err = self._get_D_mem(cc, new_reserves, amp)
		if err != nil {
			return nil, 0, err
		}
	} else if _is_deposit {
		D2 = Clone(D1)
	} else {
		return nil, 0, errors.New("Exchange: LPTOKEN_SUPPLY_0")
	}

	var amt *big.Int
	ratio := uint64(0)
	if _is_deposit {
		if D0.Cmp(Zero) == 0 {
			amt = Clone(D2)
			ratio = amount.FractionalMax
		} else {
			amt = MulDiv(Sub(D2, D0), token_supply, D0)
			ratio = uint64(MulDiv(amt, big.NewInt(int64(amount.FractionalMax)), Add(token_supply, amt)).Int64())
		}
	} else {
		amt = MulDiv(Sub(D0, D2), token_supply, D0)
		amt = Add(amt, big.NewInt(1)) // RemoveLiquidityImbalance와 맞추기 위해서
		ratio = uint64(MulDiv(amt, big.NewInt(int64(amount.FractionalMax)), token_supply).Int64())
	}

	return amt, ratio, nil
}

//  @notice Deposit tokens into the pool
//  @param _amounts List of amounts of tokens to deposit
//  @param _min_mint_amount Minimum amount of LP tokens to mint from the deposit
//  @return Amount of LP tokens received by depositing
func (self *StableSwap) addLiquidity(cc *types.ContractContext, _amounts []*big.Int, _min_mint_amount *big.Int) (*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return nil, errors.New("Exchange: KILLED") // is killed
	}

	amp := self._A(cc)
	N, old_reserves := self.cloneSlice(cc, self.reserves(cc))

	// Initial invariant
	D0, err := self._get_D_mem(cc, old_reserves, amp)
	if err != nil {
		return nil, err
	}

	token_supply := self.totalSupply(cc)
	new_reserves := CloneSlice(old_reserves)
	for k := uint8(0); k < N; k++ {
		if token_supply.Cmp(Zero) == 0 {
			if _amounts[k].Cmp(Zero) <= 0 {
				return nil, errors.New("Exchange: INITILAL_DEPOSIT") // initial deposit requires all tokens
			}
		}
		// balances store amounts of c-tokens
		if _amounts[k].Cmp(Zero) < 0 {
			return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
		}
		new_reserves[k] = Add(new_reserves[k], _amounts[k])
	}

	// Invariant after change
	D1, err := self._get_D_mem(cc, new_reserves, amp)
	if err != nil {
		return nil, err
	}

	if D1.Cmp(D0) <= 0 {
		return nil, errors.New("Exchange: D1_<=_D0")
	}

	// We need to recalculate the invariant accounting for fees
	// to calculate fair user's share
	fees := make([]*big.Int, N, N)
	mint_amount := big.NewInt(0)
	reserves := make([]*big.Int, N, N)
	if token_supply.Cmp(Zero) > 0 {
		// Only account for fees if we are not the first to deposit
		// D2 := Clone(D1)
		_fee, err := self.feeAddress(cc, cc.From())
		if err != nil {
			return nil, err
		}
		fee := MulDiv(big.NewInt(int64(_fee)), big.NewInt(int64(N)), big.NewInt(int64(4*(N-1))))
		admin_fee := big.NewInt(int64(self.adminFee(cc)))
		for k := uint8(0); k < N; k++ {
			idead_balance := MulDiv(D1, old_reserves[k], D0)
			new_balance := new_reserves[k]
			difference := Abs(Sub(idead_balance, new_balance))

			fees[k] = MulDivC(fee, difference, FEE_DENOMINATOR)
			reserves[k] = Sub(new_balance, MulDivC(fees[k], admin_fee, FEE_DENOMINATOR))
			new_reserves[k] = Sub(new_reserves[k], fees[k])
		}
		self.setReserves(cc, reserves)
		D2, err := self._get_D_mem(cc, new_reserves, amp)
		if err != nil {
			return nil, err
		}
		mint_amount = MulDiv(token_supply, Sub(D2, D0), D0)
	} else {
		self.setReserves(cc, new_reserves)
		mint_amount.Set(D1) // Take the dust if there was any
	}

	if mint_amount.Cmp(Zero) < 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_LIQUIDITY_MINTED")
	}

	if mint_amount.Cmp(_min_mint_amount) < 0 {
		return nil, errors.New("Exchange: SLIPPAGE")
	}

	// Take tokens from the sender
	tokens := self.tokens(cc)
	for k := uint8(0); k < N; k++ {
		if _amounts[k].Cmp(Zero) > 0 {
			if err := SafeTransferFrom(cc, tokens[k], cc.From(), self.Address(), _amounts[k]); err != nil {
				return nil, err
			}
		}
	}

	// Mint pool tokens
	self._mint(cc, cc.From(), mint_amount)

	return mint_amount, nil
}

// Calculate x[out] if one makes x[in] = x
// Done by solving quadratic equation iteratively.
// x_1**2 + x_1 * (sum' - (A*n**n - 1) * D / (A * n**n)) = D ** (n + 1) / (n ** (2 * n) * prod' * A)
// x_1**2 + b*x_1 = c
// x_1 = (x_1**2 + c) / (2*x_1 + b)
func (self *StableSwap) _getY(cc types.ContractLoader, in, out uint8, x *big.Int, _xp []*big.Int) (*big.Int, error) {
	N := self.nTokens(cc)
	// x in the input is converted to the same price/precision
	if in == out || out < 0 || out >= N { // same coin, out below Zero,  out above N
		return nil, errors.New("Exchange: OUT")
	}

	// should be unreachable, but good for safety
	if in < 0 || in >= N {
		return nil, errors.New("Exchange: IN")
	}

	A := Clone(self._A(cc))
	D, err := self._get_D(cc, _xp, A)
	if err != nil {
		return nil, err
	}
	Ann := MulC(A, int64(N))
	c := Clone(D)
	S := big.NewInt(0)
	_x := big.NewInt(0)
	yPrev := big.NewInt(0)

	for k := uint8(0); k < N; k++ {
		if k == in {
			_x.Set(x)
		} else if k != out {
			_x.Set(_xp[k])
		} else {
			continue
		}
		S = Add(S, _x)
		c = MulDiv(c, D, MulC(_x, int64(N)))
	}
	c = MulDiv(Mul(c, D), big.NewInt(A_PRECISION), MulC(Ann, int64(N)))
	b := Sub(Add(S, MulDiv(D, big.NewInt(A_PRECISION), Ann)), D)
	y := Clone(D)
	for k := 0; k < 255; k++ {
		yPrev.Set(y)
		y = Div(Add(Mul(y, y), c), Add(MulC(y, 2), b))
		// Equality with the precision of 1
		if Abs(Sub(y, yPrev)).Cmp(big.NewInt(1)) <= 0 {
			return y, nil
		}
	}

	return nil, errors.New("Exchange: Y")
}

// whitelist fee 반영
func (self *StableSwap) _get_dy_mem(cc *types.ContractContext, in, out uint8, _dx *big.Int, _reserves []*big.Int, from common.Address) (*big.Int, *big.Int, *big.Int, error) {
	if _dx.Cmp(Zero) <= 0 {
		return nil, nil, nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	xp := self._xp_mem(cc, _reserves)
	rates := self.rates(cc)

	x := Add(xp[in], MulDivC(_dx, rates[in], PRECISION))
	y, err := self._getY(cc, in, out, x, xp)
	if err != nil {
		return nil, nil, nil, err
	}
	dy := Sub(Sub(xp[out], y), big.NewInt(1))

	_fee, err := self._feeAddress(cc, from)
	if err != nil {
		return nil, nil, nil, err
	}
	fee := MulDivCC(dy, int64(_fee), FEE_DENOMINATOR)
	admin_fee := MulDivCC(fee, int64(self.adminFee(cc)), FEE_DENOMINATOR)

	return MulDiv(Sub(dy, fee), big.NewInt(PRECISION), rates[out]), MulDiv(fee, big.NewInt(PRECISION), rates[out]), MulDiv(admin_fee, big.NewInt(PRECISION), rates[out]), nil
}

// whitelist fee 반영
func (self *StableSwap) get_dy(cc *types.ContractContext, in, out uint8, dx *big.Int, from common.Address) (*big.Int, *big.Int, *big.Int, error) {
	old_reserves := self.reserves(cc)
	return self._get_dy_mem(cc, in, out, dx, old_reserves, from)
}

// whitelist fee 반영
// no transfer but reserves change
func (self *StableSwap) _exchange(cc *types.ContractContext, in, out uint8, _dx *big.Int, _min_dy *big.Int, from common.Address) (*big.Int, error) {
	old_reserves := self.reserves(cc)
	xp := self._xp_mem(cc, old_reserves)

	rates := self.rates(cc)
	x := Add(xp[in], MulDivC(_dx, rates[in], PRECISION))
	y, err := self._getY(cc, in, out, x, xp)
	if err != nil {
		return nil, err
	}
	dy := SubC(Sub(xp[out], y), 1) // -1 just in case there were some rounding errors

	fee, err := self._feeAddress(cc, from)
	if err != nil {
		return nil, err
	}
	dy_fee := MulDivCC(dy, int64(fee), FEE_DENOMINATOR)

	// Convert all to real units
	dy = MulDiv(Sub(dy, dy_fee), big.NewInt(PRECISION), rates[out])

	if dy.Cmp(_min_dy) < 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_OUTPUT_AMOUNT")
	}

	dy_admin_fee := MulDivCC(dy_fee, int64(self.adminFee(cc)), FEE_DENOMINATOR)
	dy_admin_fee = MulDiv(dy_admin_fee, big.NewInt(PRECISION), rates[out])

	//Change reserves exactly in same way as we change actual ERC20 coin amounts
	reserves := self.reserves(cc)
	//When rounding errors happen, we undercharge admin fee in favor of LP
	reserves[in] = Add(old_reserves[in], _dx)
	reserves[out] = Sub(Sub(old_reserves[out], dy), dy_admin_fee)
	self.setReserves(cc, reserves)

	return dy, nil
}

// @notice Perform an exchange between two tokens
// @dev Index values can be found via the `tokens` public getter method
// @param in Index value for the coin to send
// @param out Index valie of the coin to recieve
// @param _dx Amount of `in` being exchanged
// @param _min_dy Minimum amount of `out` to receive
// @return Actual amount of `out` received
// whitelist fee 반영
func (self *StableSwap) exchange(cc *types.ContractContext, in, out uint8, _dx *big.Int, _min_dy *big.Int, from common.Address) (*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return nil, errors.New("Exchange: KILLED")
	}

	N := self.nTokens(cc)
	if in == out || out < 0 || out >= N { // same coin, out below Zero,  out above N
		return nil, errors.New("Exchange: OUT")
	}
	if in < 0 || in >= N {
		return nil, errors.New("Exchange: IN")
	}

	if _dx.Cmp(Zero) <= 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	dy, err := self._exchange(cc, in, out, _dx, _min_dy, from)
	if err != nil {
		return nil, err
	}

	_coins := self.tokens(cc)
	if err := SafeTransferFrom(cc, _coins[in], cc.From(), self.Address(), _dx); err != nil {
		return nil, err
	}
	if err := SafeTransfer(cc, _coins[out], cc.From(), dy); err != nil {
		return nil, err
	}

	return dy, nil

}
func (self *StableSwap) calcWithdrawCoins(cc types.ContractLoader, _amount *big.Int) ([]*big.Int, error) {
	if _amount.Cmp(Zero) <= 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	total_supply := self.totalSupply(cc)
	N, amounts := self.makeSlice(cc)

	reserves := self.reserves(cc)
	for k := uint8(0); k < N; k++ {
		old_reserve := reserves[k]
		value := MulDiv(old_reserve, _amount, total_supply)
		if value.Cmp(Zero) < 0 {
			return nil, errors.New("Exchange: WITHDRAWAL_RESULTED_IN_MINUS")
		}
		amounts[k] = value
	}
	return amounts, nil
}

// @notice Withdraw tokens from the pool
// @dev Withdrawal amounts are based on current deposit ratios
// @param _amount Quantity of LP tokens to burn in the withdrawal
// @param _min_amounts Minimum amounts of underlying tokens to receive
// @return List of amounts of tokens that were withdrawn
func (self *StableSwap) removeLiquidity(cc *types.ContractContext, _amount *big.Int, _min_amounts []*big.Int) ([]*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if _amount.Cmp(Zero) <= 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	total_supply := self.totalSupply(cc)
	N, amounts := self.makeSlice(cc)

	reserves := self.reserves(cc)
	tokens := self.tokens(cc)
	for k := uint8(0); k < N; k++ {
		old_reserve := reserves[k]
		value := MulDiv(old_reserve, _amount, total_supply)
		if value.Cmp(_min_amounts[k]) < 0 {
			return nil, errors.New("Exchange: WITHDRAWAL_RESULTED_IN_FEWER_COINS_THAN_EXPECTED")
		}
		reserves[k] = Sub(old_reserve, value)
		amounts[k] = value
		if err := SafeTransfer(cc, tokens[k], cc.From(), value); err != nil {
			return nil, err
		}
	}

	self.setReserves(cc, reserves)

	if err := self._burn(cc, cc.From(), _amount); err != nil {
		return nil, err
	}

	return amounts, nil
}

// @notice Withdraw tokens from the pool in an imbalanced amount
// @param _amounts List of amounts of underlying tokens to withdraw
// @param _max_burn_amount Maximum amount of LP token to burn in the withdrawal
// @return Actual amount of the LP token burned in the withdrawal
func (self *StableSwap) removeLiquidityImbalance(cc *types.ContractContext, _amounts []*big.Int, _max_burn_amount *big.Int) (*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return nil, errors.New("Exchange: KILLED") // is killed
	}

	N := self.nTokens(cc)
	for k := uint8(0); k < N; k++ {
		if _amounts[k].Cmp(Zero) < 0 {
			return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
		}
	}

	amp := self._A(cc)
	old_reserves := self.reserves(cc)
	D0, err := self._get_D_mem(cc, old_reserves, amp)
	if err != nil {
		return nil, err
	}
	if D0.Cmp(Zero) == 0 {
		return nil, errors.New("Exchange: D0_IS_ZERO") // is killed
	}

	N, new_reserves := self.cloneSlice(cc, old_reserves)
	for k := uint8(0); k < N; k++ {
		new_reserves[k] = Sub(new_reserves[k], _amounts[k])
	}
	D1, err := self._get_D_mem(cc, new_reserves, amp)
	if err != nil {
		return nil, err
	}

	_fee, err := self.feeAddress(cc, cc.From())
	if err != nil {
		return nil, err
	}
	fee := MulDiv(big.NewInt(int64(_fee)), big.NewInt(int64(N)), big.NewInt(int64(4*(N-1))))
	admin_fee := big.NewInt(int64(self.adminFee(cc)))
	fees := make([]*big.Int, N, N)
	reserves := make([]*big.Int, N, N)
	for k := uint8(0); k < N; k++ {
		new_balance := new_reserves[k]
		ideal_balance := MulDiv(D1, old_reserves[k], D0)
		difference := Abs(Sub(ideal_balance, new_balance))

		fees[k] = MulDivC(fee, difference, FEE_DENOMINATOR)
		reserves[k] = Sub(new_balance, MulDivC(fees[k], admin_fee, FEE_DENOMINATOR))
		new_reserves[k] = Sub(new_balance, fees[k])
	}
	self.setReserves(cc, reserves)

	D2, err := self._get_D_mem(cc, new_reserves, amp)
	if err != nil {
		return nil, err
	}

	token_supply := self.totalSupply(cc)
	token_amount := MulDiv(Sub(D0, D2), token_supply, D0)
	if token_amount.Cmp(Zero) == 0 {
		return nil, errors.New("Exchange: ZERO_TOKEN_BURN") // Zero tokens burned
	}
	token_amount = Add(token_amount, big.NewInt(1)) // In case of rounding errors - make it unfavorable for the "attacker"
	if token_amount.Cmp(_max_burn_amount) > 0 {
		return nil, errors.New("Exchange: SLIPPAGE")
	}

	if err := self._burn(cc, cc.From(), token_amount); err != nil {
		return nil, err
	}

	_coins := self.tokens(cc)
	for k := uint8(0); k < N; k++ {
		if _amounts[k].Cmp(Zero) != 0 {
			if err := SafeTransfer(cc, _coins[k], cc.From(), _amounts[k]); err != nil {
				return nil, err
			}
		}
	}

	return token_amount, nil
}

// Calculate x[in] if one reduces D from being calculated for xp to D
// Done by solving quadratic equation iteratively.
// x_1**2 + x_1 * (sum' - (A*n**n - 1) * D / (A * n**n)) = D ** (n + 1) / (n ** (2 * n) * prod' * A)
// x_1**2 + b*x_1 = c
// x_1 = (x_1**2 + c) / (2*x_1 + b)
func (self *StableSwap) _getYD(cc types.ContractLoader, A *big.Int, idx uint8, _xp []*big.Int, D *big.Int) (*big.Int, error) {
	N := self.nTokens(cc)
	// x in the input is converted to the same price/precision
	if idx < 0 || idx >= N { // idx below Zero, idx above N
		return nil, errors.New("Exchange: IDX")
	}

	Ann := MulC(A, int64(N))
	c := Clone(D)
	S := big.NewInt(0)
	_x := big.NewInt(0)
	y_prev := big.NewInt(0)

	for k := uint8(0); k < N; k++ {
		if k != idx {
			_x.Set(_xp[k])
		} else {
			continue
		}
		S = Add(S, _x)
		c = MulDiv(c, D, Mul(_x, big.NewInt(int64(N))))
	}
	c = MulDiv(c, Mul(D, big.NewInt(A_PRECISION)), MulC(Ann, int64(N)))
	b := Add(S, MulDiv(D, big.NewInt(A_PRECISION), Ann))
	y := Clone(D)
	for k := 0; k < 255; k++ {
		y_prev.Set(y)
		y = Div(Add(Mul(y, y), c), Sub(Add(MulC(y, 2), b), D))
		// Equality with the precision of 1
		if Abs(Sub(y, y_prev)).Cmp(big.NewInt(1)) <= 0 {
			return y, nil
		}
	}
	return nil, errors.New("Exchange: YD")
}

func (self *StableSwap) _calcWithdrawOneCoin(cc *types.ContractContext, _token_amount *big.Int, out uint8) (*big.Int, *big.Int, *big.Int, error) {
	// First, need to calculate
	// * Get current D
	// * Solve Eqn against y_i for D - _token_amount

	total_supply := self.totalSupply(cc)
	if total_supply.Cmp(Zero) <= 0 {
		return nil, nil, nil, errors.New("Exchange: TOTALSUPPLY_0")
	}

	amp := self._A(cc)
	xp := self._xp(cc)
	D0, err := self._get_D(cc, xp, amp)
	if err != nil {
		return nil, nil, nil, err
	}
	D1 := Sub(D0, MulDiv(_token_amount, D0, total_supply))

	new_y, err := self._getYD(cc, amp, out, xp, D1)
	if err != nil {
		return nil, nil, nil, err
	}
	N, xp_reduced := self.cloneSlice(cc, xp)
	_fee, err := self.feeAddress(cc, cc.From())
	if err != nil {
		return nil, nil, nil, err
	}
	fee := MulDiv(big.NewInt(int64(_fee)), big.NewInt(int64(N)), big.NewInt(int64(4*(N-1))))
	for k := uint8(0); k < uint8(N); k++ {
		dx_expected := big.NewInt(0)
		if k == out {
			dx_expected = Sub(MulDiv(xp[k], D1, D0), new_y)
		} else {
			dx_expected = Sub(xp[k], MulDiv(xp[k], D1, D0))
		}
		xp_reduced[k] = Sub(xp_reduced[k], MulDivC(fee, dx_expected, FEE_DENOMINATOR))
	}
	yD, err := self._getYD(cc, amp, out, xp_reduced, D1)
	if err != nil {
		return nil, nil, nil, err
	}
	dy := Sub(xp_reduced[out], yD)
	precisions := self.precisionMul(cc)
	dy = DivC(Sub(dy, big.NewInt(1)), int64(precisions[out])) // Withdraw less to account for rounding errors
	dy_0 := DivC(Sub(xp[out], new_y), int64(precisions[out])) // w/o fees

	return dy, Sub(dy_0, dy), total_supply, nil
}

// @notice Calculate the amount received when withdrawing a single coin
// @param _token_amount Amount of LP tokens to burn in the withdrawal
// @param out Index value of the coin to withdraw
// @return Amount of coin received
func (self *StableSwap) calcWithdrawOneCoin(cc *types.ContractContext, _token_amount *big.Int, out uint8) (*big.Int, *big.Int, *big.Int, error) {
	if _token_amount.Cmp(Zero) <= 0 {
		return nil, nil, nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	dy, dy_fee, token_supply, err := self._calcWithdrawOneCoin(cc, _token_amount, out)
	return dy, dy_fee, token_supply, err
}

// @notice Withdraw a single coin from the pool
// @param _token_amount Amount of LP tokens to burn in the withdrawal
// @param out Index value of the coin to withdraw
// @param _min_amount Minimum amount of coin to receive
// @return Amount of coin received
func (self *StableSwap) removeLiquidityOneCoin(cc *types.ContractContext, _token_amount *big.Int, out uint8, _min_amount *big.Int) (*big.Int, error) {
	self.Lock()
	defer self.Unlock()

	if self.isKilled(cc) {
		return nil, errors.New("Exchange: KILLED") // is killed
	}

	if _token_amount.Cmp(Zero) <= 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_INPUT")
	}

	dy, dy_fee, _, err := self._calcWithdrawOneCoin(cc, _token_amount, out)
	if err != nil {
		return nil, err
	}
	if dy.Cmp(_min_amount) < 0 {
		return nil, errors.New("Exchange: INSUFFICIENT_OUTPUT_AMOUNT")
	}

	reserves := self.reserves(cc)
	reserves[out] = Sub(reserves[out], Add(dy, MulDivCC(dy_fee, int64(self.adminFee(cc)), FEE_DENOMINATOR)))
	self.setReserves(cc, reserves)

	if err := self._burn(cc, cc.From(), _token_amount); err != nil { // insufficient funds
		return nil, err
	}

	// 추가 :  dy가 0인 경우 error를 발생시키지 않고 종료
	if dy.Cmp(Zero) == 0 {
		return dy, nil
	}

	if err := SafeTransfer(cc, self.tokens(cc)[out], cc.From(), dy); err != nil {
		return nil, err
	}
	return dy, nil
}
func (self *StableSwap) rampA(cc *types.ContractContext, _future_A *big.Int, _future_time uint64) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	lastTimeStamp := cc.LastTimestamp() / uint64(time.Second)

	if lastTimeStamp < (self.initialATime(cc) + MIN_RAMP_TIME) {
		return errors.New("Exchange: Ramp_A_SMALL")
	}

	if _future_time < (lastTimeStamp + MIN_RAMP_TIME) { // insufficient time
		return errors.New("Exchange: Ramp_A_BIG")
	}

	initial_A := self._A(cc)
	future_A_p := MulC(_future_A, int64(A_PRECISION))

	if !(_future_A.Cmp(Zero) > 0 && _future_A.Cmp(big.NewInt(int64(MAX_A))) < 0) {
		return errors.New("Exchange: FUTURE_A")
	}
	if future_A_p.Cmp(initial_A) < 0 {
		if MulC(future_A_p, MAX_A_CHANGE).Cmp(initial_A) < 0 {
			return errors.New("Exchange: FUTURE_A_MAX")
		}
	} else {
		if future_A_p.Cmp(MulC(initial_A, MAX_A_CHANGE)) > 0 {
			return errors.New("Exchange: FUTURE_A_CHANGE")
		}
	}

	self.setInitialA(cc, initial_A)
	self.setFutureA(cc, future_A_p)
	self.setInitialATime(cc, lastTimeStamp)
	self.setFutureATime(cc, _future_time)
	return nil
}
func (self *StableSwap) stopRampA(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}

	current_A := self._A(cc)
	self.setInitialA(cc, current_A)
	self.setFutureA(cc, current_A)
	self.setInitialATime(cc, cc.LastTimestamp()/uint64(time.Second))
	self.setFutureATime(cc, cc.LastTimestamp()/uint64(time.Second))

	// now (block.timestamp < t1) is always False, so we return saved A
	return nil
}
func (self *StableSwap) adminBalances(cc *types.ContractContext, idx uint8) (*big.Int, error) {
	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	if idx >= N {
		return nil, errors.New("Exchange: IDX")
	}

	_payToken := self.payToken(cc)
	if _payToken == ZeroAddress { // payToken  설정 안된 경우
		tokenBalance, err := TokenBalanceOf(cc, tokens[idx], self.Address())
		if err != nil {
			return nil, err
		}
		balance := Sub(tokenBalance, self.reserves(cc)[idx])
		return balance, nil
	}

	pi, err := self.payTokenIndex(cc)
	if err != nil {
		return nil, err
	}
	if pi != idx {
		return big.NewInt(0), nil
	}

	adminfeeTotal := big.NewInt(0)
	_reserves := self.reserves(cc)
	for k := uint8(0); k < N; k++ {
		tokenBalance, err := TokenBalanceOf(cc, tokens[k], self.Address())
		if err != nil {
			return nil, err
		}
		delta := Sub(tokenBalance, _reserves[k])
		if delta.Cmp(Zero) > 0 {
			if k == pi {
				adminfeeTotal = Add(adminfeeTotal, delta)
			} else {
				dy, _, admin_fee, err := self._get_dy_mem(cc, k, pi, delta, _reserves, self.owner(cc))
				if err != nil {
					return nil, err
				}
				// 살제 _exchange에서 reserve 변경 반영
				_reserves[k].Set(Add(_reserves[k], delta))
				_reserves[pi].Set(Sub(Sub(_reserves[pi], dy), admin_fee))
				adminfeeTotal = Add(adminfeeTotal, dy)
			}
		}
	}
	return adminfeeTotal, nil
}
func (self *StableSwap) withdrawAdminFees(cc *types.ContractContext) ([]*big.Int, []*big.Int, error) {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return nil, nil, err
	}

	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	payToken := self.payToken(cc)
	adminFees := MakeSlice(N)
	if payToken == ZeroAddress {
		reserves := self.reserves(cc)
		for k := 0; k < int(N); k++ {
			tokenBalance, err := TokenBalanceOf(cc, tokens[k], self.Address())
			if err != nil {
				return nil, nil, err
			}
			adminFees[k].Set(Sub(tokenBalance, reserves[k]))
		}
	} else {
		feeTotal := big.NewInt(0)
		pi, err := self.payTokenIndex(cc)
		if err != nil {
			return nil, nil, err
		}
		for k := uint8(0); k < N; k++ {
			reserves := self.reserves(cc) // _exchange에서 reserve 변경
			tokenBalance, err := TokenBalanceOf(cc, tokens[k], self.Address())
			if err != nil {
				return nil, nil, err
			}
			balance := Sub(tokenBalance, reserves[k])
			if balance.Cmp(Zero) > 0 {
				if k == pi {
					feeTotal = Add(feeTotal, balance)
				} else {
					dy, err := self._exchange(cc, k, pi, balance, big.NewInt(0), cc.From()) // no transfer, from = cc.From() or ZeroAddress Not Concern
					if err != nil {
						return nil, nil, err
					}
					feeTotal = Add(feeTotal, dy)
				}
			}
		}
		adminFees[pi].Set(feeTotal)
	}

	ownnerFees, winnerFees, err := self._divideFee(cc, adminFees) // transfer
	if err != nil {
		return nil, nil, err
	}
	return ownnerFees, winnerFees, nil
}
func (self *StableSwap) donateAdminFees(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}

	N, balances := self.makeSlice(cc)
	tokens := self.tokens(cc)

	for k := 0; k < int(N); k++ {
		tokenBalance, err := TokenBalanceOf(cc, tokens[k], self.Address())
		if err != nil {
			return err
		}
		balances[k].Set(tokenBalance)
	}
	self.setReserves(cc, balances)
	return nil
}
