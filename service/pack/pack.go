package pack

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/common/amount"
)

var bigIntType = reflect.TypeOf(&big.Int{})
var amountType = reflect.TypeOf(&amount.Amount{})
var byteSliceType = reflect.TypeOf([]byte{})
var addressType = reflect.TypeOf(common.Address{})

// methodToString convert the given golang method to appropriate solidity-type function
// contract function의 경우 index = 2 부터 실제 argument 이다.
// ex. router.AddLiqudity : func(*router.RounterFront, *types.ContractContext, common.Address, common.Address, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount) (*amount.Amount, *amount.Amount, *amount.Amount, common.Address, error)
func ArgsToString(start int, method reflect.Method) (string, error) {
	methodType := method.Type
	args := ""
	for j := start; j < methodType.NumIn(); j++ {
		param := methodType.In(j)
		arg, err := argToString(param)
		if err != nil {
			return "", err
		}
		if j > start {
			args += ","
		}
		args += arg
	}

	return args, nil
}

// ArgsToString2 convert the given arguments to appropriate solidity-type function arguments
func ArgsToString2(args []interface{}) (string, error) {
	strArgs := ""
	for j := 0; j < len(args); j++ {
		arg, err := argToString(reflect.TypeOf(args[j]))
		if err != nil {
			return "", err
		}
		if j > 0 {
			strArgs += ","
		}
		strArgs += arg
	}

	return strArgs, nil
}

// argToString convert the given type to appropriate solidity-type
func argToString(typ reflect.Type) (string, error) {
	switch kind := typ.Kind(); kind {
	case reflect.Bool:
		return "bool", nil
	case reflect.Uint:
		return "uint256", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint64:
		return "uint64", nil
	case reflect.Int:
		return "int256", nil
	case reflect.Int8:
		return "int8", nil
	case reflect.Int16:
		return "int16", nil
	case reflect.Int32:
		return "int32", nil
	case reflect.Int64:
		return "int64", nil
	case reflect.Pointer:
		switch typ {
		case bigIntType, amountType:
			return "uint256", nil
		default:
			return "", ErrInvalidType
		}
	case reflect.String:
		return "string", nil
	case reflect.Interface:
		return "interface{}", nil

	case reflect.Slice:
		switch typ {
		case byteSliceType:
			return "bytes", nil
		default:
			r, err := argToString(typ.Elem())
			if err != nil {
				return "", err
			}
			return "[]" + r, nil
		}
	case reflect.Array:
		switch typ {
		case addressType:
			return "address", nil
		default:
			r, err := argToString(typ.Elem())
			if err != nil {
				return "", err
			}
			return "[" + strconv.Itoa(typ.Len()) + "]" + r, nil
		}
	default:
		return "", ErrInvalidType
	}
}

// toUint256Bytes convert the given reflect-value to appropriate uint256 bytes representation and function-type.
func ToUint256Bytes(topics [][]byte, value reflect.Value) ([][]byte, error) {
	switch kind := value.Kind(); kind {
	case reflect.Bool:
		if value.Bool() {
			topics = append(topics, math.PaddedBigBytes(common.Big1, 32))
		} else {
			topics = append(topics, math.PaddedBigBytes(common.Big0, 32))
		}
		return topics, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		topics = append(topics, math.U256Bytes(new(big.Int).SetUint64(value.Uint())))
		return topics, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		topics = append(topics, math.U256Bytes(big.NewInt(value.Int())))
		return topics, nil
	case reflect.Pointer:
		switch typ := value.Type(); typ {
		case bigIntType:
			topics = append(topics, math.U256Bytes(value.Interface().(*big.Int)))
			return topics, nil
		case amountType:
			topics = append(topics, math.U256Bytes(value.Interface().(*amount.Amount).Int))
			return topics, nil
		default:
			return nil, ErrInvalidType
		}
	case reflect.Slice:
		switch typ := value.Type(); typ {
		case byteSliceType:
			topics = append(topics, crypto.Keccak256(value.Bytes()))
			return topics, nil
		default: // ex. []interface{}
			for i := 0; i < value.Len(); i++ {
				var err error
				topics, err = ToUint256Bytes(topics, value.Index(i))
				if err != nil {
					return nil, err
				}
			}
			return topics, nil
		}
	case reflect.Array:
		switch typ := value.Type(); typ {
		case addressType:
			value = mustArrayToByteSlice(value)
			topics = append(topics, common.LeftPadBytes(value.Bytes(), 32))
			return topics, nil
		default:
			for i := 0; i < value.Len(); i++ {
				var err error
				topics, err = ToUint256Bytes(topics, value.Index(i))
				if err != nil {
					return nil, err
				}
			}
			return topics, nil
		}
	case reflect.String:
		topics = append(topics, crypto.Keccak256([]byte(value.String())))
		return topics, nil
	case reflect.Interface:
		var err error
		topics, err = ToUint256Bytes(topics, value.Elem())
		if err != nil {
			return nil, err
		}
		return topics, nil
	default:
		return nil, ErrInvalidType
	}
}

