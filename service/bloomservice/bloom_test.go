package bloomservice

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/assert"

	emath "github.com/ethereum/go-ethereum/common/math"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/pack"
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
		have, err := pack.ArgsToString(1, method)
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
		topics, err = pack.ToUint256Bytes(topics, tt.value)
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
		topics, err = pack.ToUint256Bytes(topics, tt.value)
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

// chainID 1337 - hardhat test purpose
func TestMutliTranactionEvents(t *testing.T) {
	chainID := big.NewInt(1337)
	version := uint16(1)

	userKeys, err := getSingers(chainID)
	if err != nil {
		t.Fatal(err)
	}
	aliceKey, bobKey, charlieKey := userKeys[0], userKeys[1], userKeys[2]
	alice, bob, charlie := aliceKey.PublicKey().Address(), bobKey.PublicKey().Address(), charlieKey.PublicKey().Address()

	intialize := func(ctx *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {
		tRet, err := mevInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}

		dRet, err := dexInitialize(ctx, classMap, args)
		if err != nil {
			return nil, err
		}
		return append(tRet, dRet...), nil
	}

	// alice(admin), bob, charlie
	args := []interface{}{alice, bob, charlie}
	tb, ret, err := prepare("_data", true, chainID, version, alice, args, intialize, &initContextInfo{})
	if err != nil {
		t.Fatal(err)
	}

	provider := tb.chain.Provider()

	stepMiliSeconds := uint64(1000)
	//timestamp := provider.LastTimestamp() + stepMiliSeconds*uint64(time.Millisecond)
	ctx := tb.newContext()

	//ctx := tb.nextContext(provider.LastHash(), timestamp)

	mev := ret[0].(common.Address)
	_, router, _, token0, token1, pair := ret[1].(common.Address), ret[2].(common.Address), ret[3].(common.Address), ret[4].(common.Address), ret[5].(common.Address), ret[6].(common.Address)

	// block 1 - StorageWithEvent Contract 생성
	tx0, err := StorageWithEventContractCreation(tb)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err = tb.addBlockAndSleep(ctx, []*txWithSigner{{tx0, aliceKey}}, stepMiliSeconds)
	if err != nil {
		t.Fatal(err)
	}

	storageWithEvent := crypto.CreateAddress(alice, tx0.Seq)
	fmt.Println("storageWithEvent  Address", storageWithEvent.String())

	// block 2
	tx0, err = StorageWithEventSet(tb)
	if err != nil {
		t.Fatal(err)
	}

	transferAmount := amount.NewAmount(1, 0)
	tx1 := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        mev,
		Method:    "Transfer",
		Args:      bin.TypeWriteAll(bob, transferAmount),
	}

	tx2 := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        mev,
		Method:    "Approve",
		Args:      bin.TypeWriteAll(charlie, MaxUint256),
	}

	token0Amount := amount.NewAmount(1, 0)
	token1Amount := amount.NewAmount(4, 0)
	tx3 := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        router,
		Method:    "UniAddLiquidity",
		Args:      bin.TypeWriteAll(token0, token1, token0Amount, token1Amount, amount.ZeroCoin, amount.ZeroCoin),
	}

	ctx, err = tb.addBlockAndSleep(ctx, []*txWithSigner{{tx0, aliceKey}, {tx1, aliceKey}, {tx2, bobKey}, {tx3, aliceKey}}, stepMiliSeconds)
	if err != nil {
		t.Fatal(err)
	}

	b, err := provider.Block(provider.Height())
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)
	var tIdx int
	var txBloom etypes.Bloom

	//tx 0
	tIdx = 0
	bHeight := provider.Height()
	receipts, err := provider.Receipts(bHeight)
	if err != nil {
		t.Fatal(err)
	}
	receipt := receipts[tIdx]
	txBloom = etypes.CreateBloom(etypes.Receipts{receipt})

	assertStorageWithEventSet := func(bloom etypes.Bloom) {
		// positive
		assert.Equal(bloom.Test(storageWithEvent[:]), true, "storageWithEvent :", storageWithEvent)
		eventFunc := "Set1(uint256,address,uint256,string)"
		assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1234))), true, "1234")
		assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)

		eventFunc = "Set2(string,address,uint256)"
		assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("set5678"))), true, "set5678")
		assert.Equal(bloom.Test(common.LeftPadBytes(common.HexToAddress("0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc").Bytes(), 32)), true, "0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc : ")

		// negative : unindexed arguments
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(10000))), false, "10000")
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("abcd"))), false, "abcd")
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(20000))), false, "20000")
	}

	assertStorageWithEventSet(txBloom)

	// tx 1
	tIdx = 1
	tx := b.Body.Transactions[tIdx]
	evs, err := FindTransactionsEvents(b.Body.Transactions, b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = CreateEventBloom(tb.newContext(), evs)
	if err != nil {
		t.Fatal(err)
	}

	assertMevTransfer := func(bloom etypes.Bloom) {
		// positive
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(bob.Bytes(), 32)), true, "to : ", bob)
			assert.Equal(bloom.Test(emath.U256Bytes(transferAmount.Int)), true, "transferAmount : ", transferAmount)
		}

		// negative
		{
			assert.Equal(bloom.Test(crypto.Keccak256([]byte("Transfer(address,address)"))), false)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)
		}
	}

	assertMevTransfer(txBloom)

	// tx 2
	tIdx = 2
	tx = b.Body.Transactions[tIdx]
	evs, err = FindTransactionsEvents(b.Body.Transactions, b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = CreateEventBloom(tb.newContext(), evs)
	if err != nil {
		t.Fatal(err)
	}

	assertMevApprove := func(bloom etypes.Bloom) {

		// positive
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "Approval(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(charlie.Bytes(), 32)), true, "to : ", charlie)
			assert.Equal(bloom.Test(emath.U256Bytes(MaxUint256.Int)), true, "approveAmount : ", MaxUint256)
		}

		// negative
		{
			assert.Equal(bloom.Test(crypto.Keccak256([]byte("Approve(address,uint256)"))), false)
			assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)
		}
	}

	assertMevApprove(txBloom)

	// tx 3
	tIdx = 3
	tx = b.Body.Transactions[tIdx]
	evs, err = FindTransactionsEvents(b.Body.Transactions, b.Body.Events, tIdx)
	if err != nil {
		t.Fatal(err)
	}
	txBloom, err = CreateEventBloom(tb.newContext(), evs)
	if err != nil {
		t.Fatal(err)
	}

	assertUniswapAddLiquidity := func(bloom etypes.Bloom) {

		// positive
		// 0 : router.UniAddLiquidity
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(tx.From.Bytes(), 32)), true, "tx.From : ", tx.From)
			assert.Equal(bloom.Test(tx.To[:]), true, "tx.To :", tx.To)
			eventFunc := "UniAddLiquidity(address,address,uint256,uint256,uint256,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(token0.Bytes(), 32)), true, "token0 : ", token0)
			assert.Equal(bloom.Test(common.LeftPadBytes(token1.Bytes(), 32)), true, "token1 : ", token1)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token0Amount : ", token0Amount)
			assert.Equal(bloom.Test(emath.U256Bytes(token1Amount.Int)), true, "token0Amount : ", token1Amount)
			assert.Equal(bloom.Test(emath.U256Bytes(ZeroAmount.Int)), true, "ZeroAmount : ", ZeroAmount)
		}
		// 1 : pair.Reserve
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(pair[:]), true, "to : ", pair)
			eventFunc := "Reserves()"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
		}
		// 2 : TransferFrom router -> token0.Transferfrom(alice,pair,amt)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(token0[:]), true, "to : ", token0)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token0Amount : ", token0Amount)
		}
		// 3 : TransferFrom router -> token1.Transferfrom(alice,pair,amt)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(token1[:]), true, "to : ", token1)
			eventFunc := "Transfer(address,address,uint256)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
			assert.Equal(bloom.Test(emath.U256Bytes(token0Amount.Int)), true, "token1Amount : ", token1Amount)
		}
		// 4 : pair.Mint
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(router.Bytes(), 32)), true, "from : ", router)
			assert.Equal(bloom.Test(pair[:]), true, "to : ", pair)
			eventFunc := "Mint(address)" // "Mint(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(alice.Bytes(), 32)), true, "alice : ", alice)
		}

		// 5 : BalanceOf pair : token0.BalanceOf(pair)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "from : ", pair)
			assert.Equal(bloom.Test(token0[:]), true, "to : ", token0)
			eventFunc := "BalanceOf(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
		}

		// 6 : BalanceOf pair : token1.BalanceOf(pair)
		{
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "from : ", pair)
			assert.Equal(bloom.Test(token1[:]), true, "to : ", token1)
			eventFunc := "BalanceOf(address)"
			assert.Equal(bloom.Test(crypto.Keccak256([]byte(eventFunc))), true, eventFunc)
			assert.Equal(bloom.Test(common.LeftPadBytes(pair.Bytes(), 32)), true, "pair : ", pair)
		}

		// negative
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("UniAddLiquiditi(address,address,uint256,uint256,uint256,uint256)"))), false)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("BalanceOg(address)"))), false)
		assert.Equal(bloom.Test(crypto.Keccak256([]byte("Transfer(address,address)"))), false)
		assert.Equal(bloom.Test(emath.U256Bytes(big.NewInt(1))), false)

	}

	assertUniswapAddLiquidity(txBloom)

	blockBloom, err := BlockLogsBloom(tb.chain, b)
	if err != nil {
		t.Fatal(err)
	}
	assertStorageWithEventSet(blockBloom)
	assertMevTransfer(blockBloom)
	assertMevApprove(blockBloom)
	assertUniswapAddLiquidity(blockBloom)

	removeChainData(tb.path)
}

