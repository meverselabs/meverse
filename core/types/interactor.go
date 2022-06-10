package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
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

var bigIntType = reflect.TypeOf(&big.Int{}).String()
var amountType = reflect.TypeOf(&amount.Amount{}).String()

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
	cont, err := i.getContract(ContAddr)
	if err != nil {
		return nil, err
	}

	/*
		SetOwner(cc *types.ContractContext, NewOwner common.Address) error {
		Update(cc *types.ContractContext, contract []byte) error {
		ContractInvoke(cc *types.ContractContext, method string, params []interface{}) ([]interface{}, error) {
		IsUpdateable(cc *types.ContractContext) bool {
	*/

	var is interface{} = cont
	if _, ok := is.(InvokeableContract); ok &&
		(MethodName != "InitContract" &&
			MethodName != "IsUpdateable" &&
			MethodName != "Update" &&
			MethodName != "ContractInvoke" &&
			MethodName != "SetOwner") {
		Args = []interface{}{
			MethodName,
			Args,
		}
		MethodName = "ContractInvoke"
	} else {
		MethodName = strings.ToUpper(string(MethodName[0])) + MethodName[1:]
	}

	if i.saveEvent {
		en = i.addCallEvent(ecc, ContAddr, MethodName, Args)
	}

	rMethod, err := i.methodByName(cont, ContAddr, MethodName)
	if err != nil {
		return nil, err
	}

	// in := make([]reflect.Value, rMethod.Type().NumIn())
	// in[0] = reflect.ValueOf(ecc)
	//fmt.Printf("parm.Type() in Exec: %v", param.Type())
	//fmt.Printf("mType in Exec : %v", mType)
	in, err := ContractInputsConv(Args, rMethod)
	if err != nil {
		return nil, err
	}
	in = append([]reflect.Value{reflect.ValueOf(ecc)}, in...)

	sn := ecc.ctx.Snapshot()

	vs, err := func() (vs []reflect.Value, err error) {
		defer func() {
			v := recover()
			if v != nil {
				fmt.Println(v)
				if MethodName == "ContractInvoke" && len(Args) > 0 {
					MethodName = fmt.Sprintf("ci %v", Args[0])
				}
				err = fmt.Errorf("occur error call method(%v) of contract(%v) message: %v", MethodName, ContAddr.String(), v)
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

func ContractInputsConv(Args []interface{}, rMethod reflect.Value) ([]reflect.Value, error) {
	if rMethod.Type().NumIn() != len(Args)+1 {
		return nil, errors.Errorf("invalid inputs count got %v want %v", len(Args), rMethod.Type().NumIn()-1)
	}
	if rMethod.Type().NumIn() < 1 {
		return nil, errors.New("not found")
	}
	in := make([]reflect.Value, len(Args))
	for i, v := range Args {
		param := reflect.ValueOf(v)
		mType := rMethod.Type().In(i + 1)

		if param.Type() != mType {
			if param.Kind() == reflect.Array && mType.Kind() == reflect.Slice {
				l := param.Type().Len()
				bs := []byte{}
				for i := 0; i < l; i++ {
					rvc := param.Index(i)
					if !rvc.CanInterface() {
						return nil, errors.Errorf("array type only support bytes. index(%v) get not interface %v want %v", i, rvc.Type(), mType)
					}
					val := rvc.Interface()
					b, ok := val.(byte)
					if !ok {
						return nil, errors.Errorf("array type only support bytes. index(%v) get not byte %v want %v", i, val, mType)
					}
					bs = append(bs, b)
				}
				param = reflect.ValueOf(bs)
				if param.Type() != mType {
					return nil, errors.Errorf("array type only support bytes. index(%v) get %v want %v", i, param.Type(), mType)
				}
			} else {
				switch param.Type().String() {
				case bigIntType:
					switch mType.String() {
					case amountType:
						param = reflect.ValueOf(amount.NewAmountFromBytes(v.(*big.Int).Bytes()))
					case reflect.Uint.String():
						param = reflect.ValueOf(uint(v.(*big.Int).Uint64()))
					case reflect.Uint8.String():
						param = reflect.ValueOf(uint8(v.(*big.Int).Uint64()))
					case reflect.Uint16.String():
						param = reflect.ValueOf(uint16(v.(*big.Int).Uint64()))
					case reflect.Uint32.String():
						param = reflect.ValueOf(uint32(v.(*big.Int).Uint64()))
					case reflect.Uint64.String():
						param = reflect.ValueOf(v.(*big.Int).Uint64())
					case reflect.Int.String():
						param = reflect.ValueOf(int(v.(*big.Int).Int64()))
					case reflect.Int8.String():
						param = reflect.ValueOf(int8(v.(*big.Int).Int64()))
					case reflect.Int16.String():
						param = reflect.ValueOf(int16(v.(*big.Int).Int64()))
					case reflect.Int32.String():
						param = reflect.ValueOf(int32(v.(*big.Int).Int64()))
					case reflect.Int64.String():
						param = reflect.ValueOf(v.(*big.Int).Int64())
					}
				case amountType:
					switch mType.String() {
					case bigIntType:
						param = reflect.ValueOf(big.NewInt(0).SetBytes(v.(*amount.Amount).Bytes()))
					case "string":
						param = reflect.ValueOf(big.NewInt(0).SetBytes(v.(*amount.Amount).Bytes()).String())
					}
				case "[]interface {}":
					if tv, ok := v.([]interface{}); ok {
						switch mType.String() {
						case "[]common.Address":
							as := []common.Address{}
							for _, t := range tv {
								addr, ok := t.(common.Address)
								if !ok {
									return nil, errors.Errorf("invalid input addr type(%v) get %v want %v(%v)", i, param.Type(), mType, mType.String())
								}
								as = append(as, addr)
							}
							param = reflect.ValueOf(as)
						case "[]string":
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
						case "[]*amount.Amount":
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
						}
					}
				case "[]*big.Int":
					if tv, ok := v.([]*big.Int); ok {
						switch mType.String() {
						case "[]*amount.Amount":
							as := []*amount.Amount{}
							for _, t := range tv {
								as = append(as, amount.NewAmountFromBytes(t.Bytes()))
							}
							param = reflect.ValueOf(as)
						}
					}
				case "string":
					if tv, ok := v.(string); ok {
						switch mType.String() {
						case "bool":
							if strings.ToLower(tv) == "true" {
								param = reflect.ValueOf(true)
							} else {
								param = reflect.ValueOf(false)
							}
						case "common.Hash":
							param = reflect.ValueOf(hash.HexToHash(tv))
						case "common.Address":
							param = reflect.ValueOf(common.HexToAddress(tv))
						case "*amount.Amount":
							am, err := amount.ParseAmount(tv)
							if err == nil {
								param = reflect.ValueOf(am)
							} else {
								tv2 := strings.Replace(tv, "0x", "", -1)
								if len(tv2)%2 == 1 {
									tv2 = "0" + tv2
								}
								var bs []byte
								if bs, err = hex.DecodeString(tv2); err == nil {
									am = amount.NewAmountFromBytes(bs)
									param = reflect.ValueOf(am)
								}
							}
						case "[]byte", "[]uint8":
							bs, err := hex.DecodeString(tv)
							if err == nil {
								param = reflect.ValueOf(bs)
							}
						default:
							if bi, ok := big.NewInt(0).SetString(tv, 10); ok {
								switch mType.String() {
								case "*big.Int":
									param = reflect.ValueOf(bi)
								case "int":
									param = reflect.ValueOf(int(bi.Int64()))
								case "int8":
									param = reflect.ValueOf(int8(bi.Int64()))
								case "int16":
									param = reflect.ValueOf(int16(bi.Int64()))
								case "int32":
									param = reflect.ValueOf(int32(bi.Int64()))
								case "int64":
									param = reflect.ValueOf(int64(bi.Int64()))
								case "uint":
									param = reflect.ValueOf(uint(bi.Uint64()))
								case "uint8":
									param = reflect.ValueOf(uint8(bi.Uint64()))
								case "uint16":
									param = reflect.ValueOf(uint16(bi.Uint64()))
								case "uint32":
									param = reflect.ValueOf(uint32(bi.Uint64()))
								case "uint64":
									param = reflect.ValueOf(uint64(bi.Uint64()))
								}
							}
						}
					}
				case "[]byte", "[]uint8":
					if bs, ok := v.([]byte); ok {
						switch mType.String() {
						case "common.Hash":
							h := hash.Hash256{}
							copy(h[:], bs)
							param = reflect.ValueOf(h)
						case "common.Address":
							param = reflect.ValueOf(common.BytesToAddress(bs))
						case "*amount.Amount":
							param = reflect.ValueOf(amount.NewAmountFromBytes(bs))
						case "*big.Int":
							param = reflect.ValueOf(big.NewInt(0).SetBytes(bs))
						}
					}
				default:
				}
			}
		}
		if param.Type() != mType {
			cy := param.Type().String() + mType.String()
			return nil, errors.Errorf("invalid input view(%v) type(%v) get %v want %v(%v)", cy, i, param.Type(), mType, mType.String())
		}

		in[i] = param
	}
	return in, nil
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

func (i *interactor) methodByName(cont Contract, Addr common.Address, MethodName string) (reflect.Value, error) {
	_cont := cont.Front()

	method, err := contractMethod(_cont, Addr, MethodName)
	if err != nil {
		return reflect.Value{}, err
	}
	return method, nil
}

func (i *interactor) getContract(Addr common.Address) (Contract, error) {
	var cont Contract
	if _cont, ok := i.conMap[Addr]; ok {
		cont = _cont
	} else {
		_cont, err := i.ctx.Contract(Addr)
		if err != nil {
			return nil, err
		}
		i.conMap[Addr] = _cont
		cont = _cont
	}
	return cont, nil
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