// Pack packs MethodCallEvent arguments. When unpacking, the function(event) definition is necessary
func Pack(args []interface{}) ([]byte, error) {

	var ret []byte
	for _, arg := range args {
		argByte, err := PackElement(reflect.ValueOf(arg))
		if err != nil {
			return nil, err
		}
		ret = append(ret, argByte...)
	}
	return ret, nil
}

// packList packs only slice and array, excludes address, byte-slice([]byte), byte-array([k]byte)
func packList(v reflect.Value) ([]byte, error) {

	var ret []byte
	ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)

	// slice, and length = 0
	if v.Len() == 0 {
		return ret, nil
	}

	// calculate offset if any
	offset := 0
	offsetReq := isDynamicType(v.Index(0))
	if offsetReq {
		offset = getTypeSize(v.Index(0)) * v.Len()
	}
	var tail []byte
	for i := 0; i < v.Len(); i++ {
		val, err := PackElement(v.Index(i))
		if err != nil {
			return nil, err
		}
		if !offsetReq {
			ret = append(ret, val...)
			continue
		}
		ret = append(ret, packNum(reflect.ValueOf(offset))...)
		offset += len(val)
		tail = append(tail, val...)
	}
	return append(ret, tail...), nil
}

// isDynamicType returns true if the type is dynamic.
// The following types are called “dynamic”:
// * bytes
// * string
// * T[] for any T
// * T[k] for any dynamic T and any k >= 0
// * (T1,...,Tk) if Ti is dynamic for some 1 <= i <= k
// reflect.TypeOf(v).Elem().Kind() 의 경우 하위로 계속 내려갈 수
func isDynamicType(v reflect.Value) bool {
	switch kind := v.Kind(); kind {
	case reflect.String:
		return true
	case reflect.Slice:
		return true
	case reflect.Array:
		if v.Len() > 0 && isDynamicType(v.Index(0)) {
			return true
		}
		return false
	default:
		return false
	}
}

// getTypeSize returns the size that this type needs to occupy.
// We distinguish static and dynamic types. Static types are encoded in-place
// and dynamic types are encoded at a separately allocated location after the
// current block.
// So for a static variable, the size returned represents the size that the
// variable actually occupies.
// For a dynamic variable, the returned size is fixed 32 bytes, which is used
// to store the location reference for actual value storage.
func getTypeSize(v reflect.Value) int {
	kind := v.Kind()
	if kind == reflect.Array {
		if v.Len() > 0 && !isDynamicType(v.Index(0)) {
			if v.Index(0).Kind() == reflect.Array {
				return v.Len() * getTypeSize(v.Index(0))
			}
			return v.Len() * 32
		}
	}
	return 32
}

// PackElement packs the given reflect value according to the abi specification in v.
func PackElement(v reflect.Value) ([]byte, error) {
	switch kind := v.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return packNum(v), nil
	case reflect.Ptr:
		switch typ := v.Type(); typ {
		case bigIntType:
			return math.U256Bytes(new(big.Int).Set(v.Interface().(*big.Int))), nil
		case amountType:
			return math.U256Bytes(new(big.Int).Set(v.Interface().(*amount.Amount).Int)), nil
		default:
			return nil, fmt.Errorf("PackElement, unknown type: %v", v.Type())
		}
	case reflect.String:
		return packBytesSlice([]byte(v.String()), v.Len()), nil
	case reflect.Array:
		switch typ := v.Type(); typ {
		case addressType:
			v = mustArrayToByteSlice(v)
			return common.LeftPadBytes(v.Bytes(), 32), nil
		default:
			// ex. [220]byte
			elem := v.Index(0)
			if elem.Kind() == reflect.Uint8 {
				v = mustArrayToByteSlice(v)
				return common.RightPadBytes(v.Bytes(), 32), nil
			}
			return packList(v)
		}
	case reflect.Slice:
		switch typ := v.Type(); typ {
		case byteSliceType:
			v = mustArrayToByteSlice(v)
			return packBytesSlice(v.Bytes(), v.Len()), nil
		}
		return packList(v)
	case reflect.Bool:
		if v.Bool() {
			return math.PaddedBigBytes(common.Big1, 32), nil
		}
		return math.PaddedBigBytes(common.Big0, 32), nil
	case reflect.Interface:
		ret, err := PackElement(v.Elem())
		if err != nil {
			return nil, err
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("PackElement, unknown type: %v", v.Type())
	}
}

// packBytesSlice packs the given bytes as [L, V] as the canonical representation
// bytes slice.
func packBytesSlice(bytes []byte, l int) []byte {
	len := packNum(reflect.ValueOf(l))
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...)
}

// packNum packs the given number (using the reflect value) and will cast it to appropriate number representation.
func packNum(value reflect.Value) []byte {
	switch kind := value.Kind(); kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return math.U256Bytes(new(big.Int).SetUint64(value.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return math.U256Bytes(big.NewInt(value.Int()))
	default:
		panic("abi: fatal error")
	}
}

// mustArrayToByteSlice creates a new byte slice with the exact same size as value
// and copies the bytes in value to the new slice.
func mustArrayToByteSlice(value reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.TypeOf([]byte{}), value.Len(), value.Len())
	reflect.Copy(slice, value)
	return slice
}
