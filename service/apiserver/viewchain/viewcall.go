package viewchain

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

var bigIntType = reflect.TypeOf(&big.Int{})
var amountType = reflect.TypeOf(&amount.Amount{})

type ViewCaller struct {
	cn *chain.Chain
}

func NewViewCaller(cn *chain.Chain) *ViewCaller {
	return &ViewCaller{
		cn: cn,
	}
}

var errType = reflect.TypeOf((*error)(nil)).Elem()

func (m *ViewCaller) getContract(conAddr common.Address, from string) (interface{}, *types.ContractContext, error) {
	ctx := m.cn.NewContext()
	cont, err := ctx.Contract(conAddr)
	if err != nil {
		return nil, nil, err
	}
	var cc *types.ContractContext
	if from == "" {
		cc = ctx.ContractContext(cont, cont.Address())
	} else {
		cc = ctx.ContractContext(cont, common.HexToAddress(from))
	}
	intr := types.NewInteractor(ctx, cont, cc, "000000000000", false)
	cc.Exec = intr.Exec
	return cont.Front(), cc, nil
}

func (m *ViewCaller) Execute(addr common.Address, from, method string, inputs []interface{}) ([]interface{}, error) {
	contract, cc, err := m.getCallContract(addr, from)
	if err != nil {
		return nil, err
	}
	return m._execute(contract, cc, method, inputs)
}

func (m *ViewCaller) MultiExecute(addr common.Address, from string, methods []string, inputss [][]interface{}) ([][]interface{}, error) {
	contract, cc, err := m.getCallContract(addr, from)
	if err != nil {
		return nil, err
	}

	if len(methods) != len(inputss) {
		return nil, errors.New("not match params count")
	}

	result := make([][]interface{}, len(methods))
	for i, method := range methods {
		r, err := m._execute(contract, cc, method, inputss[i])
		if err != nil {
			return nil, err
		}
		result[i] = r
	}
	return result, nil
}

func (m *ViewCaller) getCallContract(addr common.Address, from string) (contract reflect.Value, cc *types.ContractContext, err error) {
	var cont interface{}
	cont, cc, err = m.getContract(addr, from)
	if err != nil {
		return
	}
	contract = reflect.ValueOf(cont)
	if contract.IsNil() {
		err = errors.New("not exist contract")
		return
	}
	return
}

