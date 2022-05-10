package test

import (
	"errors"
	"math/big"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	. "github.com/meverselabs/meverse/contract/exchange/util"
)

type Curve struct {
	A      int64
	n      uint8
	fee    *big.Int
	p      []*big.Int
	x      []*big.Int
	tokens *big.Int //token_supply
}

/// Python model of Curve pool math.
/// curve-contract/simulation.py

/*
A: Amplification coefficient
D: Total deposit size
n: number of currencies
p: target prices
*/
func (self *Curve) init(A int64, D []*big.Int, n uint8, p []*big.Int, tokens *big.Int) {

	self.A = A // actually A * n ** (n - 1) because it's an invariant
	self.n = n
	//bN = big.NewInt(int64(n))
	self.fee = big.NewInt(10000000)
	if p != nil {
		self.p = CloneSlice(p)
	} else {
		self.p = MakeSlice(n)
		for k := uint8(0); k < n; k++ {
			self.p[k].Set(big.NewInt(trade.PRECISION))
		}
	}

	if len(D) == int(n) {
		self.x = CloneSlice(D)
	} else {
		self.x = MakeSlice(n)
		for k := uint8(0); k < n; k++ {
			self.x[k].Set(Div(MulDivCC(D[0], trade.PRECISION, int64(n)), self.p[k]))
		}
	}
	self.tokens = tokens
}

func (self *Curve) xp() []*big.Int {
	result := MakeSlice(self.n)
	for k := uint8(0); k < self.n; k++ {
		result[k].Set(MulDivC(self.x[k], self.p[k], trade.PRECISION))
	}
	return result
}

/*
D invariant calculation in non-overflowing integer operations
iteratively

A * Sum(x_i) * n**n + D = A * D * n**n + D**(n+1) / (n**n * prod(x_i))

Converging solution:
D[j+1] = (A * n**n * Sum(x_i) - D[j]**(n+1) / (n**n prod(x_i))) / (A * n**n - 1)
*/
func (self *Curve) D() *big.Int {
	Dprev := big.NewInt(0)
	xp := self.xp()
	S := Sum(xp)
	D := Clone(S)
	Ann := big.NewInt(self.A * int64(self.n))
	for Abs(Sub(D, Dprev)).Cmp(big.NewInt(1)) > 0 {
		D_P := Clone(D)
		for k := uint8(0); k < self.n; k++ {
			D_P = MulDiv(D_P, D, MulC(xp[k], int64(self.n)))
		}
		Dprev.Set(D)
		D.Set(MulDiv(Add(Mul(Ann, S), MulC(D_P, int64(self.n))), D, Add(Mul(SubC(Ann, 1), D), MulC(D_P, int64(self.n+1)))))
	}

	return D
}

/*
Calculate x[j] if one makes x[i] = x

Done by solving quadratic equation iteratively.
x_1**2 + x1 * (Sum' - (A*n**n - 1) * D / (A * n**n)) = D ** (n+1)/(n ** (2 * n) * prod' * A)
x_1**2 + b*x_1 = c

x_1 = (x_1**2 + c) / (2*x_1 + b)
*/
func (self *Curve) y(i, j uint8, x *big.Int) *big.Int {
	D := self.D()
	xx := self.xp()
	xx[i].Set(x) // x is quantity of underlying asset brought to 1e18 precision
	xx[j].Set(big.NewInt(0))
	Ann := big.NewInt(self.A * int64(self.n))
	c := Clone(D)
	for k := uint8(0); k < self.n; k++ {
		if k == j {
			continue
		}
		c = MulDiv(c, D, MulC(xx[k], int64(self.n)))
	}
	c = MulDiv(c, D, MulC(Ann, int64(self.n)))

	b := Sub(Add(Sum(xx), Div(D, Ann)), D)

	y_prev := big.NewInt(0)
	y := Clone(D)
	for Abs(Sub(y, y_prev)).Cmp(big.NewInt(1)) > 0 {
		y_prev.Set(y)
		y = Div(Add(Mul(y, y), c), Add(Mul(big.NewInt(2), y), b))
	}
	return y // the result is in underlying units too
}

