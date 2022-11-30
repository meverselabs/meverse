package bloomservice

import (
	"bytes"
	"errors"
	"math/big"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

var bigIntType = reflect.TypeOf(&big.Int{})
var amountType = reflect.TypeOf(&amount.Amount{})
var byteSliceType = reflect.TypeOf([]byte{})
var addressType = reflect.TypeOf(common.Address{})

// methodToString convert the given golang method to appropriate solidity-type function
// contract function의 경우 index = 2 부터 실제 argument 이다.
// ex. router.AddLiqudity : func(*router.RounterFront, *types.ContractContext, common.Address, common.Address, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount) (*amount.Amount, *amount.Amount, *amount.Amount, common.Address, error)
func argsToString(start int, method reflect.Method) (string, error) {
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
func toUint256Bytes(topics [][]byte, value reflect.Value) ([][]byte, error) {
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
				topics, err = toUint256Bytes(topics, value.Index(i))
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
				topics, err = toUint256Bytes(topics, value.Index(i))
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
		topics, err = toUint256Bytes(topics, value.Elem())
		if err != nil {
			return nil, err
		}
		return topics, nil
	default:
		return nil, ErrInvalidType
	}
}

// mustArrayToByteSlice creates a new byte slice with the exact same size as value
// and copies the bytes in value to the new slice.
func mustArrayToByteSlice(value reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.TypeOf([]byte{}), value.Len(), value.Len())
	reflect.Copy(slice, value)
	return slice
}

// AppendBloom appends the given Receipts (+Logs) to bloom
func AppendBloom(bin *etypes.Bloom, receipts etypes.Receipts) {
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			bin.Add(log.Address.Bytes())
			for _, b := range log.Topics {
				bin.Add(b[:])
			}
		}
	}
}

// CreateEventBloom creates a bloom filter from the given go-type contract events
func CreateEventBloom(ctx *types.Context, events []*types.Event) (etypes.Bloom, error) {
	var blm etypes.Bloom
	topics := [][]byte{}

	for _, event := range events {

		switch event.Type {
		case types.EventTagCallHistory:

			mc := &types.MethodCallEvent{}
			_, err := mc.ReadFrom(bytes.NewReader(event.Result))
			if err != nil {
				return etypes.Bloom{}, err
			}
			//fmt.Println(mc)

			if mc.To == (common.Address{}) {
				return etypes.Bloom{}, errors.New("event To is null")
			}

			blm.Add(mc.To.Bytes())

			evTopics, err := makeEventTopics(ctx, mc)
			if err != nil {
				return etypes.Bloom{}, err
			}

			topics = append(topics, evTopics...)
		default:
			continue
		}
	}
	for _, topic := range topics {
		blm.Add(topic)
	}

	return blm, nil
}

// makeEventTopics makes topics from event
func makeEventTopics(ctx *types.Context, mc *types.MethodCallEvent) ([][]byte, error) {
	topics := [][]byte{}

	event, err := functionToEvent(ctx, mc)
	if err != nil {
		return nil, err
	}

	topics = append(topics, crypto.Keccak256([]byte(event)))          // event Hash
	topics = append(topics, common.LeftPadBytes(mc.From.Bytes(), 32)) // Transfer, Approval Event의 경우 mc.From이 들어간다.
	if topics, err = toUint256Bytes(topics, reflect.ValueOf(mc.Args)); err != nil {
		return nil, err
	}
	return topics, nil
}

