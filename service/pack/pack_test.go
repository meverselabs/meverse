package pack

import (
	"bytes"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/meverselabs/meverse/common/amount"
)

type testStruct struct{}

func (t *testStruct) SetBool(b bool)                                               {}
func (t *testStruct) SetUint(u8 uint8, u16 uint16, u32 uint32, u64 uint64, u uint) {}
func (t *testStruct) SetInt(i8 int8, i16 int16, i32 int32, i64 int64, i int)       {}
func (t *testStruct) SetPointer(b *big.Int, a *amount.Amount)                      {}
func (t *testStruct) SetString(s string)                                           {}
func (t *testStruct) SetInterface(i interface{})                                   {}
func (t *testStruct) SetAddress(a common.Address)                                  {}

// slice
func (t *testStruct) SetBoolSlice(b []bool)                                                       {}
func (t *testStruct) SetUintSlice(u8 []uint8, u16 []uint16, u32 []uint32, u64 []uint64, u []uint) {}
func (t *testStruct) SetIntSlice(i8 []int8, i16 []int16, i32 []int32, i64 []int64, i []int)       {}
func (t *testStruct) SetPointerSlice(b []*big.Int, a []*amount.Amount)                            {}
func (t *testStruct) SetStringSlice(s []string)                                                   {}
func (t *testStruct) SetInterfaceSlice(i []interface{})                                           {}
func (t *testStruct) SetAddressSlice(a []common.Address)                                          {}

// array
func (t *testStruct) SetBoolArray(b [2]bool) {}
func (t *testStruct) SetUintArray(u8 [2]uint8, u16 [3]uint16, u32 [4]uint32, u64 [5]uint64, u [6]uint) {
}
func (t *testStruct) SetIntArray(i8 [2]int8, i16 [3]int16, i32 [4]int32, i64 [5]int64, i [6]int) {}
func (t *testStruct) SetPointerArray(b [2]*big.Int, a [3]*amount.Amount)                         {}
func (t *testStruct) SetStringArray(s [4]string)                                                 {}
func (t *testStruct) SetInterfaceArray(i [5]interface{})                                         {}
func (t *testStruct) SetAddressArray(i [5]common.Address)                                        {}

// nested slice
func (t *testStruct) SetBoolNestedSlice(b [][]bool) {}
func (t *testStruct) SetUintNestedSlice(u8 [][]uint8, u16 [][]uint16, u32 [][]uint32, u64 [][]uint64, u [][]uint) {
}
func (t *testStruct) SetIntNestedSlice(i8 [][]int8, i16 [][]int16, i32 [][]int32, i64 [][]int64, i [][]int) {
}
func (t *testStruct) SetPointerNestedSlice(b [][]*big.Int, a [][]*amount.Amount) {}
func (t *testStruct) SetStringNestedSlice(s [][]string)                          {}
func (t *testStruct) SetInterfaceNestedSlice(i [][]interface{})                  {}
func (t *testStruct) SetAddressNestedSlice(a [][]common.Address)                 {}

// nested array
func (t *testStruct) SetBoolNestedArray(b [2][7]bool) {}
func (t *testStruct) SetUintNestedArray(u8 [2][7]uint8, u16 [3][7]uint16, u32 [4][7]uint32, u64 [5][7]uint64, u [6][7]uint) {
}
func (t *testStruct) SetIntNestedArray(i8 [2][7]int8, i16 [3][7]int16, i32 [4][7]int32, i64 [5][7]int64, i [6][7]int) {
}
func (t *testStruct) SetPointerNestedArray(b [2][7]*big.Int, a [3][7]*amount.Amount) {}
func (t *testStruct) SetStringNestedArray(s [4][7]string)                            {}
func (t *testStruct) SetInterfaceNestedArray(i [5][7]interface{})                    {}
func (t *testStruct) SetAddressNestedArray(a [5][7]common.Address)                   {}