// StorageWithEventContractCreation deploy StorageWithEvent contract by alice with nonce = 0
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventContractCreation(tb *testBlockChain) (*types.Transaction, error) {
	rlp := "0x02f90cca8205398080842c9d07ea841dcd65008080b90c72608060405234801561001057600080fd5b50336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550610c12806100606000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80632ecb20d31461005157806342526e4e1461008157806360fe47b1146100b15780636d4ce63c146100cd575b600080fd5b61006b60048036038101906100669190610710565b6100eb565b604051610078919061085e565b60405180910390f35b61009b600480360381019061009691906106a6565b61045a565b6040516100a891906107df565b60405180910390f35b6100cb60048036038101906100c691906106e7565b610503565b005b6100d561060a565b6040516100e29190610843565b60405180910390f35b60007f30000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156101cb57507f39000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610206577f300000000000000000000000000000000000000000000000000000000000000060f81c826101ff91906109ba565b9050610455565b7f61000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916101580156102e457507f66000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b1561032b577f610000000000000000000000000000000000000000000000000000000000000060f81c82600a61031a9190610935565b61032491906109ba565b9050610455565b7f41000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19161015801561040957507f46000000000000000000000000000000000000000000000000000000000000007effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff19168260f81b7effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff191611155b15610450577f410000000000000000000000000000000000000000000000000000000000000060f81c82600a61043f9190610935565b61044991906109ba565b9050610455565b600090505b919050565b600080600090506000805b60288160ff1610156104f85760108361047e919061096c565b92506104d2858260ff16815181106104bf577f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b602001015160f81c60f81b60f81c6100eb565b60ff16915081836104e391906108eb565b925080806104f090610a9b565b915050610465565b508192505050919050565b80600181905550600061052d604051806060016040528060288152602001610bb56028913961045a565b905060008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166104d27f0ca6778676cf982a32c6c02a2aa57966620f6581c6f016a2837cb60c2e9f5c5361271060405161059a91906107fa565b60405180910390a38073ffffffffffffffffffffffffffffffffffffffff166040516105c5906107ca565b60405180910390207f1658817f2f384793b2fe883b68c3df91c0122e6b4ed86878f4c10e494082d8cc614e206040516105fe9190610828565b60405180910390a35050565b6000600154905090565b60006106276106228461089e565b610879565b90508281526020810184848401111561063f57600080fd5b61064a848285610a5b565b509392505050565b600082601f83011261066357600080fd5b8135610673848260208601610614565b91505092915050565b60008135905061068b81610b86565b92915050565b6000813590506106a081610b9d565b92915050565b6000602082840312156106b857600080fd5b600082013567ffffffffffffffff8111156106d257600080fd5b6106de84828501610652565b91505092915050565b6000602082840312156106f957600080fd5b60006107078482850161067c565b91505092915050565b60006020828403121561072257600080fd5b600061073084828501610691565b91505092915050565b610742816109ee565b82525050565b61075181610a37565b82525050565b61076081610a49565b82525050565b60006107736004836108cf565b915061077e82610b34565b602082019050919050565b60006107966007836108e0565b91506107a182610b5d565b600782019050919050565b6107b581610a20565b82525050565b6107c481610a2a565b82525050565b60006107d582610789565b9150819050919050565b60006020820190506107f46000830184610739565b92915050565b600060408201905061080f6000830184610748565b818103602083015261082081610766565b905092915050565b600060208201905061083d6000830184610757565b92915050565b600060208201905061085860008301846107ac565b92915050565b600060208201905061087360008301846107bb565b92915050565b6000610883610894565b905061088f8282610a6a565b919050565b6000604051905090565b600067ffffffffffffffff8211156108b9576108b8610af4565b5b6108c282610b23565b9050602081019050919050565b600082825260208201905092915050565b600081905092915050565b60006108f682610a00565b915061090183610a00565b92508273ffffffffffffffffffffffffffffffffffffffff0382111561092a57610929610ac5565b5b828201905092915050565b600061094082610a2a565b915061094b83610a2a565b92508260ff0382111561096157610960610ac5565b5b828201905092915050565b600061097782610a00565b915061098283610a00565b92508173ffffffffffffffffffffffffffffffffffffffff04831182151516156109af576109ae610ac5565b5b828202905092915050565b60006109c582610a2a565b91506109d083610a2a565b9250828210156109e3576109e2610ac5565b5b828203905092915050565b60006109f982610a00565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600060ff82169050919050565b6000610a4282610a20565b9050919050565b6000610a5482610a20565b9050919050565b82818337600083830152505050565b610a7382610b23565b810181811067ffffffffffffffff82111715610a9257610a91610af4565b5b80604052505050565b6000610aa682610a2a565b915060ff821415610aba57610ab9610ac5565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b7f6162636400000000000000000000000000000000000000000000000000000000600082015250565b7f7365743536373800000000000000000000000000000000000000000000000000600082015250565b610b8f81610a20565b8114610b9a57600080fd5b50565b610ba681610a2a565b8114610bb157600080fd5b5056fe33633434636464646236613930306661326235383564643239396530336431326661343239336263a26469706673582212200227ddd9f59da8a18fa533bb272780c8c899d4fa64a28dc408b7d65c3430c50364736f6c63430008040033c080a01d6bb58ee614b7906d7903798b4e8dc635627f9e1b4840e781dc814491cea11ea004a4e1dae0b64bfbbd27a9966192c1d0d23eb756692d38e4a36a26a2103b156c"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         0,
		To:          ZeroAddress,
		Method:      "",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}

// StorageWithEventSet returns StorageWithEvent's Set function by alice with nonce = 1
// source code : evm-client/contracts/StorageWithEvent.sol
func StorageWithEventSet(tb *testBlockChain) (*types.Transaction, error) {
	rlp := "0x02f88e8205390180842c9d07ea841dcd6500945fbdb2315678afecb367f032d93f642f64180aa380a460fe47b1000000000000000000000000000000000000000000000000000000003ade68b1c001a06e1ab9f8e0b00a24946a95475d04adae3fd4cd5d92baf14178a5112247b554d1a01d30e5a83a548f34543b8a986404332bf34bb0cdce3d9a2bc0522194406c669f"

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	tx := &types.Transaction{
		ChainID:     tb.chainID,
		Timestamp:   uint64(time.Now().UnixNano()),
		Seq:         1,
		To:          common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"),
		Method:      "",
		GasPrice:    big.NewInt(748488682),
		UseSeq:      true,
		IsEtherType: true,
		VmType:      types.Evm,
		Args:        rlpBytes,
	}

	return tx, nil

}
