package bloomservice

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/pack"
)

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
func CreateEventBloom(provider types.Provider, events []*ctypes.Event) (etypes.Bloom, error) {
	var blm etypes.Bloom
	topics := [][]byte{}

	for _, event := range events {

		switch event.Type {
		case ctypes.EventTagCallHistory:

			mc := &ctypes.MethodCallEvent{}
			_, err := mc.ReadFrom(bytes.NewReader(event.Result))
			if err != nil {
				return etypes.Bloom{}, err
			}

			if mc.To == (common.Address{}) {
				return etypes.Bloom{}, errors.New("event To is null")
			}

			// address
			blm.Add(mc.To.Bytes())

			// topics
			evTopics, err := makeEventTopics(provider, mc)
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
func makeEventTopics(provider types.Provider, mc *ctypes.MethodCallEvent) ([][]byte, error) {
	topics := [][]byte{}

	event, err := functionToEvent(provider, mc)
	if err != nil {
		return nil, err
	}

	topics = append(topics, crypto.Keccak256([]byte(event))) // event Hash

	// mc.Method에 따라 topic 추가
	// Mint : from = zero address 추가 <- Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount)
	// Burn : from = mc.From(), to = zeroaddress 추가 <- Burn(cc *types.ContractContext, am *amount.Amount)
	// Transfer, Approve  :  mc.From 추가
	// TransferFrom :  추가 없음
	endArg := len(mc.Args)
	switch mc.Method {
	case "Transfer", "Approve":
		topics = append(topics, common.LeftPadBytes(mc.From.Bytes(), 32))
		endArg = 1
	case "TransferFrom":
		endArg = 2
	case "Mint":
		topics = append(topics, common.LeftPadBytes([]byte{}, 32))
		endArg = 1
	case "Burn":
		endArg = 0
		topics = append(topics, common.LeftPadBytes(mc.From.Bytes(), 32))
		topics = append(topics, common.LeftPadBytes([]byte{}, 32))
	}
	if endArg >= 0 {
		if topics, err = pack.ToUint256Bytes(topics, reflect.ValueOf(mc.Args[:endArg])); err != nil {
			return nil, err
		}
	}
	return topics, nil
}

// makeEventData makes data from args
func makeEventData(provider types.Provider, mc *ctypes.MethodCallEvent) []interface{} {
	args := []interface{}{}

	// erc20에서 indexed 인 값을 제외하고 추가(amount 값만 들어감)
	// Burn : 1번째
	// Transfer, Approve, Mint  :  2번째
	// TransferFrom :  3번째
	switch mc.Method {
	case "TransferFrom":
		args = append(args, mc.Args[2])
	case "Transfer", "Approve", "Mint":
		args = append(args, mc.Args[1])
	case "Burn":
		args = append(args, mc.Args[0])
	default:
		return mc.Args
	}
	return args
}

// functionToEvent convert function name and arguments to event
// ex1. router.AddLiqudity : func(*router.RounterFront, *types.ContractContext, common.Address, common.Address, *amount.Amount, *amount.Amount, *amount.Amount, *amount.Amount) -> AddLiquidity(address,address,uint256,uint256,uint256,uint256)
// ex2. Approve(address,uint256) -> Approval(address,address,uint256)
func functionToEvent(provider types.Provider, mc *ctypes.MethodCallEvent) (string, error) {

	contract, err := provider.Contract(mc.To)
	if err != nil {
		return "", err
	}

	contractName := reflect.TypeOf(contract).String()
	contractName = contractName[1:]

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
	argsStr, err := pack.ArgsToString(2, method)
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
func FindCallHistoryEvents(evs []*ctypes.Event, idx uint16) ([]*ctypes.Event, error) {

	if len(evs) == 0 {
		return nil, nil
	}

	var ret []*ctypes.Event
	for i := 0; i < len(evs); i++ {
		if evs[i].Index != idx || evs[i].Type != ctypes.EventTagCallHistory {
			continue
		}
		ret = append(ret, evs[i])
	}

	return ret, nil
}

// BlockLogsBloom retrurn block's logsBloom
func BlockLogsBloom(cn *chain.Chain, b *types.Block) (etypes.Bloom, error) {

	// go-type Contract event
	bloom, err := CreateEventBloom(cn.Provider(), b.Body.Events)
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

// TxLogsBloom retrurn transaction's logsBloom  ( idx-th tx in block b)
func TxLogsBloom(cn *chain.Chain, b *types.Block, idx uint16, receipt *etypes.Receipt) (etypes.Bloom, []*etypes.Log, error) {

	//func BloLogsBloom(cn *chain.Chain, b *types.Block, idx uint16, receipt *etypes.Receipt) (etypes.Bloom, []*etypes.Log, error) {

	var bloom etypes.Bloom
	logs := []*etypes.Log{}

	// combine logs and logsBloom
	evs, err := FindCallHistoryEvents(b.Body.Events, idx)
	if err != nil {
		return etypes.Bloom{}, nil, err
	}
	if evs != nil {
		bloom, err = CreateEventBloom(cn.Provider(), evs)
		if err != nil {
			return etypes.Bloom{}, nil, err
		}
		logs, err = EventsToFullLogs(cn, &b.Header, b.Body.Transactions[idx], evs, idx)
		if err != nil {
			return etypes.Bloom{}, nil, err
		}
	}

	if receipt != nil {
		evsLen := len(logs)
		for i, log := range receipt.Logs {
			log.Index = uint(evsLen + i)
			logs = append(logs, log)
		}
		rBloom := etypes.CreateBloom(etypes.Receipts{receipt})
		for i := 0; i < len(rBloom); i++ {
			bloom[i] |= rBloom[i]
		}
	}
	return bloom, logs, nil
}