// nested array slice
func (t *testStruct) SetBoolNestedArraySlice(b [2][]bool) {}
func (t *testStruct) SetUintNestedArraySlice(u8 [2][]uint8, u16 [3][]uint16, u32 [4][]uint32, u64 [5][]uint64, u [6][]uint) {
}
func (t *testStruct) SetIntNestedArraySlice(i8 [2][]int8, i16 [3][]int16, i32 [4][]int32, i64 [5][]int64, i [6][]int) {
}
func (t *testStruct) SetPointerNestedArraySlice(b [2][]*big.Int, a [3][]*amount.Amount) {}
func (t *testStruct) SetStringNestedArraySlice(s [4][]string)                           {}
func (t *testStruct) SetInterfaceNestedArraySlice(i [5][]interface{})                   {}
func (t *testStruct) SetAddressNestedArraySlice(a [5][]common.Address)                  {}

// nested slice array
func (t *testStruct) SetBoolNestedSliceArray(b [][7]bool) {}
func (t *testStruct) SetUintNestedSliceArray(u8 [][7]uint8, u16 [][7]uint16, u32 [][7]uint32, u64 [][7]uint64, u [][7]uint) {
}
func (t *testStruct) SetIntNestedSliceArray(i8 [][7]int8, i16 [][7]int16, i32 [][7]int32, i64 [][7]int64, i [][7]int) {
}
func (t *testStruct) SetPointerNestedSliceArray(b [][7]*big.Int, a [][7]*amount.Amount) {}
func (t *testStruct) SetStringNestedSliceArray(s [][7]string)                           {}
func (t *testStruct) SetInterfaceNestedSliceArray(i [][7]interface{})                   {}
func (t *testStruct) SetAddressNestedSliceArray(a [][7]common.Address)                  {}

func TestMethodToString(t *testing.T) {

	tests := map[string]string{
		// basic type
		"SetBool":      "bool",
		"SetUint":      "uint8,uint16,uint32,uint64,uint256",
		"SetInt":       "int8,int16,int32,int64,int256",
		"SetPointer":   "uint256,uint256",
		"SetString":    "string",
		"SetInterface": "interface{}",
		"SetAddress":   "address",

		// slice
		"SetBoolSlice":      "[]bool",
		"SetUintSlice":      "bytes,[]uint16,[]uint32,[]uint64,[]uint256",
		"SetIntSlice":       "[]int8,[]int16,[]int32,[]int64,[]int256",
		"SetPointerSlice":   "[]uint256,[]uint256",
		"SetStringSlice":    "[]string",
		"SetInterfaceSlice": "[]interface{}",
		"SetAddressSlice":   "[]address",

		// array
		"SetBoolArray":      "[2]bool",
		"SetUintArray":      "[2]uint8,[3]uint16,[4]uint32,[5]uint64,[6]uint256",
		"SetIntArray":       "[2]int8,[3]int16,[4]int32,[5]int64,[6]int256",
		"SetPointerArray":   "[2]uint256,[3]uint256",
		"SetStringArray":    "[4]string",
		"SetInterfaceArray": "[5]interface{}",
		"SetAddressArray":   "[5]address",

		// nested slice
		"SetBoolNestedSlice":      "[][]bool",
		"SetUintNestedSlice":      "[]bytes,[][]uint16,[][]uint32,[][]uint64,[][]uint256",
		"SetIntNestedSlice":       "[][]int8,[][]int16,[][]int32,[][]int64,[][]int256",
		"SetPointerNestedSlice":   "[][]uint256,[][]uint256",
		"SetStringNestedSlice":    "[][]string",
		"SetInterfaceNestedSlice": "[][]interface{}",
		"SetAddressNestedSlice":   "[][]address",

		// nested array
		"SetBoolNestedArray":      "[2][7]bool",
		"SetUintNestedArray":      "[2][7]uint8,[3][7]uint16,[4][7]uint32,[5][7]uint64,[6][7]uint256",
		"SetIntNestedArray":       "[2][7]int8,[3][7]int16,[4][7]int32,[5][7]int64,[6][7]int256",
		"SetPointerNestedArray":   "[2][7]uint256,[3][7]uint256",
		"SetStringNestedArray":    "[4][7]string",
		"SetInterfaceNestedArray": "[5][7]interface{}",
		"SetAddressNestedArray":   "[5][7]address",

		// nested array slice
		"SetBoolNestedArraySlice":      "[2][]bool",
		"SetUintNestedArraySlice":      "[2]bytes,[3][]uint16,[4][]uint32,[5][]uint64,[6][]uint256",
		"SetIntNestedArraySlice":       "[2][]int8,[3][]int16,[4][]int32,[5][]int64,[6][]int256",
		"SetPointerNestedArraySlice":   "[2][]uint256,[3][]uint256",
		"SetStringNestedArraySlice":    "[4][]string",
		"SetInterfaceNestedArraySlice": "[5][]interface{}",
		"SetAddressNestedArraySlice":   "[5][]address",

		// nested slice array
		"SetBoolNestedSliceArray":      "[][7]bool",
		"SetUintNestedSliceArray":      "[][7]uint8,[][7]uint16,[][7]uint32,[][7]uint64,[][7]uint256",
		"SetIntNestedSliceArray":       "[][7]int8,[][7]int16,[][7]int32,[][7]int64,[][7]int256",
		"SetPointerNestedSliceArray":   "[][7]uint256,[][7]uint256",
		"SetStringNestedSliceArray":    "[][7]string",
		"SetInterfaceNestedSliceArray": "[][7]interface{}",
		"SetAddressNestedSliceArray":   "[][7]address",
	}

	typ := reflect.TypeOf(&testStruct{})

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)

		// first argument = *testStruct
		have, err := ArgsToString(1, method)
		if err != nil {
			t.Errorf("err : %s, method : %s", err.Error(), method.Name)
		}
		exp := tests[method.Name]

		if have != exp {
			t.Errorf("got %s, want %s", have, exp)
		}

	}
}