// functionToEvent convert function name and arguments to event
// ex1. router.AddLiqudity : func(*router.RounterFront, *types.ContractContext, common.Address, common.Address, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount) -> AddLiquidity(address,address,uint256,uint256,uint256,uint256)
// ex2. Approve(address,uint256) -> Approval(address,address,uint256)
func functionToEvent(ctx *types.Context, mc *types.MethodCallEvent) (string, error) {
	contract, err := ctx.Contract(mc.To)
	if err != nil {
		return "", err
	}

	contractName := reflect.TypeOf(contract).String()
	contractName = contractName[1:]
	// log.Println("reflect.TypeOf(contract).String()", reflect.TypeOf(contract).String())

	if c, ok := convertMap[contractName]; ok {
		if f, ok := c[mc.Method]; ok {
			return f, nil
		}
	}

	frontType := reflect.TypeOf(contract.Front())
	method, ok := frontType.MethodByName(mc.Method)
	if !ok {
		return "", err
	}
	argsStr, err := argsToString(2, method)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	b.WriteString(mc.Method)
	b.WriteString("(")
	b.WriteString(argsStr)
	b.WriteString(")")

	return b.String(), nil

}

// hashTopics converts  []byte slice topics to hash-type slice topics
func hashTopics(topics [][]byte) []common.Hash {
	var hashTopics []common.Hash
	for i := 0; i < len(topics); i++ {
		hashTopics = append(hashTopics, common.BytesToHash(topics[i]))
	}
	return hashTopics
}

// findTransactionsEvents find tx's EventTagCallHistory-type events from given event list
func FindTransactionsEvents(txs []*types.Transaction, evs []*types.Event, idx int) ([]*types.Event, error) {

	if len(evs) == 0 {
		return nil, nil
	}

	eIdx := 0
	start := 0
	startIdx := 0
	endIdx := len(evs)

	found := false

	for _, tx := range txs {
		// tx 가 Evm Type 이거나 tx.To가 null인 경우 (즉 admin tx인 경우)
		if tx.VmType == types.Evm || (tx.VmType != types.Evm && tx.To == (common.Address{})) {
			if idx == eIdx {
				return nil, nil
			}
			eIdx++
			continue
		}

		// nonce만 다르고 똑같은 transaction이 여러번 올 수 있는 경우 고려
		// types.EventTagReward 는 tx없이 실행됨 blockCreator.Finalize
		for i := start; i < len(evs); i++ {
			if evs[i].Type != types.EventTagCallHistory {
				continue
			}
			mc := &types.MethodCallEvent{}
			_, err := mc.ReadFrom(bytes.NewReader(evs[i].Result))
			if err != nil {
				return nil, err
			}
			// tx에 해당하는 events중 첫번째 types.EventTagCallHistory
			if tx.From == mc.From && tx.To == mc.To && tx.Method == mc.Method {
				if idx == eIdx {
					found = true
					startIdx = i
					start = i + 1
					eIdx++
					break
				} else {
					// evs에 끝에 도달하지 않으면, 마지막 event는 제거
					eIdx++
					if found {
						endIdx = i
						goto ListEvent
					}
				}
			}
		}
	}

ListEvent:
	//중간에 다른 Event가 끼워 들어갈 수 있는 부분 방지
	var ret []*types.Event
	for i := startIdx; i < endIdx; i++ {
		if evs[i].Type != types.EventTagCallHistory {
			continue
		}
		ret = append(ret, evs[i])
	}

	return ret, nil
}

// BlockLogsBloom retrurn block's logsBloom
func BlockLogsBloom(cn *chain.Chain, b *types.Block) (etypes.Bloom, error) {

	// go-type Contract event
	bloom, err := CreateEventBloom(cn.NewContext(), b.Body.Events)
	if err != nil {
		return etypes.Bloom{}, err
	}

	// evm-type Contract event
	var receipts types.Receipts
	if b.Header.Height <= cn.Provider().InitHeight() {
		receipts = types.Receipts{}
	} else {
		receipts, err = cn.Provider().Receipts(b.Header.Height)
		if err != nil {
			return etypes.Bloom{}, err
		}
	}

	// receipt.log.Address, Topics는 이미 설정되어 DeriveReceiptFields는 필요없음
	AppendBloom(&bloom, append(etypes.Receipts{}, receipts...))
	return bloom, nil
}
