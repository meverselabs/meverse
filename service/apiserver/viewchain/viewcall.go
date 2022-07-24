package viewchain

import (
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/prefix"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

// var bigIntType = reflect.TypeOf(&big.Int{})
// var amountType = reflect.TypeOf(&amount.Amount{})

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

func (m *ViewCaller) MultiExecute(addr []common.Address, from string, methods []string, inputss [][]interface{}) ([][]interface{}, error) {
	if len(addr) != len(inputss) {
		return nil, errors.New("not match params count")
	}
	if len(methods) != len(inputss) {
		return nil, errors.New("not match params count")
	}

	result := make([][]interface{}, len(methods))
	for i, method := range methods {
		contract, cc, err := m.getCallContract(addr[i], from)
		if err != nil {
			return nil, err
		}

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
	types.ExecLock.Lock()
	defer types.ExecLock.Unlock()

	for _, v := range inputs {
		if v == nil {
			return nil, errors.New("nil params")
		}
	}
	if len(method) < 1 {
		return nil, errors.New("invalid method name")
	}
	var is interface{} = contract.Interface()
	if _, ok := is.(types.InvokeableContract); ok &&
		(method != "InitContract" &&
			method != "IsUpdateable" &&
			method != "Update" &&
			method != "ContractInvoke" &&
			method != "SetOwner") {
		inputs = []interface{}{
			method,
			inputs,
		}
		method = "ContractInvoke"
	} else {
		method = strings.ToUpper(string(method[0])) + method[1:]
	}

	rMethod := contract.MethodByName(method)
	if rMethod.Kind() == reflect.Invalid {
		return nil, errors.New("not exist method")
	}
	in, err := types.ContractInputsConv(inputs, rMethod)
	if err != nil {
		return nil, err
	}
	in = append([]reflect.Value{reflect.ValueOf(cc)}, in...)

	vs, err := func() (vs []reflect.Value, err error) {
		defer func() {
			v := recover()
			if v != nil {
				fmt.Println(v)
				if method == "ContractInvoke" && len(inputs) > 0 {
					method = fmt.Sprintf("%v", inputs[0])
				}
				err = fmt.Errorf("occur error call method(%v) of message: %v", method, v)
			}
		}()
		return rMethod.Call(in), nil
	}()
	if err != nil {
		return nil, err
	}

	mtype := rMethod.Type()
	result := []interface{}{}
	for i, v := range vs {
		vi := v.Interface()
		if mtype.Out(i).Kind() == reflect.Interface && mtype.Out(i).Implements(errType) {
			if _err, ok := vi.(error); ok {
				if method == "ContractInvoke" && len(inputs) > 0 {
					method = fmt.Sprintf("%v", inputs[0])
				}
				err = fmt.Errorf("%v method %v", _err.Error(), method)
			}
			continue
		}
		switch v := vi.(type) {
		case *big.Int:
			result = append(result, v.String())
		case []*big.Int:
			res := []string{}
			for _, bi := range v {
				res = append(res, bi.String())
			}
			result = append(result, res)
		default:
			result = append(result, vi)
		}
	}
	return result, err
}

func GetVersion() string {
	sb := strings.Builder{}
	sb.WriteString("MEVerse/")
	sb.WriteString(prefix.ClientVersion)
	sb.WriteString("/")
	sb.WriteString(runtime.GOOS)
	sb.WriteString("-")
	sb.WriteString(runtime.GOARCH)
	sb.WriteString("/")
	sb.WriteString(runtime.Version())
	return sb.String()
}

// var (
// 	strToAddr   string = "stringcommon.Address"
// 	strToAmount string = "string*amount.Amount"
// 	strToBigInt string = "string*big.Int"
// 	strToInt    string = "stringint"
// 	strToInt8   string = "stringint8"
// 	strToInt16  string = "stringint16"
// 	strToInt32  string = "stringint32"
// 	strToInt64  string = "stringint64"
// 	strToUint   string = "stringuint"
// 	strToUint8  string = "stringuint8"
// 	strToUint16 string = "stringuint16"
// 	strToUint32 string = "stringuint32"
// 	strToUint64 string = "stringuint64"

// 	bytesToAddr   string = "[]bytecommon.Address"
// 	bytesToAmount string = "[]byte*amount.Amount"
// 	bytesToBigInt string = "[]byte*big.Int"
// )

// func (m ViewCaller) parse(t string, v interface{}) (interface{}, bool) {
// 	if str, ok := v.(string); ok {
// 		switch t {
// 		case strToAddr:
// 			str = strings.Replace(str, "0x", "", -1)
// 			return common.HexToAddress(str), true
// 		case strToAmount:
// 			if am, err := amount.ParseAmount(str); err == nil {
// 				return am, true
// 			}
// 		case strToBigInt:
// 			if bi, ok := big.NewInt(0).SetString(str, 10); ok {
// 				return bi, true
// 			}
// 		default:
// 			if bi, ok := big.NewInt(0).SetString(str, 10); ok {
// 				switch t {
// 				case strToInt:
// 					return int(bi.Int64()), true
// 				case strToInt8:
// 					return int8(bi.Int64()), true
// 				case strToInt16:
// 					return int16(bi.Int64()), true
// 				case strToInt32:
// 					return int32(bi.Int64()), true
// 				case strToInt64:
// 					return int64(bi.Int64()), true
// 				case strToUint:
// 					return uint(bi.Uint64()), true
// 				case strToUint8:
// 					return uint8(bi.Uint64()), true
// 				case strToUint16:
// 					return uint16(bi.Uint64()), true
// 				case strToUint32:
// 					return uint32(bi.Uint64()), true
// 				case strToUint64:
// 					return uint64(bi.Uint64()), true
// 				}
// 			}
// 		}
// 	}
// 	if bs, ok := v.([]byte); ok {
// 		switch t {
// 		case bytesToAddr:
// 			return common.BytesToAddress(bs), true
// 		case bytesToAmount:
// 			return amount.NewAmountFromBytes(bs), true
// 		case bytesToBigInt:
// 			return big.NewInt(0).SetBytes(bs), true
// 		}
// 	}
// 	return v, false
// }
