package types

import (
	"math/big"
	"reflect"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/pkg/errors"
)

type IInteractor interface {
	Distroy()
	Exec(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) ([]interface{}, error)
}

type ExecFunc = func(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) ([]interface{}, error)

type interactor struct {
	ctx    *Context
	cont   Contract
	ccMap  map[common.Address]*ContractContext
	conMap map[common.Address]Contract
	exit   bool
}

var bigIntType = reflect.TypeOf(&big.Int{})
var amountType = reflect.TypeOf(&amount.Amount{})

func NewInteractor(ctx *Context, cont Contract, cc *ContractContext) IInteractor {
	return &interactor{
		ctx:  ctx,
		cont: cont,
		ccMap: map[common.Address]*ContractContext{
			cc.cont: cc,
		},
		conMap: map[common.Address]Contract{},
	}
}

func (i *interactor) Distroy() {
	i.exit = true
}

func (i *interactor) Exec(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) ([]interface{}, error) {
	if i.exit {
		return nil, errors.New("expired")
	}
	if MethodName == "" {
		return nil, errors.New("method not given")
	}

	ecc := i.currentContractContext(Cc, Addr)
	rMethod, err := i.methodByName(Addr, MethodName)
	if err != nil {
		return nil, err
	}

	if rMethod.Type().NumIn() != len(Args)+1 {
		return nil, errors.Errorf("invalid inputs count got %v want %v", len(Args), rMethod.Type().NumIn()-1)
	}
	if rMethod.Type().NumIn() < 1 {
		return nil, errors.New("not found")
	}
	in := make([]reflect.Value, rMethod.Type().NumIn())
	in[0] = reflect.ValueOf(ecc)
	for i, v := range Args {
		param := reflect.ValueOf(v)
		mType := rMethod.Type().In(i + 1)

		if param.Type() != mType {
			switch param.Type().String() + mType.String() {
			case bigIntType.String() + amountType.String():
				param = reflect.ValueOf(amount.NewAmountFromBytes(v.(*big.Int).Bytes()))
			case amountType.String() + bigIntType.String():
				param = reflect.ValueOf(big.NewInt(0).SetBytes(v.(*amount.Amount).Bytes()))
			default:
				return nil, errors.Errorf("invalid input type get %v want %v", param.Type(), mType)
			}
		}
		in[i+1] = param
	}

	sn := ecc.ctx.Snapshot()
	vs := rMethod.Call(in)
	params := []interface{}{}
	for _, v := range vs {
		params = append(params, v.Interface())
		if v.Interface() != nil {
			if _err, ok := v.Interface().(error); ok {
				err = _err
			}
		}
	}
	if err != nil {
		ecc.ctx.Revert(sn)
	}
	ecc.ctx.Commit(sn)
	return params, err
}

func (i *interactor) methodByName(Addr common.Address, MethodName string) (reflect.Value, error) {
	var cont Contract
	if _cont, ok := i.conMap[Addr]; ok {
		cont = _cont
	} else {
		_cont, err := i.ctx.Contract(Addr)
		if err != nil {
			return reflect.Value{}, err
		}
		i.conMap[Addr] = _cont
		cont = _cont
	}
	_cont := cont.Front()

	method, err := contractMethod(_cont, MethodName)
	if err != nil {
		return reflect.Value{}, err
	}
	return method, nil
}

func contractMethod(cont interface{}, MethodName string) (reflect.Value, error) {
	vo := reflect.ValueOf(cont)
	if !vo.IsValid() {
		return reflect.Value{}, errors.New("wrong contract")
	}
	if vo.IsNil() {
		return reflect.Value{}, errors.New("nil contract")
	}

	method := vo.MethodByName(MethodName)
	if !method.IsValid() || method.IsNil() {
		return reflect.Value{}, errors.New("method not exist: " + MethodName)
	}
	return method, nil
}

func (i *interactor) currentContractContext(Cc *ContractContext, Addr common.Address) *ContractContext {
	if i.cont.Address() == Addr {
		return Cc
	}
	var ecc *ContractContext
	if _cc, ok := i.ccMap[Addr]; ok {
		ecc = _cc
	} else {
		ecc = &ContractContext{
			cont: Addr,
			from: Cc.cont,
			ctx:  Cc.ctx,
			Exec: i.Exec,
		}
		i.ccMap[Addr] = ecc
	}
	return ecc
}
