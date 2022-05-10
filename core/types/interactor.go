package types

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"reflect"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/pkg/errors"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

type IInteractor interface {
	Distroy()
	Exec(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) ([]interface{}, error)
	EventList() []*Event
}

type ExecFunc = func(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) ([]interface{}, error)

type interactor struct {
	ctx       *Context
	cont      Contract
	conMap    map[common.Address]Contract
	exit      bool
	index     uint16
	eventList []*Event
	saveEvent bool
}

var bigIntType = reflect.TypeOf(&big.Int{})
var amountType = reflect.TypeOf(&amount.Amount{})

func NewInteractor(ctx *Context, cont Contract, cc *ContractContext, TXID string, saveEvent bool) IInteractor {
	_, i, _ := ParseTransactionID(TXID)
	return &interactor{
		ctx:       ctx,
		cont:      cont,
		conMap:    map[common.Address]Contract{},
		index:     i,
		eventList: []*Event{},
		saveEvent: saveEvent,
	}
}

func (i *interactor) Distroy() {
	i.exit = true
}

func (i *interactor) Exec(Cc *ContractContext, ContAddr common.Address, MethodName string, Args []interface{}) ([]interface{}, error) {
	if i.exit {
		return nil, errors.New("expired")
	}
	if MethodName == "" {
		return nil, errors.New("method not given")
	}

	ecc := i.currentContractContext(Cc, ContAddr)

	var en *Event
	if i.saveEvent {
		en = i.addCallEvent(ecc, ContAddr, MethodName, Args)
	}

	rMethod, err := i.methodByName(ContAddr, MethodName)
	if err != nil {
		return nil, err
	}

	if rMethod.Type().NumIn() != len(Args)+1 {
		log.Println("MethodName in Exec", MethodName)
		log.Println("Args in Exec", Args)
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
			if param.Kind() == reflect.Array && mType.Kind() == reflect.Slice {
				l := param.Type().Len()
				bs := []byte{}
				for i := 0; i < l; i++ {
					val := param.Index(i).Interface()
					b := val.(byte)
					bs = append(bs, b)
				}
				param = reflect.ValueOf(bs)
				if param.Type() != mType {
					return nil, errors.Errorf("invalid input type(%v) get %v want %v", i, param.Type(), mType)
				}
			} else {
				//fmt.Printf("parm.Type() in Exec: %v", param.Type())
				//fmt.Printf("mType in Exec : %v", mType)
				switch param.Type().String() + mType.String() {
				case bigIntType.String() + amountType.String():
					param = reflect.ValueOf(amount.NewAmountFromBytes(v.(*big.Int).Bytes()))
				case amountType.String() + bigIntType.String():
					param = reflect.ValueOf(big.NewInt(0).SetBytes(v.(*amount.Amount).Bytes()))
				case bigIntType.String() + reflect.Uint.String():
					param = reflect.ValueOf(uint(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint8.String():
					param = reflect.ValueOf(uint8(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint16.String():
					param = reflect.ValueOf(uint16(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint32.String():
					param = reflect.ValueOf(uint32(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint64.String():
					param = reflect.ValueOf(v.(*big.Int).Uint64())

				case bigIntType.String() + reflect.Int.String():
					param = reflect.ValueOf(int(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int8.String():
					param = reflect.ValueOf(int8(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int16.String():
					param = reflect.ValueOf(int16(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int32.String():
					param = reflect.ValueOf(int32(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int64.String():
					param = reflect.ValueOf(v.(*big.Int).Int64())
				case "[]interface {}[]common.Address":
					if tv, ok := v.([]interface{}); ok {
						as := []common.Address{}
						for _, t := range tv {
							addr, ok := t.(common.Address)
							if !ok {
								return nil, errors.Errorf("invalid input addr type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
							}
							as = append(as, addr)
						}
						param = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
					}
				case "[]interface {}[]string":
					if tv, ok := v.([]interface{}); ok {
						as := []string{}
						for _, t := range tv {
							addr, ok := t.(string)
							if !ok {
								if str, ok := t.(fmt.Stringer); !ok {
									trfv := reflect.ValueOf(t)
									return nil, errors.Errorf("invalid input addr type get %v(%v, %v) want string", t, trfv.Type().String(), trfv.Kind().String())
								} else {
									addr = str.String()
								}
							}
							as = append(as, addr)
						}
						param = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
					}
				case "[]interface {}[]*amount.Amount":
					if tv, ok := v.([]interface{}); ok {
						as := []*amount.Amount{}
						for _, t := range tv {
							addr, ok := t.(*amount.Amount)
							if !ok {
								trfv := reflect.ValueOf(t)
								return nil, errors.Errorf("invalid input addr type get %v(%v, %v) want *amount.Amount", t, trfv.Type().String(), trfv.Kind().String())
							}
							as = append(as, addr)
						}
						param = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
					}
				case "[]*big.Int[]*amount.Amount":
					if tv, ok := v.([]*big.Int); ok {
						as := []*amount.Amount{}
						for _, t := range tv {
							as = append(as, amount.NewAmountFromBytes(t.Bytes()))
						}
						param = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
					}

				default:
					cy := param.Type().String() + mType.String()
					return nil, errors.Errorf("invalid input (%v) type(%v) get %v want %v(%v)", cy, i, param.Type(), mType, mType.String())
				}
			}
		}
		in[i+1] = param
	}

	sn := ecc.ctx.Snapshot()

	vs, err := func() (vs []reflect.Value, err error) {
		defer func() {
			v := recover()
			if v != nil {
				fmt.Println(v)
				err = errors.New("occur error call method(" + MethodName + ") of contract(" + ContAddr.String() + ") message: " + fmt.Sprintf("%v", v))
			}
		}()
		return rMethod.Call(in), nil
	}()
	if err != nil {
		return nil, err
	}

	mtype := rMethod.Type()
	params, resultsWithoutError, err := i.getResults(mtype, vs)
	if err != nil {
		ecc.ctx.Revert(sn)
	}
	ecc.ctx.Commit(sn)

	if i.saveEvent {
		_err := i.insertResultEvent(en, resultsWithoutError, err)
		if _err != nil {
			return nil, _err
		}
	}

	return params, err
}

func (i *interactor) EventList() []*Event {
	return i.eventList
}

func (i *interactor) getResults(mType reflect.Type, vs []reflect.Value) (params []interface{}, result []interface{}, err error) {
	params = []interface{}{}
	result = []interface{}{}
	for i, v := range vs {
		vi := v.Interface()
		params = append(params, v.Interface())
		if mType.Out(i).Kind() == reflect.Interface && mType.Out(i).Implements(errType) { // is error type
			if _err, ok := vi.(error); ok {
				err = _err
			}
			continue
		}
		result = append(result, vi)
	}
	return
}

func (i *interactor) addCallEvent(Cc *ContractContext, Addr common.Address, MethodName string, Args []interface{}) *Event {
	mc := MethodCallEvent{
		From:   Cc.From(),
		To:     Addr,
		Method: MethodName,
		Args:   Args,
	}
	bf := &bytes.Buffer{}
	_, err := mc.WriteTo(bf)
	if err != nil {
		panic(err)
	}
	rv := &Event{
		Index:  i.index,
		Type:   EventTagCallHistory,
		Result: bf.Bytes(),
	}
	i.eventList = append(i.eventList, rv)
	return rv
}

func (i *interactor) insertResultEvent(en *Event, Results []interface{}, Err error) error {
	bf := bytes.NewBuffer(en.Result)

	mc := &MethodCallEvent{}
	_, err := mc.ReadFrom(bf)
	if err != nil {
		return err
	}

	if Err != nil {
		mc.Error = Err.Error()
	} else {
		mc.Result = Results
	}

	wbf := &bytes.Buffer{}
	_, err = mc.WriteTo(wbf)
	if err != nil {
		panic(err)
	}
	en.Result = wbf.Bytes()
	return err
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
	rt := reflect.TypeOf(cont)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	_cont := cont.Front()

	method, err := contractMethod(_cont, Addr, MethodName)
	if err != nil {
		return reflect.Value{}, err
	}
	return method, nil
}

func contractMethod(cont interface{}, addr common.Address, MethodName string) (reflect.Value, error) {
	vo := reflect.ValueOf(cont)
	if !vo.IsValid() {
		return reflect.Value{}, errors.New("wrong contract")
	}
	if vo.IsNil() {
		return reflect.Value{}, errors.New("nil contract")
	}

	method := vo.MethodByName(MethodName)
	if !method.IsValid() || method.IsNil() {
		return reflect.Value{}, errors.New("method not exist: " + MethodName + " cont" + addr.String())
	}
	return method, nil
}

func (i *interactor) currentContractContext(Cc *ContractContext, Addr common.Address) *ContractContext {
	if i.cont.Address() == Addr && Cc.cont == Addr {
		return Cc
	}
	return &ContractContext{
		cont: Addr,
		from: Cc.cont,
		ctx:  Cc.ctx,
		Exec: i.Exec,
	}
}