/*
Calculate x[j] if one makes x[i] = x

Done by solving quadratic equation iteratively.
x_1**2 + x1 * (Sum' - (A*n**n - 1) * D / (A * n**n)) = D ** (n+1)/(n ** (2 * n) * prod' * A)
x_1**2 + b*x_1 = c

x_1 = (x_1**2 + c) / (2*x_1 + b)
*/
func (self *Curve) y_D(i uint8, _D *big.Int) *big.Int {

	xx := self.xp()
	xx[i].Set(big.NewInt(0))
	S := Sum(xx)
	Ann := big.NewInt(self.A * int64(self.n))
	c := Clone(_D)
	for k := uint8(0); k < self.n; k++ {
		if k == i {
			continue
		}
		c = MulDiv(c, _D, MulC(xx[k], int64(self.n)))
	}
	c = MulDiv(c, _D, MulC(Ann, int64(self.n)))
	b := Add(S, Div(_D, Ann))
	y_prev := big.NewInt(0)
	y := Clone(_D)
	for Abs(Sub(y, y_prev)).Cmp(big.NewInt(1)) > 0 {
		y_prev.Set(y)
		y.Set(Div(Add(Mul(y, y), c), Sub(Add(Mul(big.NewInt(2), y), b), _D)))
	}
	return y // the result is in underlying units too
}

func (self *Curve) dy(i, j uint8, dx *big.Int) *big.Int {
	//dx and dy are in underlying units
	xp := self.xp()
	y := self.y(i, j, Add(xp[i], dx))
	return Sub(xp[j], y)
}

func (self *Curve) exchange(i, j uint8, dx *big.Int) (*big.Int, error) {
	xp := self.xp()
	x := Add(xp[i], dx)
	y := self.y(i, j, x)
	dy := Sub(xp[j], y)
	if !(dy.Cmp(Zero) > 0) {
		return nil, errors.New("Curve: EXCHANGE_DY_NOT_POSITIVE")
	}
	fee := MulDivC(dy, self.fee, trade.FEE_DENOMINATOR)

	self.x[i].Set(MulDiv(x, big.NewInt(trade.PRECISION), self.p[i]))
	self.x[j].Set(MulDiv(Add(y, fee), big.NewInt(trade.PRECISION), self.p[j]))
	return Sub(dy, fee), nil
}

func (self *Curve) remove_liquidity_imbalance(amounts []*big.Int) *big.Int {

	_fee := MulDivCC(self.fee, int64(self.n), int64(4*(self.n-1)))

	old_balances := CloneSlice(self.x)
	new_balances := CloneSlice(self.x)
	D0 := self.D()
	for k := uint8(0); k < self.n; k++ {
		new_balances[k] = Sub(new_balances[k], amounts[k])
	}
	self.x = CloneSlice(new_balances)
	D1 := self.D()
	self.x = old_balances
	fees := MakeSlice(self.n)
	for k := uint8(0); k < self.n; k++ {
		ideal_balance := MulDiv(D1, old_balances[k], D0)
		difference := Abs(Sub(ideal_balance, new_balances[k]))
		fees[k].Set(MulDivC(_fee, difference, trade.FEE_DENOMINATOR))
		new_balances[k] = Sub(new_balances[k], fees[k])
	}
	self.x = CloneSlice(new_balances)
	D2 := self.D()
	self.x = CloneSlice(old_balances)

	return MulDiv(Sub(D0, D2), self.tokens, D0)
}

func (self *Curve) calc_withdraw_one_coin(token_amount *big.Int, i uint8) *big.Int {

	xp := self.xp()
	var fee *big.Int
	if self.fee.Cmp(Zero) != 0 {
		fee = Add(Sub(self.fee, MulDiv(self.fee, xp[i], Sum(xp))), big.NewInt(5*100000))
	} else {
		fee = big.NewInt(0)
	}

	D0 := self.D()
	D1 := Sub(D0, MulDiv(token_amount, D0, self.tokens))
	dy := Sub(xp[i], self.y_D(i, D1))

	return Sub(dy, MulDivC(dy, fee, trade.FEE_DENOMINATOR))
}