func (m *ViewCaller) _execute(contract reflect.Value, cc *types.ContractContext, method string, inputs []interface{}) ([]interface{}, error) {
	for _, v := range inputs {
		if v == nil {
			return nil, errors.New("nil params")
		}
	}
	if len(method) < 1 {
		return nil, errors.New("invalid method name")
	}
	method = strings.ToUpper(string(method[0])) + method[1:]

	rMethod := contract.MethodByName(method)
	if rMethod.Kind() == reflect.Invalid {
		return nil, errors.New("not exist method")
	}

	mtype := rMethod.Type()
	if mtype.NumIn() != len(inputs)+1 {
		return nil, errors.Errorf("invalid inputs count got %v want %v", len(inputs), mtype.NumIn()-1)
	}
	if mtype.NumIn() < 1 {
		return nil, errors.New("not found")
	}
	in := make([]reflect.Value, mtype.NumIn())
	in[0] = reflect.ValueOf(cc)
	for i, v := range inputs {
		rv := reflect.ValueOf(v)
		mType := mtype.In(i + 1)
		_v, ok := m.parse(rv.Type().String()+mType.String(), v)
		if ok {
			rv = reflect.ValueOf(_v)
		}
		if rv.Type() != mType {
			if rv.Kind() == reflect.Array && mType.Kind() == reflect.Slice {
				l := rv.Type().Len()
				bs := []byte{}
				for i := 0; i < l; i++ {
					val := rv.Index(i).Interface()
					b := val.(byte)
					bs = append(bs, b)
				}
				rv = reflect.ValueOf(bs)
			} else {
				//fmt.Printf("parm.Type() in Exec: %v", rv.Type())
				//fmt.Printf("mType in Exec : %v", mType)
				switch rv.Type().String() + mType.String() {
				case bigIntType.String() + amountType.String():
					rv = reflect.ValueOf(amount.NewAmountFromBytes(v.(*big.Int).Bytes()))
				case amountType.String() + bigIntType.String():
					rv = reflect.ValueOf(big.NewInt(0).SetBytes(v.(*amount.Amount).Bytes()))
				case bigIntType.String() + reflect.Uint.String():
					rv = reflect.ValueOf(uint(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint8.String():
					rv = reflect.ValueOf(uint8(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint16.String():
					rv = reflect.ValueOf(uint16(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint32.String():
					rv = reflect.ValueOf(uint32(v.(*big.Int).Uint64()))
				case bigIntType.String() + reflect.Uint64.String():
					rv = reflect.ValueOf(v.(*big.Int).Uint64())

				case bigIntType.String() + reflect.Int.String():
					rv = reflect.ValueOf(int(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int8.String():
					rv = reflect.ValueOf(int8(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int16.String():
					rv = reflect.ValueOf(int16(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int32.String():
					rv = reflect.ValueOf(int32(v.(*big.Int).Int64()))
				case bigIntType.String() + reflect.Int64.String():
					rv = reflect.ValueOf(v.(*big.Int).Int64())
				case "[]interface {}[]common.Address":
					if tv, ok := v.([]interface{}); ok {
						as := []common.Address{}
						for _, t := range tv {
							addr, ok := t.(common.Address)
							if !ok {
								return nil, errors.Errorf("invalid input addr type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
							}
							as = append(as, addr)
						}
						rv = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
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
						rv = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
					}
				case "[]interface {}[]*amount.Amount":
					if tv, ok := v.([]interface{}); ok {
						as := []*amount.Amount{}
						for _, t := range tv {
							addr, ok := t.(*amount.Amount)
							if !ok {
								return nil, errors.Errorf("invalid input addr type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
							}
							as = append(as, addr)
						}
						rv = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
					}
				case "[]*big.Int[]*amount.Amount":
					if tv, ok := v.([]*big.Int); ok {
						as := []*amount.Amount{}
						for _, t := range tv {
							as = append(as, amount.NewAmountFromBytes(t.Bytes()))
						}
						rv = reflect.ValueOf(as)
					} else {
						return nil, errors.Errorf("invalid input addrs type(%v) get %v want %v(%v)", i, rv.Type(), mType, mType.String())
					}
				default:
					cy := rv.Type().String() + mType.String()
					return nil, errors.Errorf("invalid input view(%v) type(%v) get %v want %v(%v)", cy, i, rv.Type(), mType, mType.String())
				}
			}
			if rv.Type() != mType {
				return nil, errors.Errorf("invalid input type get %v want %v", rv.Type(), mType)
			}
		}
		in[i+1] = rv
	}

	vs, err := func() (vs []reflect.Value, err error) {
		defer func() {
			v := recover()
			if v != nil {
				fmt.Println(v)
				err = errors.New("occur error call method(" + method + ") of message: " + fmt.Sprintf("%v", v))
			}
		}()
		return rMethod.Call(in), nil
	}()
	if err != nil {
		return nil, err
	}

	result := []interface{}{}
	for i, v := range vs {
		vi := v.Interface()
		if mtype.Out(i).Kind() == reflect.Interface && mtype.Out(i).Implements(errType) {
			if _err, ok := vi.(error); ok {
				err = _err
			}
			continue
		}
		result = append(result, vi)
	}
	return result, err
}

var (
	strToAddr   string = "stringcommon.Address"
	strToAmount string = "string*amount.Amount"
	strToBigInt string = "string*big.Int"
	strToInt    string = "stringint"
	strToInt8   string = "stringint8"
	strToInt16  string = "stringint16"
	strToInt32  string = "stringint32"
	strToInt64  string = "stringint64"
	strToUint   string = "stringuint"
	strToUint8  string = "stringuint8"
	strToUint16 string = "stringuint16"
	strToUint32 string = "stringuint32"
	strToUint64 string = "stringuint64"

	bytesToAddr   string = "[]bytecommon.Address"
	bytesToAmount string = "[]byte*amount.Amount"
	bytesToBigInt string = "[]byte*big.Int"
)

func (m ViewCaller) parse(t string, v interface{}) (interface{}, bool) {
	if str, ok := v.(string); ok {
		switch t {
		case strToAddr:
			str = strings.Replace(str, "0x", "", -1)
			return common.HexToAddress(str), true
		case strToAmount:
			if am, err := amount.ParseAmount(str); err == nil {
				return am, true
			}
		case strToBigInt:
			if bi, ok := big.NewInt(0).SetString(str, 10); ok {
				return bi, true
			}
		default:
			if bi, ok := big.NewInt(0).SetString(str, 10); ok {
				switch t {
				case strToInt:
					return int(bi.Int64()), true
				case strToInt8:
					return int8(bi.Int64()), true
				case strToInt16:
					return int16(bi.Int64()), true
				case strToInt32:
					return int32(bi.Int64()), true
				case strToInt64:
					return int64(bi.Int64()), true
				case strToUint:
					return uint(bi.Uint64()), true
				case strToUint8:
					return uint8(bi.Uint64()), true
				case strToUint16:
					return uint16(bi.Uint64()), true
				case strToUint32:
					return uint32(bi.Uint64()), true
				case strToUint64:
					return uint64(bi.Uint64()), true
				}
			}
		}
	}
	if bs, ok := v.([]byte); ok {
		switch t {
		case bytesToAddr:
			return common.BytesToAddress(bs), true
		case bytesToAmount:
			return amount.NewAmountFromBytes(bs), true
		case bytesToBigInt:
			return big.NewInt(0).SetBytes(bs), true
		}
	}
	return v, false
}