func TestUint256Bytes(t *testing.T) {
	tests := []struct {
		value     reflect.Value
		converted []byte
	}{
		// bool
		{reflect.ValueOf(false), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(true), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},

		// Protocol limits
		{reflect.ValueOf(0), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(1), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},
		{reflect.ValueOf(-1), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},

		// Type corner cases
		{reflect.ValueOf(uint8(math.MaxUint8)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000ff")},
		{reflect.ValueOf(uint16(math.MaxUint16)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000ffff")},
		{reflect.ValueOf(uint32(math.MaxUint32)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000ffffffff")},
		{reflect.ValueOf(uint64(math.MaxUint64)), common.Hex2Bytes("000000000000000000000000000000000000000000000000ffffffffffffffff")},

		{reflect.ValueOf(int8(math.MaxInt8)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000007f")},
		{reflect.ValueOf(int16(math.MaxInt16)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000007fff")},
		{reflect.ValueOf(int32(math.MaxInt32)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000007fffffff")},
		{reflect.ValueOf(int64(math.MaxInt64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},

		{reflect.ValueOf(int8(math.MinInt8)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80")},
		{reflect.ValueOf(int16(math.MinInt16)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000")},
		{reflect.ValueOf(int32(math.MinInt32)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000")},
		{reflect.ValueOf(int64(math.MinInt64)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000")},

		// big.Int, amount.Amount
		{reflect.ValueOf(big.NewInt(math.MaxInt64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},
		{reflect.ValueOf(amount.NewAmount(0, math.MaxInt64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},

		// []byte, string
		{reflect.ValueOf("set5678"), common.Hex2Bytes("63e3e34f1519e390d4f69aa701eb66f1606a135404028a735c685cd12a5886fc")},
		{reflect.ValueOf([]byte("set5678")), common.Hex2Bytes("63e3e34f1519e390d4f69aa701eb66f1606a135404028a735c685cd12a5886fc")},

		// address
		{reflect.ValueOf(common.HexToAddress("f39fd6e51aad88f6f4ce6ab8827279cfffb92266")), common.Hex2Bytes("000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266")},
	}

	topics := [][]byte{}
	var err error
	for _, tt := range tests {
		topics, err = ToUint256Bytes(topics, tt.value)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	// t.Log("length", len(topics))

	// for _, topic := range topics {
	// 	t.Log(topic)
	// }

	for i, tt := range tests {
		if !bytes.Equal(topics[i], tt.converted) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, topics[i], tt.converted)
		}
	}
}

func TestUint256BytesArray(t *testing.T) {
	tests := []struct {
		value     reflect.Value
		converted [][]byte
	}{
		// array
		{reflect.ValueOf([3]int{1, 2, 3}), [][]byte{common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002"), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000003")}},

		// big.Int slice
		{reflect.ValueOf([]*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300)}), [][]byte{common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000064"), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000C8"), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000012C")}},

		// amount.Amount slice
		{reflect.ValueOf([]*amount.Amount{amount.NewAmount(0, 1000), amount.NewAmount(0, 2000), amount.NewAmount(0, 3000)}), [][]byte{common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000003E8"), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000007D0"), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000BB8")}},
	}

	topics := [][]byte{}
	var err error
	for _, tt := range tests {
		topics, err = ToUint256Bytes(topics, tt.value)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	t.Log("length", len(topics))

	// for _, topic := range topics {
	// 	t.Log(topic)
	// }

	k := 0
	for i, tt := range tests {
		for _, cc := range tt.converted {
			//t.Log(topics[k], cc)
			if !bytes.Equal(topics[k], cc) {
				t.Errorf("test %d: pack mismatch: have %x, want %x", i, topics[i], tt.converted[i])
			}
			k++
		}
	}
}

func TestPackBoolAndNumber(t *testing.T) {
	tests := []struct {
		value  reflect.Value
		packed []byte
	}{
		{reflect.ValueOf(false), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(true), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},

		// Protocol limits
		{reflect.ValueOf(0), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf(1), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001")},
		{reflect.ValueOf(-1), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")},

		// Type corner cases
		{reflect.ValueOf(uint8(math.MaxUint8)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000ff")},
		{reflect.ValueOf(uint16(math.MaxUint16)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000ffff")},
		{reflect.ValueOf(uint32(math.MaxUint32)), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000ffffffff")},
		{reflect.ValueOf(uint64(math.MaxUint64)), common.Hex2Bytes("000000000000000000000000000000000000000000000000ffffffffffffffff")},

		{reflect.ValueOf(int8(math.MaxInt8)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000000007f")},
		{reflect.ValueOf(int16(math.MaxInt16)), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000007fff")},
		{reflect.ValueOf(int32(math.MaxInt32)), common.Hex2Bytes("000000000000000000000000000000000000000000000000000000007fffffff")},
		{reflect.ValueOf(int64(math.MaxInt64)), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},

		{reflect.ValueOf(int8(math.MinInt8)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80")},
		{reflect.ValueOf(int16(math.MinInt16)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000")},
		{reflect.ValueOf(int32(math.MinInt32)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffffffffffff80000000")},
		{reflect.ValueOf(int64(math.MinInt64)), common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffffffffffffff8000000000000000")},

		{reflect.ValueOf(big.NewInt(int64(math.MaxInt64))), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},
		{reflect.ValueOf(amount.NewAmount(0, uint64(math.MaxInt64))), common.Hex2Bytes("0000000000000000000000000000000000000000000000007fffffffffffffff")},
	}
	for i, tt := range tests {
		packed, err := PackElement(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(packed, tt.packed) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, packed, tt.packed)
		}
	}
}

func TestPack(t *testing.T) {
	tests := []struct {
		value  reflect.Value
		packed []byte
	}{
		{reflect.ValueOf(common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")), common.Hex2Bytes("000000000000000000000000f39fd6e51aad88f6f4ce6ab8827279cfffb92266")},
		{reflect.ValueOf("a"), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000016100000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([2]string{"a", "b"}), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([]string{"a", "b"}), common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000001610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([2]byte{10, 20}), common.Hex2Bytes("0a14000000000000000000000000000000000000000000000000000000000000")},
		{reflect.ValueOf([]byte{10, 20}), common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000000020a14000000000000000000000000000000000000000000000000000000000000")},
	}
	for i, tt := range tests {
		packed, err := PackElement(tt.value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(packed, tt.packed) {
			t.Errorf("test %d: pack mismatch: have %x, want %x", i, packed, tt.packed)
		}
	}
}
